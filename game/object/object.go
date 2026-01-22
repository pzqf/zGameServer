package object

import (
	"sync"

	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object/component"
	"github.com/pzqf/zGameServer/game/systems/combat"
	"github.com/pzqf/zGameServer/game/systems/property"
)

// GameObject 基础游戏对象类
type GameObject struct {
	mu           sync.RWMutex
	id           uint64
	name         string
	objectType   int
	position     common.Vector3
	isActive     bool
	eventEmitter *zEvent.EventBus
	components   *component.ComponentManager
}

// NewGameObject 创建新的游戏对象
func NewGameObject(id uint64, name string) *GameObject {
	goObj := &GameObject{
		id:           id,
		name:         name,
		objectType:   GameObjectTypeBasic,
		position:     common.NewVector3(0, 0, 0),
		isActive:     true,
		eventEmitter: zEvent.GetGlobalEventBus(),
	}
	goObj.components = component.NewComponentManager(goObj)
	return goObj
}

// NewGameObjectWithType 创建指定类型的游戏对象
func NewGameObjectWithType(id uint64, name string, objectType int) *GameObject {
	goObj := &GameObject{
		id:           id,
		name:         name,
		objectType:   objectType,
		position:     common.NewVector3(0, 0, 0),
		isActive:     true,
		eventEmitter: zEvent.GetGlobalEventBus(),
	}
	goObj.components = component.NewComponentManager(goObj)
	return goObj
}

// GetID 获取唯一标识
func (goObj *GameObject) GetID() uint64 {
	return goObj.id
}

// GetName 获取名称
func (goObj *GameObject) GetName() string {
	return goObj.name
}

// GetType 获取对象类型
func (goObj *GameObject) GetType() int {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.objectType
}

// SetType 设置对象类型
func (goObj *GameObject) SetType(objectType int) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.objectType = objectType
}

// GetPosition 获取位置信息
func (goObj *GameObject) GetPosition() common.Vector3 {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.position
}

// SetPosition 设置位置
func (goObj *GameObject) SetPosition(pos common.Vector3) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.position = pos
}

// Update 更新逻辑
func (goObj *GameObject) Update(deltaTime float64) {
	// 基础游戏对象更新逻辑
	// 移动、战斗等功能由独立系统处理

	// 更新所有组件
	goObj.components.Update(deltaTime)
}

// Destroy 销毁对象
func (goObj *GameObject) Destroy() {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.isActive = false

	// 销毁所有组件
	goObj.components.Destroy()
	// TODO: 释放其他资源
}

// IsActive 检查是否存活
func (goObj *GameObject) IsActive() bool {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.isActive
}

// SetActive 设置是否激活
func (goObj *GameObject) SetActive(active bool) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.isActive = active
}

// GetEventEmitter 获取事件总线
func (goObj *GameObject) GetEventEmitter() *zEvent.EventBus {
	return goObj.eventEmitter
}

// AddComponent 添加组件
func (goObj *GameObject) AddComponent(component common.IComponent) {
	// 直接添加组件到管理器
	goObj.components.AddComponent(component)
}

// AddComponentWithName 添加带有名称的组件（兼容旧接口）
func (goObj *GameObject) AddComponentWithName(name string, component common.IComponent) {
	// 直接添加组件到管理器
	goObj.components.AddComponent(component)
}

// GetComponent 获取组件
func (goObj *GameObject) GetComponent(componentID string) common.IComponent {
	return goObj.components.GetComponent(componentID)
}

// RemoveComponent 移除组件
func (goObj *GameObject) RemoveComponent(componentID string) {
	goObj.components.RemoveComponent(componentID)
}

// HasComponent 检查是否存在指定组件
func (goObj *GameObject) HasComponent(componentID string) bool {
	return goObj.components.HasComponent(componentID)
}

// GetAllComponents 获取所有组件
func (goObj *GameObject) GetAllComponents() []common.IComponent {
	return goObj.components.GetAllComponents()
}

// LivingObject 生命对象类
type LivingObject struct {
	GameObject
	mu        sync.RWMutex
	health    float32
	maxHealth float32
}

// NewLivingObject 创建新的生命对象
func NewLivingObject(id uint64, name string) *LivingObject {
	gameObj := NewGameObject(id, name)

	// 使用属性系统设置默认属性
	property.GlobalPropertySystem.SetProperty(id, "health", 100)
	property.GlobalPropertySystem.SetProperty(id, "max_health", 100)
	property.GlobalPropertySystem.SetProperty(id, "attack", 10)
	property.GlobalPropertySystem.SetProperty(id, "defense", 5)
	property.GlobalPropertySystem.SetProperty(id, "speed", 3)
	property.GlobalPropertySystem.SetProperty(id, "mana", 50)
	property.GlobalPropertySystem.SetProperty(id, "max_mana", 50)

	// 创建生命对象
	livingObj := &LivingObject{
		GameObject: *gameObj,
		health:     100,
		maxHealth:  100,
	}

	return livingObj
}

// GetHealth 获取生命值
func (lo *LivingObject) GetHealth() float32 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.health
}

// SetHealth 设置生命值
func (lo *LivingObject) SetHealth(health float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	// 确保生命值在0到maxHealth之间
	lo.health = max(0, min(health, lo.maxHealth))
}

// GetMaxHealth 获取最大生命值
func (lo *LivingObject) GetMaxHealth() float32 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.maxHealth
}

// SetMaxHealth 设置最大生命值
func (lo *LivingObject) SetMaxHealth(maxHealth float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.maxHealth = max(1, maxHealth)
	// 确保当前生命值不超过新的最大生命值
	if lo.health > lo.maxHealth {
		lo.health = lo.maxHealth
	}
}

// GetType 获取对象类型
func (lo *LivingObject) GetType() int {
	return GameObjectTypeLiving
}

// TakeDamage 受到伤害
func (lo *LivingObject) TakeDamage(damage float32, attacker common.IGameObject) {
	// 使用战斗系统计算最终伤害
	finalDamage := combat.GlobalCombatSystem.CalculateDamage(attacker.GetID(), lo.GetID())

	currentHealth := lo.GetHealth()
	newHealth := currentHealth - finalDamage

	lo.SetHealth(newHealth)

	// TODO: 实现事件触发
	// lo.GetEventEmitter().Emit(event.EventPlayerTakeDamage, lo, finalDamage, attacker)

	// 如果生命值为0，触发死亡事件
	if lo.GetHealth() <= 0 {
		lo.Die(attacker)
	}
}

// Heal 回复生命值
func (lo *LivingObject) Heal(amount float32) {
	currentHealth := lo.GetHealth()
	lo.SetHealth(currentHealth + amount)
}

// Die 死亡处理
func (lo *LivingObject) Die(killer common.IGameObject) {
	lo.SetActive(false)

	// TODO: 实现事件触发
	// lo.GetEventEmitter().Emit(event.EventPlayerDie, lo, killer)

	// 结束战斗状态，使用独立的战斗系统
	combat.GlobalCombatSystem.EndCombat(lo.GetID())
}

// IsAlive 检查是否存活
func (lo *LivingObject) IsAlive() bool {
	return lo.IsActive() && lo.GetHealth() > 0
}

// Helper functions
func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
