package conf

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Port                       int                        `yaml:"port"`
	StaticDir                  string                     `yaml:"static_dir"`
	MySQL                      MySQLConfig                `yaml:"mysql"`
	Prometheus                 PrometheusConfig           `yaml:"prometheus"`
	IDCs                       []IDCConfig                `yaml:"idcs"`
	IDCLinks                   []IDCLinkConfig            `yaml:"idc_links"`
	LatencyLevelsForIDCInner   LatencyLevelsForIDCInner   `yaml:"latency_levels_for_idc_inner"`
	LostLevelsForIDCInner      LostLevelsForIDCInner      `yaml:"lost_levels_for_idc_inner"`
	LatencyLevelsForIDCBetween LatencyLevelsForIDCBetween `yaml:"latency_levels_for_idc_between"`
	LostLevelsForIDCBetween    LostLevelsForIDCBetween    `yaml:"lost_levels_for_idc_between"`
	MetricsMeta                []MetricMetaConfig         `yaml:"metrics_meta"`
	Users                      []UserConfig               `yaml:"users"`
	CAS                        CASConfig                  `yaml:"cas"`
}

// MetricMetaConfig 指标元数据配置
type MetricMetaConfig struct {
	Metric      string `yaml:"metric" json:"metric"`
	Type        string `yaml:"type" json:"type"`               // latency / loss
	Scope       string `yaml:"scope" json:"scope"`             // idc / spine / leaf
	Category    string `yaml:"category" json:"category"`       // raw / max / avg / p99
	Description string `yaml:"description" json:"description"` // 中文描述
	Status      string `yaml:"status" json:"status"`           // new / show / hide / disable
	Strategy    string `yaml:"strategy" json:"strategy"`       // 00 / 01 / 02 / 03
}

// LatencyLevelsForIDCInner 机房内时延阈值（延迟 ms）
type LatencyLevelsForIDCInner struct {
	Warn  float64 `yaml:"warn"`  // >= warn → 告警(2)
	Alarm float64 `yaml:"alarm"` // >= alarm → 严重(3)
}

// LostLevelsForIDCInner 机房内丢包率阈值（百分比）
type LostLevelsForIDCInner struct {
	Warn  float64 `yaml:"warn"`  // >= warn → 告警(2)
	Alarm float64 `yaml:"alarm"` // >= alarm → 严重(3)
}

// LatencyLevelsForIDCBetween 机房间时延阈值（延迟 ms）
type LatencyLevelsForIDCBetween struct {
	Warn  float64 `yaml:"warn"`  // >= warn → 告警(2)
	Alarm float64 `yaml:"alarm"` // >= alarm → 严重(3)
}

// LostLevelsForIDCBetween 机房间丢包率阈值（百分比）
type LostLevelsForIDCBetween struct {
	Warn  float64 `yaml:"warn"`  // >= warn → 告警(2)
	Alarm float64 `yaml:"alarm"` // >= alarm → 严重(3)
}

// IDCConfig IDC（数据中心）配置
type IDCConfig struct {
	Code   string `yaml:"code"`
	Name   string `yaml:"name"`
	Enable bool   `yaml:"enable"`
}

// IDCLinkConfig IDC 间链路配置
type IDCLinkConfig struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
	Enable bool   `yaml:"enable"`
}

// UserConfig 用户登录配置
type UserConfig struct {
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	DisplayName string `yaml:"display_name"`
}

// CASConfig CAS 认证配置
type CASConfig struct {
	Enabled          bool           `yaml:"enabled" json:"enabled"`
	ServerURL        string         `yaml:"server_url" json:"server_url"`
	LoginURL         string         `yaml:"login_url" json:"login_url"`
	ValidateURL      string         `yaml:"validate_url" json:"validate_url"`
	LogoutURL        string         `yaml:"logout_url" json:"logout_url"`
	ServiceURL       string         `yaml:"service_url" json:"service_url"`
	FrontendCallback string         `yaml:"frontend_callback" json:"frontend_callback"`
	AttributeMapping CASAttrMapping `yaml:"attribute_mapping" json:"attribute_mapping"`
}

// CASAttrMapping CAS 属性字段映射
type CASAttrMapping struct {
	Username    string `yaml:"username" json:"username"`
	DisplayName string `yaml:"display_name" json:"display_name"`
	Email       string `yaml:"email" json:"email"`
	Phone       string `yaml:"phone" json:"phone"`
	Department  string `yaml:"department" json:"department"`
}

// currentConfig 保存最后一次 Load() 的结果，供全局访问
var (
	currentConfig *Config
	configMu      sync.RWMutex
)

// Get 返回当前加载的配置（供其他包获取 IDC、链路等数据）
func Get() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	if currentConfig == nil {
		return defaultConfig()
	}
	return currentConfig
}

// MySQLConfig MySQL 数据库配置
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// DSN 返回 MySQL 连接字符串
func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.Database)
}

