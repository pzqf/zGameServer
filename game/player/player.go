package player

import (
	"sync/atomic"
	"time"

	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object"
	"go.uber.org/zap"
)

// PlayerStatus 玩家状态枚举
type PlayerStatus int

// 玩家状态常量
const (
	PlayerStatusOffline PlayerStatus = iota
	PlayerStatusOnline
	PlayerStatusBusy
	PlayerStatusAFK
)

// Player 玩家对象类 (继承 LivingObject)
type Player struct {
	*object.LivingObject
	playerId     int64
	name         string
	session      *zNet.TcpServerSession
	status       atomic.Int32
	inventory    *Inventory
	equipment    *Equipment
	mailbox      *Mailbox
	taskManager  *TaskManager
	skillManager *SkillManager
	exp          atomic.Int64
	gold         atomic.Int64
	level        atomic.Int32
	vipLevel     atomic.Int32
	serverId     int
	createTime   int64
}

// NewPlayer 创建新玩家对象
func NewPlayer(playerId int64, name string, session *zNet.TcpServerSession) *Player {
	livingObj := object.NewLivingObject(uint64(playerId), name)
	livingObj.SetType(common.GameObjectTypePlayer)

	player := &Player{
		LivingObject: livingObj,
		playerId:     playerId,
		name:         name,
		session:      session,
		serverId:     1,
		createTime:   time.Now().UnixMilli(),
	}

	player.status.Store(int32(PlayerStatusOnline))
	player.exp.Store(0)
	player.gold.Store(1000)
	player.level.Store(1)
	player.vipLevel.Store(0)

	player.initComponents()

	return player
}

// initComponents 初始化玩家组件
func (p *Player) initComponents() {
	// 背包组件
	p.inventory = NewInventory(p.playerId)
	p.AddComponent(p.inventory)

	// 装备组件
	p.equipment = NewEquipment(p.playerId)
	p.AddComponent(p.equipment)

	// 邮箱组件
	p.mailbox = NewMailbox(p.playerId)
	p.AddComponent(p.mailbox)

	// 任务组件
	p.taskManager = NewTaskManager(p.playerId)
	p.AddComponent(p.taskManager)

	// 技能组件
	p.skillManager = NewSkillManager(p.playerId)
	p.AddComponent(p.skillManager)
}

// GetPlayerId 获取玩家ID
func (p *Player) GetPlayerId() int64 {
	return p.playerId
}

// GetName 获取玩家名称
func (p *Player) GetName() string {
	return p.name
}

// SetName 设置玩家名称
func (p *Player) SetName(name string) {
	p.name = name
}

// GetSession 获取会话
func (p *Player) GetSession() *zNet.TcpServerSession {
	return p.session
}

// SetSession 设置会话
func (p *Player) SetSession(session *zNet.TcpServerSession) {
	p.session = session
}

// GetStatus 获取玩家状态
func (p *Player) GetStatus() PlayerStatus {
	return PlayerStatus(p.status.Load())
}

func (p *Player) SetStatus(status PlayerStatus) {
	p.status.Store(int32(status))
}

// IsOnline 检查玩家是否在线
func (p *Player) IsOnline() bool {
	return p.GetStatus() == PlayerStatusOnline && p.session != nil
}

// IsBusy 检查玩家是否忙
func (p *Player) IsBusy() bool {
	return p.GetStatus() == PlayerStatusBusy
}

// IsAFK 检查玩家是否挂机
func (p *Player) IsAFK() bool {
	return p.GetStatus() == PlayerStatusAFK
}

// GetLevel 获取玩家等级
func (p *Player) GetLevel() int {
	return int(p.level.Load())
}

// SetLevel 设置玩家等级
func (p *Player) SetLevel(level int) {
	p.level.Store(int32(level))
}

// GetExp 获取玩家经验
func (p *Player) GetExp() int64 {
	return p.exp.Load()
}

// SetExp 设置玩家经验
func (p *Player) SetExp(exp int64) {
	p.exp.Store(exp)
	p.checkLevelUp()
}

