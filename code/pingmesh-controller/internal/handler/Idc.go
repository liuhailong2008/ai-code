package handler

import (
	"encoding/json"
	"net/http"

	"pingmesh-controller/conf"
	"pingmesh-controller/internal/common"
	"pingmesh-controller/internal/service"
)

// Handler 聚合所有 handler，持有 service 依赖
type Handler struct {
	Dashboard  *DashboardHandler
	Leaf       *LeafHandler
	Latency    *LatencyHandler
	Alert      *AlertHandler
	Statistics *StatisticsHandler
}

// NewHandler 创建 Handler 实例
func NewHandler(dashboardSvc *service.DashboardService, leafSvc *service.LeafService,
	latencySvc *service.LatencyService, alertSvc *service.AlertService,
	statisticsSvc *service.StatisticsService,
) *Handler {
	return &Handler{
		Dashboard:  &DashboardHandler{svc: dashboardSvc},
		Leaf:       &LeafHandler{svc: leafSvc},
		Latency:    &LatencyHandler{svc: latencySvc},
		Alert:      &AlertHandler{svc: alertSvc},
		Statistics: &StatisticsHandler{svc: statisticsSvc},
	}
}

// IDCs 获取或更新机房列表
// GET  -> 返回机房列表
// POST -> 更新机房 name / enable 并持久化
func (h *Handler) IDCs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.updateIDCs(w, r)
		return
	}

	cfg := conf.Get()
	result := make([]common.IDC, len(cfg.IDCs))
	for i, idc := range cfg.IDCs {
		result[i] = common.IDC{Code: idc.Code, Name: idc.Name, Enable: idc.Enable}
	}
	common.WriteJSON(w, result)
}

func (h *Handler) updateIDCs(w http.ResponseWriter, r *http.Request) {
	var idcs []common.IDC
	if err := json.NewDecoder(r.Body).Decode(&idcs); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}

	cfgIDCs := make([]conf.IDCConfig, len(idcs))
	for i, idc := range idcs {
		cfgIDCs[i] = conf.IDCConfig{
			Code:   idc.Code,
			Name:   idc.Name,
			Enable: idc.Enable,
		}
	}

	if err := conf.UpdateIDCs(cfgIDCs); err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}

	common.WriteJSON(w, map[string]string{"status": "ok"})
}
