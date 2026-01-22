package gameserver

import (
	"sync"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"go.uber.org/zap"
)

type PacketHandlerFunc func(session *zNet.TcpServerSession, packet *zNet.NetPacket) error

type PacketRouter struct {
	handlers sync.Map // key: int32(protoId), value: PacketHandlerFunc
}

func NewPacketRouter() *PacketRouter {
	return &PacketRouter{}
}

func (pr *PacketRouter) RegisterHandler(protoId int32, handler PacketHandlerFunc) {
	pr.handlers.Store(protoId, handler)
	zLog.Info("Registered packet handler", zap.Int32("protoId", protoId))
}

func (pr *PacketRouter) UnregisterHandler(protoId int32) {
	pr.handlers.Delete(protoId)
	zLog.Info("Unregistered packet handler", zap.Int32("protoId", protoId))
}

func (pr *PacketRouter) Route(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	if session == nil || packet == nil {
		return nil
	}

	// 查找对应的处理程序
	handlerInterface, exists := pr.handlers.Load(packet.ProtoId)
	if !exists {
		zLog.Warn("No handler found for packet", zap.Int32("protoId", packet.ProtoId))
		return nil
	}

	// 调用处理程序
	handler, ok := handlerInterface.(PacketHandlerFunc)
	if !ok {
		zLog.Error("Invalid packet handler type", zap.Int32("protoId", packet.ProtoId))
		return nil
	}

	return handler(session, packet)
}
