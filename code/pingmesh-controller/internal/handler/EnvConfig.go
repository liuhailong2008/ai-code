package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pingmesh-controller/conf"
	"pingmesh-controller/internal/common"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLView 前端展示用的 MySQL 配置（密码不回显）
type MySQLView struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// EnvConfigGetMySQL GET /api/env-config/mysql
func EnvConfigGetMySQL(w http.ResponseWriter, r *http.Request) {
	cfg := conf.Get()
	v := MySQLView{
		Host:     cfg.MySQL.Host,
		Port:     cfg.MySQL.Port,
		User:     cfg.MySQL.User,
		Password: "",
		Database: cfg.MySQL.Database,
	}
	// 密码不回显，仅提示是否已设置
	if cfg.MySQL.Password != "" {
		v.Password = "******"
	}
	common.WriteJSON(w, v)
}

// EnvConfigSaveMySQL POST /api/env-config/mysql
func EnvConfigSaveMySQL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}

	cfg := conf.Get()
	// 如果传入密码为 "******" 或空，保留原密码
	password := body.Password
	if password == "******" || password == "" {
		password = cfg.MySQL.Password
	}

	newCfg := conf.MySQLConfig{
		Host:     body.Host,
		Port:     body.Port,
		User:     body.User,
		Password: password,
		Database: body.Database,
	}

	if err := conf.UpdateMySQL(newCfg); err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]string{"status": "ok"})
}

// EnvConfigTestMySQL POST /api/env-config/mysql/test
func EnvConfigTestMySQL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}

	cfg := conf.Get()
	password := body.Password
	if password == "******" || password == "" {
		password = cfg.MySQL.Password
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&timeout=5s",
		body.User, password, body.Host, body.Port, body.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("打开数据库连接失败: %v", err)})
		return
	}
	defer db.Close()

	db.SetConnMaxLifetime(5 * time.Second)
	if err := db.Ping(); err != nil {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("数据库连接失败: %v", err)})
		return
	}

	// 查询 alert_record 表最新1条数据
	row := db.QueryRow("SELECT id, alert_idc, alert_spine, alert_leaf, alert_rack, alert_type, alert_status, create_by, create_time, update_by, update_time FROM alert_record ORDER BY create_time DESC LIMIT 1")

	var alert struct {
		ID          int64     `json:"id"`
		AlertIDC    string    `json:"alert_idc"`
		AlertSpine  *string   `json:"alert_spine"`
		AlertLeaf   *string   `json:"alert_leaf"`
		AlertRack   *string   `json:"alert_rack"`
		AlertType   *string   `json:"alert_type"`
		AlertStatus *string   `json:"alert_status"`
		CreateBy    *string   `json:"create_by"`
		CreateTime  time.Time `json:"create_time"`
		UpdateBy    *string   `json:"update_by"`
		UpdateTime  time.Time `json:"update_time"`
	}
	err = row.Scan(&alert.ID, &alert.AlertIDC, &alert.AlertSpine, &alert.AlertLeaf, &alert.AlertRack,
		&alert.AlertType, &alert.AlertStatus, &alert.CreateBy, &alert.CreateTime,
		&alert.UpdateBy, &alert.UpdateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteJSON(w, map[string]interface{}{
				"status":  "ok",
				"message": "数据库连接测试成功，alert_record 表中暂无数据",
				"alert":   nil,
			})
			return
		}
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("查询 alert_record 表失败: %v", err)})
		return
	}

	common.WriteJSON(w, map[string]interface{}{
		"status":  "ok",
		"message": "数据库连接测试成功，已获取 alert_record 最新1条数据",
		"alert":   alert,
	})
}

// EnvConfigGetCAS GET /api/env-config/cas
func EnvConfigGetCAS(w http.ResponseWriter, r *http.Request) {
	cfg := conf.Get()
	common.WriteJSON(w, cfg.CAS)
}

// EnvConfigSaveCAS POST /api/env-config/cas
func EnvConfigSaveCAS(w http.ResponseWriter, r *http.Request) {
	var casCfg conf.CASConfig
	if err := json.NewDecoder(r.Body).Decode(&casCfg); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}

	if err := conf.UpdateCAS(casCfg); err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]string{"status": "ok"})
}

