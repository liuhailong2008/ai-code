package handler

import (
	"encoding/json"
	"net/http"

	"pingmesh-controller/conf"
	"pingmesh-controller/internal/common"
	"pingmesh-controller/internal/service"
)

// StatisticsHandler 统计数据处理器
type StatisticsHandler struct {
	svc *service.StatisticsService
}

// ListMetrics 获取指标列表（GET /api/statistics/metrics）
func (h *StatisticsHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteJSON(w, map[string]string{"error": "method not allowed"})
		return
	}

	metrics, err := h.svc.ListMetrics()
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, metrics)
}

// QueryMetricData 查询指标数据（GET /api/statistics/metric-data?metric=xxx&label_xxx=yyy）
func (h *StatisticsHandler) QueryMetricData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteJSON(w, map[string]string{"error": "method not allowed"})
		return
	}

	metric := r.URL.Query().Get("metric")
	if metric == "" {
		common.WriteJSON(w, map[string]string{"error": "metric parameter is required"})
		return
	}

	// 从 query params 构建 label 筛选条件
	labels := make(map[string]string)
	for key, values := range r.URL.Query() {
		if key != "metric" && len(values) > 0 && values[0] != "" {
			labels[key] = values[0]
		}
	}

	data, err := h.svc.QueryMetricData(metric, labels)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

// GetMetricLabels 获取指标标签信息（GET /api/statistics/metric-labels?metric=xxx）
func (h *StatisticsHandler) GetMetricLabels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteJSON(w, map[string]string{"error": "method not allowed"})
		return
	}

	metric := r.URL.Query().Get("metric")
	if metric == "" {
		common.WriteJSON(w, map[string]string{"error": "metric parameter is required"})
		return
	}

	labels, err := h.svc.GetMetricLabels(metric)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, labels)
}

// MetricsMetaConfig 获取或更新指标元数据配置
// GET  -> 返回当前配置
// POST -> 更新配置并持久化
func (h *StatisticsHandler) MetricsMetaConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.updateMetricsMeta(w, r)
		return
	}

	meta := h.svc.GetMetricsMetaConfig()
	common.WriteJSON(w, meta)
}

func (h *StatisticsHandler) updateMetricsMeta(w http.ResponseWriter, r *http.Request) {
	var meta []conf.MetricMetaConfig
	if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.svc.UpdateMetricsMetaConfig(meta); err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}

	common.WriteJSON(w, map[string]string{"status": "ok"})
}
