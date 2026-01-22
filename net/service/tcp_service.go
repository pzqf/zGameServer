package service

import (
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/net/metrics"
	"github.com/pzqf/zGameServer/net/protolayer"
	"github.com/pzqf/zGameServer/net/router"
	"go.uber.org/zap"
)

type TcpService struct {
	zObject.BaseObject
	netServer    *zNet.TcpServer
	netConfig    *zNet.TcpConfig
	packetRouter *router.PacketRouter
	metrics      *metrics.NetworkMetrics
}

// 全局网络指标监控实例
var GlobalNetworkMetrics = metrics.NewNetworkMetrics()

// GetNetworkMetrics 获取网络指标监控实例
func GetNetworkMetrics() *metrics.NetworkMetrics {
	return GlobalNetworkMetrics
}

func NewTcpService(router *router.PacketRouter) *TcpService {
	ts := &TcpService{
		packetRouter: router,
		metrics:      GlobalNetworkMetrics,
	}
	ts.SetId(ServiceIdTcpServer)
	return ts
}

func (ts *TcpService) Init() error {
	serverCfg := config.GetServerConfig()
	ts.netConfig = &zNet.TcpConfig{
		ListenAddress:  serverCfg.ListenAddress,
		ChanSize:       serverCfg.ChanSize,
		MaxClientCount: serverCfg.MaxClientCount,
	}
	zLog.Info("Initializing TCP service...", zap.String("listen_address", ts.netConfig.ListenAddress))

	// 使用标准日志接口
	logger := zLog.GetStandardLogger()
	ts.netServer = zNet.NewTcpServer(ts.netConfig, zNet.WithLogger(logger))
	ts.netServer.RegisterDispatcher(ts.dispatchPacket, 100)

	// 设置网络指标监控实例到protocol层
	protolayer.SetNetworkMetrics(ts.metrics)

	return nil
}

func (ts *TcpService) Close() error {
	zLog.Info("Closing TCP service...")
	ts.netServer.Close()
	return nil
}

func (ts *TcpService) Serve() {
	zLog.Info("Starting TCP service...")

	// 启动定期打印网络指标的协程
	go ts.startMetricsPrinter()

	if err := ts.netServer.Start(); err != nil {
		zLog.Error("Failed to start TCP service", zap.Error(err))
		return
	}
}

// startMetricsPrinter 启动定期打印网络指标的协程
func (ts *TcpService) startMetricsPrinter() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒打印一次
	defer ticker.Stop()

	for range ticker.C {
		stats := ts.metrics.GetStats()
		zLog.Info("Network metrics",
			zap.Int("active_connections", stats["active_connections"].(int)),
			zap.Float64("avg_latency_ms", stats["avg_latency_ms"].(float64)),
			zap.Float64("throughput_sent_bps", stats["throughput_sent_bps"].(float64)),
			zap.Float64("throughput_received_bps", stats["throughput_received_bps"].(float64)),
			zap.Int64("total_packets_sent", stats["total_packets_sent"].(int64)),
			zap.Int64("total_packets_received", stats["total_packets_received"].(int64)),
		)

		// 重置统计信息
		ts.metrics.Reset()
	}
}

func (ts *TcpService) dispatchPacket(session zNet.Session, packet *zNet.NetPacket) error {
	// 记录接收的数据包大小
	ts.metrics.RecordBytesReceived(len(packet.Data) + zNet.NetPacketHeadSize)

	// 记录开始处理时间
	startTime := time.Now()

	// 将 Session 接口转换为具体的 TcpServerSession 类型
	tcpSession, ok := session.(*zNet.TcpServerSession)
	if !ok {
		zLog.Error("Failed to convert session to TcpServerSession")
		ts.metrics.IncDecodingErrors()
		return nil
	}

	// 路由数据包到相应的处理程序
	err := ts.packetRouter.Route(tcpSession, packet)

	// 记录处理延迟
	latency := time.Since(startTime)
	ts.metrics.RecordLatency(latency)

	if err != nil {
		zLog.Error("Failed to route packet", zap.Error(err))
		ts.metrics.IncDecodingErrors()
	}

	return err
}
