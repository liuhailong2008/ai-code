package common

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// WriteJSON 统一 JSON 响应写入
func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}

// ReadJSON 从 HTTP 请求读取 JSON body 到目标结构体
func ReadJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// NewRand 创建新的随机数生成器
func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandFloat 生成 [min, max) 范围内保留两位小数的随机数
func RandFloat(r *rand.Rand, min, max float64) float64 {
	return math.Round((min+r.Float64()*(max-min))*100) / 100
}

// ── 告警状态映射 ──
// DB 存储英文原始值，接口返回中文展示值

// MapAlertStatus 将内部状态映射为中文展示值
func MapAlertStatus(internal string) string {
	switch internal {
	case "PROCESSING":
		return "处置中"
	case "PROCESSED":
		return "已处置"
	default: // "" "UNPROCESSED"
		return "告警中"
	}
}

// MapAlertStatusReverse 将中文展示值反映射为英文内部值（用于写入 DB）
func MapAlertStatusReverse(display string) string {
	switch display {
	case "告警中":
		return "UNPROCESSED"
	case "处置中":
		return "PROCESSING"
	case "已处置":
		return "PROCESSED"
	default:
		return "UNPROCESSED"
	}
}

// AllDisplayStatuses 返回所有展示状态列表（按顺序）
func AllDisplayStatuses() []string {
	return []string{"告警中", "处置中", "已处置"}
}
