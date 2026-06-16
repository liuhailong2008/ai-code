package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pingmesh-controller/conf"
	"pingmesh-controller/internal/common"
	"pingmesh-controller/internal/handler"
	"pingmesh-controller/internal/service"
	"pingmesh-controller/internal/utils"
)

func main() {
	cfg := conf.Load()

	// ── 初始化数据源 ──
	var db *service.DB
	var promClient *service.PromClient

	// 连接 MySQL（配置信息数据源）
	if cfg.MySQL.Host != "" {
		var err error
		db, err = service.NewDB(cfg.MySQL)
		if err != nil {
			log.Printf("MySQL connection failed: %v, continuing without database", err)
		}
	}
	if db == nil {
		log.Println("Warning: running without MySQL, config data will fallback to defaults")
	}

	// 创建 Prometheus 客户端（实时监控数据源）
	promClient = service.NewPromClient(cfg.Prometheus)

	// ── 初始化 Service 层 ──
	dashboardSvc := service.NewDashboardService(db, promClient)
	leafSvc := service.NewLeafService(db, promClient)
	latencySvc := service.NewLatencyService(db, promClient)
	alertSvc := service.NewAlertService(db, promClient)
	statisticsSvc := service.NewStatisticsService(promClient)

	// ── 初始化 Handler 层 ──
	h := handler.NewHandler(dashboardSvc, leafSvc, latencySvc, alertSvc, statisticsSvc)

	// ── 路由注册 ──
	mux := http.NewServeMux()

	// Auth（无需认证）
	mux.HandleFunc("/api/auth/login", handler.Login)
	mux.HandleFunc("/api/auth/config", handler.CasGetLoginConfig)
	mux.HandleFunc("/api/auth/cas/callback", handler.CasCallback)
	mux.HandleFunc("/api/auth/cas/user", handler.CasGetUserInfo)
	mux.HandleFunc("/api/auth/logout", handler.CasLogout)
	mux.HandleFunc("/api/health", handler.Health)

	// ── 受保护的 API 路由（CAS 启用时需要认证）──
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/dashboard-monitor/graph-idc-status", h.Dashboard.SummaryIDCStatus)
	apiMux.HandleFunc("/api/dashboard-monitor/idc-link-status", h.Dashboard.IDCLinkStatus)
	apiMux.HandleFunc("/api/dashboard-monitor/idc-status", h.Dashboard.IDCStatus)
	apiMux.HandleFunc("/api/dashboard-monitor/idc-nodes", h.Dashboard.IDCNodes)
	apiMux.HandleFunc("/api/dashboard-monitor/summary-status", h.Dashboard.SummaryStatus)
	apiMux.HandleFunc("/api/dashboard-monitor/overall-stats", h.Dashboard.OverallStats)
	apiMux.HandleFunc("/api/dashboard-monitor/alert-page", h.Dashboard.AlertPage)
	apiMux.HandleFunc("/api/dashboard-monitor/unresolved-alert-page", h.Dashboard.UnresolvedAlertPage)
	apiMux.HandleFunc("/api/dashboard-monitor/alert-type-stats", h.Dashboard.AlertTypeStats)
	apiMux.HandleFunc("/api/dashboard-monitor/alert-status-stats", h.Dashboard.AlertStatusStats)
	apiMux.HandleFunc("/api/dashboard-monitor/unresolved-alerts", h.Dashboard.UnresolvedAlerts)
	apiMux.HandleFunc("/api/dashboard-monitor/thresholds", h.Dashboard.Thresholds)

	// ── 公共 API（白名单，无需认证）──
	apiMux.HandleFunc("/api/public/graph-idc-status", h.Dashboard.SummaryIDCStatus)
	apiMux.HandleFunc("/api/public/idc-link-status", h.Dashboard.IDCLinkStatus)
	apiMux.HandleFunc("/api/public/summary-status", h.Dashboard.SummaryStatus)
	apiMux.HandleFunc("/api/leaf-connectivity/leafs", h.Leaf.Leafs)
	apiMux.HandleFunc("/api/leaf-connectivity/heatmap", h.Leaf.Heatmap)
	apiMux.HandleFunc("/api/leaf-connectivity/heatmap-range", h.Leaf.HeatmapRange)
	apiMux.HandleFunc("/api/leaf-connectivity/top-links", h.Leaf.TopLinks)
	apiMux.HandleFunc("/api/leaf-connectivity/max-all-links", h.Leaf.MaxAllLinks)
	apiMux.HandleFunc("/api/leaf-connectivity/time-options", h.Leaf.TimeOptions)
	apiMux.HandleFunc("/api/latency/scatter", h.Latency.Scatter)
	apiMux.HandleFunc("/api/alerts", h.Alert.AlertsList)
	apiMux.HandleFunc("/api/alerts/", h.Alert.AlertsList)
	apiMux.HandleFunc("/api/alerts/stats", h.Alert.AlertsStats)
	apiMux.HandleFunc("/api/alerts/update-status", h.Alert.UpdateStatus)
	apiMux.HandleFunc("/api/alerts/unresolved-count", h.Alert.UnresolvedCount)
	apiMux.HandleFunc("/api/idcs", h.IDCs)
	apiMux.HandleFunc("/api/statistics/metrics", h.Statistics.ListMetrics)
	apiMux.HandleFunc("/api/statistics/metric-data", h.Statistics.QueryMetricData)
	apiMux.HandleFunc("/api/statistics/metric-labels", h.Statistics.GetMetricLabels)
	apiMux.HandleFunc("/api/statistics/metric-meta", h.Statistics.MetricsMetaConfig)

	// 环境管理配置 API
	apiMux.HandleFunc("/api/env-config/mysql", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.EnvConfigGetMySQL(w, r)
		case http.MethodPost:
			handler.EnvConfigSaveMySQL(w, r)
		default:
			common.WriteJSON(w, map[string]string{"error": "method not allowed"})
		}
	})
	apiMux.HandleFunc("/api/env-config/mysql/test", handler.EnvConfigTestMySQL)
	apiMux.HandleFunc("/api/env-config/cas", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.EnvConfigGetCAS(w, r)
		case http.MethodPost:
			handler.EnvConfigSaveCAS(w, r)
		default:
			common.WriteJSON(w, map[string]string{"error": "method not allowed"})
		}
	})
	apiMux.HandleFunc("/api/env-config/cas/test", handler.EnvConfigTestCAS)
	apiMux.HandleFunc("/api/env-config/prometheus", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.EnvConfigGetPrometheus(w, r)
		case http.MethodPost:
			handler.EnvConfigSavePrometheus(w, r)
		default:
			common.WriteJSON(w, map[string]string{"error": "method not allowed"})
		}
	})
	apiMux.HandleFunc("/api/env-config/prometheus/test", handler.EnvConfigTestPrometheus)
	apiMux.HandleFunc("/api/env-config/prometheus/metric-test", handler.EnvConfigTestMetric)
	apiMux.HandleFunc("/api/env-config/prometheus/query", handler.EnvConfigQueryMetric)

	mux.Handle("/api/dashboard-monitor/", utils.AuthRequired(apiMux))
	mux.Handle("/api/leaf-connectivity/", utils.AuthRequired(apiMux))
	mux.Handle("/api/latency/", utils.AuthRequired(apiMux))
	mux.Handle("/api/alerts/", utils.AuthRequired(apiMux))
	mux.Handle("/api/alerts", utils.AuthRequired(apiMux))
	mux.Handle("/api/idcs", utils.AuthRequired(apiMux))
	mux.Handle("/api/statistics/", utils.AuthRequired(apiMux))
	mux.Handle("/api/env-config/", utils.AuthRequired(apiMux))
	mux.Handle("/api/public/", utils.AuthRequired(apiMux))

	// ── Static Files (frontend dist) ──
	distPath := cfg.StaticDir
	if _, err := os.Stat(distPath); err == nil {
		fs := http.FileServer(http.Dir(distPath))
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := filepath.Join(distPath, r.URL.Path)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
				return
			}
			fs.ServeHTTP(w, r)
		}))
		log.Printf("Serving static files from: %s", distPath)
	} else {
		log.Printf("Static files not found at %s, API-only mode", distPath)
	}

	// Apply middleware
	app := utils.Logging(utils.CORS(mux))

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      app,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf(strings.Repeat("=", 60) + "\n")
	fmt.Printf("  PingMesh Controller Server\n")
	fmt.Printf("  Listening on: http://localhost%s\n", addr)
	fmt.Printf(strings.Repeat("=", 60) + "\n")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
