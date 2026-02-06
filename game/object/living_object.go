package object

import (
	"math/rand"
	"sync"
	"time"

	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/systems/buff"
	"github.com/pzqf/zGameServer/game/systems/combat"
	"github.com/pzqf/zGameServer/game/systems/movement"
	"github.com/pzqf/zGameServer/game/systems/property"
)

// 默认属性常量
const (
	DefaultHealth float64 = 100 // 默认生命值
	DefaultMana   float64 = 50  // 默认魔法值
)

// Property 属性结构
// 用于存储角色的各种属性（如攻击力、防御力等）
type Property struct {
	Name  string  // 属性名称
	Value float64 // 属性值
}

// LivingObject 活体对象
// 继承自GameObject，增加了生命、魔法、属性、战斗等活体特有的功能
// 是玩家、怪物、NPC等的基类
type LivingObject struct {
	*GameObject                                   // 继承基础游戏对象
	mu                sync.RWMutex                // 读写锁
	health            float64                     // 当前生命值
	maxHealth         float64                     // 最大生命值
	mana              float64                     // 当前魔法值
	maxMana           float64                     // 最大魔法值
	properties        map[string]float64          // 属性表（存储各种属性如攻击力、防御力等）
	inCombat          bool                        // 是否在战斗中
	targetID          common.ObjectIdType         // 目标对象ID
	lastAttack        time.Time                   // 上次攻击时间
	onDeath           func()                      // 死亡回调函数
	buffComponent     *buff.BuffComponent         // Buff组件
	combatComponent   *combat.CombatComponent     // 战斗组件
	movementComponent *movement.MovementComponent // 移动组件
}

// NewLivingObject 创建活体对象
// 参数:
//   - id: 对象ID
//   - name: 对象名称
//
// 返回:
//   - *LivingObject: 新创建的活体对象
func NewLivingObject(id common.ObjectIdType, name string) *LivingObject {
	goObj := NewGameObject(id, name)
	goObj.SetType(gamecommon.GameObjectTypeLiving)

	livingObj := &LivingObject{
		GameObject: goObj,
		health:     DefaultHealth,
		maxHealth:  DefaultHealth,
		mana:       DefaultMana,
		maxMana:    DefaultMana,
		properties: make(map[string]float64),
		inCombat:   false,
		targetID:   0,
	}

	livingObj.buffComponent = buff.NewBuffComponent(livingObj)
	livingObj.combatComponent = combat.NewCombatComponent(livingObj)
	livingObj.movementComponent = movement.NewMovementComponent(livingObj)

	return livingObj
}

// GetHealth 获取当前生命值
func (lo *LivingObject) GetHealth() float64 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.health
}

// SetHealth 设置当前生命值
// 自动限制不超过最大生命值
func (lo *LivingObject) SetHealth(health float64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	if health > lo.maxHealth {
		health = lo.maxHealth
	}
	lo.health = health
}

// GetMaxHealth 获取最大生命值
func (lo *LivingObject) GetMaxHealth() float64 {
	return lo.maxHealth
}

// SetMaxHealth 设置最大生命值
func (lo *LivingObject) SetMaxHealth(maxHealth float64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.maxHealth = maxHealth
}

// GetMana 获取当前魔法值
func (lo *LivingObject) GetMana() float64 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.mana
}

// SetMana 设置当前魔法值
// 自动限制不超过最大魔法值
func (lo *LivingObject) SetMana(mana float64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	if mana > lo.maxMana {
		mana = lo.maxMana
	}
	lo.mana = mana
}

// GetMaxMana 获取最大魔法值
func (lo *LivingObject) GetMaxMana() float64 {
	return lo.maxMana
}

// SetMaxMana 设置最大魔法值
func (lo *LivingObject) SetMaxMana(maxMana float64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.maxMana = maxMana
}

// GetProperty 获取属性值
// 参数:
//   - propType: 属性类型
//
// 返回:
//   - float64: 属性值
func (lo *LivingObject) GetProperty(propType property.PropertyType) float64 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.properties[string(propType)]
}

// GetPropertyByName 通过属性名称获取属性值
// 参数:
//   - name: 属性名称
//
// 返回:
//   - float64: 属性值
func (lo *LivingObject) GetPropertyByName(name string) float64 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.properties[name]
}

