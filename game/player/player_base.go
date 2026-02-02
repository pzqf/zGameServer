package player

import (
	"sync/atomic"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/game/object/component"
	"go.uber.org/zap"
)

// BaseInfo 玩家基础信息组件
type BaseInfo struct {
	*component.BaseComponent
	name       string
	session    *zNet.TcpServerSession
	status     atomic.Int32
	exp        atomic.Int64
	gold       atomic.Int64
	level      atomic.Int32
	vipLevel   atomic.Int32
	serverId   int
	createTime int64
}

// NewBaseInfo 创建新的基础信息组件
func NewBaseInfo(name string, session *zNet.TcpServerSession) *BaseInfo {
	return &BaseInfo{
		BaseComponent: component.NewBaseComponent("baseinfo"),
		name:          name,
		session:       session,
		serverId:      1,
		createTime:    time.Now().UnixMilli(),
	}
}

func (b *BaseInfo) Init() error {
	return nil
}

func (b *BaseInfo) Destroy() {
}

// GetName 获取玩家名称
func (b *BaseInfo) GetName() string {
	return b.name
}

// SetName 设置玩家名称
func (b *BaseInfo) SetName(name string) {
	b.name = name
}

// GetServerId 获取服务器ID
func (b *BaseInfo) GetServerId() int {
	return b.serverId
}

// SetServerId 设置服务器ID
func (b *BaseInfo) SetServerId(serverId int) {
	b.serverId = serverId
}

// GetSession 获取会话
func (b *BaseInfo) GetSession() *zNet.TcpServerSession {
	return b.session
}

// SetSession 设置会话
func (b *BaseInfo) SetSession(session *zNet.TcpServerSession) {
	b.session = session
}

// GetStatus 获取玩家状态
func (b *BaseInfo) GetStatus() PlayerStatus {
	return PlayerStatus(b.status.Load())
}

func (b *BaseInfo) SetStatus(status PlayerStatus) {
	b.status.Store(int32(status))
}

// IsOnline 检查玩家是否在线
func (b *BaseInfo) IsOnline() bool {
	return b.GetStatus() == PlayerStatusOnline && b.session != nil
}

// IsBusy 检查玩家是否忙
func (b *BaseInfo) IsBusy() bool {
	return b.GetStatus() == PlayerStatusBusy
}

// IsAFK 检查玩家是否挂机
func (b *BaseInfo) IsAFK() bool {
	return b.GetStatus() == PlayerStatusAFK
}

// GetLevel 获取玩家等级
func (b *BaseInfo) GetLevel() int {
	return int(b.level.Load())
}

// SetLevel 设置玩家等级
func (b *BaseInfo) SetLevel(level int) {
	b.level.Store(int32(level))
}

// GetExp 获取玩家经验
func (b *BaseInfo) GetExp() int64 {
	return b.exp.Load()
}

// SetExp 设置玩家经验
func (b *BaseInfo) SetExp(exp int64) {
	b.exp.Store(exp)
}

// AddExp 增加玩家经验
func (b *BaseInfo) AddExp(exp int64) {
	currentExp := b.exp.Load()
	newExp := currentExp + exp
	b.exp.Store(newExp)

	zLog.Info("Player gained exp",
		zap.Int64("exp", exp),
		zap.Int64("totalExp", newExp))
}

// GetGold 获取玩家金币
func (b *BaseInfo) GetGold() int64 {
	return b.gold.Load()
}

// SetGold 设置玩家金币
func (b *BaseInfo) SetGold(gold int64) {
	if gold < 0 {
		gold = 0
	}
	b.gold.Store(gold)
}

// AddGold 增加玩家金币
func (b *BaseInfo) AddGold(gold int64) {
	currentGold := b.gold.Load()
	newGold := currentGold + gold
	b.gold.Store(newGold)

	zLog.Info("Player gained gold",
		zap.Int64("gold", gold),
		zap.Int64("totalGold", newGold))
}

// SubGold 减少玩家金币
func (b *BaseInfo) SubGold(gold int64) bool {
	currentGold := b.gold.Load()
	if currentGold < gold {
		return false
	}

	newGold := currentGold - gold
	b.gold.Store(newGold)

	return true
}

// GetVIPLevel 获取VIP等级
func (b *BaseInfo) GetVIPLevel() int {
	return int(b.vipLevel.Load())
}

// SetVIPLevel 设置VIP等级
func (b *BaseInfo) SetVIPLevel(vipLevel int) {
	b.vipLevel.Store(int32(vipLevel))
}

// GetCreateTime 获取创建时间
func (b *BaseInfo) GetCreateTime() int64 {
	return b.createTime
}

// SendPacket 发送数据包
func (b *BaseInfo) SendPacket(packetId int32, data []byte) error {
	if b.session == nil {
		zLog.Warn("Player session is nil")
		return nil
	}
	return b.session.Send(packetId, data)
}

// SendText 发送文本消息
func (b *BaseInfo) SendText(message string) error {
	return b.SendPacket(1001, []byte(message))
}
