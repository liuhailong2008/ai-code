package service

import (
	"fmt"
	"pingmesh-controller/conf"
	"strconv"
	"strings"
)

// DashboardService 大屏监控页面 Service
type DashboardService struct {
	db   *DB
	prom *PromClient
}

// NewDashboardService 创建大屏监控 Service
func NewDashboardService(db *DB, prom *PromClient) *DashboardService {
	return &DashboardService{db: db, prom: prom}
}

// ── 1. 机房拓扑状态 ──

// GetIdcStatusForSvgGraph 获取机房状态（用于 SVG 拓扑图），基于机房间丢包率阈值推导（批量正则查询优化）
func (s *DashboardService) GetIdcStatusForSvgGraph(idcs []IDC) (map[string]interface{}, error) {
	th := conf.Get().LostLevelsForIDCBetween
	if th.Warn <= 0 {
		th.Warn = 5
	}
	if th.Alarm <= 0 {
		th.Alarm = 10
	}

	result := make(map[string]interface{})
	for _, idc := range idcs {
		result[idc.Code] = map[string]interface{}{"state": 1, "enable": idc.Enable}
	}

	// 批量查询所有 IDC
	regex := buildIDCRegex(idcs)
	if lostResults, err := s.prom.QueryWithLabels(fmt.Sprintf(`ping_idc_lost{tidc=~"%s"}`, regex)); err == nil {
		for _, r := range lostResults {
			tidc := r.Metric["tidc"]
			if tidc == "" {
				continue
			}
			val := parseMetricFloat(r)
			state := 1
			switch {
			case val >= th.Alarm:
				state = 3
			case val >= th.Warn:
				state = 2
			}
			result[tidc].(map[string]interface{})["state"] = state
		}
	}

	return result, nil
}

// ── 2. 机房间链路状态 ──

// GetIDCLinkStatus 获取链路状态
func (s *DashboardService) GetIDCLinkStatus(links []IDCLink) ([][]interface{}, error) {
	// 构建 source+target -> enable 的映射
	linkEnableMap := make(map[string]bool)
	for _, link := range links {
		linkEnableMap[link.Source+"->"+link.Target] = link.Enable
	}

	results, err := s.prom.QueryIDCLinkStatusResults()
	if err != nil {
		// 降级：全部返回 1（正常）
		statuses := make([][]interface{}, 0, len(links))
		for _, link := range links {
			statuses = append(statuses, []interface{}{link.Source, link.Target, 1, link.Enable})
		}
		return statuses, nil
	}

	statusMap := make(map[string]int)
	for _, r := range results {
		src := r.Metric["sidc"]
		dst := r.Metric["tidc"]
		val, _ := strconv.Atoi(fmt.Sprintf("%v", r.Value[1]))
		if val == 0 {
			val = 1
		}
		statusMap[src+"->"+dst] = val
	}

	statuses := make([][]interface{}, 0, len(links))
	for _, link := range links {
		state := 1
		if s, ok := statusMap[link.Source+"->"+link.Target]; ok {
			state = s
		}
		enable := linkEnableMap[link.Source+"->"+link.Target]
		statuses = append(statuses, []interface{}{link.Source, link.Target, state, enable})
	}
	return statuses, nil
}

// buildIDCRegex 构建 tidc 正则匹配表达式
func buildIDCRegex(idcs []IDC) string {
	codes := make([]string, len(idcs))
	for i, idc := range idcs {
		codes[i] = idc.Code
	}
	return strings.Join(codes, "|")
}

// ── 3. 机房面板详情 ──

// GetIDCStatusPanels 获取机房面板数据（批量正则查询优化）
func (s *DashboardService) GetIDCStatusPanels(idcs []IDC) (map[string]interface{}, error) {
	regex := buildIDCRegex(idcs)
	result := make(map[string]interface{})

	// 初始化所有 IDC 面板
	idcEnable := make(map[string]bool)
	for _, idc := range idcs {
		idcEnable[idc.Code] = idc.Enable
		result[idc.Code] = map[string]interface{}{
			"state":  1,
			"enable": idc.Enable,
			"latency": map[string]interface{}{
				"max": "0.00", "avg": "0.00", "p99": "0.00",
			},
			"package-lost": map[string]interface{}{
				"max": "0.00",
			},
		}
	}

	// 批量查询：一次拉取所有 IDC 的延迟数据
	queryMetric := func(metric string, setter func(tidc string, val float64)) {
		if results, err := s.prom.QueryWithLabels(fmt.Sprintf(`%s{tidc=~"%s"}`, metric, regex)); err == nil {
			for _, r := range results {
				tidc := r.Metric["tidc"]
				if tidc == "" {
					continue
				}
				setter(tidc, parseMetricFloat(r))
			}
		}
	}

	// 状态：p99 延迟阈值
	queryMetric("ping_idc_latency_p99", func(tidc string, val float64) {
		state := 1
		switch {
		case val > 100:
			state = 3
		case val > 30:
			state = 2
		}
		result[tidc].(map[string]interface{})["state"] = state
	})

	// 延迟 max
	queryMetric("ping_latency_by_idc_max", func(tidc string, val float64) {
		result[tidc].(map[string]interface{})["latency"].(map[string]interface{})["max"] = fmt.Sprintf("%.2f", val)
	})

	// 延迟 avg
	queryMetric("ping_latency_by_idc_avg", func(tidc string, val float64) {
		result[tidc].(map[string]interface{})["latency"].(map[string]interface{})["avg"] = fmt.Sprintf("%.2f", val)
	})

	// 延迟 p99
	queryMetric("ping_latency_by_idc_p99", func(tidc string, val float64) {
		result[tidc].(map[string]interface{})["latency"].(map[string]interface{})["p99"] = fmt.Sprintf("%.2f", val)
	})

	// 丢包率 max
	queryMetric("ping_lost_by_idc_max", func(tidc string, val float64) {
		result[tidc].(map[string]interface{})["package-lost"].(map[string]interface{})["max"] = fmt.Sprintf("%.2f", val)
	})

	return result, nil
}

