package component

import "github.com/pzqf/zGameServer/game/common"

// PositionComponent 位置组件
type PositionComponent struct {
	*BaseComponent
	position common.Vector3
}

// NewPositionComponent 创建位置组件
func NewPositionComponent() *PositionComponent {
	return &PositionComponent{
		BaseComponent: NewBaseComponent("PositionComponent"),
		position:      common.NewVector3(0, 0, 0),
	}
}

// GetPosition 获取位置
func (pc *PositionComponent) GetPosition() common.Vector3 {
	return pc.position
}

// SetPosition 设置位置
func (pc *PositionComponent) SetPosition(pos common.Vector3) {
	pc.position = pos
}

// Init 初始化组件
func (pc *PositionComponent) Init() error {
	return pc.BaseComponent.Init()
}

// Update 更新组件
func (pc *PositionComponent) Update(deltaTime float64) {
	pc.BaseComponent.Update(deltaTime)
}

// Destroy 销毁组件
func (pc *PositionComponent) Destroy() {
	pc.BaseComponent.Destroy()
}
