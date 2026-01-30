package guild

import (
	"github.com/pzqf/zEngine/zActor"
	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zEngine/zLog"
	"go.uber.org/zap"
)

// GuildActorMessageType GuildActor消息类型
type GuildActorMessageType int

const (
	// GuildActorMessageTypeCreate 创建公会消息
	GuildActorMessageTypeCreate GuildActorMessageType = iota
	// GuildActorMessageTypeJoin 加入公会消息
	GuildActorMessageTypeJoin
	// GuildActorMessageTypeLeave 离开公会消息
	GuildActorMessageTypeLeave
	// GuildActorMessageTypeKick 踢出公会消息
	GuildActorMessageTypeKick
	// GuildActorMessageTypeSetPosition 设置职位消息
	GuildActorMessageTypeSetPosition
	// GuildActorMessageTypeUpdateNotice 更新公告消息
	GuildActorMessageTypeUpdateNotice
	// GuildActorMessageTypeUpgrade 升级公会消息
	GuildActorMessageTypeUpgrade
	// GuildActorMessageTypeApply 申请加入公会消息
	GuildActorMessageTypeApply
	// GuildActorMessageTypeProcessApply 处理公会申请消息
	GuildActorMessageTypeProcessApply
	// GuildActorMessageTypeUpdateContribution 更新贡献消息
	GuildActorMessageTypeUpdateContribution
)

// GuildActorCreateMessage 创建公会消息
type GuildActorCreateMessage struct {
	zActor.BaseActorMessage
	GuildId    int64
	GuildName  string
	LeaderId   int64
	LeaderName string
}

// GuildActorJoinMessage 加入公会消息
type GuildActorJoinMessage struct {
	zActor.BaseActorMessage
	PlayerId   int64
	PlayerName string
}

// GuildActorLeaveMessage 离开公会消息
type GuildActorLeaveMessage struct {
	zActor.BaseActorMessage
	PlayerId int64
}

// GuildActorKickMessage 踢出公会消息
type GuildActorKickMessage struct {
	zActor.BaseActorMessage
	PlayerId int64
	Reason   string
}

// GuildActorSetPositionMessage 设置职位消息
type GuildActorSetPositionMessage struct {
	zActor.BaseActorMessage
	PlayerId    int64
	NewPosition int
}

// GuildActorUpdateNoticeMessage 更新公告消息
type GuildActorUpdateNoticeMessage struct {
	zActor.BaseActorMessage
	Notice string
}

// GuildActorUpgradeMessage 升级公会消息
type GuildActorUpgradeMessage struct {
	zActor.BaseActorMessage
	OperatorId int64
}

// GuildActorApplyMessage 申请加入公会消息
type GuildActorApplyMessage struct {
	zActor.BaseActorMessage
	ApplyId    int64
	PlayerId   int64
	PlayerName string
	Remark     string
}

// GuildActorProcessApplyMessage 处理公会申请消息
type GuildActorProcessApplyMessage struct {
	zActor.BaseActorMessage
	ApplyId int64
	Accept  bool
	Remark  string
}

// GuildActorUpdateContributionMessage 更新贡献消息
type GuildActorUpdateContributionMessage struct {
	zActor.BaseActorMessage
	PlayerId int64
	Amount   int64
}

// GuildActor 公会Actor
type GuildActor struct {
	*zActor.BaseActor
	Guild   *Guild
	Service *GuildService
}

// NewGuildActor 创建公会Actor
func NewGuildActor(guildId int64, service *GuildService) *GuildActor {
	// 创建基础Actor
	baseActor := zActor.NewBaseActor(guildId, 100)

	// 创建并返回GuildActor
	guildActor := &GuildActor{
		BaseActor: baseActor,
		Service:   service,
	}

	return guildActor
}

// ProcessMessage 处理消息
func (ga *GuildActor) ProcessMessage(msg zActor.ActorMessage) {
	// 根据消息类型进行处理
	switch msg := msg.(type) {
	case *GuildActorCreateMessage:
		ga.handleCreateMessage(msg)
	case *GuildActorJoinMessage:
		ga.handleJoinMessage(msg)
	case *GuildActorLeaveMessage:
		ga.handleLeaveMessage(msg)
	case *GuildActorKickMessage:
		ga.handleKickMessage(msg)
	case *GuildActorSetPositionMessage:
		ga.handleSetPositionMessage(msg)
	case *GuildActorUpdateNoticeMessage:
		ga.handleUpdateNoticeMessage(msg)
	case *GuildActorUpgradeMessage:
		ga.handleUpgradeMessage(msg)
	case *GuildActorApplyMessage:
		ga.handleApplyMessage(msg)
	case *GuildActorProcessApplyMessage:
		ga.handleProcessApplyMessage(msg)
	case *GuildActorUpdateContributionMessage:
		ga.handleUpdateContributionMessage(msg)
	default:
		zLog.Warn("Unknown guild actor message type", zap.Any("message", msg))
	}
}

