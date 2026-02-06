package player

import (
	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object"
	"go.uber.org/zap"
)

// PlayerStatus 玩家状态枚举
// 用于标识玩家的当前在线/忙碌状态
type PlayerStatus int

// 玩家状态常量
const (
	PlayerStatusOffline PlayerStatus = iota // 离线
	PlayerStatusOnline                      // 在线
	PlayerStatusBusy                        // 忙碌
	PlayerStatusAFK                         // 挂机/离开
)

// Player 玩家对象类
// 继承自 LivingObject，是游戏中玩家的核心实体
// 管理玩家的所有组件：基础信息、背包、装备、邮箱、任务、技能等
type Player struct {
	*object.LivingObject                        // 继承活体对象（包含生命、魔法、属性等）
	playerId             common.PlayerIdType    // 玩家唯一ID
	session              *zNet.TcpServerSession // 网络会话（用于与客户端通信）
}

// NewPlayer 创建新玩家对象
// 参数:
//   - playerId: 玩家ID
//   - name: 玩家名称
//   - session: 网络会话
//
// 返回:
//   - *Player: 新创建的玩家对象
func NewPlayer(playerId common.PlayerIdType, name string, session *zNet.TcpServerSession) *Player {
	livingObj := object.NewLivingObject(common.ObjectIdType(playerId), name)
	livingObj.SetType(gamecommon.GameObjectTypePlayer)

	player := &Player{
		LivingObject: livingObj,
		playerId:     playerId,
		session:      session,
	}

	player.initComponents(name)

	return player
}

// initComponents 初始化玩家组件
// 创建并添加所有必要的组件到玩家对象
func (p *Player) initComponents(name string) {
	// 基础信息组件（名称、等级、金币等）
	baseInfo := NewBaseInfo(name, p.session)
	p.AddComponent(baseInfo)

	// 背包组件（物品管理）
	inventory := NewInventory(p.GetPlayerId())
	p.AddComponent(inventory)

	// 装备组件（装备管理）
	equipment := NewEquipment(p.GetPlayerId())
	p.AddComponent(equipment)

	// 邮箱组件（邮件管理）
	mailbox := NewMailbox(p.GetPlayerId())
	p.AddComponent(mailbox)

	// 任务管理器组件（任务进度）
	taskManager := NewTaskManager(p.GetPlayerId())
	p.AddComponent(taskManager)

	// 技能管理器组件（技能学习和使用）
	skillManager := NewSkillManager(p.GetPlayerId())
	p.AddComponent(skillManager)
}

// Update 更新玩家状态
// 每帧调用，更新玩家及其组件的状态
func (p *Player) Update(deltaTime float64) {
	p.LivingObject.Update(deltaTime)
}

// GetPlayerId 获取玩家ID
func (p *Player) GetPlayerId() common.PlayerIdType {
	return p.playerId
}

// SetPlayerId 设置玩家ID
func (p *Player) SetPlayerId(playerId common.PlayerIdType) {
	p.playerId = playerId
}

// GetName 获取玩家名称
// 从baseinfo组件获取
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

// GetSession 获取网络会话
func (p *Player) GetSession() *zNet.TcpServerSession {
	return p.session
}

// SetSession 设置网络会话
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

// IsBusy 检查玩家是否忙碌
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

// GetExp 获取玩家经验值
func (p *Player) GetExp() int64 {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return 0
	}
	return baseInfo.(*BaseInfo).GetExp()
}

// SetExp 设置玩家经验值
// 自动检查是否升级
func (p *Player) SetExp(exp int64) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetExp(exp)
		p.checkLevelUp()
	}
}

// AddExp 增加玩家经验值
// 自动检查是否升级并发布事件
func (p *Player) AddExp(exp int64) {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).AddExp(exp)
		p.checkLevelUp()

		// 发布经验增加事件（事件ID=1）
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

		// 发布金币变化事件（事件ID=2）
		p.PublishEvent(zEvent.NewEvent(2, p, map[string]interface{}{
			"playerId": p.GetPlayerId(),
			"oldGold":  baseInfo.(*BaseInfo).GetGold() - gold,
			"newGold":  baseInfo.(*BaseInfo).GetGold(),
		}))
	}
}

// SubGold 减少玩家金币
// 返回: true表示扣除成功，false表示金币不足
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

// GetInventory 获取背包组件
func (p *Player) GetInventory() *Inventory {
	inventory := p.GetComponent("inventory")
	if inventory == nil {
		return nil
	}
	return inventory.(*Inventory)
}

// GetEquipment 获取装备组件
func (p *Player) GetEquipment() *Equipment {
	equipment := p.GetComponent("equipment")
	if equipment == nil {
		return nil
	}
	return equipment.(*Equipment)
}

