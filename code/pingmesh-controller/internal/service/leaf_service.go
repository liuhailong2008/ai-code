package service

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

// LeafService 机房网络监控（Leaf Connectivity） Service
type LeafService struct {
	db   *DB
	prom *PromClient
}

// NewLeafService 创建 Leaf 网络监控 Service
func NewLeafService(db *DB, prom *PromClient) *LeafService {
	return &LeafService{db: db, prom: prom}
}

// ── 1. Leaf 节点列表 ──

// GetLeafs 获取指定机房的 Leaf 节点名称列表
// 通过查询 ping_leaf_avg recording rule，从 sleaf/tleaf 标签中提取去重后的 leaf 名称
func (s *LeafService) GetLeafs(idcCode string) ([]string, error) {
	leafSet := make(map[string]bool)

	// 从 Prometheus 查询 recording rule 获取 sleaf 和 tleaf 标签值
	results, err := s.prom.QueryLeafLatencyMatrix(idcCode, "avg")
	if err == nil {
		for _, r := range results {
			if sleaf := r.Metric["sleaf"]; sleaf != "" {
				leafSet[sleaf] = true
			}
			if tleaf := r.Metric["tleaf"]; tleaf != "" {
				leafSet[tleaf] = true
			}
		}
	}

	// 如果 Prometheus 返回了数据，使用这些名称
	if len(leafSet) > 0 {
		names := make([]string, 0, len(leafSet))
		for name := range leafSet {
			names = append(names, name)
		}
		sort.Strings(names)
		return names, nil
	}

	// 降级：尝试从数据库获取
	leafs, err := s.db.GetLeafsByIDC(idcCode)
	if err == nil && len(leafs) > 0 {
		names := make([]string, len(leafs))
		for i, l := range leafs {
			names[i] = l.Name
		}
		return names, nil
	}

	// 最终降级：生成默认名称
	names := make([]string, 15)
	for i := 0; i < 15; i++ {
		names[i] = fmt.Sprintf("Leaf-%02d", i+1)
	}
	return names, nil
}

// ── 2. 连通性热力图数据 ──

// GetHeatmapData 获取 Leaf 间连通性矩阵数据（优化：消除 GetLeafs 导致的重复查询）
// metric: avg / p99 / max (对应 recording rule ping_leaf_avg/ping_leaf_p99/ping_leaf_max)
func (s *LeafService) GetHeatmapData(idcCode, metric string) ([][]interface{}, error) {
	if metric == "" {
		metric = "avg"
	}

	// 一次查询同时获取 leaf 名称和延迟数据
	results, err := s.prom.QueryLeafLatencyMatrix(idcCode, metric)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("no latency data for idc=%s metric=%s", idcCode, metric)
	}

	// 从查询结果中提取 leaf 名称并排序
	leafSet := make(map[string]bool)
	for _, r := range results {
		if sleaf := r.Metric["sleaf"]; sleaf != "" {
			leafSet[sleaf] = true
		}
		if tleaf := r.Metric["tleaf"]; tleaf != "" {
			leafSet[tleaf] = true
		}
	}
	leafs := make([]string, 0, len(leafSet))
	for name := range leafSet {
		leafs = append(leafs, name)
	}
	sort.Strings(leafs)
	leafCount := len(leafs)

	// 构建矩阵：先初始化全 0
	data := make([][]interface{}, leafCount*leafCount)
	idx := 0
	for i := 0; i < leafCount; i++ {
		for j := 0; j < leafCount; j++ {
			data[idx] = []interface{}{j, i, float64(0)}
			idx++
		}
	}

	// 建立名称到索引的映射
	nameToIdx := make(map[string]int)
	for i, name := range leafs {
		nameToIdx[name] = i
	}

	// 填充查询结果（复用已有的 results，无需二次查询）
	for _, r := range results {
		srcLeaf := r.Metric["sleaf"]
		dstLeaf := r.Metric["tleaf"]
		srcIdx, ok1 := nameToIdx[srcLeaf]
		dstIdx, ok2 := nameToIdx[dstLeaf]
		if !ok1 || !ok2 {
			continue
		}
		val := parseFloat(r.Value[1])
		data[srcIdx*leafCount+dstIdx] = []interface{}{dstIdx, srcIdx, val}
	}

	return data, nil
}