// handleCreateMessage 处理创建公会消息
func (ga *GuildActor) handleCreateMessage(msg *GuildActorCreateMessage) {
	// 创建公会
	guild, err := ga.Service.CreateGuild(msg.GuildId, msg.GuildName, msg.LeaderId, msg.LeaderName)
	if err != nil {
		zLog.Error("Failed to create guild", zap.Error(err))
		return
	}

	ga.Guild = guild
	zLog.Info("Guild created (Actor)", zap.Int64("guildId", msg.GuildId), zap.String("guildName", msg.GuildName))

	// 发布公会创建事件
	ga.publishEvent(zEvent.EventType(EventGuildCreate), map[string]interface{}{
		"guildId":    msg.GuildId,
		"guildName":  msg.GuildName,
		"leaderId":   msg.LeaderId,
		"leaderName": msg.LeaderName,
	})
}

// handleJoinMessage 处理加入公会消息
func (ga *GuildActor) handleJoinMessage(msg *GuildActorJoinMessage) {
	if ga.Guild == nil {
		zLog.Error("Guild not found", zap.Int64("guildId", ga.ID()))
		return
	}

	// 加入公会
	err := ga.Service.JoinGuild(msg.PlayerId, msg.PlayerName, ga.Guild.GuildId)
	if err != nil {
		zLog.Error("Failed to join guild", zap.Error(err))
		return
	}

	zLog.Info("Player joined guild (Actor)", zap.Int64("playerId", msg.PlayerId), zap.Int64("guildId", ga.Guild.GuildId))

	// 发布加入公会事件
	ga.publishEvent(zEvent.EventType(EventGuildJoin), map[string]interface{}{
		"guildId":    ga.Guild.GuildId,
		"playerId":   msg.PlayerId,
		"playerName": msg.PlayerName,
	})
}

// handleLeaveMessage 处理离开公会消息
func (ga *GuildActor) handleLeaveMessage(msg *GuildActorLeaveMessage) {
	if ga.Guild == nil {
		zLog.Error("Guild not found", zap.Int64("guildId", ga.ID()))
		return
	}

	// 离开公会
	err := ga.Service.LeaveGuild(msg.PlayerId)
	if err != nil {
		zLog.Error("Failed to leave guild", zap.Error(err))
		return
	}

	zLog.Info("Player left guild (Actor)", zap.Int64("playerId", msg.PlayerId), zap.Int64("guildId", ga.Guild.GuildId))

	// 发布离开公会事件
	ga.publishEvent(zEvent.EventType(EventGuildLeave), map[string]interface{}{
		"guildId":  ga.Guild.GuildId,
		"playerId": msg.PlayerId,
	})
}

// handleKickMessage 处理踢出公会消息
func (ga *GuildActor) handleKickMessage(msg *GuildActorKickMessage) {
	if ga.Guild == nil {
		zLog.Error("Guild not found", zap.Int64("guildId", ga.ID()))
		return
	}

	// 踢出公会
	err := ga.Service.KickGuildMember(ga.Guild.LeaderId, msg.PlayerId, msg.Reason)
	if err != nil {
		zLog.Error("Failed to kick guild member", zap.Error(err))
		return
	}

	zLog.Info("Player kicked from guild (Actor)", zap.Int64("playerId", msg.PlayerId), zap.Int64("guildId", ga.Guild.GuildId))

	// 发布踢出公会事件
	ga.publishEvent(zEvent.EventType(EventGuildKick), map[string]interface{}{
		"guildId":  ga.Guild.GuildId,
		"playerId": msg.PlayerId,
		"reason":   msg.Reason,
	})
}

// handleSetPositionMessage 处理设置职位消息
func (ga *GuildActor) handleSetPositionMessage(msg *GuildActorSetPositionMessage) {
	if ga.Guild == nil {
		zLog.Error("Guild not found", zap.Int64("guildId", ga.ID()))
		return
	}

	// 设置职位
	err := ga.Service.SetGuildMemberPosition(ga.Guild.LeaderId, msg.PlayerId, msg.NewPosition)
	if err != nil {
		zLog.Error("Failed to set guild member position", zap.Error(err))
		return
	}

	zLog.Info("Guild member position updated (Actor)", zap.Int64("playerId", msg.PlayerId), zap.Int("newPosition", msg.NewPosition))

	// 发布职位变更事件
	ga.publishEvent(zEvent.EventType(EventGuildPositionChange), map[string]interface{}{
		"guildId":     ga.Guild.GuildId,
		"playerId":    msg.PlayerId,
		"newPosition": msg.NewPosition,
	})
}

// handleUpdateNoticeMessage 处理更新公告消息
func (ga *GuildActor) handleUpdateNoticeMessage(msg *GuildActorUpdateNoticeMessage) {
	if ga.Guild == nil {
		zLog.Error("Guild not found", zap.Int64("guildId", ga.ID()))
		return
	}

	// 更新公告
	err := ga.Service.UpdateGuildNotice(ga.Guild.LeaderId, msg.Notice)
	if err != nil {
		zLog.Error("Failed to update guild notice", zap.Error(err))
		return
	}

	zLog.Info("Guild notice updated (Actor)", zap.Int64("guildId", ga.Guild.GuildId))

	// 发布公告更新事件
	ga.publishEvent(zEvent.EventType(EventGuildNoticeUpdate), map[string]interface{}{
		"guildId": ga.Guild.GuildId,
		"notice":  msg.Notice,
	})
}

