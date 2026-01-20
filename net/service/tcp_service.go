package service

import (
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/net/router"
	"go.uber.org/zap"
)

type TcpService struct {
	zObject.BaseObject
	logger       *zap.Logger
	netServer    *zNet.TcpServer
	netConfig    *zNet.TcpConfig
	packetRouter *router.PacketRouter
}

func NewTcpService(router *router.PacketRouter) *TcpService {
	ts := &TcpService{
		logger:       zLog.GetLogger(),
		packetRouter: router,
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
	ts.logger.Info("Initializing TCP service...", zap.String("listen_address", ts.netConfig.ListenAddress))
	ts.netServer = zNet.NewTcpServer(ts.netConfig)
	ts.netServer.RegisterHandler(ts.handlePacket, 100)
	return nil
}

func (ts *TcpService) Close() error {
	ts.logger.Info("Closing TCP service...")
	ts.netServer.Close()
	return nil
}

func (ts *TcpService) Serve() {
	ts.logger.Info("Starting TCP service...")
	if err := ts.netServer.Start(); err != nil {
		ts.logger.Error("Failed to start TCP service", zap.Error(err))
		return
	}
}

func (ts *TcpService) handlePacket(session zNet.Session, packet *zNet.NetPacket) error {
	// 将 Session 接口转换为具体的 TcpServerSession 类型
	tcpSession, ok := session.(*zNet.TcpServerSession)
	if !ok {
		ts.logger.Error("Failed to convert session to TcpServerSession")
		return nil
	}
	// 路由数据包到相应的处理程序
	return ts.packetRouter.Route(tcpSession, packet)
}