// AddExp 增加玩家经验
func (p *Player) AddExp(exp int64) {
	currentExp := p.exp.Load()
	newExp := currentExp + exp
	p.exp.Store(newExp)
	p.checkLevelUp()

	zLog.Info("Player gained exp",
		zap.Int64("playerId", p.playerId),
		zap.Int64("exp", exp),
		zap.Int64("totalExp", newExp))

	p.PublishEvent(zEvent.NewEvent(1, p, map[string]interface{}{
		"playerId": p.playerId,
		"exp":      exp,
	}))
}

// GetGold 获取玩家金币
func (p *Player) GetGold() int64 {
	return p.gold.Load()
}

// SetGold 设置玩家金币
func (p *Player) SetGold(gold int64) {
	if gold < 0 {
		gold = 0
	}
	p.gold.Store(gold)
}

// AddGold 增加玩家金币
func (p *Player) AddGold(gold int64) {
	currentGold := p.gold.Load()
	newGold := currentGold + gold
	p.gold.Store(newGold)

	zLog.Info("Player gained gold",
		zap.Int64("playerId", p.playerId),
		zap.Int64("gold", gold),
		zap.Int64("totalGold", newGold))

	p.PublishEvent(zEvent.NewEvent(2, p, map[string]interface{}{
		"playerId": p.playerId,
		"oldGold":  currentGold,
		"newGold":  newGold,
	}))
}

// SubGold 减少玩家金币
func (p *Player) SubGold(gold int64) bool {
	currentGold := p.gold.Load()
	if currentGold < gold {
		return false
	}

	newGold := currentGold - gold
	p.gold.Store(newGold)

	return true
}

// GetInventory 获取背包
func (p *Player) GetInventory() *Inventory {
	return p.inventory
}

// GetEquipment 获取装备
func (p *Player) GetEquipment() *Equipment {
	return p.equipment
}

// GetMailbox 获取邮箱
func (p *Player) GetMailbox() *Mailbox {
	return p.mailbox
}

// GetTaskManager 获取任务管理器
func (p *Player) GetTaskManager() *TaskManager {
	return p.taskManager
}

// GetSkillManager 获取技能管理器
func (p *Player) GetSkillManager() *SkillManager {
	return p.skillManager
}

// SendPacket 发送数据包
func (p *Player) SendPacket(packetId int32, data []byte) error {
	if p.session == nil {
		zLog.Warn("Player session is nil", zap.Int64("playerId", p.playerId))
		return nil
	}
	return p.session.Send(packetId, data)
}

// SendText 发送文本消息
func (p *Player) SendText(message string) error {
	return p.SendPacket(1001, []byte(message))
}

// Login 玩家登录
func (p *Player) Login() {
	if p.session == nil {
		zLog.Error("Login failed: no session", zap.Int64("playerId", p.playerId))
		return
	}

	p.SetStatus(PlayerStatusOnline)
	p.SetActive(true)

	zLog.Info("Player logged in",
		zap.Int64("playerId", p.playerId),
		zap.String("name", p.name))

	p.PublishEvent(zEvent.NewEvent(0, p, map[string]interface{}{
		"playerId": p.playerId,
		"name":     p.name,
		"level":    p.GetLevel(),
	}))
}

// Logout 玩家登出
func (p *Player) Logout() {
	p.SetStatus(PlayerStatusOffline)

	if p.session != nil {
		p.session.Close()
		p.session = nil
	}

	zLog.Info("Player logged out", zap.Int64("playerId", p.playerId), zap.String("name", p.name))

	p.PublishEvent(zEvent.NewEvent(3, p, map[string]interface{}{
		"playerId": p.playerId,
	}))
}

// OnDisconnect 玩家断开连接
func (p *Player) OnDisconnect() {
	p.SetStatus(PlayerStatusOffline)

	if p.session != nil {
		p.session.Close()
		p.session = nil
	}

	zLog.Info("Player disconnected", zap.Int64("playerId", p.playerId))

	p.PublishEvent(zEvent.NewEvent(4, p, map[string]interface{}{
		"playerId": p.playerId,
	}))
}

// Attack 玩家攻击目标
func (p *Player) Attack(target common.IGameObject) {
	if !p.IsOnline() {
		return
	}

	livingTarget, ok := target.(*object.LivingObject)
	if ok {
		livingTarget.TakeDamage(10, p)
	}

	p.PublishEvent(zEvent.NewEvent(5, p, map[string]interface{}{
		"playerId": p.playerId,
		"targetId": target.GetID(),
	}))
}

