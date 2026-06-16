package handler

import (
	"net/http"
	"strconv"

	"pingmesh-controller/conf"
	"pingmesh-controller/internal/common"
	"pingmesh-controller/internal/service"
)

// DashboardHandler 大屏监控 Handler
type DashboardHandler struct {
	svc *service.DashboardService
}

func (h *DashboardHandler) SummaryIDCStatus(w http.ResponseWriter, r *http.Request) {
	idcs := idcsFromConfig()
	data, err := h.svc.GetIdcStatusForSvgGraph(idcs)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

func (h *DashboardHandler) IDCLinkStatus(w http.ResponseWriter, r *http.Request) {
	links := idcLinksFromConfig()
	if len(links) == 0 {
		common.WriteJSON(w, []interface{}{})
		return
	}
	data, err := h.svc.GetIDCLinkStatus(links)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

func (h *DashboardHandler) IDCStatus(w http.ResponseWriter, r *http.Request) {
	idcs := idcsFromConfig()
	data, err := h.svc.GetIDCStatusPanels(idcs)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

func (h *DashboardHandler) IDCNodes(w http.ResponseWriter, r *http.Request) {
	idcs := idcsFromConfig()
	data, err := h.svc.GetTreeNodesForIdc(idcs)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

func (h *DashboardHandler) SummaryStatus(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetSummaryStatus()
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

func (h *DashboardHandler) OverallStats(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetOverallStats()
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

func (h *DashboardHandler) AlertPage(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 4
	}
	data, err := h.svc.GetAlertPage(page, pageSize)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

// UnresolvedAlertPage 未处置告警分页（排除"已处置"），用于监控大屏轮播
func (h *DashboardHandler) UnresolvedAlertPage(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 4
	}
	data, err := h.svc.GetUnresolvedAlertPage(page, pageSize)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

// AlertTypeStats 按类型统计
func (h *DashboardHandler) AlertTypeStats(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetAlertTypeStats()
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

// AlertStatusStats 按状态统计
func (h *DashboardHandler) AlertStatusStats(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetAlertStatusStats()
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

// Thresholds 返回机房内和机房间时延和丢包率判定阈值
func (h *DashboardHandler) Thresholds(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.updateThresholds(w, r)
		return
	}
	cfg := conf.Get()
	common.WriteJSON(w, map[string]interface{}{
		"latency_levels_for_idc_inner": map[string]float64{
			"warn":  cfg.LatencyLevelsForIDCInner.Warn,
			"alarm": cfg.LatencyLevelsForIDCInner.Alarm,
		},
		"lost_levels_for_idc_inner": map[string]float64{
			"warn":  cfg.LostLevelsForIDCInner.Warn,
			"alarm": cfg.LostLevelsForIDCInner.Alarm,
		},
		"latency_levels_for_idc_between": map[string]float64{
			"warn":  cfg.LatencyLevelsForIDCBetween.Warn,
			"alarm": cfg.LatencyLevelsForIDCBetween.Alarm,
		},
		"lost_levels_for_idc_between": map[string]float64{
			"warn":  cfg.LostLevelsForIDCBetween.Warn,
			"alarm": cfg.LostLevelsForIDCBetween.Alarm,
		},
	})
}

func (h *DashboardHandler) updateThresholds(w http.ResponseWriter, r *http.Request) {
	var body struct {
		LatencyLevelsForIDCInner   conf.LatencyLevelsForIDCInner   `json:"latency_levels_for_idc_inner"`
		LostLevelsForIDCInner      conf.LostLevelsForIDCInner      `json:"lost_levels_for_idc_inner"`
		LatencyLevelsForIDCBetween conf.LatencyLevelsForIDCBetween `json:"latency_levels_for_idc_between"`
		LostLevelsForIDCBetween    conf.LostLevelsForIDCBetween    `json:"lost_levels_for_idc_between"`
	}
	if err := common.ReadJSON(r, &body); err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	if err := conf.UpdateThresholds(body.LatencyLevelsForIDCInner, body.LostLevelsForIDCInner, body.LatencyLevelsForIDCBetween, body.LostLevelsForIDCBetween); err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]string{"status": "ok"})
}

// UnresolvedAlerts 未解决告警名称
func (h *DashboardHandler) UnresolvedAlerts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	data, err := h.svc.GetUnresolvedAlertNames(limit)
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]interface{}{"names": data})
}

// idcsFromConfig 从 config.yaml 获取 IDC 列表（唯一数据源）
func idcsFromConfig() []service.IDC {
	cfg := conf.Get()
	idcs := make([]service.IDC, len(cfg.IDCs))
	for i, idc := range cfg.IDCs {
		idcs[i] = service.IDC{Code: idc.Code, Name: idc.Name, Enable: idc.Enable}
	}
	return idcs
}

// idcLinksFromConfig 从 config.yaml 获取 IDC 间链路列表（唯一数据源）
func idcLinksFromConfig() []service.IDCLink {
	cfg := conf.Get()
	links := make([]service.IDCLink, len(cfg.IDCLinks))
	for i, link := range cfg.IDCLinks {
		links[i] = service.IDCLink{Source: link.Source, Target: link.Target, Enable: link.Enable}
	}
	return links
}
