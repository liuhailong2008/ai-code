package common

// IDC 统一机房信息
type IDC struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Enable bool   `json:"enable"`
}

// AllIDCs 统一机房列表（所有页面共用，仅启用状态机房）
var AllIDCs = []IDC{
	{Code: "BJ12", Name: "北京昌平", Enable: true},
	{Code: "BJ13", Name: "北京亦庄", Enable: true},
	{Code: "SH23", Name: "上海宝山", Enable: true},
	{Code: "SZ32", Name: "深圳宝安", Enable: true},
}

// LeafCountPerIDC 每个 IDC 的 Leaf 节点数量
const LeafCountPerIDC = 15
