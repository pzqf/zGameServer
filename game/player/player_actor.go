package player

import (
	"github.com/pzqf/zEngine/zActor"
	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/event"
	"go.uber.org/zap"
)

// PlayerActorMessageType PlayerActor消息类型
type PlayerActorMessageType int

const (
	// PlayerActorMessageTypeConnect 连接消息
	PlayerActorMessageTypeConnect PlayerActorMessageType = iota
	// PlayerActorMessageTypeDisconnect 断开连接消息
	PlayerActorMessageTypeDisconnect
	// PlayerActorMessageTypeAddExp 增加经验消息
	PlayerActorMessageTypeAddExp
	// PlayerActorMessageTypeAddGold 增加金币消息
	PlayerActorMessageTypeAddGold
	// PlayerActorMessageTypeUseItem 使用物品消息
	PlayerActorMessageTypeUseItem
	// PlayerActorMessageTypeEquipItem 装备物品消息
	PlayerActorMessageTypeEquipItem
	// PlayerActorMessageTypeUnequipItem 卸下装备消息
	PlayerActorMessageTypeUnequipItem
	// PlayerActorMessageTypeReceiveMail 接收邮件消息
	PlayerActorMessageTypeReceiveMail
	// PlayerActorMessageTypeClaimMail 领取邮件附件消息
	PlayerActorMessageTypeClaimMail
)

// PlayerActorConnectMessage 连接消息
type PlayerActorConnectMessage struct {
	zActor.BaseActorMessage
	Session *zNet.TcpServerSession
}

// PlayerActorDisconnectMessage 断开连接消息
type PlayerActorDisconnectMessage struct {
	zActor.BaseActorMessage
}

// PlayerActorAddExpMessage 增加经验消息
type PlayerActorAddExpMessage struct {
	zActor.BaseActorMessage
	Exp int64
}

// PlayerActorAddGoldMessage 增加金币消息
type PlayerActorAddGoldMessage struct {
	zActor.BaseActorMessage
	Gold int64
}

// PlayerActorUseItemMessage 使用物品消息
type PlayerActorUseItemMessage struct {
	zActor.BaseActorMessage
	ItemID int64
	Slot   int
}

// PlayerActorEquipItemMessage 装备物品消息
type PlayerActorEquipItemMessage struct {
	zActor.BaseActorMessage
	ItemID int64
	Slot   int
	Pos    int
}

// PlayerActorUnequipItemMessage 卸下装备消息
type PlayerActorUnequipItemMessage struct {
	zActor.BaseActorMessage
	Pos int
}

// PlayerActorReceiveMailMessage 接收邮件消息
type PlayerActorReceiveMailMessage struct {
	zActor.BaseActorMessage
	MailID int64
}

// PlayerActorClaimMailMessage 领取邮件附件消息
type PlayerActorClaimMailMessage struct {
	zActor.BaseActorMessage
	MailID int64
}

// PlayerActor 玩家Actor
type PlayerActor struct {
	*zActor.BaseActor
	*Player
}

// NewPlayerActor 创建玩家Actor
func NewPlayerActor(playerId int64, name string, session *zNet.TcpServerSession) *PlayerActor {
	// 创建基础Actor
	baseActor := zActor.NewBaseActor(playerId, 100)

	// 创建Player对象
	player := NewPlayer(playerId, name, session)

	// 创建并返回PlayerActor
	playerActor := &PlayerActor{
		BaseActor: baseActor,
		Player:    player,
	}

	return playerActor
}

// ProcessMessage 处理消息
func (pa *PlayerActor) ProcessMessage(msg zActor.ActorMessage) {
	// 根据消息类型进行处理
	switch msg := msg.(type) {
	case *PlayerActorConnectMessage:
		pa.handleConnectMessage(msg)
	case *PlayerActorDisconnectMessage:
		pa.handleDisconnectMessage(msg)
	case *PlayerActorAddExpMessage:
		pa.handleAddExpMessage(msg)
	case *PlayerActorAddGoldMessage:
		pa.handleAddGoldMessage(msg)
	case *PlayerActorUseItemMessage:
		pa.handleUseItemMessage(msg)
	case *PlayerActorEquipItemMessage:
		pa.handleEquipItemMessage(msg)
	case *PlayerActorUnequipItemMessage:
		pa.handleUnequipItemMessage(msg)
	case *PlayerActorReceiveMailMessage:
		pa.handleReceiveMailMessage(msg)
	case *PlayerActorClaimMailMessage:
		pa.handleClaimMailMessage(msg)
	default:
		zLog.Warn("Unknown message type", zap.Any("message", msg))
	}
}

