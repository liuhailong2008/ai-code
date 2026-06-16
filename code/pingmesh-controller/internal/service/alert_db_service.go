package service

import (
	"database/sql"
	"fmt"
	"pingmesh-controller/internal/common"
	"strings"
	"time"
)

// AlertRecord 告警记录
type AlertRecord struct {
	ID          int64     `json:"id"`
	AlertIDC    string    `json:"alert_idc"`
	AlertSpine  string    `json:"alert_spine"`
	AlertLeaf   string    `json:"alert_leaf"`
	AlertRack   string    `json:"alert_rack"`
	AlertType   string    `json:"alert_type"`
	AlertStatus string    `json:"alert_status"`
	CreateBy    string    `json:"create_by"`
	CreateTime  time.Time `json:"create_time"`
	UpdateBy    string    `json:"update_by"`
	UpdateTime  time.Time `json:"update_time"`
}

// AlertTypeStat 按告警类型统计
type AlertTypeStat struct {
	AlertType string `json:"alert_type"`
	Count     int    `json:"count"`
}

// AlertStatusStat 按告警状态统计
type AlertStatusStat struct {
	AlertStatus string `json:"alert_status"`
	Count       int    `json:"count"`
}

// PageResult 分页结果
type PageResult struct {
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	List     []AlertRecord `json:"list"`
}

// alertColumns 告警记录 SELECT 列，使用 COALESCE 处理可空字段
const alertColumns = `id, alert_idc,
COALESCE(alert_spine, '') AS alert_spine,
COALESCE(alert_leaf, '') AS alert_leaf,
COALESCE(alert_rack, '') AS alert_rack,
COALESCE(alert_type, '') AS alert_type,
COALESCE(alert_status, '') AS alert_status,
COALESCE(create_by, '') AS create_by,
create_time,
COALESCE(update_by, '') AS update_by,
update_time`

// scanAlert 扫描一行告警记录
func scanAlert(scanner interface {
	Scan(dest ...interface{}) error
}) (AlertRecord, error) {
	var a AlertRecord
	err := scanner.Scan(&a.ID, &a.AlertIDC, &a.AlertSpine, &a.AlertLeaf, &a.AlertRack,
		&a.AlertType, &a.AlertStatus, &a.CreateBy, &a.CreateTime, &a.UpdateBy, &a.UpdateTime)
	return a, err
}

// ListAlerts 分页查询告警，按 create_time 倒序，支持 status 筛选
// status 参数为中文展示值，内部映射为 DB 英文值
func (d *DB) ListAlerts(page, pageSize int, status string) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 5
	}
	offset := (page - 1) * pageSize

	var total int
	var err error
	if status != "" {
		internalStatus := common.MapAlertStatusReverse(status)
		err = d.conn.QueryRow("SELECT COUNT(*) FROM alert_record WHERE alert_status = ?", internalStatus).Scan(&total)
	} else {
		err = d.conn.QueryRow("SELECT COUNT(*) FROM alert_record").Scan(&total)
	}
	if err != nil {
		return nil, fmt.Errorf("count alerts: %w", err)
	}

	var rows *sql.Rows
	if status != "" {
		internalStatus := common.MapAlertStatusReverse(status)
		querySQL := `SELECT ` + alertColumns + ` FROM alert_record WHERE alert_status = ? ORDER BY create_time DESC, id DESC LIMIT ? OFFSET ?`
		rows, err = d.conn.Query(querySQL, internalStatus, pageSize, offset)
	} else {
		rows, err = d.conn.Query(`SELECT `+alertColumns+` FROM alert_record ORDER BY create_time DESC, id DESC LIMIT ? OFFSET ?`, pageSize, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("query alerts: %w", err)
	}
	defer rows.Close()

	list := make([]AlertRecord, 0, pageSize)
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		a.AlertStatus = common.MapAlertStatus(a.AlertStatus)
		list = append(list, a)
	}
	return &PageResult{Total: total, Page: page, PageSize: pageSize, List: list}, rows.Err()
}

// ListAlertsByStatuses 多状态筛选分页查询
// statuses 参数为中文展示值，内部映射为 DB 英文值后 IN 查询
func (d *DB) ListAlertsByStatuses(page, pageSize int, statuses []string) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 5
	}
	offset := (page - 1) * pageSize

	dbValues := make([]interface{}, len(statuses))
	for i, s := range statuses {
		dbValues[i] = common.MapAlertStatusReverse(s)
	}

	placeholders := make([]string, len(dbValues))
	for i := range dbValues {
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ",")

	var total int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM alert_record WHERE alert_status IN (%s)", inClause)
	if err := d.conn.QueryRow(countSQL, dbValues...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count by statuses: %w", err)
	}

	querySQL := fmt.Sprintf(`SELECT %s FROM alert_record WHERE alert_status IN (%s) ORDER BY create_time DESC, id DESC LIMIT ? OFFSET ?`, alertColumns, inClause)
	queryArgs := append(dbValues, pageSize, offset)

	rows, err := d.conn.Query(querySQL, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("query by statuses: %w", err)
	}
	defer rows.Close()

	list := make([]AlertRecord, 0, pageSize)
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, fmt.Errorf("scan statuses: %w", err)
		}
		a.AlertStatus = common.MapAlertStatus(a.AlertStatus)
		list = append(list, a)
	}
	return &PageResult{Total: total, Page: page, PageSize: pageSize, List: list}, rows.Err()
}

