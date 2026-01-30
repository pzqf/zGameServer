package object

import (
	"sync"

	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object/component"
)

// GameObjectType 游戏对象类型枚举
type GameObjectType = common.GameObjectType

// GameObject 基础游戏对象类
type GameObject struct {
	*zObject.BaseObject
	mu           sync.RWMutex
	name         string
	objectType   common.GameObjectType
	position     common.Vector3
	isActive     bool
	eventEmitter *zEvent.EventBus
	components   *component.ComponentManager
	mapObject    common.IMap
}

// NewGameObject 创建新的游戏对象
func NewGameObject(id uint64, name string) *GameObject {
	goObj := &GameObject{
		BaseObject:   &zObject.BaseObject{},
		name:         name,
		objectType:   common.GameObjectTypeBasic,
		position:     common.NewVector3(0, 0, 0),
		isActive:     true,
		eventEmitter: zEvent.NewEventBus(),
	}
	goObj.components = component.NewComponentManager(goObj)
	goObj.SetId(id)
	return goObj
}

// NewGameObjectWithType 创建指定类型的游戏对象
func NewGameObjectWithType(id uint64, name string, objectType GameObjectType) *GameObject {
	goObj := &GameObject{
		BaseObject:   &zObject.BaseObject{},
		name:         name,
		objectType:   objectType,
		position:     common.NewVector3(0, 0, 0),
		isActive:     true,
		eventEmitter: zEvent.NewEventBus(),
	}
	goObj.components = component.NewComponentManager(goObj)
	goObj.SetId(id)
	return goObj
}

// GetID 获取对象ID
func (goObj *GameObject) GetID() uint64 {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.GetId().(uint64)
}

// SetID 设置对象ID
func (goObj *GameObject) SetID(id uint64) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.SetId(id)
}

// GetName 获取对象名称
func (goObj *GameObject) GetName() string {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.name
}

// SetName 设置对象名称
func (goObj *GameObject) SetName(name string) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.name = name
}

// GetType 获取对象类型
func (goObj *GameObject) GetType() int {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return int(goObj.objectType)
}

// SetType 设置对象类型
func (goObj *GameObject) SetType(objectType GameObjectType) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.objectType = objectType
}

// GetPosition 获取对象位置
func (goObj *GameObject) GetPosition() common.Vector3 {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.position
}

// SetPosition 设置对象位置
func (goObj *GameObject) SetPosition(position common.Vector3) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.position = position
}

// IsActive 检查对象是否激活
func (goObj *GameObject) IsActive() bool {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.isActive
}

// SetActive 设置对象激活状态
func (goObj *GameObject) SetActive(active bool) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.isActive = active
}

// GetEventEmitter 获取事件总线
func (goObj *GameObject) GetEventEmitter() *zEvent.EventBus {
	return goObj.eventEmitter
}

// GetMap 获取所属地图
func (goObj *GameObject) GetMap() common.IMap {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.mapObject
}

// SetMap 设置所属地图
func (goObj *GameObject) SetMap(mapObj common.IMap) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.mapObject = mapObj
}

// AddComponent 添加组件
func (goObj *GameObject) AddComponent(component common.IComponent) {
	goObj.components.AddComponent(component)
}

// AddComponentWithName 添加带有名称的组件（兼容旧接口）
func (goObj *GameObject) AddComponentWithName(name string, component common.IComponent) {
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

// Update 更新逻辑
func (goObj *GameObject) Update(deltaTime float64) {
	goObj.components.Update(deltaTime)
}

// Destroy 销毁对象
func (goObj *GameObject) Destroy() {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()

	// 移除地图引用
	if goObj.mapObject != nil {
		goObj.mapObject.RemoveObject(goObj.GetID())
		goObj.mapObject = nil
	}

	// 销毁所有组件
	goObj.components.Destroy()
	goObj.SetActive(false)
}

// HasComponent 检查是否有指定组件
func (goObj *GameObject) HasComponent(componentID string) bool {
	return goObj.components.HasComponent(componentID)
}

// GetAllComponents 获取所有组件
func (goObj *GameObject) GetAllComponents() []common.IComponent {
	return goObj.components.GetAllComponents()
}

// Move 移动到指定位置
func (goObj *GameObject) MoveTo(targetPos common.Vector3) error {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()

	// 检查是否有地图
	if goObj.mapObject != nil {
		return goObj.mapObject.MoveObject(goObj, targetPos)
	}

	// 直接移动
	goObj.position = targetPos
	return nil
}

// Teleport 传送到指定位置
func (goObj *GameObject) Teleport(targetPos common.Vector3) error {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()

	// 检查是否有地图
	if goObj.mapObject != nil {
		return goObj.mapObject.TeleportObject(goObj, targetPos)
	}

	// 直接传送
	goObj.position = targetPos
	return nil
}

// InRange 检查是否在指定范围内
func (goObj *GameObject) InRange(targetPos common.Vector3, radius float32) bool {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()

	return goObj.position.DistanceTo(targetPos) <= radius*radius
}

// GetDistance 获取到目标的距离
func (goObj *GameObject) GetDistance(target common.IGameObject) float32 {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()

	targetPos := target.GetPosition()
	return goObj.position.DistanceTo(targetPos)
}

// IsSameMap 检查是否和目标在同一地图
func (goObj *GameObject) IsSameMap(target common.IGameObject) bool {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()

	targetMap := target.(*GameObject).GetMap()
	return goObj.mapObject == targetMap
}

// GetNeighbors 获取周围的对象
func (goObj *GameObject) GetNeighbors(radius float32) []common.IGameObject {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()

	if goObj.mapObject == nil {
		return nil
	}

	return goObj.mapObject.GetObjectsInRange(goObj.position, radius)
}
