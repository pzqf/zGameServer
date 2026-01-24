package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// globalMetricsManager 全局指标管理器实例
	globalMetricsManager *MetricsManager
	// once 确保全局指标管理器只初始化一次
	once sync.Once
)

// GetMetricsManager 获取全局指标管理器实例
func GetMetricsManager() *MetricsManager {
	once.Do(func() {
		globalMetricsManager = NewMetricsManager()
	})

	return globalMetricsManager
}

// GetNetworkMetrics 获取全局网络指标实例
func GetNetworkMetrics() *NetworkMetrics {
	return GetMetricsManager().GetNetworkMetrics()
}

// GetBusinessMetrics 获取或创建全局业务指标实例
func GetBusinessMetrics(name string) *BusinessMetrics {
	return GetMetricsManager().GetBusinessMetrics(name)
}

// ResetAllMetrics 重置所有指标
func ResetAllMetrics() {
	GetMetricsManager().ResetAll()
}

// RegisterCounter 注册一个全局的counter类型指标
func RegisterCounter(name, help string, labels map[string]string) prometheus.Counter {
	return GetMetricsManager().RegisterCounter(name, help, labels)
}

// RegisterHistogram 注册一个全局的histogram类型指标
func RegisterHistogram(name, help string, buckets []float64, labels map[string]string) prometheus.Histogram {
	return GetMetricsManager().RegisterHistogram(name, help, buckets, labels)
}

// RegisterGauge 注册一个全局的gauge类型指标
func RegisterGauge(name, help string, labels map[string]string) prometheus.Gauge {
	return GetMetricsManager().RegisterGauge(name, help, labels)
}

// GetCounter 获取一个全局的counter类型指标
func GetCounter(name string) prometheus.Counter {
	return GetMetricsManager().GetCounter(name)
}

// GetHistogram 获取一个全局的histogram类型指标
func GetHistogram(name string) prometheus.Histogram {
	return GetMetricsManager().GetHistogram(name)
}

// GetGauge 获取一个全局的gauge类型指标
func GetGauge(name string) prometheus.Gauge {
	return GetMetricsManager().GetGauge(name)
}

// GetRegistry 获取全局的prometheus registry
func GetRegistry() *prometheus.Registry {
	return GetMetricsManager().GetRegistry()
}