// EnvConfigTestCAS POST /api/env-config/cas/test
func EnvConfigTestCAS(w http.ResponseWriter, r *http.Request) {
	var casCfg conf.CASConfig
	if err := json.NewDecoder(r.Body).Decode(&casCfg); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}

	type urlResult struct {
		Name   string `json:"name"`
		URL    string `json:"url"`
		Status string `json:"status"` // "ok" or "fail"
		Error  string `json:"error,omitempty"`
	}

	results := []urlResult{}
	client := &http.Client{Timeout: 10 * time.Second}

	// 定义需要测试的链接列表
	type casURL struct {
		name string
		url  string
	}
	urls := []casURL{
		{"CAS服务器地址", casCfg.ServerURL},
		{"登录页URL", casCfg.LoginURL},
		{"验证票据URL", casCfg.ValidateURL},
		{"登出URL", casCfg.LogoutURL},
		{"Service回调URL", casCfg.ServiceURL},
		{"前端回调地址", casCfg.FrontendCallback},
	}

	allOk := true
	for _, u := range urls {
		r := urlResult{Name: u.name, URL: u.url}
		if u.url == "" {
			r.Status = "fail"
			r.Error = "未配置"
			results = append(results, r)
			allOk = false
			continue
		}

		resp, err := client.Head(u.url)
		if err != nil {
			// HEAD 失败尝试 GET
			resp, err = client.Get(u.url)
			if err != nil {
				r.Status = "fail"
				r.Error = fmt.Sprintf("连接失败: %v", err)
				results = append(results, r)
				allOk = false
				continue
			}
			resp.Body.Close()
		}

		if resp.StatusCode >= 500 {
			r.Status = "fail"
			r.Error = fmt.Sprintf("服务器错误 (HTTP %d)", resp.StatusCode)
			allOk = false
		} else {
			r.Status = "ok"
			r.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		results = append(results, r)
	}

	status := "ok"
	message := "所有 CAS 链接测试通过"
	if !allOk {
		status = "partial"
		message = "部分 CAS 链接测试失败，请检查配置"
	}

	common.WriteJSON(w, map[string]interface{}{
		"status":  status,
		"message": message,
		"results": results,
	})
}

// EnvConfigGetPrometheus GET /api/env-config/prometheus
func EnvConfigGetPrometheus(w http.ResponseWriter, r *http.Request) {
	cfg := conf.Get()
	common.WriteJSON(w, cfg.Prometheus)
}

// EnvConfigSavePrometheus POST /api/env-config/prometheus
func EnvConfigSavePrometheus(w http.ResponseWriter, r *http.Request) {
	var promCfg conf.PrometheusConfig
	if err := json.NewDecoder(r.Body).Decode(&promCfg); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}

	if err := conf.UpdatePrometheus(promCfg); err != nil {
		common.WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}
	common.WriteJSON(w, map[string]string{"status": "ok"})
}

// EnvConfigTestPrometheus POST /api/env-config/prometheus/test
func EnvConfigTestPrometheus(w http.ResponseWriter, r *http.Request) {
	var promCfg conf.PrometheusConfig
	if err := json.NewDecoder(r.Body).Decode(&promCfg); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}

	if promCfg.BaseURL == "" {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": "Prometheus BaseURL 不能为空"})
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	queryURL := strings.TrimRight(promCfg.BaseURL, "/") + "/api/v1/query?query=ping_idc_latency_avg"
	resp, err := client.Get(queryURL)
	if err != nil {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("Prometheus 连接失败: %v", err)})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 16384))
	if resp.StatusCode >= 400 {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("Prometheus 返回异常状态 %d: %s", resp.StatusCode, string(body))})
		return
	}

	// 解析响应，提取数据条数
	var promResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []any  `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &promResp); err != nil {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("Prometheus 响应解析失败: %v, 原始响应: %s", err, string(body))})
		return
	}

	if promResp.Status != "success" {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("Prometheus 查询失败, status: %s, 响应: %s", promResp.Status, string(body))})
		return
	}

	common.WriteJSON(w, map[string]string{
		"status":  "ok",
		"message": fmt.Sprintf("Prometheus 连接测试成功，ping_idc_latency_avg 指标返回 %d 条数据", len(promResp.Data.Result)),
	})
}

// EnvConfigTestMetric POST /api/env-config/prometheus/metric-test
// 测试单个指标，返回数据条数
func EnvConfigTestMetric(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Metric string `json:"metric"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Metric == "" {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": "metric 不能为空"})
		return
	}

	cfg := conf.Get()
	if cfg.Prometheus.BaseURL == "" {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": "Prometheus 未配置"})
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s",
		strings.TrimRight(cfg.Prometheus.BaseURL, "/"),
		url.QueryEscape(body.Metric))

	resp, err := client.Get(queryURL)
	if err != nil {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("Prometheus 查询失败: %v", err)})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))

	var promResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []any  `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &promResp); err != nil {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("Prometheus 响应解析失败: %v", err)})
		return
	}

	if promResp.Status != "success" {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("Prometheus 查询失败, status: %s", promResp.Status)})
		return
	}

	common.WriteJSON(w, map[string]interface{}{
		"status": "ok",
		"metric": body.Metric,
		"count":  len(promResp.Data.Result),
	})
}

// EnvConfigQueryMetric POST /api/env-config/prometheus/query
// 查询指标最新数据，返回完整结果
func EnvConfigQueryMetric(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Metric string `json:"metric"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.WriteJSON(w, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Metric == "" {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": "metric 不能为空"})
		return
	}

	cfg := conf.Get()
	if cfg.Prometheus.BaseURL == "" {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": "Prometheus 未配置"})
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s",
		strings.TrimRight(cfg.Prometheus.BaseURL, "/"),
		url.QueryEscape(body.Metric))

	resp, err := client.Get(queryURL)
	if err != nil {
		common.WriteJSON(w, map[string]string{"status": "fail", "error": fmt.Sprintf("Prometheus 查询失败: %v", err)})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 262144))

	// 返回原始 Prometheus 响应
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}
