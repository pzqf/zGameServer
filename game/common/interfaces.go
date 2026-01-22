package common

import (
	"github.com/pzqf/zEngine/zEvent"
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
	GetID() uint64
	// 获取名称
	GetName() string
	// 获取对象类型
	GetType() int
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
	// 组件管理
	AddComponent(component IComponent)
	GetComponent(componentID string) IComponent
	RemoveComponent(componentID string)
	HasComponent(componentID string) bool
	GetAllComponents() []IComponent
}

// Vector3 三维向量
type Vector3 struct {
	X, Y, Z float32
}

// NewVector3 创建三维向量
func NewVector3(x, y, z float32) Vector3 {
	return Vector3{X: x, Y: y, Z: z}
}

// Add 向量加法
func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

// Subtract 向量减法
func (v Vector3) Subtract(other Vector3) Vector3 {
	return Vector3{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

// MultiplyScalar 向量乘以标量
func (v Vector3) MultiplyScalar(scalar float32) Vector3 {
	return Vector3{
		X: v.X * scalar,
		Y: v.Y * scalar,
		Z: v.Z * scalar,
	}
}

// Dot 向量点积
func (v Vector3) Dot(other Vector3) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross 向量叉积
func (v Vector3) Cross(other Vector3) Vector3 {
	return Vector3{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

// Length 向量长度
func (v Vector3) Length() float32 {
	return float32((v.X*v.X + v.Y*v.Y + v.Z*v.Z))
}

// Normalize 向量归一化
func (v Vector3) Normalize() Vector3 {
	length := v.Length()
	if length == 0 {
		return v
	}
	return v.MultiplyScalar(1 / length)
}

// DistanceTo 计算到另一个向量的距离
func (v Vector3) DistanceTo(other Vector3) float32 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	dz := v.Z - other.Z
	return float32((dx*dx + dy*dy + dz*dz))
}
