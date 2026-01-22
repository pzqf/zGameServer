package player

import (
	"sync/atomic"

	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/event"
	"github.com/pzqf/zUtil/zTime"

	"github.com/pzqf/zGameServer/game/object"
	"github.com/pzqf/zGameServer/game/object/component"
	"go.uber.org/zap"
)

// 玩家状态定义
const (
	PlayerStatusOnline  = 1
	PlayerStatusOffline = 2
	PlayerStatusBusy    = 3
)

// Player 玩家对象
type Player struct {
	object.LivingObject
	playerId int64
	name     string
	Session  *zNet.TcpServerSession
	status   int
}

// BasicInfo 玩家基础信息
type BasicInfo struct {
	*component.BaseComponent
	Level      int
	Exp        atomic.Int64
	Gold       atomic.Int64
	VipLevel   int
	ServerId   int
	CreateTime int64
}

// Destroy 销毁基础信息组件
func (bi *BasicInfo) Destroy() {
	// 清理基础信息资源
	// 这里不需要特别清理，因为没有需要释放的资源
}

func NewPlayer(playerId int64, name string, session *zNet.TcpServerSession) *Player {
	// 创建基础生命对象
	livingObj := object.NewLivingObject(uint64(playerId), name)

	// 创建玩家对象
	player := &Player{
		LivingObject: *livingObj,
		playerId:     playerId,
		name:         name,
		Session:      session,
		status:       PlayerStatusOnline,
	}

	// 创建并添加玩家系统组件
	player.addComponents()

	return player
}

// addComponents 添加玩家系统组件
func (p *Player) addComponents() {
	// 基础信息组件
	basicInfo := &BasicInfo{
		BaseComponent: component.NewBaseComponent("basicInfo"),
		Level:         1,
		VipLevel:      0,
		ServerId:      1,
		CreateTime:    zTime.Now().Time().UnixMilli(), // 设置为当前时间
	}
	// 初始化原子字段
	basicInfo.Exp.Store(0)
	basicInfo.Gold.Store(1000)
	p.AddComponent(basicInfo)

	// 背包组件
	inventory := NewInventory(p.playerId)
	inventory.Init()
	p.AddComponent(inventory)

	// 装备组件
	equipment := NewEquipment(p.playerId)
	equipment.Init()
	p.AddComponent(equipment)

	// 邮箱组件
	mailbox := NewMailbox(p.playerId)
	mailbox.Init()
	p.AddComponent(mailbox)

	// 任务组件
	tasks := NewTaskManager(p.playerId)
	tasks.Init()
	p.AddComponent(tasks)

	// 技能组件
	skills := NewSkillManager(p.playerId)
	skills.Init()
	p.AddComponent(skills)
}

// GetPlayerId 获取玩家ID
func (p *Player) GetPlayerId() int64 {
	return p.playerId
}

// GetName 获取玩家名称
func (p *Player) GetName() string {
	return p.name
}

// GetSession 获取玩家会话
func (p *Player) GetSession() *zNet.TcpServerSession {
	return p.Session
}

// GetStatus 获取玩家状态
func (p *Player) GetStatus() int {
	return p.status
}

// SetStatus 设置玩家状态
func (p *Player) SetStatus(status int) {
	p.status = status
}

// GetType 获取玩家类型
func (p *Player) GetType() int {
	return object.GameObjectTypePlayer
}

// GetBasicInfo 获取玩家基础信息
func (p *Player) GetBasicInfo() *BasicInfo {
	component := p.GetComponent("basicInfo")
	if component != nil {
		return component.(*BasicInfo)
	}
	return nil
}

// GetInventory 获取玩家背包
func (p *Player) GetInventory() *Inventory {
	component := p.GetComponent("inventory")
	if component != nil {
		return component.(*Inventory)
	}
	return nil
}

// GetEquipment 获取玩家装备
func (p *Player) GetEquipment() *Equipment {
	component := p.GetComponent("equipment")
	if component != nil {
		return component.(*Equipment)
	}
	return nil
}

// GetMailbox 获取玩家邮箱
func (p *Player) GetMailbox() *Mailbox {
	component := p.GetComponent("mailbox")
	if component != nil {
		return component.(*Mailbox)
	}
	return nil
}

// GetTasks 获取玩家任务
func (p *Player) GetTasks() *TaskManager {
	component := p.GetComponent("tasks")
	if component != nil {
		return component.(*TaskManager)
	}
	return nil
}

// GetSkills 获取玩家技能
func (p *Player) GetSkills() *SkillManager {
	component := p.GetComponent("skills")
	if component != nil {
		return component.(*SkillManager)
	}
	return nil
}

// OnConnect 玩家连接成功
func (p *Player) OnConnect() {
	p.status = PlayerStatusOnline
	zLog.Info("Player connected", zap.Int64("playerId", p.playerId), zap.String("name", p.name))
}

// OnDisconnect 玩家断开连接
func (p *Player) OnDisconnect() {
	p.status = PlayerStatusOffline
	zLog.Info("Player disconnected", zap.Int64("playerId", p.playerId), zap.String("name", p.name))
}

// SendPacket 给玩家发送数据包
func (p *Player) SendPacket(protoId int32, data []byte) error {
	if p.Session == nil {
		return nil
	}
	return p.Session.Send(protoId, data)
}

// publishEvent 发布玩家相关事件
func (p *Player) publishEvent(eventType zEvent.EventType, data interface{}) {
	eventObj := event.NewEvent(eventType, p, data)
	// 直接调用event包的GetGlobalEventBus函数获取全局事件总线实例
	event.GetGlobalEventBus().Publish(eventObj)
}

// AddExp 增加玩家经验并发布事件
func (p *Player) AddExp(exp int64) {
	basicInfo := p.GetBasicInfo()
	if basicInfo == nil {
		return
	}

	basicInfo.Exp.Add(exp)

	// 发布经验增加事件
	eventData := &event.PlayerExpEventData{
		PlayerID: p.playerId,
		Exp:      exp,
	}
	p.publishEvent(event.EventPlayerExpAdd, eventData)

	// 检查是否升级
	// TODO: 实现升级逻辑
	// if newLevel > oldLevel {
	// 	levelUpData := &event.PlayerLevelUpEventData{
	// 		PlayerID: p.playerId,
	// 		OldLevel: oldLevel,
	// 		NewLevel: newLevel,
	// 	}
	// 	p.publishEvent(event.EventPlayerLevelUp, levelUpData)
	// }
}

// AddGold 增加玩家金币并发布事件
func (p *Player) AddGold(gold int64) {
	basicInfo := p.GetBasicInfo()
	if basicInfo == nil {
		return
	}

	basicInfo.Gold.Add(gold)

	// 发布金币增加事件
	eventData := &event.PlayerGoldEventData{
		PlayerID: p.playerId,
		Gold:     gold,
	}
	p.publishEvent(event.EventPlayerGoldAdd, eventData)
}

// SubGold 减少玩家金币并发布事件
func (p *Player) SubGold(gold int64) bool {
	basicInfo := p.GetBasicInfo()
	if basicInfo == nil {
		return false
	}

	for {
		currentGold := basicInfo.Gold.Load()
		if currentGold < gold {
			return false // 金币不足
		}
		if basicInfo.Gold.CompareAndSwap(currentGold, currentGold-gold) {
			// 发布金币减少事件
			eventData := &event.PlayerGoldEventData{
				PlayerID: p.playerId,
				Gold:     -gold,
			}
			p.publishEvent(event.EventPlayerGoldSub, eventData)
			return true
		}
	}
}
