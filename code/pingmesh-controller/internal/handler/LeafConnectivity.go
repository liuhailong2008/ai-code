package handler

import (
	"net/http"

	"pingmesh-controller/internal/common"
	"pingmesh-controller/internal/service"
)

// LeafHandler 机房网络监控 Handler
type LeafHandler struct {
	svc *service.LeafService
}

func (h *LeafHandler) Leafs(w http.ResponseWriter, r *http.Request) {
	idcCode := r.URL.Query().Get("idc")
	leafs, err := h.svc.GetLeafs(idcCode)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]interface{}{"leafs": leafs})
}

func (h *LeafHandler) Heatmap(w http.ResponseWriter, r *http.Request) {
	idcCode := r.URL.Query().Get("idc")
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "avg"
	}

	data, err := h.svc.GetHeatmapData(idcCode, metric)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}

	// 同时返回 leaf 名称列表，方便前端做坐标轴映射
	leafs, _ := h.svc.GetLeafs(idcCode)
	common.WriteJSON(w, map[string]interface{}{
		"data":   data,
		"leafs":  leafs,
		"metric": metric,
		"idc":    idcCode,
	})
}

// HeatmapRange 获取时间范围内的热力图数据（预加载用）
// 参数: idc, metric, start, end, step
func (h *LeafHandler) HeatmapRange(w http.ResponseWriter, r *http.Request) {
	idcCode := r.URL.Query().Get("idc")
	metric := r.URL.Query().Get("metric")
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")
	step := r.URL.Query().Get("step")
	if metric == "" {
		metric = "avg"
	}
	if step == "" {
		step = "60s"
	}

	points, leafs, err := h.svc.GetHeatmapRangeData(idcCode, metric, startTime, endTime, step)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]interface{}{
		"points": points,
		"leafs":  leafs,
		"metric": metric,
		"idc":    idcCode,
	})
}

func (h *LeafHandler) TopLinks(w http.ResponseWriter, r *http.Request) {
	idcCode := r.URL.Query().Get("idc")
	links, err := h.svc.GetTopLinks(idcCode)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]interface{}{"links": links})
}

func (h *LeafHandler) MaxAllLinks(w http.ResponseWriter, r *http.Request) {
	links, err := h.svc.GetMaxAllLinks()
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]interface{}{"links": links})
}

func (h *LeafHandler) TimeOptions(w http.ResponseWriter, r *http.Request) {
	options := h.svc.GetTimeOptions()
	common.WriteJSON(w, map[string]interface{}{"timeOptions": options})
}