// GetMailbox 获取邮箱组件
func (p *Player) GetMailbox() *Mailbox {
	mailbox := p.GetComponent("mailbox")
	if mailbox == nil {
		return nil
	}
	return mailbox.(*Mailbox)
}

// GetTaskManager 获取任务管理器组件
func (p *Player) GetTaskManager() *TaskManager {
	taskManager := p.GetComponent("tasks")
	if taskManager == nil {
		return nil
	}
	return taskManager.(*TaskManager)
}

// GetSkillManager 获取技能管理器组件
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

// GetCreateTime 获取账号创建时间
func (p *Player) GetCreateTime() int64 {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return 0
	}
	return baseInfo.(*BaseInfo).GetCreateTime()
}

// SendPacket 发送网络数据包
func (p *Player) SendPacket(packetId int32, data []byte) error {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil {
		return nil
	}
	return baseInfo.(*BaseInfo).SendPacket(packetId, data)
}

// SendText 发送文本消息给客户端
func (p *Player) SendText(message string) error {
	return p.SendText(message)
}

// Login 玩家登录处理
// 设置状态为在线，发布登录事件
func (p *Player) Login() {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo == nil || baseInfo.(*BaseInfo).GetSession() == nil {
		zLog.Error("Login failed: no session", zap.Int64("playerId", int64(p.GetPlayerId())))
		return
	}

	baseInfo.(*BaseInfo).SetStatus(PlayerStatusOnline)
	p.SetActive(true)

	zLog.Info("Player logged in",
		zap.Int64("playerId", int64(p.GetPlayerId())),
		zap.String("name", baseInfo.(*BaseInfo).GetName()))

	// 发布登录事件（事件ID=0）
	p.PublishEvent(zEvent.NewEvent(0, p, map[string]interface{}{
		"playerId": p.GetPlayerId(),
		"name":     baseInfo.(*BaseInfo).GetName(),
		"level":    p.GetLevel(),
	}))
}

// Logout 玩家登出处理
// 设置状态为离线，关闭会话
func (p *Player) Logout() {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetStatus(PlayerStatusOffline)

		// 关闭网络会话
		if session := baseInfo.(*BaseInfo).GetSession(); session != nil {
			session.Close()
			baseInfo.(*BaseInfo).SetSession(nil)
		}

		zLog.Info("Player logged out", zap.Int64("playerId", int64(p.GetPlayerId())), zap.String("name", baseInfo.(*BaseInfo).GetName()))

		// 发布登出事件（事件ID=3）
		p.PublishEvent(zEvent.NewEvent(3, p, map[string]interface{}{
			"playerId": p.GetPlayerId(),
		}))
	}
}

// OnDisconnect 玩家断开连接处理
// 与Logout类似，但不主动关闭会话
func (p *Player) OnDisconnect() {
	baseInfo := p.GetComponent("baseinfo")
	if baseInfo != nil {
		baseInfo.(*BaseInfo).SetStatus(PlayerStatusOffline)

		if session := baseInfo.(*BaseInfo).GetSession(); session != nil {
			session.Close()
			baseInfo.(*BaseInfo).SetSession(nil)
		}

		zLog.Info("Player disconnected", zap.Int64("playerId", int64(p.GetPlayerId())))

		// 发布断开连接事件（事件ID=4）
		p.PublishEvent(zEvent.NewEvent(4, p, map[string]interface{}{
			"playerId": p.GetPlayerId(),
		}))
	}
}

// Attack 玩家攻击目标
// 参数:
//   - target: 攻击目标
func (p *Player) Attack(target gamecommon.IGameObject) {
	if !p.IsOnline() {
		return
	}

	// 对目标造成伤害
	livingTarget, ok := target.(*object.LivingObject)
	if ok {
		livingTarget.TakeDamage(10, p)
	}

	// 发布攻击事件（事件ID=5）
	p.PublishEvent(zEvent.NewEvent(5, p, map[string]interface{}{
		"playerId": p.GetPlayerId(),
		"targetId": target.GetID(),
	}))
}

// MoveTo 移动到目标位置
// 参数:
//   - targetPos: 目标坐标
func (p *Player) MoveTo(targetPos gamecommon.Vector3) {
	if !p.IsOnline() {
		return
	}

	p.SetStatus(PlayerStatusBusy)
	p.SetPosition(targetPos)
	p.SetStatus(PlayerStatusOnline)

	// 发布移动事件（事件ID=6）
	p.PublishEvent(zEvent.NewEvent(6, p, map[string]interface{}{
		"playerId": p.GetPlayerId(),
		"position": targetPos,
	}))
}

