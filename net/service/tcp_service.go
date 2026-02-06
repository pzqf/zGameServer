package service

import (
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/metrics"
	"github.com/pzqf/zGameServer/net/protolayer"
	"github.com/pzqf/zGameServer/net/router"
	"github.com/pzqf/zGameServer/util"
	"go.uber.org/zap"
)

type TcpService struct {
	zService.BaseService
	netServer    *zNet.TcpServer
	netConfig    *zNet.TcpConfig
	packetRouter *router.PacketRouter
	metrics      *metrics.NetworkMetrics
	protocol     protolayer.Protocol
}

func NewTcpService(router *router.PacketRouter) *TcpService {
	ts := &TcpService{
		BaseService:  *zService.NewBaseService(common.ServiceIdTcpServer),
		packetRouter: router,
		metrics:      metrics.NewNetworkMetrics(),
	}
	return ts
}

func (ts *TcpService) Init() error {
	ts.SetState(zService.ServiceStateInit)

	serverCfg := config.GetServerConfig()
	ts.netConfig = &zNet.TcpConfig{
		ListenAddress:     serverCfg.ListenAddress,
		ChanSize:          serverCfg.ChanSize,
		MaxClientCount:    serverCfg.MaxClientCount,
		HeartbeatDuration: serverCfg.HeartbeatDuration,
	}
	zLog.Info("Initializing TCP service...", zap.String("listen_address", ts.netConfig.ListenAddress))

	// 根据配置创建协议实例
	protocolName := serverCfg.Protocol
	if protocolName == "" {
		protocolName = "protobuf" // 默认使用protobuf
	}

	var err error
	ts.protocol, err = protolayer.NewProtocolByName(protocolName)
	if err != nil {
		zLog.Warn("Failed to create protocol, using default protobuf", zap.Error(err))
		ts.protocol = protolayer.NewProtobufProtocol()
	}
	zLog.Info("Protocol initialized", zap.String("protocol", protocolName))

	// 配置防DDoS攻击参数
	ddosConfig := &config.GetConfig().DDoS

	// 使用标准日志接口
	logger := zLog.GetStandardLogger()
	ts.netServer = zNet.NewTcpServer(ts.netConfig, zNet.WithLogger(logger), zNet.WithDDoSConfig(ddosConfig))
	ts.netServer.RegisterDispatcher(ts.dispatchPacket)

	// 设置网络指标监控实例到protocol层
	protolayer.SetNetworkMetrics(ts.metrics)

	return nil
}

func (ts *TcpService) Close() error {
	ts.SetState(zService.ServiceStateStopping)
	zLog.Info("Closing TCP service...")
	ts.netServer.Close()
	ts.SetState(zService.ServiceStateStopped)
	return nil
}

func (ts *TcpService) Serve() {
	ts.SetState(zService.ServiceStateRunning)
	zLog.Info("Starting TCP service...")

	// 启动定期打印网络指标的协程
	go ts.startMetricsPrinter()

	if err := ts.netServer.Start(); err != nil {
		zLog.Error("Failed to start TCP service", zap.Error(err))
		ts.SetState(zService.ServiceStateStopped)
		return
	}
}

// startMetricsPrinter 启动定期打印网络指标的协程
func (ts *TcpService) startMetricsPrinter() {
	defer util.Recover(func(recover interface{}, stack string) {
		zLog.Error("Metrics printer panicked", zap.Any("panic", recover))
	})

	ticker := time.NewTicker(30 * time.Second) // 每30秒打印一次
	defer ticker.Stop()

	for range ticker.C {
		defer util.Recover(func(recover interface{}, stack string) {
			zLog.Error("Metrics printing panicked", zap.Any("panic", recover))
		})

		stats := ts.metrics.GetStats()
		zLog.Info("Network metrics",
			zap.Int("active_connections", stats["active_connections"].(int)),
			zap.Float64("avg_latency_ms", stats["avg_latency_ms"].(float64)),
			zap.Float64("throughput_sent_bps", stats["throughput_sent_bps"].(float64)),
			zap.Float64("throughput_received_bps", stats["throughput_received_bps"].(float64)),
			zap.Int64("total_packets_sent", stats["total_packets_sent"].(int64)),
			zap.Int64("total_packets_received", stats["total_packets_received"].(int64)),
			zap.Int64("dropped_packets", stats["dropped_packets"].(int64)),
		)

		// 重置统计信息
		ts.metrics.Reset()
	}
}

func (ts *TcpService) dispatchPacket(session zNet.Session, packet *zNet.NetPacket) error {
	// 记录接收的数据包大小
	ts.metrics.RecordBytesReceived(len(packet.Data) + zNet.NetPacketHeadSize)

	// 将 Session 接口转换为具体的 TcpServerSession 类型
	tcpSession, ok := session.(*zNet.TcpServerSession)
	if !ok {
		zLog.Error("Failed to convert session to TcpServerSession")
		ts.metrics.IncDecodingErrors()
		return nil
	}

	// 直接处理数据包，保证顺序
	defer util.Recover(func(recover interface{}, stack string) {
		zLog.Error("Packet processing panicked", zap.Any("panic", recover))
	})

	ts.processPacket(tcpSession, packet)

	return nil
}

// processPacket 处理数据包
func (ts *TcpService) processPacket(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	// 记录开始处理时间
	startTime := time.Now()

	// 路由数据包到相应的处理程序
	err := ts.packetRouter.Route(session, packet)

	// 记录处理延迟
	latency := time.Since(startTime)
	ts.metrics.RecordLatency(latency)

	if err != nil {
		zLog.Error("Failed to route packet", zap.Error(err))
		ts.metrics.IncDecodingErrors()
	}

	return err
}
