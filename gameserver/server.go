package gameserver

import (
	"sync"

	"github.com/pzqf/zEngine/zInject"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/net/protolayer"
	"github.com/pzqf/zGameServer/net/router"
	"go.uber.org/zap"
)

type GameServer struct {
	*zService.ServiceManager
	wg            sync.WaitGroup
	isRunning     bool
	startCalled   bool
	packetRouter  *router.PacketRouter
	protocol      protolayer.Protocol
	objectManager *zObject.ObjectManager
}

func NewGameServer() *GameServer {
	// 使用全局配置
	return NewGameServerWithConfig(config.GetServerConfig())
}

// NewGameServerWithConfig 使用配置创建游戏服务器
func NewGameServerWithConfig(serverCfg *config.ServerConfig) *GameServer {
	gs := &GameServer{
		ServiceManager: zService.NewServiceManager(),
	}

	// 注册核心依赖
	gs.RegisterCoreDependencies(serverCfg)

	// 解析依赖
	gs.resolveDependencies()

	return gs
}

// RegisterCoreDependencies 注册核心依赖
func (gs *GameServer) RegisterCoreDependencies(serverCfg *config.ServerConfig) {
	// 注册配置
	gs.RegisterSingleton("config", serverCfg)

	// 注册协议
	gs.RegisterSingleton("protocol", func() protolayer.Protocol {
		protocolName := serverCfg.Protocol
		if protocolName == "" {
			protocolName = "protobuf"
		}

		protocol, err := protolayer.NewProtocolByName(protocolName)
		if err != nil {
			zLog.Warn("Failed to create protocol, using default protobuf", zap.Error(err))
			return protolayer.NewProtobufProtocol()
		}
		return protocol
	}())

	// 注册PacketRouter
	gs.RegisterSingleton("packetRouter", router.NewPacketRouter())

	// 注册ObjectManager
	gs.RegisterSingleton("objectManager", zObject.NewObjectManager())
}

// resolveDependencies 解析依赖
func (gs *GameServer) resolveDependencies() {
	// 解析协议
	protocol, err := gs.ResolveDependency("protocol")
	if err != nil {
		zLog.Error("Failed to resolve protocol dependency", zap.Error(err))
		gs.protocol = protolayer.NewProtobufProtocol()
	} else {
		gs.protocol = protocol.(protolayer.Protocol)
	}

	// 解析PacketRouter
	packetRouter, err := gs.ResolveDependency("packetRouter")
	if err != nil {
		zLog.Error("Failed to resolve packetRouter dependency", zap.Error(err))
		gs.packetRouter = router.NewPacketRouter()
	} else {
		gs.packetRouter = packetRouter.(*router.PacketRouter)
	}

	// 解析ObjectManager
	objectManager, err := gs.ResolveDependency("objectManager")
	if err != nil {
		zLog.Error("Failed to resolve objectManager dependency", zap.Error(err))
		gs.objectManager = zObject.NewObjectManager()
	} else {
		gs.objectManager = objectManager.(*zObject.ObjectManager)
	}
}

// RegisterDependency 注册依赖
func (gs *GameServer) RegisterDependency(name string, factory interface{}) {
	gs.ServiceManager.RegisterDependency(name, factory)
}

// RegisterSingleton 注册单例依赖
func (gs *GameServer) RegisterSingleton(name string, instance interface{}) {
	gs.ServiceManager.RegisterSingleton(name, instance)
}

// ResolveDependency 解析依赖
func (gs *GameServer) ResolveDependency(name string) (interface{}, error) {
	return gs.ServiceManager.ResolveDependency(name)
}

// GetContainer 获取依赖注入容器
func (gs *GameServer) GetContainer() zInject.Container {
	return gs.ServiceManager.GetContainer()
}

func (gs *GameServer) Start() error {
	if gs.startCalled {
		return nil
	}

	// 增加等待组计数
	gs.wg.Add(1)
	gs.startCalled = true

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
	if gs.startCalled {
		gs.wg.Done()
		gs.startCalled = false
	}
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

// GetObjectManager 获取对象管理器
func (gs *GameServer) GetObjectManager() *zObject.ObjectManager {
	return gs.objectManager
}
