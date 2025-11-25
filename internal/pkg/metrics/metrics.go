package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// CounterVec 最小化实现，用 label map 组合成 key，适合少量指标。
type CounterVec struct {
	name   string
	help   string
	labels []string
	mu     sync.RWMutex
	data   map[string]uint64
}

func NewCounterVec(name, help string, labels []string) *CounterVec {
	return &CounterVec{name: name, help: help, labels: labels, data: make(map[string]uint64)}
}

func (c *CounterVec) Inc(lbls map[string]string) {
	key := buildKey(c.labels, lbls)
	c.mu.Lock()
	c.data[key]++
	c.mu.Unlock()
}

func (c *CounterVec) Export(sb *strings.Builder) {
	fmt.Fprintf(sb, "# HELP %s %s\n", c.name, c.help)
	fmt.Fprintf(sb, "# TYPE %s counter\n", c.name)
	c.mu.RLock()
	defer c.mu.RUnlock()
	for key, value := range c.data {
		fmt.Fprintf(sb, "%s{%s} %d\n", c.name, key, value)
	}
}

type Histogram struct {
	name    string
	help    string
	labels  []string
	buckets []float64

	mu     sync.RWMutex
	counts map[string][]uint64
	sum    map[string]float64
}

func NewHistogram(name, help string, labels []string, buckets []float64) *Histogram {
	// 确保升序
	cp := append([]float64{}, buckets...)
	sort.Float64s(cp)
	return &Histogram{
		name:    name,
		help:    help,
		labels:  labels,
		buckets: cp,
		counts:  make(map[string][]uint64),
		sum:     make(map[string]float64),
	}
}

func (h *Histogram) Observe(lbls map[string]string, value float64) {
	key := buildKey(h.labels, lbls)
	h.mu.Lock()
	defer h.mu.Unlock()
	bins, ok := h.counts[key]
	if !ok {
		bins = make([]uint64, len(h.buckets)+1) // +Inf bucket
		h.counts[key] = bins
	}
	for i, b := range h.buckets {
		if value <= b {
			bins[i]++
			h.sum[key] += value
			return
		}
	}
	// +Inf
	bins[len(bins)-1]++
	h.sum[key] += value
}

func (h *Histogram) Export(sb *strings.Builder) {
	fmt.Fprintf(sb, "# HELP %s %s\n", h.name, h.help)
	fmt.Fprintf(sb, "# TYPE %s histogram\n", h.name)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for key, buckets := range h.counts {
		var cumulative uint64
		for i, b := range h.buckets {
			cumulative += buckets[i]
			fmt.Fprintf(sb, "%s_bucket{%s,le=\"%.3g\"} %d\n", h.name, key, b, cumulative)
		}
		cumulative += buckets[len(buckets)-1]
		fmt.Fprintf(sb, "%s_bucket{%s,le=\"+Inf\"} %d\n", h.name, key, cumulative)
		fmt.Fprintf(sb, "%s_sum{%s} %.6f\n", h.name, key, h.sum[key])
		fmt.Fprintf(sb, "%s_count{%s} %d\n", h.name, key, cumulative)
	}
}

// Builder
func buildKey(order []string, lbls map[string]string) string {
	pairs := make([]string, 0, len(order))
	for _, k := range order {
		if v, ok := lbls[k]; ok {
			pairs = append(pairs, fmt.Sprintf("%s=\"%s\"", k, sanitize(v)))
		}
	}
	return strings.Join(pairs, ",")
}

func sanitize(v string) string {
	// 避免换行破坏格式
	return strings.ReplaceAll(strings.ReplaceAll(v, "\n", ""), "\"", "'")
}

// Registry
var (
	httpRequests  = NewCounterVec("http_requests_total", "Total HTTP requests", []string{"path", "method", "code"})
	httpDuration  = NewHistogram("http_request_duration_seconds", "HTTP request latency seconds", []string{"path", "method"}, []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2})
	seckillResult = NewCounterVec("seckill_requests_total", "Seckill business result", []string{"result"})
)

// ObserveHTTP 记录 HTTP 维度请求。
func ObserveHTTP(path, method, code string, latency time.Duration) {
	httpRequests.Inc(map[string]string{
		"path":   path,
		"method": method,
		"code":   code,
	})
	httpDuration.Observe(map[string]string{
		"path":   path,
		"method": method,
	}, latency.Seconds())
}

// IncSeckillResult 记录秒杀业务结果。
func IncSeckillResult(result string) {
	seckillResult.Inc(map[string]string{"result": result})
}

// Handler 暴露 Prometheus 文本格式。
func Handler(w http.ResponseWriter, _ *http.Request) {
	var sb strings.Builder
	httpRequests.Export(&sb)
	httpDuration.Export(&sb)
	seckillResult.Export(&sb)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(sb.String()))
}