// handleUpgradeMessage 处理升级公会消息
func (ga *GuildActor) handleUpgradeMessage(msg *GuildActorUpgradeMessage) {
	if ga.Guild == nil {
		zLog.Error("Guild not found", zap.Int64("guildId", ga.ID()))
		return
	}

	// 升级公会
	err := ga.Service.UpgradeGuild(msg.OperatorId)
	if err != nil {
		zLog.Error("Failed to upgrade guild", zap.Error(err))
		return
	}

	// 重新获取公会信息
	guild, _ := ga.Service.GetGuild(ga.Guild.GuildId)
	if guild != nil {
		ga.Guild = guild
	}

	zLog.Info("Guild upgraded (Actor)", zap.Int64("guildId", ga.Guild.GuildId), zap.Int("newLevel", ga.Guild.Level))

	// 发布公会升级事件
	ga.publishEvent(zEvent.EventType(EventGuildUpgrade), map[string]interface{}{
		"guildId":    ga.Guild.GuildId,
		"newLevel":   ga.Guild.Level,
		"operatorId": msg.OperatorId,
	})
}

// handleApplyMessage 处理申请加入公会消息
func (ga *GuildActor) handleApplyMessage(msg *GuildActorApplyMessage) {
	if ga.Guild == nil {
		zLog.Error("Guild not found", zap.Int64("guildId", ga.ID()))
		return
	}

	// 申请加入公会
	err := ga.Service.ApplyGuild(msg.ApplyId, msg.PlayerId, msg.PlayerName, ga.Guild.GuildId, msg.Remark)
	if err != nil {
		zLog.Error("Failed to apply for guild", zap.Error(err))
		return
	}

	zLog.Info("Guild application submitted (Actor)", zap.Int64("applyId", msg.ApplyId), zap.Int64("playerId", msg.PlayerId))

	// 发布申请加入公会事件
	ga.publishEvent(zEvent.EventType(EventGuildApply), map[string]interface{}{
		"applyId":    msg.ApplyId,
		"playerId":   msg.PlayerId,
		"playerName": msg.PlayerName,
		"guildId":    ga.Guild.GuildId,
		"remark":     msg.Remark,
	})
}

// handleProcessApplyMessage 处理公会申请消息
func (ga *GuildActor) handleProcessApplyMessage(msg *GuildActorProcessApplyMessage) {
	if ga.Guild == nil {
		zLog.Error("Guild not found", zap.Int64("guildId", ga.ID()))
		return
	}

	// 处理公会申请
	err := ga.Service.ProcessGuildApply(ga.Guild.LeaderId, msg.ApplyId, msg.Accept, msg.Remark)
	if err != nil {
		zLog.Error("Failed to process guild application", zap.Error(err))
		return
	}

	zLog.Info("Guild application processed (Actor)", zap.Int64("applyId", msg.ApplyId), zap.Bool("accepted", msg.Accept))

	// 发布处理公会申请事件
	ga.publishEvent(zEvent.EventType(EventGuildProcessApply), map[string]interface{}{
		"applyId":  msg.ApplyId,
		"accepted": msg.Accept,
		"remark":   msg.Remark,
		"guildId":  ga.Guild.GuildId,
	})
}

// handleUpdateContributionMessage 处理更新贡献消息
func (ga *GuildActor) handleUpdateContributionMessage(msg *GuildActorUpdateContributionMessage) {
	if ga.Guild == nil {
		zLog.Error("Guild not found", zap.Int64("guildId", ga.ID()))
		return
	}

	// 更新贡献
	err := ga.Service.UpdateGuildMemberContribution(msg.PlayerId, msg.Amount)
	if err != nil {
		zLog.Error("Failed to update guild member contribution", zap.Error(err))
		return
	}

	// 重新获取公会信息
	guild, _ := ga.Service.GetGuild(ga.Guild.GuildId)
	if guild != nil {
		ga.Guild = guild
	}

	zLog.Info("Guild member contribution updated (Actor)", zap.Int64("playerId", msg.PlayerId), zap.Int64("amount", msg.Amount))

	// 发布贡献更新事件
	ga.publishEvent(zEvent.EventType(EventGuildContributionUpdate), map[string]interface{}{
		"guildId":  ga.Guild.GuildId,
		"playerId": msg.PlayerId,
		"amount":   msg.Amount,
	})
}

// publishEvent 发布事件
func (ga *GuildActor) publishEvent(eventType zEvent.EventType, data map[string]interface{}) {
	// 这里可以使用全局事件总线发布事件
	// event.GetGlobalEventBus().Publish(event.NewEvent(eventType, data))
}

// Guild events
const (
	EventGuildCreate GuildEvent = iota
	EventGuildJoin
	EventGuildLeave
	EventGuildKick
	EventGuildPositionChange
	EventGuildNoticeUpdate
	EventGuildUpgrade
	EventGuildApply
	EventGuildProcessApply
	EventGuildContributionUpdate
)

// GuildEvent 公会事件类型
type GuildEvent int