// CountByAlertType 按告警类型分组统计（最近7天）
func (d *DB) CountByAlertType() ([]AlertTypeStat, error) {
	rows, err := d.conn.Query(`SELECT COALESCE(alert_type, ''), COUNT(*) FROM alert_record
		WHERE create_time >= DATE_SUB(NOW(), INTERVAL 7 DAY)
		GROUP BY alert_type ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, fmt.Errorf("count by type: %w", err)
	}
	defer rows.Close()

	stats := make([]AlertTypeStat, 0)
	for rows.Next() {
		var s AlertTypeStat
		if err := rows.Scan(&s.AlertType, &s.Count); err != nil {
			return nil, fmt.Errorf("scan type: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// CountByAlertStatus 按告警状态分组统计（最近7天），使用 CASE WHEN 映射为中文展示值
func (d *DB) CountByAlertStatus() ([]AlertStatusStat, error) {
	rows, err := d.conn.Query(`SELECT 
		CASE 
			WHEN alert_status = 'UNPROCESSED' OR alert_status = '' OR alert_status IS NULL THEN '告警中'
			WHEN alert_status = 'PROCESSING' THEN '处置中'
			WHEN alert_status = 'PROCESSED' THEN '已处置'
			ELSE '告警中'
		END as display_status,
		COUNT(*) 
		FROM alert_record 
		WHERE create_time >= DATE_SUB(NOW(), INTERVAL 7 DAY)
		GROUP BY display_status 
		ORDER BY FIELD(display_status, '告警中', '处置中', '已处置')`)
	if err != nil {
		return nil, fmt.Errorf("count by status: %w", err)
	}
	defer rows.Close()

	stats := make([]AlertStatusStat, 0)
	for rows.Next() {
		var s AlertStatusStat
		if err := rows.Scan(&s.AlertStatus, &s.Count); err != nil {
			return nil, fmt.Errorf("scan status: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// ListUnresolvedNames 查询未解决的告警名称，按 create_time 升序
func (d *DB) ListUnresolvedNames(limit int) ([]string, error) {
	if limit < 1 {
		limit = 50
	}
	rows, err := d.conn.Query(`SELECT alert_idc, COALESCE(alert_leaf, ''), COALESCE(alert_type, ''), create_time, COALESCE(alert_status, '') FROM alert_record
		WHERE alert_status NOT IN ('PROCESSED') AND alert_leaf IS NOT NULL AND alert_leaf <> ''
		ORDER BY create_time ASC, id ASC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query unresolved: %w", err)
	}
	defer rows.Close()

	names := make([]string, 0)
	for rows.Next() {
		var idc, leaf, atype, status string
		var createTime time.Time
		if err := rows.Scan(&idc, &leaf, &atype, &createTime, &status); err != nil {
			return nil, fmt.Errorf("scan name: %w", err)
		}
		timeStr := createTime.Format("15:04:05")
		displayStatus := common.MapAlertStatus(status)
		names = append(names, fmt.Sprintf("[%s] %s %s %s -%s-", idc, timeStr, leaf, atype, displayStatus))
	}
	return names, rows.Err()
}

// UpdateAlertStatus 更新告警状态，内部将中文展示值映射为英文后写入 DB
func (d *DB) UpdateAlertStatus(id int64, newStatus string) error {
	internalStatus := common.MapAlertStatusReverse(newStatus)
	_, err := d.conn.Exec("UPDATE alert_record SET alert_status = ?, update_by = 'system', update_time = NOW() WHERE id = ?", internalStatus, id)
	return err
}

// CountUnresolved 统计未处置的告警数量
func (d *DB) CountUnresolved() (int, error) {
	var count int
	err := d.conn.QueryRow("SELECT COUNT(*) FROM alert_record WHERE alert_status != 'PROCESSED'").Scan(&count)
	return count, err
}

// ListUnresolvedAlerts 分页查询未处置告警，用于监控大屏轮播
func (d *DB) ListUnresolvedAlerts(page, pageSize int) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 4
	}
	offset := (page - 1) * pageSize

	var total int
	if err := d.conn.QueryRow("SELECT COUNT(*) FROM alert_record WHERE alert_status != 'PROCESSED'").Scan(&total); err != nil {
		return nil, fmt.Errorf("count unresolved: %w", err)
	}

	rows, err := d.conn.Query(`SELECT `+alertColumns+` FROM alert_record WHERE alert_status != 'PROCESSED' ORDER BY create_time DESC, id DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query unresolved: %w", err)
	}
	defer rows.Close()

	list := make([]AlertRecord, 0, pageSize)
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, fmt.Errorf("scan unresolved: %w", err)
		}
		a.AlertStatus = common.MapAlertStatus(a.AlertStatus)
		list = append(list, a)
	}
	return &PageResult{Total: total, Page: page, PageSize: pageSize, List: list}, rows.Err()
}
