package player

import (
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
	playerId int64
	session  *zNet.TcpServerSession
}

// NewPlayer 创建新玩家对象
func NewPlayer(playerId int64, name string, session *zNet.TcpServerSession) *Player {
	livingObj := object.NewLivingObject(uint64(playerId), name)
	livingObj.SetType(common.GameObjectTypePlayer)

	player := &Player{
		LivingObject: livingObj,
		playerId:     playerId,
		session:      session,
	}

	player.initComponents(name)

	return player
}

// initComponents 初始化玩家组件
func (p *Player) initComponents(name string) {
	baseInfo := NewBaseInfo(name, p.session)
	p.AddComponent(baseInfo)

	inventory := NewInventory(p.GetPlayerId())
	p.AddComponent(inventory)

	equipment := NewEquipment(p.GetPlayerId())
	p.AddComponent(equipment)

	mailbox := NewMailbox(p.GetPlayerId())
	p.AddComponent(mailbox)

	taskManager := NewTaskManager(p.GetPlayerId())
	p.AddComponent(taskManager)

	skillManager := NewSkillManager(p.GetPlayerId())
	p.AddComponent(skillManager)
}

// Update 更新玩家状态
func (p *Player) Update(deltaTime float64) {
	p.LivingObject.Update(deltaTime)
}

// GetPlayerId 获取玩家ID
func (p *Player) GetPlayerId() int64 {
	return p.playerId
}

// SetPlayerId 设置玩家ID
func (p *Player) SetPlayerId(playerId int64) {
	p.playerId = playerId
}

// GetName 获取玩家名称
func (p *Player) GetName() string {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return ""
	}
	return baseInfo.(*BaseInfo).GetName()
}

// SetName 设置玩家名称
func (p *Player) SetName(name string) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetName(name)
	}
}

// GetSession 获取会话
func (p *Player) GetSession() *zNet.TcpServerSession {
	return p.session
}

// SetSession 设置会话
func (p *Player) SetSession(session *zNet.TcpServerSession) {
	p.session = session
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetSession(session)
	}
}

// GetStatus 获取玩家状态
func (p *Player) GetStatus() PlayerStatus {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return PlayerStatusOffline
	}
	return baseInfo.(*BaseInfo).GetStatus()
}

// SetStatus 设置玩家状态
func (p *Player) SetStatus(status PlayerStatus) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetStatus(status)
	}
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
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return 1
	}
	return baseInfo.(*BaseInfo).GetLevel()
}

// SetLevel 设置玩家等级
func (p *Player) SetLevel(level int) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetLevel(level)
	}
}

// GetExp 获取玩家经验
func (p *Player) GetExp() int64 {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return 0
	}
	return baseInfo.(*BaseInfo).GetExp()
}

// SetExp 设置玩家经验
func (p *Player) SetExp(exp int64) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetExp(exp)
		p.checkLevelUp()
	}
}

// AddExp 增加玩家经验
func (p *Player) AddExp(exp int64) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).AddExp(exp)
		p.checkLevelUp()

		p.PublishEvent(zEvent.NewEvent(1, p, map[string]interface{}{
			"playerId": p.GetPlayerId(),
			"exp":      exp,
		}))
	}
}

// GetGold 获取玩家金币
func (p *Player) GetGold() int64 {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return 0
	}
	return baseInfo.(*BaseInfo).GetGold()
}

// SetGold 设置玩家金币
func (p *Player) SetGold(gold int64) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetGold(gold)
	}
}

// AddGold 增加玩家金币
func (p *Player) AddGold(gold int64) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).AddGold(gold)

		p.PublishEvent(zEvent.NewEvent(2, p, map[string]interface{}{
			"playerId": p.GetPlayerId(),
			"oldGold":  baseInfo.(*BaseInfo).GetGold() - gold,
			"newGold":  baseInfo.(*BaseInfo).GetGold(),
		}))
	}
}

// SubGold 减少玩家金币
func (p *Player) SubGold(gold int64) bool {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return false
	}
	return baseInfo.(*BaseInfo).SubGold(gold)
}

// GetVIPLevel 获取VIP等级
func (p *Player) GetVIPLevel() int {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return 0
	}
	return baseInfo.(*BaseInfo).GetVIPLevel()
}

// SetVIPLevel 设置VIP等级
func (p *Player) SetVIPLevel(vipLevel int) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetVIPLevel(vipLevel)
	}
}

// GetInventory 获取背包
func (p *Player) GetInventory() *Inventory {
	inventory := p.GetComponent("inventory")
	if inventory == nil {
		return nil
	}
	return inventory.(*Inventory)
}

// GetEquipment 获取装备
func (p *Player) GetEquipment() *Equipment {
	equipment := p.GetComponent("equipment")
	if equipment == nil {
		return nil
	}
	return equipment.(*Equipment)
}

// GetMailbox 获取邮箱
func (p *Player) GetMailbox() *Mailbox {
	mailbox := p.GetComponent("mailbox")
	if mailbox == nil {
		return nil
	}
	return mailbox.(*Mailbox)
}

// GetTaskManager 获取任务管理器
func (p *Player) GetTaskManager() *TaskManager {
	taskManager := p.GetComponent("tasks")
	if taskManager == nil {
		return nil
	}
	return taskManager.(*TaskManager)
}

