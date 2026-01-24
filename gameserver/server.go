package gameserver

import (
	"sync"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/net/protolayer"
	"github.com/pzqf/zGameServer/net/router"
	"go.uber.org/zap"
)

type GameServer struct {
	*zService.ServiceManager
	wg           sync.WaitGroup
	isRunning    bool
	packetRouter *router.PacketRouter
	protocol     protolayer.Protocol
}

func NewGameServer() *GameServer {
	// 使用全局配置
	return NewGameServerWithConfig(config.GetServerConfig())
}

// NewGameServerWithConfig 使用配置创建游戏服务器
func NewGameServerWithConfig(serverCfg *config.ServerConfig) *GameServer {
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
		zLog.Warn("Unknown protocol type, using default protobuf", zap.String("protocol", serverCfg.Protocol))
		protocol = protolayer.NewProtobufProtocol()
	}

	gs := &GameServer{
		ServiceManager: zService.NewServiceManager(),
		packetRouter:   router.NewPacketRouter(),
		protocol:       protocol,
	}

	return gs
}

func (gs *GameServer) Start() error {
	// 增加等待组计数
	gs.wg.Add(1)

	// 启动所有服务（包括TCP服务）
	zLog.Info("Starting all game services...")
	gs.ServeServices()

	gs.isRunning = true
	return nil
}

func (gs *GameServer) Stop() {
	if !gs.isRunning {
		return
	}

	zLog.Info("Stopping game server...")

	// 关闭所有服务
	gs.CloseServices()

	gs.isRunning = false
	gs.wg.Done()
}

func (gs *GameServer) Wait() {
	gs.wg.Wait()
}

func (gs *GameServer) GetPacketRouter() *router.PacketRouter {
	return gs.packetRouter
}

// GetProtocol 获取协议实例
func (gs *GameServer) GetProtocol() protolayer.Protocol {
	return gs.protocol
}
