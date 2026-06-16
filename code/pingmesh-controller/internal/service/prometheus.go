package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"pingmesh-controller/conf"
)

// PromClient Prometheus HTTP API 客户端（实时监控数据源）
type PromClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPromClient 创建 Prometheus 客户端
func NewPromClient(cfg conf.PrometheusConfig) *PromClient {
	log.Printf("Prometheus client created: %s", cfg.BaseURL)
	return &PromClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ── Prometheus API 响应结构 ──

type promResponse struct {
	Status string   `json:"status"`
	Data   promData `json:"data"`
}

type promData struct {
	ResultType string       `json:"resultType"`
	Result     []promResult `json:"result"`
}

type promResult struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`  // instant query
	Values [][]interface{}   `json:"values"` // range query
}

// ── 即时查询 ──

// Query 执行 Prometheus 即时查询，返回第一个结果的值
func (p *PromClient) Query(query string) (float64, error) {
	resp, err := p.queryAPI("/api/v1/query", query)
	if err != nil {
		return 0, err
	}
	if len(resp.Data.Result) == 0 {
		return 0, nil
	}
	val, ok := resp.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("unexpected value type")
	}
	var v float64
	fmt.Sscanf(val, "%f", &v)
	return v, nil
}

// QueryWithLabels 执行 Prometheus 即时查询，返回带标签的结果列表
func (p *PromClient) QueryWithLabels(query string) ([]promResult, error) {
	resp, err := p.queryAPI("/api/v1/query", query)
	if err != nil {
		return nil, err
	}
	return resp.Data.Result, nil
}

// queryAPI 通用 Prometheus API 查询
func (p *PromClient) queryAPI(apiPath, query string) (*promResponse, error) {
	u, err := url.Parse(p.baseURL + apiPath)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned %d: %s", resp.StatusCode, string(body))
	}

	var pr promResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if pr.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s", string(body))
	}
	return &pr, nil
}

// ── 通用 API 访问方法 ──

// rawGet 对 Prometheus API 做 GET 请求，返回 response body
func (p *PromClient) rawGet(apiPath string, params url.Values) ([]byte, error) {
	u, err := url.Parse(p.baseURL + apiPath)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// ListMetrics 获取所有指标名称
func (p *PromClient) ListMetrics() ([]string, error) {
	body, err := p.rawGet("/api/v1/label/__name__/values", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("prometheus list metrics failed")
	}
	return resp.Data, nil
}

// GetMetricLabels 获取指标的标签列表
func (p *PromClient) GetMetricLabels(metric string) ([]string, error) {
	params := url.Values{}
	params.Set("match[]", metric)
	body, err := p.rawGet("/api/v1/labels", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("prometheus get labels failed")
	}
	// 过滤掉 __name__ 内置标签
	var labels []string
	for _, l := range resp.Data {
		if l != "__name__" {
			labels = append(labels, l)
		}
	}
	return labels, nil
}

// GetLabelValues 获取指标某个标签的所有值
func (p *PromClient) GetLabelValues(metric, label string) ([]string, error) {
	params := url.Values{}
	params.Set("match[]", metric)
	body, err := p.rawGet("/api/v1/label/"+label+"/values", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("prometheus get label values failed")
	}
	return resp.Data, nil
}

// ── 常用监控指标查询方法 ──

// metric helpers: 根据指标模式生成标准查询语句

// QueryLeafLatencyMatrix 查询 Leaf 间延迟矩阵（使用 recording rules）
// recording rules: ping_leaf_avg / ping_leaf_max / ping_leaf_p99
// 标签: tidc, sleaf, tleaf
func (p *PromClient) QueryLeafLatencyMatrix(idc, metric string) ([]promResult, error) {
	// metric 映射为 recording rule 名称
	ruleName := "ping_leaf_" + metric // avg / max / p99
	query := fmt.Sprintf(`%s{tidc="%s"}`, ruleName, idc)
	return p.QueryWithLabels(query)
}

// QueryLeafLatencyRange 查询 Leaf 间延迟时间范围数据
// 使用 Prometheus range query 获取指定时间范围内的 recording rule 历史值
func (p *PromClient) QueryLeafLatencyRange(idc, metric, startTime, endTime, step string) ([]promResult, error) {
	ruleName := "ping_leaf_" + metric
	query := fmt.Sprintf(`%s{tidc="%s"}`, ruleName, idc)
	return p.queryRangeAPI(query, startTime, endTime, step)
}

// queryRangeAPI 执行 Prometheus range query
func (p *PromClient) queryRangeAPI(query, startTime, endTime, step string) ([]promResult, error) {
	u, err := url.Parse(p.baseURL + "/api/v1/query_range")
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	q := u.Query()
	q.Set("query", query)
	q.Set("start", startTime)
	q.Set("end", endTime)
	q.Set("step", step)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned %d: %s", resp.StatusCode, string(body))
	}

	var pr promResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if pr.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s", string(body))
	}
	return pr.Data.Result, nil
}

// QueryIDCLinkStatus 查询机房间链路状态（recording rule: ping_idc_link_status）
// 标签: sidc (源机房), tidc (目标机房)
// 值: 1=正常, 2=告警, 3=严重
func (p *PromClient) QueryIDCLinkStatusResults() ([]promResult, error) {
	query := `ping_idc_link_status`
	return p.QueryWithLabels(query)
}

// QueryPingLeafMaxAll 查询同机房 Leaf 间最大延迟（recording rule: ping_leaf_max_all）
func (p *PromClient) QueryPingLeafMaxAll() ([]promResult, error) {
	query := `ping_leaf_max_all`
	return p.QueryWithLabels(query)
}

// QueryLatencyByLeaf 查询机房内 Leaf 时延（recording rule: ping_latency_by_leaf_avg/max/p99）
// 注意: p99 使用 sidc 标签, avg/max 使用 tidc 标签
func (p *PromClient) QueryLatencyByLeaf(idc, metric string) ([]promResult, error) {
	labelKey := "tidc"
	if metric == "p99" {
		labelKey = "sidc"
	}
	query := fmt.Sprintf(`ping_latency_by_leaf_%s{%s="%s"}`, metric, labelKey, idc)
	return p.QueryWithLabels(query)
}

// QueryLostByLeaf 查询机房内 Leaf 丢包率（recording rule: ping_lost_by_leaf）
func (p *PromClient) QueryLostByLeaf(idc, metric string) ([]promResult, error) {
	ruleName := "ping_lost_by_leaf"
	query := fmt.Sprintf(`%s{tidc="%s"}`, ruleName, idc)
	return p.QueryWithLabels(query)
}
