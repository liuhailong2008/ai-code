package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"pingmesh-controller/conf"
)

// StatisticsService 统计数据服务
type StatisticsService struct {
	promClient *PromClient
}

// NewStatisticsService 创建 StatisticsService
func NewStatisticsService(promClient *PromClient) *StatisticsService {
	log.Println("StatisticsService created")
	return &StatisticsService{
		promClient: promClient,
	}
}

// MetricMeta 指标元数据（返回给前端）
type MetricMeta struct {
	Metric      string `json:"metric"`
	Type        string `json:"type"`        // latency / loss
	Scope       string `json:"scope"`       // idc / spine / leaf
	Category    string `json:"category"`    // raw / max / avg / p99
	Description string `json:"description"` // 中文描述
	Status      string `json:"status"`      // new / show / hide / disable
	Strategy    string `json:"strategy"`    // 00 / 01 / 02 / 03
	PromType    string `json:"prom_type"`   // Prometheus 返回的 type（gauge/counter等），可选
}

// ListMetrics 获取所有指标，附带配置中的元数据
func (s *StatisticsService) ListMetrics() ([]MetricMeta, error) {
	// 从 Prometheus 获取所有指标名
	promMetrics, err := s.promClient.ListMetrics()
	if err != nil {
		// Prometheus 不可用时，直接返回配置中的指标
		log.Printf("ListMetrics: prometheus not available, using config metadata only: %v", err)
		promMetrics = nil
	}

	// 构建配置元数据索引
	cfg := conf.Get()
	metaMap := make(map[string]*conf.MetricMetaConfig)
	for i := range cfg.MetricsMeta {
		metaMap[cfg.MetricsMeta[i].Metric] = &cfg.MetricsMeta[i]
	}

	// 如果 Prometheus 可用，用它返回的指标列表
	if promMetrics != nil {
		var result []MetricMeta
		for _, name := range promMetrics {
			// 只返回 ping_ 开头的指标
			if !strings.HasPrefix(name, "ping_") {
				continue
			}
			item := MetricMeta{Metric: name, Status: "new"}
			if m, ok := metaMap[name]; ok {
				item.Type = m.Type
				item.Scope = m.Scope
				item.Category = m.Category
				item.Description = m.Description
				item.Strategy = m.Strategy
				if m.Status != "" {
					item.Status = m.Status
				} else {
					item.Status = "show"
				}
			}
			result = append(result, item)
		}
		return result, nil
	}

	// Prometheus 不可用时，返回配置中的指标
	var result []MetricMeta
	for _, m := range cfg.MetricsMeta {
		status := m.Status
		if status == "" {
			status = "show"
		}
		result = append(result, MetricMeta{
			Metric:      m.Metric,
			Type:        m.Type,
			Scope:       m.Scope,
			Category:    m.Category,
			Description: m.Description,
			Status:      status,
			Strategy:    m.Strategy,
		})
	}
	return result, nil
}

// MetricDataRequest 指标数据查询请求
type MetricDataRequest struct {
	Metric string            `json:"metric"`
	Labels map[string]string `json:"labels"` // label 筛选条件
}

// MetricDataResult 指标数据查询结果
type MetricDataResult struct {
	Metric string            `json:"metric"`
	Labels map[string]string `json:"labels"`
	Value  string            `json:"value"`
}

// QueryMetricData 查询指标数据（带 label 筛选）
func (s *StatisticsService) QueryMetricData(metric string, labels map[string]string) ([]MetricDataResult, error) {
	// 构建 PromQL 查询
	query := metric
	if len(labels) > 0 {
		var labelPairs []string
		for k, v := range labels {
			if v != "" {
				labelPairs = append(labelPairs, fmt.Sprintf(`%s="%s"`, k, v))
			}
		}
		if len(labelPairs) > 0 {
			query = fmt.Sprintf(`%s{%s}`, metric, strings.Join(labelPairs, ","))
		}
	}

	results, err := s.promClient.QueryWithLabels(query)
	if err != nil {
		return nil, fmt.Errorf("query prometheus: %w", err)
	}

	var data []MetricDataResult
	for _, r := range results {
		val := ""
		if len(r.Value) >= 2 {
			val = fmt.Sprintf("%v", r.Value[1])
		}
		data = append(data, MetricDataResult{
			Metric: metric,
			Labels: r.Metric,
			Value:  val,
		})
	}
	return data, nil
}

// GetMetricLabels 获取指标的标签列表和每个标签的值
func (s *StatisticsService) GetMetricLabels(metric string) (map[string][]string, error) {
	labelNames, err := s.promClient.GetMetricLabels(metric)
	if err != nil {
		return nil, fmt.Errorf("get metric labels: %w", err)
	}

	result := make(map[string][]string)
	for _, label := range labelNames {
		values, err := s.promClient.GetLabelValues(metric, label)
		if err != nil {
			log.Printf("GetMetricLabels: get values for %s.%s failed: %v", metric, label, err)
			continue
		}
		result[label] = values
	}
	return result, nil
}

// ── 指标元数据配置管理 ──

// GetMetricsMetaConfig 获取指标元数据配置
func (s *StatisticsService) GetMetricsMetaConfig() []conf.MetricMetaConfig {
	return conf.Get().MetricsMeta
}

// UpdateMetricsMetaConfig 更新指标元数据配置
func (s *StatisticsService) UpdateMetricsMetaConfig(meta []conf.MetricMetaConfig) error {
	return conf.UpdateMetricsMeta(meta)
}

// WriteJSON 统一 JSON 响应写入（避免循环依赖，这里自己实现）
func WriteJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
