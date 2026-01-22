package monster

import (
	"github.com/pzqf/zGameServer/game/object"
	"github.com/pzqf/zGameServer/game/object/component"
)

// Monster 怪物类
type Monster struct {
	object.LivingObject
	aiBehavior *AIBehavior
	dropConfig *DropConfig
}

// AIBehavior 怪物AI行为
type AIBehavior struct {
	*component.BaseComponent
	// AI状态：巡逻、追击、战斗、逃跑
	state string
	// 巡逻路径
	patrolPath []object.Vector3
	// 当前巡逻点
	currentPatrolPoint int
	// 感知范围
	perceptionRange float32
	// 追击范围
	chaseRange float32
	// 逃跑范围
	runawayRange float32
}

// Init 初始化AI行为组件
func (ai *AIBehavior) Init() error {
	return nil
}

// Update 更新AI行为组件
func (ai *AIBehavior) Update(deltaTime float64) {
}

// Destroy 销毁AI行为组件
func (ai *AIBehavior) Destroy() {
}

// IsActive 检查AI行为组件是否激活
func (ai *AIBehavior) IsActive() bool {
	return ai.BaseComponent.IsActive()
}

// SetActive 设置AI行为组件是否激活
func (ai *AIBehavior) SetActive(active bool) {
	ai.BaseComponent.SetActive(active)
}

// DropConfig 怪物掉落配置
type DropConfig struct {
	*component.BaseComponent
	// 掉落物品列表
	dropItems map[int32]float32 // key: 物品ID, value: 掉落概率
	// 经验值
	exp int32
	// 金币
	gold int32
}

// Init 初始化掉落配置组件
func (dc *DropConfig) Init() error {
	return nil
}

// Update 更新掉落配置组件
func (dc *DropConfig) Update(deltaTime float64) {
}

// Destroy 销毁掉落配置组件
func (dc *DropConfig) Destroy() {
}

// IsActive 检查掉落配置组件是否激活
func (dc *DropConfig) IsActive() bool {
	return dc.BaseComponent.IsActive()
}

// SetActive 设置掉落配置组件是否激活
func (dc *DropConfig) SetActive(active bool) {
	dc.BaseComponent.SetActive(active)
}

// NewMonster 创建新的怪物对象
func NewMonster(id uint64, name string) *Monster {
	// 创建基础生命对象
	livingObj := object.NewLivingObject(id, name)

	// 创建AI行为组件
	aiBehavior := &AIBehavior{
		BaseComponent:   component.NewBaseComponent("ai"),
		state:           "patrol",
		perceptionRange: 10.0,
		chaseRange:      20.0,
		runawayRange:    5.0,
	}

	// 创建掉落配置组件
	dropConfig := &DropConfig{
		BaseComponent: component.NewBaseComponent("drop"),
		dropItems:     make(map[int32]float32),
		exp:           100,
		gold:          50,
	}

	// 创建怪物对象
	monster := &Monster{
		LivingObject: *livingObj,
		aiBehavior:   aiBehavior,
		dropConfig:   dropConfig,
	}

	// 添加组件
	monster.AddComponentWithName("ai", aiBehavior)
	monster.AddComponentWithName("drop", dropConfig)

	return monster
}

// GetType 获取怪物类型
func (m *Monster) GetType() int {
	return object.GameObjectTypeMonster
}
