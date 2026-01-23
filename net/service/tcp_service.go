package service

import (
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/net/metrics"
	"github.com/pzqf/zGameServer/net/pool"
	"github.com/pzqf/zGameServer/net/protolayer"
	"github.com/pzqf/zGameServer/net/router"
	"github.com/pzqf/zGameServer/util"
	"go.uber.org/zap"
)

type TcpService struct {
	zObject.BaseObject
	netServer    *zNet.TcpServer
	netConfig    *zNet.TcpConfig
	packetRouter *router.PacketRouter
	metrics      *metrics.NetworkMetrics
	workerPool   chan *pool.PacketTask
	workerCount  int
	taskPool     *pool.PacketTaskPool
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
	ts.SetId(util.ServiceIdTcpServer)
	return ts
}

func (ts *TcpService) Init() error {
	ts.SetState(zService.ServiceStateInit)

	serverCfg := config.GetServerConfig()
	ts.netConfig = &zNet.TcpConfig{
		ListenAddress:  serverCfg.ListenAddress,
		ChanSize:       serverCfg.ChanSize,
		MaxClientCount: serverCfg.MaxClientCount,
	}
	zLog.Info("Initializing TCP service...", zap.String("listen_address", ts.netConfig.ListenAddress))

	// 配置防DDoS攻击参数
	ddosConfig := &config.GetConfig().DDoS

	// 使用标准日志接口
	logger := zLog.GetStandardLogger()
	ts.netServer = zNet.NewTcpServer(ts.netConfig, zNet.WithLogger(logger), zNet.WithDDoSConfig(ddosConfig))
	ts.netServer.RegisterDispatcher(ts.dispatchPacket, 100)

	// 初始化工作池
	ts.workerCount = 50                                // 工作线程数量
	ts.workerPool = make(chan *pool.PacketTask, 10000) // 任务队列大小

	// 初始化对象池
	ts.taskPool = pool.NewPacketTaskPool(1024)

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

	// 启动工作线程
	for i := 0; i < ts.workerCount; i++ {
		go ts.workerLoop()
	}

	// 启动定期打印网络指标的协程
	go ts.startMetricsPrinter()

	if err := ts.netServer.Start(); err != nil {
		zLog.Error("Failed to start TCP service", zap.Error(err))
		ts.SetState(zService.ServiceStateStopped)
		return
	}
}

// workerLoop 工作线程循环，处理数据包任务
func (ts *TcpService) workerLoop() {
	defer func() {
		if r := recover(); r != nil {
			zLog.Error("Worker loop panicked", zap.Any("panic", r))
		}
	}()

	for task := range ts.workerPool {
		defer func() {
			if r := recover(); r != nil {
				zLog.Error("Packet processing panicked", zap.Any("panic", r))
				ts.taskPool.PutPacketTask(task)
			}
		}()

		// 处理数据包
		tcpSession, ok := task.Session.(*zNet.TcpServerSession)
		if !ok {
			zLog.Error("Failed to convert session to TcpServerSession")
			ts.taskPool.PutPacketTask(task)
			continue
		}
		ts.processPacket(tcpSession, task.Packet)
		// 处理完成后将任务放回对象池
		ts.taskPool.PutPacketTask(task)
	}
}

// startMetricsPrinter 启动定期打印网络指标的协程
func (ts *TcpService) startMetricsPrinter() {
	defer func() {
		if r := recover(); r != nil {
			zLog.Error("Metrics printer panicked", zap.Any("panic", r))
		}
	}()

	ticker := time.NewTicker(30 * time.Second) // 每30秒打印一次
	defer ticker.Stop()

	for range ticker.C {
		defer func() {
			if r := recover(); r != nil {
				zLog.Error("Metrics printing panicked", zap.Any("panic", r))
			}
		}()

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

	// 从对象池获取数据包任务
	task := ts.taskPool.GetPacketTask()
	task.Session = tcpSession
	task.Packet = packet

	// 尝试发送任务到工作池，如果工作池已满，记录错误
	select {
	case ts.workerPool <- task:
		// 任务发送成功
	default:
		// 工作池已满，丢弃任务并放回对象池
		zLog.Warn("Worker pool is full, dropping packet", zap.Int32("protoId", packet.ProtoId))
		ts.metrics.IncDroppedPackets()
		ts.taskPool.PutPacketTask(task)
	}

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

// GetState 获取服务状态
func (ts *TcpService) GetState() zService.ServiceState {
	// 简单实现，返回服务状态
	return zService.ServiceStateUnknown
}

// SetState 设置服务状态
func (ts *TcpService) SetState(state zService.ServiceState) {
	// 简单实现，设置服务状态
}