// GetSkillManager 获取技能管理器
func (p *Player) GetSkillManager() *SkillManager {
	skillManager := p.GetComponent("skills")
	if skillManager == nil {
		return nil
	}
	return skillManager.(*SkillManager)
}

// GetBaseInfo 获取基础信息组件
func (p *Player) GetBaseInfo() *BaseInfo {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return nil
	}
	return baseInfo.(*BaseInfo)
}

// GetCreateTime 获取创建时间
func (p *Player) GetCreateTime() int64 {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return 0
	}
	return baseInfo.(*BaseInfo).GetCreateTime()
}

// SendPacket 发送数据包
func (p *Player) SendPacket(packetId int32, data []byte) error {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return nil
	}
	return baseInfo.(*BaseInfo).SendPacket(packetId, data)
}

// SendText 发送文本消息
func (p *Player) SendText(message string) error {
	return p.SendText(message)
}

// Login 玩家登录
func (p *Player) Login() {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil || baseInfo.(*BaseInfo).GetSession() == nil {
		zLog.Error("Login failed: no session", zap.Int64("playerId", p.GetPlayerId()))
		return
	}

	baseInfo.(*BaseInfo).SetStatus(PlayerStatusOnline)
	p.SetActive(true)

	zLog.Info("Player logged in",
		zap.Int64("playerId", p.GetPlayerId()),
		zap.String("name", baseInfo.(*BaseInfo).GetName()))

	p.PublishEvent(zEvent.NewEvent(0, p, map[string]interface{}{
		"playerId": p.GetPlayerId(),
		"name":     baseInfo.(*BaseInfo).GetName(),
		"level":    p.GetLevel(),
	}))
}

// Logout 玩家登出
func (p *Player) Logout() {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetStatus(PlayerStatusOffline)

		if session := baseInfo.(*BaseInfo).GetSession(); session != nil {
			session.Close()
			baseInfo.(*BaseInfo).SetSession(nil)
		}

		zLog.Info("Player logged out", zap.Int64("playerId", p.GetPlayerId()), zap.String("name", baseInfo.(*BaseInfo).GetName()))

		p.PublishEvent(zEvent.NewEvent(3, p, map[string]interface{}{
			"playerId": p.GetPlayerId(),
		}))
	}
}

// OnDisconnect 玩家断开连接
func (p *Player) OnDisconnect() {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetStatus(PlayerStatusOffline)

		if session := baseInfo.(*BaseInfo).GetSession(); session != nil {
			session.Close()
			baseInfo.(*BaseInfo).SetSession(nil)
		}

		zLog.Info("Player disconnected", zap.Int64("playerId", p.GetPlayerId()))

		p.PublishEvent(zEvent.NewEvent(4, p, map[string]interface{}{
			"playerId": p.GetPlayerId(),
		}))
	}
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
		"playerId": p.GetPlayerId(),
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
		"playerId": p.GetPlayerId(),
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
		"playerId": p.GetPlayerId(),
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
			"playerId": p.GetPlayerId(),
			"amount":   amount,
		}))
	}
}

// OnDie 玩家死亡
func (p *Player) OnDie() {
	p.SetStatus(PlayerStatusOffline)
	p.SetActive(false)

	zLog.Info("Player died", zap.Int64("playerId", p.GetPlayerId()))

	p.PublishEvent(zEvent.NewEvent(9, p, map[string]interface{}{
		"playerId": p.GetPlayerId(),
	}))
}

// checkLevelUp 检查是否升级
func (p *Player) checkLevelUp() {
	level := p.GetLevel()
	exp := p.GetExp()

	requiredExp := int64(level * 1000)

	if exp >= requiredExp {
		baseInfo := p.GetComponent("baseinfo")
		if baseInfo != nil {
			currentLevel := baseInfo.(*BaseInfo).level.Load()
			baseInfo.(*BaseInfo).level.Store(currentLevel + 1)
			newLevel := p.GetLevel()

			currentExp := baseInfo.(*BaseInfo).exp.Load()
			newExp := currentExp - requiredExp
			baseInfo.(*BaseInfo).exp.Store(newExp)

			zLog.Info("Player leveled up",
				zap.Int64("playerId", p.GetPlayerId()),
				zap.Int("oldLevel", level),
				zap.Int("newLevel", newLevel))

			p.PublishEvent(zEvent.NewEvent(10, p, map[string]interface{}{
				"playerId": p.GetPlayerId(),
				"oldLevel": level,
				"newLevel": newLevel,
			}))

			p.onLevelUp(newLevel)
		}
	}
}

// onLevelUp 升级处理
func (p *Player) onLevelUp(newLevel int) {
	currentAttack := p.GetProperty("physical_attack")
	p.SetProperty("physical_attack", currentAttack+2)
	currentDefense := p.GetProperty("physical_defense")
	p.SetProperty("physical_defense", currentDefense+1)

	p.SetHealth(p.GetMaxHealth())
	p.SetMana(p.GetMaxMana())
}

// UseSkill 使用技能
func (p *Player) UseSkill(skillId int) error {
	skillManager := p.GetComponent("skills")
	if skillManager == nil {
		return nil
	}
	return skillManager.(*SkillManager).UseSkill(p.GetPlayerId(), int64(skillId))
}

// LearnSkill 学习技能
func (p *Player) LearnSkill(skillId int) error {
	skillManager := p.GetComponent("skills")
	if skillManager == nil {
		return nil
	}
	return skillManager.(*SkillManager).LearnSkill(int64(skillId))
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