// SetProperty 设置属性值
// 参数:
//   - propType: 属性类型
//   - value: 属性值
func (lo *LivingObject) SetProperty(propType property.PropertyType, value float64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.properties[string(propType)] = value
}

// SetPropertyByName 通过属性名称设置属性值
// 参数:
//   - name: 属性名称
//   - value: 属性值
func (lo *LivingObject) SetPropertyByName(name string, value float64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.properties[name] = value
}

// GetAllProperties 获取所有属性
// 返回属性的副本（避免并发问题）
func (lo *LivingObject) GetAllProperties() map[string]float64 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	props := make(map[string]float64)
	for k, v := range lo.properties {
		props[k] = v
	}
	return props
}

// SetProperties 批量设置属性
func (lo *LivingObject) SetProperties(props map[string]float64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.properties = props
}

// RemoveProperty 移除属性
func (lo *LivingObject) RemoveProperty(name string) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	delete(lo.properties, name)
}

// ClearAllProperties 清空所有属性
func (lo *LivingObject) ClearAllProperties() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.properties = make(map[string]float64)
}

// ResetAllProperties 重置所有属性
func (lo *LivingObject) ResetAllProperties() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	for k := range lo.properties {
		delete(lo.properties, k)
	}
}

// TakeDamage 受到伤害
// 计算闪避、暴击、防御后扣除生命值
// 参数:
//   - damage: 伤害值
//   - attacker: 攻击者对象（可为nil）
func (lo *LivingObject) TakeDamage(damage float64, attacker gamecommon.IGameObject) {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if lo.health <= 0 {
		return
	}

	if attacker != nil {
		// 检查闪避
		if lo.isDodge() {
			return
		}

		// 检查攻击者暴击
		if livingAttacker, ok := attacker.(*LivingObject); ok {
			if livingAttacker.isCritical() {
				damage = damage * (1 + livingAttacker.GetProperty(property.PropertyCriticalDamage))
			}
		}
	}

	// 计算实际伤害 = 攻击 - 防御*0.5
	attack := damage
	defense := lo.GetProperty(property.PropertyPhysicalDefense)
	actualDamage := attack - defense*0.5

	if actualDamage < 1 {
		actualDamage = 1
	}

	lo.health -= actualDamage

	if lo.health <= 0 {
		lo.health = 0
		if lo.onDeath != nil {
			lo.onDeath()
		}
	}
}

// Heal 恢复生命值
// 参数:
//   - amount: 恢复量
func (lo *LivingObject) Heal(amount float64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if lo.health <= 0 {
		return
	}

	lo.health += amount
	if lo.health > lo.maxHealth {
		lo.health = lo.maxHealth
	}
}

// RegenMana 恢复魔法值
// 参数:
//   - amount: 恢复量
func (lo *LivingObject) RegenMana(amount float64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if lo.health <= 0 {
		return
	}

	lo.mana += amount
	if lo.mana > lo.maxMana {
		lo.mana = lo.maxMana
	}
}

// Revive 复活
// 恢复30%的生命和魔法，清除战斗状态
func (lo *LivingObject) Revive() {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if lo.health > 0 {
		return
	}

	lo.health = lo.maxHealth * 0.3
	lo.mana = lo.maxMana * 0.3
	lo.inCombat = false
	lo.targetID = 0
}

// Teleport 传送
func (lo *LivingObject) Teleport(targetPos gamecommon.Vector3) error {
	return lo.GameObject.Teleport(targetPos)
}

// SetTarget 设置目标
func (lo *LivingObject) SetTarget(targetID common.ObjectIdType) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.targetID = targetID
}

// GetTarget 获取目标
func (lo *LivingObject) GetTarget() common.ObjectIdType {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.targetID
}

// ClearTarget 清除目标
func (lo *LivingObject) ClearTarget() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.targetID = 0
	lo.inCombat = false
}

// IsInCombat 检查是否在战斗中
func (lo *LivingObject) IsInCombat() bool {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.inCombat
}

// SetInCombat 设置战斗状态
func (lo *LivingObject) SetInCombat(inCombat bool) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.inCombat = inCombat
}