// parseMetricFloat 从 promResult 解析 float64 值
func parseMetricFloat(r promResult) float64 {
	if sv, ok := r.Value[1].(string); ok {
		var v float64
		fmt.Sscanf(sv, "%f", &v)
		return v
	}
	return 0
}

// ── 4. 机房节点状态树 ──
// GetTreeNodesForIdc 获取交换机节点树状状态（批量正则查询优化）
// Core/Spine/Leaf 均取自对应 avg 指标。
func (s *DashboardService) GetTreeNodesForIdc(idcs []IDC) (map[string]interface{}, error) {
	regex := buildIDCRegex(idcs)
	th := conf.Get().LatencyLevelsForIDCInner
	if th.Warn <= 0 {
		th.Warn = 3
	}
	if th.Alarm <= 0 {
		th.Alarm = 10
	}

	result := make(map[string]interface{})
	for _, idc := range idcs {
		result[idc.Code] = map[string]interface{}{
			"core":   []int{1, 1},
			"spine1": []int{1, 1},
			"spine2": []int{1, 1},
			"leaf1":  []int{1, 1, 1},
			"leaf2":  []int{1, 1, 1},
			"enable": idc.Enable,
		}
	}

	// 1. Core：批量查询所有 IDC
	if coreResults, err := s.prom.QueryWithLabels(fmt.Sprintf(`ping_latency_by_idc_avg{tidc=~"%s"}`, regex)); err == nil {
		for _, r := range coreResults {
			tidc := r.Metric["tidc"]
			if tidc == "" {
				continue
			}
			val := parseMetricFloat(r)
			cs := nodeStatus(val, th.Warn, th.Alarm)
			if panel, ok := result[tidc].(map[string]interface{}); ok {
				panel["core"] = []int{cs, cs}
			}
		}
	}

	// 2. Spine：批量查询按 tspine 分组
	if spineResults, err := s.prom.QueryWithLabels(fmt.Sprintf(`ping_latency_by_spine_avg{tidc=~"%s"}`, regex)); err == nil {
		for _, r := range spineResults {
			tidc := r.Metric["tidc"]
			tspine := r.Metric["tspine"]
			if tidc == "" || tspine == "" {
				continue
			}
			val := parseMetricFloat(r)
			s := nodeStatus(val, th.Warn, th.Alarm)
			if panel, ok := result[tidc].(map[string]interface{}); ok {
				switch tspine {
				case "APP":
					panel["spine1"] = []int{s, s}
				case "OM":
					panel["spine2"] = []int{s, s}
				}
			}
		}
	}

	// 3. Leaf：批量查询按 tspine 分组统计
	if leafResults, err := s.prom.QueryWithLabels(fmt.Sprintf(`ping_latency_by_leaf_avg{tidc=~"%s"}`, regex)); err == nil {
		// 按 IDC + tspine 聚合告警/警告计数
		type idcSpineKey struct {
			idc, spine string
		}
		alarmCount := make(map[idcSpineKey]int)
		warnCount := make(map[idcSpineKey]int)
		for _, r := range leafResults {
			tidc := r.Metric["tidc"]
			tspine := r.Metric["tspine"]
			if tidc == "" || tspine == "" {
				continue
			}
			key := idcSpineKey{tidc, tspine}
			val := parseMetricFloat(r)
			if val >= th.Alarm {
				alarmCount[key]++
			} else if val >= th.Warn {
				warnCount[key]++
			}
		}
		// 填充 leaf 状态
		for _, idc := range idcs {
			if panel, ok := result[idc.Code].(map[string]interface{}); ok {
				panel["leaf1"] = fillLeafStatus(3, alarmCount[idcSpineKey{idc.Code, "APP"}], warnCount[idcSpineKey{idc.Code, "APP"}])
				panel["leaf2"] = fillLeafStatus(3, alarmCount[idcSpineKey{idc.Code, "OM"}], warnCount[idcSpineKey{idc.Code, "OM"}])
			}
		}
	}

	return result, nil
}

