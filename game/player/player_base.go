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
// 管理玩家的基本属性：名称、等级、经验、金币、状态等
type BaseInfo struct {
	*component.BaseComponent      // 继承基础组件
	name       string             // 玩家名称
	session    *zNet.TcpServerSession // 网络会话
	status     atomic.Int32       // 玩家状态（原子操作）
	exp        atomic.Int64       // 经验值（原子操作）
	gold       atomic.Int64       // 金币（原子操作）
	level      atomic.Int32       // 等级（原子操作）
	vipLevel   atomic.Int32       // VIP等级（原子操作）
	serverId   int                // 服务器ID
	createTime int64              // 创建时间戳
}

// NewBaseInfo 创建新的基础信息组件
// 参数:
//   - name: 玩家名称
//   - session: 网络会话
//
// 返回:
//   - *BaseInfo: 新创建的组件
func NewBaseInfo(name string, session *zNet.TcpServerSession) *BaseInfo {
	return &BaseInfo{
		BaseComponent: component.NewBaseComponent("baseinfo"),
		name:          name,
		session:       session,
		serverId:      1,
		createTime:    time.Now().UnixMilli(),
	}
}

// Init 初始化组件
func (b *BaseInfo) Init() error {
	return nil
}

// Destroy 销毁组件
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

// GetSession 获取网络会话
func (b *BaseInfo) GetSession() *zNet.TcpServerSession {
	return b.session
}

// SetSession 设置网络会话
func (b *BaseInfo) SetSession(session *zNet.TcpServerSession) {
	b.session = session
}

// GetStatus 获取玩家状态
// 返回PlayerStatus枚举值
func (b *BaseInfo) GetStatus() PlayerStatus {
	return PlayerStatus(b.status.Load())
}

// SetStatus 设置玩家状态
func (b *BaseInfo) SetStatus(status PlayerStatus) {
	b.status.Store(int32(status))
}

// IsOnline 检查玩家是否在线
// 返回: true表示在线
func (b *BaseInfo) IsOnline() bool {
	return b.GetStatus() == PlayerStatusOnline && b.session != nil
}

// IsBusy 检查玩家是否忙碌
// 返回: true表示忙碌
func (b *BaseInfo) IsBusy() bool {
	return b.GetStatus() == PlayerStatusBusy
}

// IsAFK 检查玩家是否挂机
// 返回: true表示挂机
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

// GetExp 获取玩家经验值
func (b *BaseInfo) GetExp() int64 {
	return b.exp.Load()
}

// SetExp 设置玩家经验值
func (b *BaseInfo) SetExp(exp int64) {
	b.exp.Store(exp)
}

// AddExp 增加玩家经验值
// 参数:
//   - exp: 增加的经验值
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
// 自动限制最小值为0
func (b *BaseInfo) SetGold(gold int64) {
	if gold < 0 {
		gold = 0
	}
	b.gold.Store(gold)
}

// AddGold 增加玩家金币
// 参数:
//   - gold: 增加的金币数量
func (b *BaseInfo) AddGold(gold int64) {
	currentGold := b.gold.Load()
	newGold := currentGold + gold
	b.gold.Store(newGold)

	zLog.Info("Player gained gold",
		zap.Int64("gold", gold),
		zap.Int64("totalGold", newGold))
}

// SubGold 减少玩家金币
// 参数:
//   - gold: 减少的金币数量
//
// 返回:
//   - bool: 扣除是否成功（金币不足时返回false）
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

// GetCreateTime 获取账号创建时间
// 返回Unix毫秒时间戳
func (b *BaseInfo) GetCreateTime() int64 {
	return b.createTime
}

// SendPacket 发送网络数据包
// 参数:
//   - packetId: 数据包ID
//   - data: 数据内容
//
// 返回:
//   - error: 发送错误
func (b *BaseInfo) SendPacket(packetId int32, data []byte) error {
	if b.session == nil {
		zLog.Warn("Player session is nil")
		return nil
	}
	return b.session.Send(packetId, data)
}

// SendText 发送文本消息
// 使用packetId=1001作为文本消息协议
// 参数:
//   - message: 消息内容
//
// 返回:
//   - error: 发送错误
func (b *BaseInfo) SendText(message string) error {
	return b.SendPacket(1001, []byte(message))
}