// ── 2.5. 热力图时间范围数据（预加载） ──

// HeatmapRangePoint 单个时间点的热力图数据
type HeatmapRangePoint struct {
	Timestamp int64           `json:"timestamp"` // Unix 毫秒时间戳
	TimeLabel string          `json:"timeLabel"` // 显示标签 "HH:MM"
	Data      [][]interface{} `json:"data"`      // 热力图矩阵数据 [xIdx, yIdx, value]
}

// GetHeatmapRangeData 获取时间范围内的热力图数据（用于预加载）
// startTime/endTime: Unix 秒时间戳字符串
// step: 步长（如 "60s"）
func (s *LeafService) GetHeatmapRangeData(idcCode, metric, startTime, endTime, step string) ([]HeatmapRangePoint, []string, error) {
	if metric == "" {
		metric = "avg"
	}

	// 先获取 leaf 列表
	leafs, err := s.GetLeafs(idcCode)
	if err != nil {
		return nil, nil, err
	}
	leafCount := len(leafs)

	// 建立名称到索引的映射
	nameToIdx := make(map[string]int)
	for i, name := range leafs {
		nameToIdx[name] = i
	}

	// 查询 Prometheus range 数据
	results, err := s.prom.QueryLeafLatencyRange(idcCode, metric, startTime, endTime, step)
	if err != nil {
		return nil, nil, fmt.Errorf("range query failed: %w", err)
	}

	// 按时间戳组织数据: timestamp -> [(sleaf, tleaf, value), ...]
	timeSeriesMap := make(map[int64][]struct {
		sleaf string
		tleaf string
		value float64
	})

	for _, r := range results {
		sleaf := r.Metric["sleaf"]
		tleaf := r.Metric["tleaf"]
		if r.Values == nil {
			continue
		}
		for _, v := range r.Values {
			if len(v) < 2 {
				continue
			}
			ts := parseTimestamp(v[0])
			val := parseFloat(v[1])
			timeSeriesMap[ts] = append(timeSeriesMap[ts], struct {
				sleaf string
				tleaf string
				value float64
			}{sleaf, tleaf, val})
		}
	}

	// 转换为有序的时间点列表
	timestamps := make([]int64, 0, len(timeSeriesMap))
	for ts := range timeSeriesMap {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})

	// 构建每个时间点的热力图数据
	points := make([]HeatmapRangePoint, 0, len(timestamps))
	for _, ts := range timestamps {
		entries := timeSeriesMap[ts]

		// 初始化矩阵
		matrix := make([][]interface{}, leafCount*leafCount)
		idx := 0
		for i := 0; i < leafCount; i++ {
			for j := 0; j < leafCount; j++ {
				matrix[idx] = []interface{}{j, i, float64(0)}
				idx++
			}
		}

		// 填充数据
		for _, e := range entries {
			srcIdx, ok1 := nameToIdx[e.sleaf]
			dstIdx, ok2 := nameToIdx[e.tleaf]
			if ok1 && ok2 {
				matrix[srcIdx*leafCount+dstIdx] = []interface{}{dstIdx, srcIdx, e.value}
			}
		}

		t := time.Unix(ts, 0)
		points = append(points, HeatmapRangePoint{
			Timestamp: ts * 1000, // 转为毫秒
			TimeLabel: fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute()),
			Data:      matrix,
		})
	}

	return points, leafs, nil
}

// ── 3. 最大耗时链路 TOP ──

// LinkStat 链路统计
type LinkStat struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Avg    float64 `json:"avg"`
	P99    float64 `json:"p99"`
	Max    float64 `json:"max"`
}

