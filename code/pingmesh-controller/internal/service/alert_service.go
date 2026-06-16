package service

import (
	"fmt"
)

// AlertService 告警 Service
type AlertService struct {
	db   *DB
	prom *PromClient
}

// NewAlertService 创建告警 Service
func NewAlertService(db *DB, prom *PromClient) *AlertService {
	return &AlertService{db: db, prom: prom}
}

// ── DB 操作（告警管理页面用） ──

// GetAlertPage 分页查询，支持 status 筛选
func (s *AlertService) GetAlertPage(page, pageSize int, status string) (*PageResult, error) {
	if s.db == nil {
		return &PageResult{Total: 0, Page: page, PageSize: pageSize, List: []AlertRecord{}}, nil
	}
	return s.db.ListAlerts(page, pageSize, status)
}

// GetAlertPageByStatuses 多状态筛选分页查询
func (s *AlertService) GetAlertPageByStatuses(page, pageSize int, statuses []string) (*PageResult, error) {
	if s.db == nil {
		return &PageResult{Total: 0, Page: page, PageSize: pageSize, List: []AlertRecord{}}, nil
	}
	return s.db.ListAlertsByStatuses(page, pageSize, statuses)
}

// UpdateAlertStatus 更新告警状态
func (s *AlertService) UpdateAlertStatus(id int64, newStatus string) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}
	return s.db.UpdateAlertStatus(id, newStatus)
}

// GetUnresolvedCount 获取未处置告警数量
func (s *AlertService) GetUnresolvedCount() (int, error) {
	if s.db == nil {
		return 0, nil
	}
	return s.db.CountUnresolved()
}

// GetStatusStats 获取按状态统计
func (s *AlertService) GetStatusStats() ([]AlertStatusStat, error) {
	if s.db == nil {
		return []AlertStatusStat{}, nil
	}
	return s.db.CountByAlertStatus()
}
