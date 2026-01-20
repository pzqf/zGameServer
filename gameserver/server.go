package gameserver

import (
	"sync"

	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/net/protolayer"
	"github.com/pzqf/zGameServer/net/router"
	"github.com/pzqf/zGameServer/net/service"
	"go.uber.org/zap"
)

type GameServer struct {
	*zService.ServiceManager
	logger       *zap.Logger
	wg           sync.WaitGroup
	isRunning    bool
	packetRouter *router.PacketRouter
	protocol     protolayer.Protocol
}

func NewGameServer(logger *zap.Logger) *GameServer {
	// 使用全局配置
	return NewGameServerWithConfig(logger, config.GetServerConfig())
}

// NewGameServerWithConfig 使用配置创建游戏服务器
func NewGameServerWithConfig(logger *zap.Logger, serverCfg *config.ServerConfig) *GameServer {
	// 根据配置选择协议类型
	var protocol protolayer.Protocol
	switch serverCfg.Protocol {
	case "protobuf":
		protocol = protolayer.NewProtobufProtocol()
	case "json":
		protocol = protolayer.NewJSONProtocol()
	case "xml":
		protocol = protolayer.NewXMLProtocol()
	default:
		logger.Warn("Unknown protocol type, using default protobuf", zap.String("protocol", serverCfg.Protocol))
		protocol = protolayer.NewProtobufProtocol()
	}

	gs := &GameServer{
		ServiceManager: zService.NewServiceManager(),
		logger:         logger,
		packetRouter:   router.NewPacketRouter(logger),
		protocol:       protocol,
	}

	// 创建并注册TCP服务
	tcpService := service.NewTcpService(gs.packetRouter)
	if err := gs.AddService(tcpService); err != nil {
		logger.Fatal("Failed to add TCP service", zap.Error(err))
	}

	// 创建并注册HTTP服务
	httpService := service.NewHTTPService()
	if err := gs.AddService(httpService); err != nil {
		logger.Fatal("Failed to add HTTP service", zap.Error(err))
	}

	return gs
}

func (gs *GameServer) Start() error {
	// 启动所有服务（包括TCP服务）
	gs.logger.Info("Starting all game services...")
	gs.ServeServices()

	gs.isRunning = true
	return nil
}

func (gs *GameServer) Stop() {
	if !gs.isRunning {
		return
	}

	gs.logger.Info("Stopping game server...")

	// 关闭所有服务
	gs.CloseServices()

	gs.isRunning = false
	gs.wg.Done()
}

func (gs *GameServer) Wait() {
	gs.wg.Add(1)
	gs.wg.Wait()
}

func (gs *GameServer) GetPacketRouter() *router.PacketRouter {
	return gs.packetRouter
}

// GetProtocol 获取协议实例
func (gs *GameServer) GetProtocol() protolayer.Protocol {
	return gs.protocol
}
