package common

import (
	"math"

	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zGameServer/common"
)

// IComponent 组件接口
type IComponent interface {
	// 获取组件ID
	GetID() string
	// 获取组件所属的游戏对象
	GetGameObject() IGameObject
	// 设置组件所属的游戏对象
	SetGameObject(obj IGameObject)
	// 初始化组件
	Init() error
	// 更新组件
	Update(deltaTime float64)
	// 销毁组件
	Destroy()
	// 检查组件是否激活
	IsActive() bool
	// 设置组件是否激活
	SetActive(active bool)
}

// IGameObject 所有游戏对象的最基本行为接口
type IGameObject interface {
	// 获取唯一标识
	GetID() common.ObjectIdType
	// 获取名称
	GetName() string
	// 获取对象类型
	GetType() GameObjectType
	// 获取位置信息
	GetPosition() Vector3
	// 设置位置
	SetPosition(pos Vector3)
	// 更新逻辑
	Update(deltaTime float64)
	// 销毁对象
	Destroy()
	// 检查是否存活
	IsActive() bool
	// 设置是否激活
	SetActive(active bool)
	// 获取事件总线
	GetEventEmitter() *zEvent.EventBus
	// 获取所属地图
	GetMap() IMap
	// 设置所属地图
	SetMap(mapObj IMap)
	// 组件管理
	AddComponent(component IComponent)
	GetComponent(componentID string) IComponent
	RemoveComponent(componentID string)
	HasComponent(componentID string) bool
	GetAllComponents() []IComponent
}

// IMap 地图接口
type IMap interface {
	// 获取地图ID
	GetID() common.MapIdType
	// 获取地图名称
	GetName() string
	// 获取指定范围内的对象
	GetObjectsInRange(pos Vector3, radius float32) []IGameObject
	// 获取指定类型的对象
	GetObjectsByType(objectType GameObjectType) []IGameObject
	// 添加对象
	AddObject(object IGameObject)
	// 移除对象
	RemoveObject(objectID common.ObjectIdType)
	// 移动对象
	MoveObject(object IGameObject, targetPos Vector3) error
	// 传送对象
	TeleportObject(object IGameObject, targetPos Vector3) error
}

// GameObjectType 游戏对象类型
type GameObjectType int

// 游戏对象类型常量
const (
	GameObjectTypeBasic    GameObjectType = 0
	GameObjectTypeLiving   GameObjectType = 1
	GameObjectTypePlayer   GameObjectType = 2
	GameObjectTypeNPC      GameObjectType = 3
	GameObjectTypeMonster  GameObjectType = 4
	GameObjectTypePet      GameObjectType = 5
	GameObjectTypeItem     GameObjectType = 6
	GameObjectTypeBuilding GameObjectType = 7
)

// Vector3 三维向量
type Vector3 struct {
	X, Y, Z float32
}

// NewVector3 创建三维向量
func NewVector3(x, y, z float32) Vector3 {
	return Vector3{X: x, Y: y, Z: z}
}

// Add 向量相加
func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

// Subtract 向量相减
func (v Vector3) Subtract(other Vector3) Vector3 {
	return Vector3{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

// MultiplyScalar 向量乘以标量
func (v Vector3) MultiplyScalar(scalar float32) Vector3 {
	return Vector3{X: v.X * scalar, Y: v.Y * scalar, Z: v.Z * scalar}
}

// DivideScalar 向量除以标量
func (v Vector3) DivideScalar(scalar float32) Vector3 {
	return Vector3{X: v.X / scalar, Y: v.Y / scalar, Z: v.Z / scalar}
}

// DistanceTo 计算两个向量之间的距离
func (v Vector3) DistanceTo(other Vector3) float32 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	dz := v.Z - other.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// Dot 点积
func (v Vector3) Dot(other Vector3) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Length 向量长度
func (v Vector3) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

// Normalize 标准化向量
func (v Vector3) Normalize() Vector3 {
	length := v.Length()
	if length == 0 {
		return Vector3{}
	}
	return Vector3{
		X: v.X / length,
		Y: v.Y / length,
		Z: v.Z / length,
	}
}
