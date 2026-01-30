package pet

import (
	"github.com/pzqf/zGameServer/game/common"
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

// GetLevel 获取等级
func (pgs *PetGrowthSystem) GetLevel() int32 {
	return pgs.level
}

// SetLevel 设置等级
func (pgs *PetGrowthSystem) SetLevel(level int32) {
	pgs.level = level
}

// GetExp 获取经验值
func (pgs *PetGrowthSystem) GetExp() int32 {
	return pgs.exp
}

// SetExp 设置经验值
func (pgs *PetGrowthSystem) SetExp(exp int32) {
	pgs.exp = exp
}

// GetMaxExp 获取最大经验值
func (pgs *PetGrowthSystem) GetMaxExp() int32 {
	return pgs.maxExp
}

// SetMaxExp 设置最大经验值
func (pgs *PetGrowthSystem) SetMaxExp(maxExp int32) {
	pgs.maxExp = maxExp
}

// GetGrowthRate 获取成长率
func (pgs *PetGrowthSystem) GetGrowthRate() float32 {
	return pgs.growthRate
}

// SetGrowthRate 设置成长率
func (pgs *PetGrowthSystem) SetGrowthRate(rate float32) {
	pgs.growthRate = rate
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

// GetIntimacy 获取亲密度
func (is *IntimacySystem) GetIntimacy() int32 {
	return is.intimacy
}

// SetIntimacy 设置亲密度
func (is *IntimacySystem) SetIntimacy(intimacy int32) {
	is.intimacy = intimacy
}

// GetMaxIntimacy 获取最大亲密度
func (is *IntimacySystem) GetMaxIntimacy() int32 {
	return is.maxIntimacy
}

// SetMaxIntimacy 设置最大亲密度
func (is *IntimacySystem) SetMaxIntimacy(maxIntimacy int32) {
	is.maxIntimacy = maxIntimacy
}

// GetMood 获取心情
func (is *IntimacySystem) GetMood() string {
	return is.mood
}

// SetMood 设置心情
func (is *IntimacySystem) SetMood(mood string) {
	is.mood = mood
}

// 宠物类
type Pet struct {
	*object.LivingObject
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
		LivingObject: livingObj,
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
	return int(common.GameObjectTypePet)
}

// GetPetGrowth 获取宠物成长系统
func (p *Pet) GetPetGrowth() *PetGrowthSystem {
	return p.petGrowth
}

// GetIntimacy 获取亲密度系统
func (p *Pet) GetIntimacy() *IntimacySystem {
	return p.intimacy
}