// handleConnectMessage 处理连接消息
func (pa *PlayerActor) handleConnectMessage(msg *PlayerActorConnectMessage) {
	pa.status.Store(int32(PlayerStatusOnline))
	pa.session = msg.Session
	zLog.Info("Player connected (Actor)", zap.Int64("playerId", pa.playerId), zap.String("name", pa.name))
}

// handleDisconnectMessage 处理断开连接消息
func (pa *PlayerActor) handleDisconnectMessage(msg *PlayerActorDisconnectMessage) {
	pa.status.Store(int32(PlayerStatusOffline))
	pa.session = nil
	zLog.Info("Player disconnected (Actor)", zap.Int64("playerId", pa.playerId), zap.String("name", pa.name))
}

// handleAddExpMessage 处理增加经验消息
func (pa *PlayerActor) handleAddExpMessage(msg *PlayerActorAddExpMessage) {
	if msg.Exp <= 0 {
		return
	}

	// 增加经验
	oldExp := pa.exp.Load()
	newExp := oldExp + msg.Exp
	pa.exp.Store(newExp)

	zLog.Debug("Player exp increased", zap.Int64("playerId", pa.playerId), zap.Int64("oldExp", oldExp), zap.Int64("newExp", newExp))
}

// handleAddGoldMessage 处理增加金币消息
func (pa *PlayerActor) handleAddGoldMessage(msg *PlayerActorAddGoldMessage) {
	if msg.Gold <= 0 {
		return
	}

	// 增加金币
	oldGold := pa.gold.Load()
	newGold := oldGold + msg.Gold
	pa.gold.Store(newGold)

	zLog.Debug("Player gold increased", zap.Int64("playerId", pa.playerId), zap.Int64("oldGold", oldGold), zap.Int64("newGold", newGold))
}

// handleUseItemMessage 处理使用物品消息
func (pa *PlayerActor) handleUseItemMessage(msg *PlayerActorUseItemMessage) {
	// TODO: 实现使用物品逻辑
}

// handleEquipItemMessage 处理装备物品消息
func (pa *PlayerActor) handleEquipItemMessage(msg *PlayerActorEquipItemMessage) {
	// TODO: 实现装备物品逻辑
	inventory := pa.GetInventory()
	equipment := pa.GetEquipment()
	if inventory == nil || equipment == nil {
		return
	}

	if item, exists := inventory.items.Get(msg.Slot); exists {
		equipment.Equip(msg.Pos, item.(*Item))
	}
}

// handleUnequipItemMessage 处理卸下装备消息
func (pa *PlayerActor) handleUnequipItemMessage(msg *PlayerActorUnequipItemMessage) {
	// TODO: 实现卸下装备逻辑
	equipment := pa.GetEquipment()
	if equipment == nil {
		return
	}

	equipment.Unequip(msg.Pos)
}

// handleReceiveMailMessage 处理接收邮件消息
func (pa *PlayerActor) handleReceiveMailMessage(msg *PlayerActorReceiveMailMessage) {
	// TODO: 实现接收邮件逻辑
	// 注意：ReceiveMail方法不存在，通常是通过服务器直接调用SendMail
	// 这里只是为了演示消息处理

	// 发布接收邮件事件
	pa.publishEvent(event.EventPlayerMailReceived, &event.PlayerMailEventData{
		PlayerID: pa.playerId,
		MailID:   msg.MailID,
	})
}

// handleClaimMailMessage 处理领取邮件附件消息
func (pa *PlayerActor) handleClaimMailMessage(msg *PlayerActorClaimMailMessage) {
	// TODO: 实现领取邮件附件逻辑
	mailbox := pa.GetMailbox()
	if mailbox == nil {
		return
	}

	mailbox.ClaimAttachments(msg.MailID)

	// 发布领取邮件附件事件
	pa.publishEvent(event.EventPlayerMailClaimed, &event.PlayerMailEventData{
		PlayerID: pa.playerId,
		MailID:   msg.MailID,
	})
}

// 所有Get方法都从Player结构体继承，无需重复定义

// publishEvent 发布事件的通用方法
func (pa *PlayerActor) publishEvent(eventType zEvent.EventType, data interface{}) {
	eventObj := event.NewEvent(
		eventType,
		pa,
		data,
	)
	event.GetGlobalEventBus().Publish(eventObj)
}
