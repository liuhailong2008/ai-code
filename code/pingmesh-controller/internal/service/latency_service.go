package service

import (
	"fmt"
)

// LatencyService 机房耗时监控 Service
type LatencyService struct {
	db   *DB
	prom *PromClient
}

// NewLatencyService 创建机房耗时监控 Service
func NewLatencyService(db *DB, prom *PromClient) *LatencyService {
	return &LatencyService{db: db, prom: prom}
}

// ── 1. 耗时-丢包率散点图 ──

// ScatterPoint 散点数据
type ScatterPoint struct {
	Value     []interface{}          `json:"value"`
	ItemStyle map[string]interface{} `json:"itemStyle"`
}

// AbnormalDevice 异常设备信息
type AbnormalDevice struct {
	Name   string  `json:"name"`
	Loss   float64 `json:"loss"`
	Avg    float64 `json:"avg"`
	P99    float64 `json:"p99"`
	Max    float64 `json:"max"`
	Status string  `json:"status"`
}

// GetScatterData 获取散点图数据
// stat: avg / p99 / max
func (s *LatencyService) GetScatterData(idcCode, stat string) (map[string]interface{}, error) {
	// 查询 recording rules: 时延和丢包率
	latencyResults, err := s.prom.QueryLatencyByLeaf(idcCode, stat)
	if err != nil {
		return nil, fmt.Errorf("query latency: %w", err)
	}
	lostResults, err := s.prom.QueryLostByLeaf(idcCode, stat)
	if err != nil {
		return nil, fmt.Errorf("query lost: %w", err)
	}

	// 按 tleaf 建立丢包率索引
	lostByLeaf := make(map[string]float64)
	for _, r := range lostResults {
		leaf := r.Metric["tleaf"]
		if leaf == "" {
			continue
		}
		if v, ok := r.Value[1].(string); ok {
			var val float64
			fmt.Sscanf(v, "%f", &val)
			lostByLeaf[leaf] = val
		}
	}

	// 构建散点数据
	points := make([]ScatterPoint, 0, len(latencyResults))
	abnormal := make([]AbnormalDevice, 0)
	var normal, warning, critical int

	for _, r := range latencyResults {
		leaf := r.Metric["tleaf"]
		if leaf == "" {
			continue
		}

		// 解析时延值
		var latency float64
		if v, ok := r.Value[1].(string); ok {
			fmt.Sscanf(v, "%f", &latency)
		}

		// 获取丢包率
		loss := lostByLeaf[leaf]

		status := "normal"
		color := "#39ff7e"

		if latency > 150 && loss > 2.5 {
			status = "critical"
			color = "#ff4560"
			critical++
			abnormal = append(abnormal, AbnormalDevice{
				Name: leaf, Loss: loss, Avg: latency,
				P99: latency, Max: latency, Status: "critical",
			})
		} else if latency > 150 || loss > 2.5 {
			status = "warning"
			color = "#ffab00"
			warning++
			abnormal = append(abnormal, AbnormalDevice{
				Name: leaf, Loss: loss, Avg: latency,
				P99: latency, Max: latency, Status: "warning",
			})
		} else {
			normal++
		}

		points = append(points, ScatterPoint{
			Value:     []interface{}{latency, loss, leaf, status},
			ItemStyle: map[string]interface{}{"color": color},
		})
	}

	total := len(points)

	return map[string]interface{}{
		"points":          points,
		"total":           total,
		"normal":          normal,
		"warning":         warning,
		"critical":        critical,
		"abnormalDevices": abnormal,
	}, nil
}
