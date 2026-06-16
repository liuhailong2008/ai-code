package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pingmesh-controller/internal/common"
	"pingmesh-controller/internal/service"
)

// AlertHandler 告警页面 Handler
type AlertHandler struct {
	svc *service.AlertService
}

// AlertsList 分页查询告警列表（DB），支持 status 或 statuses（逗号分隔多选）筛选
func (h *AlertHandler) AlertsList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	status := r.URL.Query().Get("status")
	statusesStr := r.URL.Query().Get("statuses")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 5
	}

	var data *service.PageResult
	var err error
	if statusesStr != "" {
		statuses := strings.Split(statusesStr, ",")
		data, err = h.svc.GetAlertPageByStatuses(page, pageSize, statuses)
	} else {
		data, err = h.svc.GetAlertPage(page, pageSize, status)
	}
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, data)
}

// AlertsStats 按状态统计
func (h *AlertHandler) AlertsStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStatusStats()
	if err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, stats)
}

// UpdateStatus 更新告警状态
func (h *AlertHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteJSON(w, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request"})
		return
	}
	if err := h.svc.UpdateAlertStatus(req.ID, req.Status); err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]string{"ok": "updated"})
}

// UnresolvedCount 未处置告警数量（侧边栏 badge）
func (h *AlertHandler) UnresolvedCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.svc.GetUnresolvedCount()
	if err != nil {
		common.WriteJSON(w, map[string]interface{}{"count": 0})
		return
	}
	common.WriteJSON(w, map[string]interface{}{"count": count})
}
