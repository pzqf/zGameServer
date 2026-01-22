package pet

import (
	"github.com/pzqf/zGameServer/game/object"
	"github.com/pzqf/zGameServer/game/object/component"
)

// PetGrowthSystem 宠物成长系统
type PetGrowthSystem struct {
	*component.BaseComponent
	level      int32
	exp        int32
	maxExp     int32
	growthRate float32
}

// Init 初始化宠物成长系统组件
func (pgs *PetGrowthSystem) Init() error {
	return nil
}

// Update 更新宠物成长系统组件
func (pgs *PetGrowthSystem) Update(deltaTime float64) {
}

// Destroy 销毁宠物成长系统组件
func (pgs *PetGrowthSystem) Destroy() {
}

// IsActive 检查宠物成长系统组件是否激活
func (pgs *PetGrowthSystem) IsActive() bool {
	return pgs.BaseComponent.IsActive()
}

// SetActive 设置宠物成长系统组件是否激活
func (pgs *PetGrowthSystem) SetActive(active bool) {
	pgs.BaseComponent.SetActive(active)
}

// IntimacySystem 宠物亲密度系统
type IntimacySystem struct {
	*component.BaseComponent
	intimacy    int32
	maxIntimacy int32
	mood        string
}

// Init 初始化宠物亲密度系统组件
func (is *IntimacySystem) Init() error {
	return nil
}

// Update 更新宠物亲密度系统组件
func (is *IntimacySystem) Update(deltaTime float64) {
}

// Destroy 销毁宠物亲密度系统组件
func (is *IntimacySystem) Destroy() {
}

// IsActive 检查宠物亲密度系统组件是否激活
func (is *IntimacySystem) IsActive() bool {
	return is.BaseComponent.IsActive()
}

// SetActive 设置宠物亲密度系统组件是否激活
func (is *IntimacySystem) SetActive(active bool) {
	is.BaseComponent.SetActive(active)
}

// 宠物类
type Pet struct {
	object.LivingObject
	petGrowth *PetGrowthSystem
	intimacy  *IntimacySystem
}

// NewPet 创建新的宠物对象
func NewPet(id uint64, name string) *Pet {
	// 创建基础生命对象
	livingObj := object.NewLivingObject(id, name)

	// 创建宠物成长系统
	growthSystem := &PetGrowthSystem{
		BaseComponent: component.NewBaseComponent("growth"),
		level:         1,
		exp:           0,
		maxExp:        100,
		growthRate:    1.2,
	}

	// 创建宠物亲密度系统
	intimacySystem := &IntimacySystem{
		BaseComponent: component.NewBaseComponent("intimacy"),
		intimacy:      50,
		maxIntimacy:   100,
		mood:          "happy",
	}

	// 创建宠物对象
	pet := &Pet{
		LivingObject: *livingObj,
		petGrowth:    growthSystem,
		intimacy:     intimacySystem,
	}

	// 添加组件
	pet.AddComponentWithName("growth", growthSystem)
	pet.AddComponentWithName("intimacy", intimacySystem)

	return pet
}

// GetType 获取宠物类型
func (p *Pet) GetType() int {
	return object.GameObjectTypePet
}
