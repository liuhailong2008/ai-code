package handler

import (
	"net/http"

	"pingmesh-controller/internal/common"
	"pingmesh-controller/internal/service"
)

// LatencyHandler 机房耗时监控 Handler
type LatencyHandler struct {
	svc *service.LatencyService
}

func (h *LatencyHandler) Scatter(w http.ResponseWriter, r *http.Request) {
	idcCode := r.URL.Query().Get("idc")
	stat := r.URL.Query().Get("stat")
	if stat == "" {
		stat = "avg"
	}
	data, err := h.svc.GetScatterData(idcCode, stat)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}
