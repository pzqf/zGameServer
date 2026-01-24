package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricsManager 指标管理管理器
type MetricsManager struct {
	mu              sync.RWMutex
	networkMetrics  *NetworkMetrics
	businessMetrics map[string]*BusinessMetrics

	// Prometheus 相关
	registry   *prometheus.Registry
	counters   map[string]prometheus.Counter
	histograms map[string]prometheus.Histogram
	gauges     map[string]prometheus.Gauge
}

// BusinessMetrics 业务指标监控
type BusinessMetrics struct {
	mu sync.RWMutex

	// 计数器
	counters map[string]int64

	// 计时器
	timers map[string]time.Duration

	// 采样时间
	lastSampleTime time.Time
}

// NewMetricsManager 创建指标管理器实例
func NewMetricsManager() *MetricsManager {
	return &MetricsManager{
		networkMetrics:  NewNetworkMetrics(),
		businessMetrics: make(map[string]*BusinessMetrics),
		registry:        prometheus.NewRegistry(),
		counters:        make(map[string]prometheus.Counter),
		histograms:      make(map[string]prometheus.Histogram),
		gauges:          make(map[string]prometheus.Gauge),
	}
}

// GetNetworkMetrics 获取网络指标实例
func (m *MetricsManager) GetNetworkMetrics() *NetworkMetrics {
	return m.networkMetrics
}

// GetBusinessMetrics 获取或创建业务指标实例
func (m *MetricsManager) GetBusinessMetrics(name string) *BusinessMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics, exists := m.businessMetrics[name]; exists {
		return metrics
	}

	metrics := NewBusinessMetrics()
	m.businessMetrics[name] = metrics
	return metrics
}

// GetAllBusinessMetrics 获取所有业务指标实例
func (m *MetricsManager) GetAllBusinessMetrics() map[string]*BusinessMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本，避免并发修改
	metrics := make(map[string]*BusinessMetrics, len(m.businessMetrics))
	for k, v := range m.businessMetrics {
		metrics[k] = v
	}

	return metrics
}

// ResetAll 重置所有指标
func (m *MetricsManager) ResetAll() {
	m.networkMetrics.Reset()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, metrics := range m.businessMetrics {
		metrics.Reset()
	}
}

// RegisterCounter 注册一个counter类型的指标
func (m *MetricsManager) RegisterCounter(name, help string, labels map[string]string) prometheus.Counter {
	m.mu.Lock()
	defer m.mu.Unlock()

	if counter, exists := m.counters[name]; exists {
		return counter
	}

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        name,
		Help:        help,
		ConstLabels: labels,
	})

	m.registry.MustRegister(counter)
	m.counters[name] = counter
	return counter
}

// RegisterHistogram 注册一个histogram类型的指标
func (m *MetricsManager) RegisterHistogram(name, help string, buckets []float64, labels map[string]string) prometheus.Histogram {
	m.mu.Lock()
	defer m.mu.Unlock()

	if histogram, exists := m.histograms[name]; exists {
		return histogram
	}

	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        name,
		Help:        help,
		Buckets:     buckets,
		ConstLabels: labels,
	})

	m.registry.MustRegister(histogram)
	m.histograms[name] = histogram
	return histogram
}

// RegisterGauge 注册一个gauge类型的指标
func (m *MetricsManager) RegisterGauge(name, help string, labels map[string]string) prometheus.Gauge {
	m.mu.Lock()
	defer m.mu.Unlock()

	if gauge, exists := m.gauges[name]; exists {
		return gauge
	}

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        name,
		Help:        help,
		ConstLabels: labels,
	})

	m.registry.MustRegister(gauge)
	m.gauges[name] = gauge
	return gauge
}

// GetCounter 获取一个counter类型的指标
func (m *MetricsManager) GetCounter(name string) prometheus.Counter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.counters[name]
}

// GetHistogram 获取一个histogram类型的指标
func (m *MetricsManager) GetHistogram(name string) prometheus.Histogram {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.histograms[name]
}

// GetGauge 获取一个gauge类型的指标
func (m *MetricsManager) GetGauge(name string) prometheus.Gauge {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.gauges[name]
}

// GetRegistry 获取prometheus的registry
func (m *MetricsManager) GetRegistry() *prometheus.Registry {
	return m.registry
}

// NewBusinessMetrics 创建业务指标实例
func NewBusinessMetrics() *BusinessMetrics {
	return &BusinessMetrics{
		counters:       make(map[string]int64),
		timers:         make(map[string]time.Duration),
		lastSampleTime: time.Now(),
	}
}

// IncCounter 增加计数器
func (m *BusinessMetrics) IncCounter(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[name]++
}

// AddCounter 增加计数器指定值
func (m *BusinessMetrics) AddCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[name] += value
}

// SetCounter 设置计数器值
func (m *BusinessMetrics) SetCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[name] = value
}

// GetCounter 获取计数器值
func (m *BusinessMetrics) GetCounter(name string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.counters[name]
}

// RecordTimer 记录计时器
func (m *BusinessMetrics) RecordTimer(name string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.timers[name] = duration
}

// GetTimer 获取计时器值
func (m *BusinessMetrics) GetTimer(name string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.timers[name]
}

// GetAllCounters 获取所有计数器
func (m *BusinessMetrics) GetAllCounters() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本，避免并发修改
	counters := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		counters[k] = v
	}

	return counters
}

// GetAllTimers 获取所有计时器
func (m *BusinessMetrics) GetAllTimers() map[string]time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本，避免并发修改
	timers := make(map[string]time.Duration, len(m.timers))
	for k, v := range m.timers {
		timers[k] = v
	}

	return timers
}

// Reset 重置业务指标
func (m *BusinessMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters = make(map[string]int64)
	m.timers = make(map[string]time.Duration)
	m.lastSampleTime = time.Now()
}