// nodeStatus 根据延迟值和阈值返回节点状态：1=normal, 2=warn, 3=alarm
func nodeStatus(val, warnTh, alarmTh float64) int {
	if val >= alarmTh {
		return 3
	}
	if val >= warnTh {
		return 2
	}
	return 1
}

// fillLeafStatus 按 alarm/warn 计数填充 Leaf 节点状态数组
func fillLeafStatus(n, alarmCount, warnCount int) []int {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		switch {
		case i < alarmCount:
			arr[i] = 3
		case i < alarmCount+warnCount:
			arr[i] = 2
		default:
			arr[i] = 1
		}
	}
	return arr
}

func getIndexFromMetric(r promResult, key string, defaultVal int) int {
	if val, ok := r.Metric[key]; ok {
		var idx int
		fmt.Sscanf(val, "%d", &idx)
		return idx
	}
	return defaultVal
}

// ── 5. 机房间网络概要 ──

// GetSummaryStatus 获取各机房间网络时延和丢包率
func (s *DashboardService) GetSummaryStatus() (map[string]interface{}, error) {
	// 各机房时延
	idcLatency := make(map[string]string)
	if rs, _ := s.prom.QueryWithLabels(`ping_idc_latency_avg`); len(rs) > 0 {
		for _, r := range rs {
			tidc := r.Metric["tidc"]
			if tidc == "" {
				continue
			}
			if sv, ok := r.Value[1].(string); ok {
				if v, err := strconv.ParseFloat(sv, 64); err == nil {
					idcLatency[tidc] = fmt.Sprintf("%.2f", v)
				}
			}
		}
	}

	// 各机房丢包率
	idcLoss := make(map[string]string)
	if rs, _ := s.prom.QueryWithLabels(`ping_idc_lost`); len(rs) > 0 {
		for _, r := range rs {
			tidc := r.Metric["tidc"]
			if tidc == "" {
				continue
			}
			if sv, ok := r.Value[1].(string); ok {
				if v, err := strconv.ParseFloat(sv, 64); err == nil {
					idcLoss[tidc] = fmt.Sprintf("%.2f", v)
				}
			}
		}
	}

	return map[string]interface{}{
		"idcLatency": idcLatency,
		"idcLoss":    idcLoss,
	}, nil
}

// ── 6. 整体统计 ──

// GetOverallStats 获取整体时延和丢包率
func (s *DashboardService) GetOverallStats() (map[string]interface{}, error) {
	avgLat, _ := s.prom.Query(`ping_latency_overall_avg`)
	p99Lat, _ := s.prom.Query(`ping_latency_overall_p99`)
	loss, _ := s.prom.Query(`ping_lost_overall`)

	return map[string]interface{}{
		"overall_avg_latency": fmt.Sprintf("%.2f", avgLat),
		"overall_p99_latency": fmt.Sprintf("%.2f", p99Lat),
		"overall_loss":        fmt.Sprintf("%.2f", loss),
	}, nil
}

// ── DB 告警查询 ──

// GetAlertPage 分页查询告警（DB），fallback 到空结果
func (s *DashboardService) GetAlertPage(page, pageSize int) (*PageResult, error) {
	if s.db == nil {
		return &PageResult{Total: 0, Page: page, PageSize: pageSize, List: []AlertRecord{}}, nil
	}
	return s.db.ListAlerts(page, pageSize, "")
}

// GetUnresolvedAlertPage 分页查询未处置告警（排除"已处置"），用于监控大屏轮播
func (s *DashboardService) GetUnresolvedAlertPage(page, pageSize int) (*PageResult, error) {
	if s.db == nil {
		return &PageResult{Total: 0, Page: page, PageSize: pageSize, List: []AlertRecord{}}, nil
	}
	return s.db.ListUnresolvedAlerts(page, pageSize)
}

// GetAlertTypeStats 按类型统计（DB），fallback 到空
func (s *DashboardService) GetAlertTypeStats() ([]AlertTypeStat, error) {
	if s.db == nil {
		return []AlertTypeStat{}, nil
	}
	return s.db.CountByAlertType()
}

// GetAlertStatusStats 按状态统计（DB），fallback 到空
func (s *DashboardService) GetAlertStatusStats() ([]AlertStatusStat, error) {
	if s.db == nil {
		return []AlertStatusStat{}, nil
	}
	return s.db.CountByAlertStatus()
}

// GetUnresolvedAlertNames 获取未解决告警名称列表（DB），fallback 到空
func (s *DashboardService) GetUnresolvedAlertNames(limit int) ([]string, error) {
	if s.db == nil {
		return []string{}, nil
	}
	return s.db.ListUnresolvedNames(limit)
}
