package object

import (
	"sync"

	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object/component"
)

// GameObjectType 游戏对象类型枚举
// 用于区分不同类型的游戏对象
type GameObjectType = gamecommon.GameObjectType

// GameObject 基础游戏对象类
// 实现了common.IGameObject接口，是所有游戏对象的基类
// 包含基本的属性管理、组件系统、事件系统和地图关联
type GameObject struct {
	*zObject.BaseObject                             // 继承基础对象（提供ID管理）
	mu                  sync.RWMutex                // 读写锁（保护并发访问）
	name                string                      // 对象名称
	objectType          gamecommon.GameObjectType   // 对象类型
	position            gamecommon.Vector3          // 三维位置坐标
	isActive            bool                        // 激活状态（false表示已销毁或暂停）
	eventEmitter        *zEvent.EventBus            // 事件发射器（发布/订阅事件）
	components          *component.ComponentManager // 组件管理器（ECS模式）
	mapObject           gamecommon.IMap             // 所属地图引用
}

// NewGameObject 创建新的游戏对象
// 参数:
//   - id: 对象唯一ID
//   - name: 对象名称
//
// 返回:
//   - *GameObject: 新创建的游戏对象
func NewGameObject(id common.ObjectIdType, name string) *GameObject {
	goObj := &GameObject{
		BaseObject:   &zObject.BaseObject{},
		name:         name,
		objectType:   gamecommon.GameObjectTypeBasic,
		position:     gamecommon.NewVector3(0, 0, 0),
		isActive:     true,
		eventEmitter: zEvent.NewEventBus(),
	}
	goObj.components = component.NewComponentManager(goObj)
	goObj.SetId(id)
	return goObj
}

// NewGameObjectWithType 创建指定类型的游戏对象
// 参数:
//   - id: 对象唯一ID
//   - name: 对象名称
//   - objectType: 对象类型
//
// 返回:
//   - *GameObject: 新创建的游戏对象
func NewGameObjectWithType(id common.ObjectIdType, name string, objectType GameObjectType) *GameObject {
	goObj := &GameObject{
		BaseObject:   &zObject.BaseObject{},
		name:         name,
		objectType:   objectType,
		position:     gamecommon.NewVector3(0, 0, 0),
		isActive:     true,
		eventEmitter: zEvent.NewEventBus(),
	}
	goObj.components = component.NewComponentManager(goObj)
	goObj.SetId(id)
	return goObj
}

// GetID 获取对象ID
func (goObj *GameObject) GetID() common.ObjectIdType {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.GetId().(common.ObjectIdType)
}

// SetID 设置对象ID
func (goObj *GameObject) SetID(id common.ObjectIdType) {
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
func (goObj *GameObject) GetType() GameObjectType {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.objectType
}

// SetType 设置对象类型
func (goObj *GameObject) SetType(objectType GameObjectType) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.objectType = objectType
}

// GetPosition 获取对象位置
func (goObj *GameObject) GetPosition() gamecommon.Vector3 {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.position
}

// SetPosition 设置对象位置
func (goObj *GameObject) SetPosition(position gamecommon.Vector3) {
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
// 用于发布和订阅游戏事件
func (goObj *GameObject) GetEventEmitter() *zEvent.EventBus {
	return goObj.eventEmitter
}

// GetMap 获取所属地图
func (goObj *GameObject) GetMap() gamecommon.IMap {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()
	return goObj.mapObject
}

// SetMap 设置所属地图
func (goObj *GameObject) SetMap(mapObj gamecommon.IMap) {
	goObj.mu.Lock()
	defer goObj.mu.Unlock()
	goObj.mapObject = mapObj
}

// AddComponent 添加组件
// 组件模式允许动态扩展对象功能
func (goObj *GameObject) AddComponent(component gamecommon.IComponent) {
	goObj.components.AddComponent(component)
}

// AddComponentWithName 添加带有名称的组件（兼容旧接口）
func (goObj *GameObject) AddComponentWithName(name string, component gamecommon.IComponent) {
	goObj.components.AddComponent(component)
}

// GetComponent 获取组件
func (goObj *GameObject) GetComponent(componentID string) gamecommon.IComponent {
	return goObj.components.GetComponent(componentID)
}

// RemoveComponent 移除组件
func (goObj *GameObject) RemoveComponent(componentID string) {
	goObj.components.RemoveComponent(componentID)
}

// Update 更新逻辑
// 每帧调用，更新所有组件
func (goObj *GameObject) Update(deltaTime float64) {
	goObj.components.Update(deltaTime)
}

// Destroy 销毁对象
// 移除地图引用，销毁所有组件，标记为非激活
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
func (goObj *GameObject) GetAllComponents() []gamecommon.IComponent {
	return goObj.components.GetAllComponents()
}

// MoveTo 移动到指定位置
// 如果对象在地图中，通过地图系统移动
func (goObj *GameObject) MoveTo(targetPos gamecommon.Vector3) error {
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
// 立即移动到目标位置，不检查路径
func (goObj *GameObject) Teleport(targetPos gamecommon.Vector3) error {
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
// 参数:
//   - targetPos: 目标位置
//   - radius: 半径
//
// 返回:
//   - bool: 是否在范围内
func (goObj *GameObject) InRange(targetPos gamecommon.Vector3, radius float32) bool {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()

	return goObj.position.DistanceTo(targetPos) <= radius*radius
}

// GetDistance 获取到目标对象的距离
func (goObj *GameObject) GetDistance(target gamecommon.IGameObject) float32 {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()

	targetPos := target.GetPosition()
	return goObj.position.DistanceTo(targetPos)
}

// IsSameMap 检查是否和目标在同一地图
func (goObj *GameObject) IsSameMap(target gamecommon.IGameObject) bool {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()

	targetMap := target.(*GameObject).GetMap()
	return goObj.mapObject == targetMap
}

// GetNeighbors 获取周围的对象
// 参数:
//   - radius: 搜索半径
//
// 返回:
//   - []common.IGameObject: 范围内的对象列表
func (goObj *GameObject) GetNeighbors(radius float32) []gamecommon.IGameObject {
	goObj.mu.RLock()
	defer goObj.mu.RUnlock()

	if goObj.mapObject == nil {
		return nil
	}

	return goObj.mapObject.GetObjectsInRange(goObj.position, radius)
}