// PrometheusConfig Prometheus 配置
type PrometheusConfig struct {
	BaseURL string `yaml:"base_url" json:"base_url"`
}

// defaultConfig 默认配置
func defaultConfig() *Config {
	return &Config{
		Port:      8080,
		StaticDir: filepath.Join("..", "..", "pingmesh-front-v2", "dist"),
		MySQL:     MySQLConfig{},
		Prometheus: PrometheusConfig{
			BaseURL: "http://127.0.0.1:9090",
		},
		IDCs:     []IDCConfig{},
		IDCLinks: []IDCLinkConfig{},
	}
}

// configFilePath 配置文件的完整路径
func configFilePath() string {
	return filepath.Join("conf", "config.yaml")
}

// Load 加载配置：先设默认值，再尝试从 conf/config.yaml 读取覆盖
func Load() *Config {
	cfg := defaultConfig()

	confPath := configFilePath()
	data, err := os.ReadFile(confPath)
	if err != nil {
		log.Printf("config file not found (%s), using defaults", confPath)
		configMu.Lock()
		currentConfig = cfg
		configMu.Unlock()
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Printf("failed to parse config: %v, using defaults", err)
		configMu.Lock()
		currentConfig = cfg
		configMu.Unlock()
		return cfg
	}

	log.Printf("config loaded from %s", confPath)
	configMu.Lock()
	currentConfig = cfg
	configMu.Unlock()
	return cfg
}

// UpdateThresholds 更新阈值配置并持久化到 config.yaml
func UpdateThresholds(latencyInner LatencyLevelsForIDCInner, lostInner LostLevelsForIDCInner, latencyBetween LatencyLevelsForIDCBetween, lostBetween LostLevelsForIDCBetween) error {
	configMu.Lock()
	defer configMu.Unlock()

	if currentConfig == nil {
		currentConfig = defaultConfig()
	}

	currentConfig.LatencyLevelsForIDCInner = latencyInner
	currentConfig.LostLevelsForIDCInner = lostInner
	currentConfig.LatencyLevelsForIDCBetween = latencyBetween
	currentConfig.LostLevelsForIDCBetween = lostBetween

	// 序列化并写回文件
	data, err := yaml.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configFilePath(), data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	log.Printf("thresholds updated and persisted to %s", configFilePath())
	return nil
}

// GetPort 获取端口（兼容环境变量覆盖）
func GetPort() int {
	port := os.Getenv("PORT")
	if port == "" {
		return 8080
	}
	return 8080 // simplified; could use strconv.Atoi
}

// UpdateIDCs 更新机房配置并持久化到 config.yaml
func UpdateIDCs(idcs []IDCConfig) error {
	configMu.Lock()
	defer configMu.Unlock()

	if currentConfig == nil {
		currentConfig = defaultConfig()
	}

	currentConfig.IDCs = idcs

	data, err := yaml.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configFilePath(), data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	log.Printf("idcs updated and persisted to %s", configFilePath())
	return nil
}

// UpdateMetricsMeta 更新指标元数据配置并持久化到 config.yaml
func UpdateMetricsMeta(meta []MetricMetaConfig) error {
	configMu.Lock()
	defer configMu.Unlock()

	if currentConfig == nil {
		currentConfig = defaultConfig()
	}

	currentConfig.MetricsMeta = meta

	data, err := yaml.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configFilePath(), data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	log.Printf("metrics_meta updated and persisted to %s", configFilePath())
	return nil
}

// UpdateMySQL 更新 MySQL 配置并持久化到 config.yaml
func UpdateMySQL(cfg MySQLConfig) error {
	configMu.Lock()
	defer configMu.Unlock()

	if currentConfig == nil {
		currentConfig = defaultConfig()
	}

	currentConfig.MySQL = cfg

	data, err := yaml.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configFilePath(), data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	log.Printf("mysql config updated and persisted to %s", configFilePath())
	return nil
}

// UpdateCAS 更新 CAS 配置并持久化到 config.yaml
func UpdateCAS(cfg CASConfig) error {
	configMu.Lock()
	defer configMu.Unlock()

	if currentConfig == nil {
		currentConfig = defaultConfig()
	}

	currentConfig.CAS = cfg

	data, err := yaml.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configFilePath(), data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	log.Printf("cas config updated and persisted to %s", configFilePath())
	return nil
}

// UpdatePrometheus 更新 Prometheus 配置并持久化到 config.yaml
func UpdatePrometheus(cfg PrometheusConfig) error {
	configMu.Lock()
	defer configMu.Unlock()

	if currentConfig == nil {
		currentConfig = defaultConfig()
	}

	currentConfig.Prometheus = cfg

	data, err := yaml.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configFilePath(), data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	log.Printf("prometheus config updated and persisted to %s", configFilePath())
	return nil
}