// TakeDamage 受到伤害
// 参数:
//   - damage: 伤害值
//   - attacker: 攻击者
func (p *Player) TakeDamage(damage float64, attacker gamecommon.IGameObject) {
	currentHealth := p.GetHealth()
	newHealth := currentHealth - damage
	if newHealth < 0 {
		newHealth = 0
	}

	p.SetHealth(newHealth)

	// 生命值为0时触发死亡
	if newHealth <= 0 {
		p.OnDie()
	}

	// 发布受伤事件（事件ID=7）
	p.PublishEvent(zEvent.NewEvent(7, p, map[string]interface{}{
		"playerId": p.GetPlayerId(),
		"damage":   damage,
		"health":   p.GetHealth(),
	}))
}

// OnHeal 接受治疗
// 参数:
//   - healer: 治疗者
//   - amount: 治疗量
func (p *Player) OnHeal(healer gamecommon.IGameObject, amount float64) {
	currentHealth := p.GetHealth()
	maxHealth := p.GetMaxHealth()
	newHealth := currentHealth + amount
	if newHealth > maxHealth {
		newHealth = maxHealth
	}

	p.SetHealth(newHealth)

	// 发布治疗事件（事件ID=8）
	if healer != nil {
		p.PublishEvent(zEvent.NewEvent(8, p, map[string]interface{}{
			"playerId": p.GetPlayerId(),
			"amount":   amount,
		}))
	}
}

// OnDie 玩家死亡处理
// 设置状态为离线，发布死亡事件
func (p *Player) OnDie() {
	p.SetStatus(PlayerStatusOffline)
	p.SetActive(false)

	zLog.Info("Player died", zap.Int64("playerId", int64(p.GetPlayerId())))

	// 发布死亡事件（事件ID=9）
	p.PublishEvent(zEvent.NewEvent(9, p, map[string]interface{}{
		"playerId": p.GetPlayerId(),
	}))
}

// checkLevelUp 检查是否满足升级条件
// 升级公式: 所需经验 = 等级 * 1000
func (p *Player) checkLevelUp() {
	level := p.GetLevel()
	exp := p.GetExp()

	requiredExp := int64(level * 1000)

	if exp >= requiredExp {
		baseInfo := p.GetComponent("baseinfo")
		if baseInfo != nil {
			// 增加等级
			currentLevel := baseInfo.(*BaseInfo).level.Load()
			baseInfo.(*BaseInfo).level.Store(currentLevel + 1)
			newLevel := p.GetLevel()

			// 扣除经验
			currentExp := baseInfo.(*BaseInfo).exp.Load()
			newExp := currentExp - requiredExp
			baseInfo.(*BaseInfo).exp.Store(newExp)

			zLog.Info("Player leveled up",
				zap.Int64("playerId", int64(p.GetPlayerId())),
				zap.Int("oldLevel", level),
				zap.Int("newLevel", newLevel))

			// 发布升级事件（事件ID=10）
			p.PublishEvent(zEvent.NewEvent(10, p, map[string]interface{}{
				"playerId": p.GetPlayerId(),
				"oldLevel": level,
				"newLevel": newLevel,
			}))

			p.onLevelUp(newLevel)
		}
	}
}

// onLevelUp 升级后处理
// 增加属性，恢复生命和魔法
func (p *Player) onLevelUp(newLevel int) {
	// 增加攻击力+2
	currentAttack := p.GetProperty("physical_attack")
	p.SetProperty("physical_attack", currentAttack+2)
	// 增加防御力+1
	currentDefense := p.GetProperty("physical_defense")
	p.SetProperty("physical_defense", currentDefense+1)

	// 恢复满生命和魔法
	p.SetHealth(p.GetMaxHealth())
	p.SetMana(p.GetMaxMana())
}

// UseSkill 使用技能
// 参数:
//   - skillId: 技能ID
//
// 返回:
//   - error: 使用失败返回错误
func (p *Player) UseSkill(skillId int) error {
	skillManager := p.GetComponent("skills")
	if skillManager == nil {
		return nil
	}
	return skillManager.(*SkillManager).UseSkill(int64(p.GetPlayerId()), int64(skillId))
}

// LearnSkill 学习技能
// 参数:
//   - skillId: 技能ID
//
// 返回:
//   - error: 学习失败返回错误
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

// PublishEvent 发布事件到事件总线
func (p *Player) PublishEvent(event *zEvent.Event) {
	if eventEmitter := p.GetEventEmitter(); eventEmitter != nil {
		eventEmitter.Publish(event)
	}
}

// GetTarget 获取当前攻击目标
func (p *Player) GetTarget() gamecommon.IGameObject {
	return nil
}

// SetTarget 设置当前攻击目标
func (p *Player) SetTarget(target gamecommon.IGameObject) {
}