// GetTopLinks 获取最大耗时链路排行（批量查询优化，避免 O(n²) 逐对查询）
func (s *LeafService) GetTopLinks(idcCode string) ([]LinkStat, error) {
	type pairKey struct {
		src, dst string
	}

	avgMap := make(map[pairKey]float64)
	p99Map := make(map[pairKey]float64)
	maxMap := make(map[pairKey]float64)
	leafSet := make(map[string]bool)

	// 批量查询：一次拉取全部 pair 数据，替代逐对查询
	collectLeaf := func(results []promResult, m map[pairKey]float64) {
		for _, r := range results {
			src := r.Metric["sleaf"]
			dst := r.Metric["tleaf"]
			if src == "" || dst == "" {
				continue
			}
			leafSet[src] = true
			leafSet[dst] = true
			m[pairKey{src, dst}] = parseFloat(r.Value[1])
		}
	}

	if avgResults, err := s.prom.QueryWithLabels(fmt.Sprintf(`ping_leaf_avg{tidc="%s"}`, idcCode)); err == nil {
		collectLeaf(avgResults, avgMap)
	}
	if p99Results, err := s.prom.QueryWithLabels(fmt.Sprintf(`ping_leaf_p99{tidc="%s"}`, idcCode)); err == nil {
		collectLeaf(p99Results, p99Map)
	}
	if maxResults, err := s.prom.QueryWithLabels(fmt.Sprintf(`ping_leaf_max{tidc="%s"}`, idcCode)); err == nil {
		collectLeaf(maxResults, maxMap)
	}

	if len(leafSet) == 0 {
		return nil, fmt.Errorf("no leaf data for idc=%s", idcCode)
	}

	// 排序 leaf 名称
	leafs := make([]string, 0, len(leafSet))
	for name := range leafSet {
		leafs = append(leafs, name)
	}
	sort.Strings(leafs)

	// 构建链路列表（只取上三角避免重复）
	links := make([]LinkStat, 0)
	for i := 0; i < len(leafs); i++ {
		for j := i + 1; j < len(leafs); j++ {
			src := leafs[i]
			dst := leafs[j]
			key := pairKey{src, dst}
			revKey := pairKey{dst, src}

			avg := avgMap[key]
			if avg == 0 {
				avg = avgMap[revKey]
			}
			p99 := p99Map[key]
			if p99 == 0 {
				p99 = p99Map[revKey]
			}
			maxVal := maxMap[key]
			if maxVal == 0 {
				maxVal = maxMap[revKey]
			}

			links = append(links, LinkStat{
				Source: src, Target: dst,
				Avg: avg, P99: p99, Max: maxVal,
			})
		}
	}

	// 按 max 降序排列
	sort.Slice(links, func(i, j int) bool {
		return links[i].Max > links[j].Max
	})

	topCount := 8
	if len(links) < topCount {
		topCount = len(links)
	}
	return links[:topCount], nil
}

// ── 3.5. ping_leaf_max_all 最大延迟查询 ──

// IDCLinkStat 同机房 Leaf 间最大延迟统计
type IDCLinkStat struct {
	Idc    string  `json:"idc"`
	SLeaf  string  `json:"sleaf"`
	TLeaf  string  `json:"tleaf"`
	Max    float64 `json:"max"`
}

// GetMaxAllLinks 查询 ping_leaf_max_all 并按 max 降序返回
func (s *LeafService) GetMaxAllLinks() ([]IDCLinkStat, error) {
	results, err := s.prom.QueryPingLeafMaxAll()
	if err != nil {
		return nil, err
	}

	links := make([]IDCLinkStat, 0, len(results))
	for _, r := range results {
		tidc := r.Metric["tidc"]
		sleaf := r.Metric["sleaf"]
		tleaf := r.Metric["tleaf"]
		val := parseFloat(r.Value[1])
		links = append(links, IDCLinkStat{
			Idc:   tidc,
			SLeaf: sleaf,
			TLeaf: tleaf,
			Max:   val,
		})
	}

	// 按 max 降序排序
	sort.Slice(links, func(i, j int) bool {
		return links[i].Max > links[j].Max
	})

	return links, nil
}

// ── 4. 时间选项 ──

// GetTimeOptions 获取最近 60 分钟时间选项
func (s *LeafService) GetTimeOptions() []string {
	options := make([]string, 60)
	now := time.Now()
	for i := 0; i < 60; i++ {
		t := now.Add(-time.Duration(59-i) * time.Minute)
		options[i] = fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
	}
	return options
}

// parseFloat 解析 interface{} 为 float64
func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// parseTimestamp 解析 interface{} 为 Unix 秒时间戳 (int64)
func parseTimestamp(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return int64(f)
	default:
		return 0
	}
}
