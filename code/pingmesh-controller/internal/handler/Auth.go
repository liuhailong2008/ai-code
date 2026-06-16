package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"pingmesh-controller/conf"
	"pingmesh-controller/internal/common"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ── CAS 相关数据结构 ──

type casServiceResponse struct {
	XMLName               xml.Name          `xml:"serviceResponse"`
	AuthenticationSuccess *casAuthSuccess   `xml:"authenticationSuccess"`
	AuthenticationFailure *casAuthFailure   `xml:"authenticationFailure"`
}

type casAuthSuccess struct {
	User string `xml:"user"`
}

type casAuthFailure struct {
	Code    string `xml:"code,attr"`
	Message string `xml:",chardata"`
}

// CASUserInfo CAS 认证成功后的用户信息
type CASUserInfo struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Department  string `json:"department"`
}

// ── Token 管理 ──

var tokenStore = struct {
	sync.RWMutex
	tokens map[string]*CASUserInfo
}{tokens: make(map[string]*CASUserInfo)}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ── 原有本地登录 ──

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteJSON(w, map[string]interface{}{
			"success": false,
			"message": "method not allowed",
		})
		return
	}

	var req loginRequest
	if err := common.ReadJSON(r, &req); err != nil {
		common.WriteJSON(w, map[string]interface{}{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	if req.Username == "" || req.Password == "" {
		common.WriteJSON(w, map[string]interface{}{
			"success": false,
			"message": "用户名和密码不能为空",
		})
		return
	}

	cfg := conf.Get()
	for _, u := range cfg.Users {
		if u.Username == req.Username && u.Password == req.Password {
			common.WriteJSON(w, map[string]interface{}{
				"success":      true,
				"message":      "登录成功",
				"display_name": u.DisplayName,
			})
			return
		}
	}

	common.WriteJSON(w, map[string]interface{}{
		"success": false,
		"message": "用户名或密码错误",
	})
}

// ── CAS 相关 Handler ──

// CasGetLoginConfig 前端获取认证配置
func CasGetLoginConfig(w http.ResponseWriter, r *http.Request) {
	cfg := conf.Get()
	common.WriteJSON(w, map[string]interface{}{
		"cas_enabled":   cfg.CAS.Enabled,
		"cas_login_url": cfg.CAS.LoginURL,
	})
}

// CasCallback CAS 票据回调：验证 ticket 并返回 token
func CasCallback(w http.ResponseWriter, r *http.Request) {
	cfg := conf.Get()

	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		common.WriteJSON(w, map[string]interface{}{
			"success": false,
			"message": "缺少 CAS 票据参数",
		})
		return
	}

	userInfo, err := validateCasTicket(cfg, ticket)
	if err != nil {
		common.WriteJSON(w, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("CAS 票据验证失败: %v", err),
		})
		return
	}

	token := generateToken()
	tokenStore.Lock()
	tokenStore.tokens[token] = userInfo
	tokenStore.Unlock()

	// 302 重定向到前端 CAS 回调页，携带 token
	redirectURL := fmt.Sprintf("%s/#/cas/callback?token=%s",
		cfg.CAS.FrontendCallback, token)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// CasGetUserInfo 通过 token 获取当前用户信息
func CasGetUserInfo(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		common.WriteJSON(w, map[string]interface{}{
			"success": false,
			"message": "未提供认证令牌",
		})
		return
	}

	tokenStore.RLock()
	userInfo, ok := tokenStore.tokens[token]
	tokenStore.RUnlock()

	if !ok {
		common.WriteJSON(w, map[string]interface{}{
			"success": false,
			"message": "令牌无效或已过期",
		})
		return
	}

	common.WriteJSON(w, map[string]interface{}{
		"success": true,
		"user":    userInfo,
	})
}

// CasLogout 登出
func CasLogout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token != "" {
		tokenStore.Lock()
		delete(tokenStore.tokens, token)
		tokenStore.Unlock()
	}

	cfg := conf.Get()
	if cfg.CAS.Enabled && cfg.CAS.LogoutURL != "" {
		logoutURL := fmt.Sprintf("%s?service=%s",
			cfg.CAS.LogoutURL,
			url.QueryEscape(cfg.CAS.FrontendCallback+"/#/login"))
		http.Redirect(w, r, logoutURL, http.StatusFound)
		return
	}

	http.Redirect(w, r, cfg.CAS.FrontendCallback+"/#/login", http.StatusFound)
}

// ── 辅助函数 ──

// validateCasTicket 向 CAS 服务器验证 ticket
func validateCasTicket(cfg *conf.Config, ticket string) (*CASUserInfo, error) {
	validateURL := fmt.Sprintf("%s?service=%s&ticket=%s",
		cfg.CAS.ValidateURL,
		url.QueryEscape(cfg.CAS.ServiceURL),
		url.QueryEscape(ticket),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(validateURL)
	if err != nil {
		return nil, fmt.Errorf("请求 CAS 验证服务失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 CAS 响应失败: %w", err)
	}

	var casResp casServiceResponse
	if err := xml.Unmarshal(body, &casResp); err != nil {
		return nil, fmt.Errorf("解析 CAS 响应失败: %w", err)
	}

	if casResp.AuthenticationFailure != nil {
		return nil, fmt.Errorf("[%s] %s",
			casResp.AuthenticationFailure.Code,
			strings.TrimSpace(casResp.AuthenticationFailure.Message))
	}

	if casResp.AuthenticationSuccess == nil {
		return nil, fmt.Errorf("CAS 响应中缺少认证结果")
	}

	userInfo := &CASUserInfo{
		Username: casResp.AuthenticationSuccess.User,
	}

	// 解析 CAS XML 中的属性字段
	// 需要外层 serviceResponse wrapper，因为 XML 根元素是 <cas:serviceResponse>
	type rawCasAttrs struct {
		User        string `xml:"user"`
		DisplayName string `xml:"displayName"`
		Mail        string `xml:"mail"`
		Phone       string `xml:"phone"`
		Department  string `xml:"department"`
	}
	type rawCasResp struct {
		AuthSuccess rawCasAttrs `xml:"authenticationSuccess"`
	}
	var rawResp rawCasResp
	if err := xml.Unmarshal(body, &rawResp); err == nil {
		userInfo.DisplayName = rawResp.AuthSuccess.DisplayName
		userInfo.Email = rawResp.AuthSuccess.Mail
		userInfo.Phone = rawResp.AuthSuccess.Phone
		userInfo.Department = rawResp.AuthSuccess.Department
	}

	return userInfo, nil
}

// ExtractToken 从请求中提取 token（供中间件使用）
func ExtractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:]
	}
	return r.URL.Query().Get("token")
}

// extractToken 内部使用的 token 提取
func extractToken(r *http.Request) string {
	return ExtractToken(r)
}

// ValidateToken 验证 token 是否有效（供中间件使用）
func ValidateToken(token string) bool {
	tokenStore.RLock()
	_, ok := tokenStore.tokens[token]
	tokenStore.RUnlock()
	return ok
}