// StartCombat 开始战斗
func (lo *LivingObject) StartCombat() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.inCombat = true
}

// EndCombat 结束战斗
func (lo *LivingObject) EndCombat() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.inCombat = false
	lo.targetID = 0
}

// HasTarget 检查是否有目标
func (lo *LivingObject) HasTarget() bool {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.targetID != 0
}

// GetNearestTarget 获取最近的指定类型目标
// 参数:
//   - objectType: 目标类型
//
// 返回:
//   - common.IGameObject: 最近的目标对象
func (lo *LivingObject) GetNearestTarget(objectType gamecommon.GameObjectType) gamecommon.IGameObject {
	mapObject := lo.GetMap()
	if mapObject == nil {
		return nil
	}

	position := lo.GetPosition()
	nearestObject := gamecommon.IGameObject(nil)
	minDistance := float32(0)

	objects := mapObject.GetObjectsByType(objectType)
	for _, obj := range objects {
		distance := position.DistanceTo(obj.GetPosition())
		if nearestObject == nil || distance < minDistance {
			nearestObject = obj
			minDistance = distance
		}
	}

	return nearestObject
}

// IsInRangeWithTarget 检查是否在目标范围内
// 参数:
//   - target: 目标对象
//   - radius: 范围半径
//
// 返回:
//   - bool: 是否在范围内
func (lo *LivingObject) IsInRangeWithTarget(target *LivingObject, radius float32) bool {
	lo.mu.RLock()
	defer lo.mu.RUnlock()

	distance := lo.GetPosition().DistanceTo(target.GetPosition())
	return distance <= radius*radius
}

// GetSpeed 获取移动速度
// 默认值为3.0
func (lo *LivingObject) GetSpeed() float64 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	speed := lo.properties["move_speed"]
	if speed <= 0 {
		speed = 3.0
	}
	return speed
}

// GetRange 获取攻击范围
// 默认值为2.5
func (lo *LivingObject) GetRange() float64 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	rangeValue := lo.properties["attack_range"]
	if rangeValue <= 0 {
		rangeValue = 2.5
	}
	return rangeValue
}

// isDodge 检查是否闪避
// 基于dodge属性的随机判定
func (lo *LivingObject) isDodge() bool {
	dodgeChance := lo.GetProperty(property.PropertyDodge)
	if dodgeChance < 0 {
		dodgeChance = 0
	}
	return dodgeChance > rand.Float64()
}

// isCritical 检查是否暴击
// 基于critical_rate属性的随机判定
func (lo *LivingObject) isCritical() bool {
	criticalChance := lo.GetProperty(property.PropertyCriticalRate)
	if criticalChance < 0 {
		criticalChance = 0
	}
	return criticalChance > rand.Float64()
}

// Update 更新活体对象逻辑
// 处理攻击冷却等
func (lo *LivingObject) Update(deltaTime float64) {
	lo.GameObject.Update(deltaTime)

	now := time.Now()
	cooldown := 1.0 / float64(lo.GetSpeed())
	if now.Sub(lo.lastAttack).Seconds() < cooldown {
		return
	}

	lo.lastAttack = now

	if lo.buffComponent != nil {
		lo.buffComponent.Update(deltaTime)
	}

	if lo.combatComponent != nil {
		lo.combatComponent.Update(deltaTime)
	}

	if lo.movementComponent != nil {
		lo.movementComponent.Update(deltaTime)
	}
}

// SetOnDeath 设置死亡回调函数
// 参数:
//   - callback: 死亡回调函数
func (lo *LivingObject) SetOnDeath(callback func()) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.onDeath = callback
}

// IsDead 检查是否死亡
func (lo *LivingObject) IsDead() bool {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.health <= 0
}

// GetBuffComponent 获取Buff组件
func (lo *LivingObject) GetBuffComponent() *buff.BuffComponent {
	return lo.buffComponent
}

// GetCombatComponent 获取战斗组件
func (lo *LivingObject) GetCombatComponent() *combat.CombatComponent {
	return lo.combatComponent
}

// GetMovementComponent 获取移动组件
func (lo *LivingObject) GetMovementComponent() *movement.MovementComponent {
	return lo.movementComponent
}
