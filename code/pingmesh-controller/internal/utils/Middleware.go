package utils

import (
	"log"
	"net/http"
	"strings"
	"time"

	"pingmesh-controller/conf"
	"pingmesh-controller/internal/common"
	"pingmesh-controller/internal/handler"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// AuthRequired 认证中间件：CAS 启用时验证 token
func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := conf.Get()

		if !cfg.CAS.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// 白名单路径：登录相关接口和静态资源不需要认证
		if isAuthWhitelist(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		token := handler.ExtractToken(r)
		if token == "" || !handler.ValidateToken(token) {
			common.WriteJSON(w, map[string]interface{}{
				"code":    401,
				"success": false,
				"message": "未登录，请先登录",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isAuthWhitelist 认证白名单路径
func isAuthWhitelist(path string) bool {
	whitelist := []string{
		"/api/auth",
		"/api/health",
		"/api/public",
	}
	for _, prefix := range whitelist {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