// Move 移动到目标位置
func (p *Player) MoveTo(targetPos common.Vector3) {
	if !p.IsOnline() {
		return
	}

	p.SetStatus(PlayerStatusBusy)

	p.SetPosition(targetPos)

	p.SetStatus(PlayerStatusOnline)

	p.PublishEvent(zEvent.NewEvent(6, p, map[string]interface{}{
		"playerId": p.playerId,
		"position": targetPos,
	}))
}

// TakeDamage 受到伤害
func (p *Player) TakeDamage(damage float32, attacker common.IGameObject) {
	currentHealth := p.GetHealth()
	newHealth := currentHealth - damage
	if newHealth < 0 {
		newHealth = 0
	}

	p.SetHealth(newHealth)

	if newHealth <= 0 {
		p.OnDie()
	}

	p.PublishEvent(zEvent.NewEvent(7, p, map[string]interface{}{
		"playerId": p.playerId,
		"damage":   damage,
		"health":   p.GetHealth(),
	}))
}

// OnHeal 接受治疗
func (p *Player) OnHeal(healer common.IGameObject, amount float32) {
	currentHealth := p.GetHealth()
	maxHealth := p.GetMaxHealth()
	newHealth := currentHealth + amount
	if newHealth > maxHealth {
		newHealth = maxHealth
	}

	p.SetHealth(newHealth)

	if healer != nil {
		p.PublishEvent(zEvent.NewEvent(8, p, map[string]interface{}{
			"playerId": p.playerId,
			"amount":   amount,
		}))
	}
}

// OnDie 玩家死亡
func (p *Player) OnDie() {
	p.SetStatus(PlayerStatusOffline)
	p.SetActive(false)

	zLog.Info("Player died", zap.Int64("playerId", p.playerId))

	p.PublishEvent(zEvent.NewEvent(9, p, map[string]interface{}{
		"playerId": p.playerId,
	}))
}

// checkLevelUp 检查是否升级
func (p *Player) checkLevelUp() {
	level := p.GetLevel()
	exp := p.GetExp()

	requiredExp := int64(level * 1000)

	if exp >= requiredExp {
		p.level.Add(1)
		newLevel := p.GetLevel()

		currentExp := p.exp.Load()
		newExp := currentExp - requiredExp
		p.exp.Store(newExp)

		zLog.Info("Player leveled up",
			zap.Int64("playerId", p.playerId),
			zap.Int("oldLevel", level),
			zap.Int("newLevel", newLevel))

		p.PublishEvent(zEvent.NewEvent(10, p, map[string]interface{}{
			"playerId": p.playerId,
			"oldLevel": level,
			"newLevel": newLevel,
		}))

		p.onLevelUp(newLevel)
	}
}

// onLevelUp 升级处理
func (p *Player) onLevelUp(newLevel int) {
	// 增加属性
	currentAttack := p.GetProperty("physical_attack")
	p.SetProperty("physical_attack", currentAttack+2)
	currentDefense := p.GetProperty("physical_defense")
	p.SetProperty("physical_defense", currentDefense+1)

	// 恢复生命值和魔法值
	p.SetHealth(p.GetMaxHealth())
	p.SetMana(p.GetMaxMana())
}

// UseSkill 使用技能
func (p *Player) UseSkill(skillId int) error {
	return p.skillManager.UseSkill(p.playerId, int64(skillId))
}

// LearnSkill 学习技能
func (p *Player) LearnSkill(skillId int) error {
	return p.skillManager.LearnSkill(int64(skillId))
}

// GetSkillPoints 获取技能点数
func (p *Player) GetSkillPoints() int {
	return 0
}

// PublishEvent 发布事件
func (p *Player) PublishEvent(event *zEvent.Event) {
	if eventEmitter := p.GetEventEmitter(); eventEmitter != nil {
		eventEmitter.Publish(event)
	}
}

// GetTarget 获取当前目标
func (p *Player) GetTarget() common.IGameObject {
	return nil
}

// SetTarget 设置当前目标
func (p *Player) SetTarget(target common.IGameObject) {
}
