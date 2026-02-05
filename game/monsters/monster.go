package monster

import (
	"github.com/pzqf/zEngine/zScript"
	"github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object"
	"github.com/pzqf/zGameServer/game/object/component"
)

// Monster 怪物类
type Monster struct {
	*object.LivingObject
	aiBehavior   *AIBehavior
	dropConfig   *DropConfig
	scriptHolder *zScript.ScriptHolder
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

// GetState 获取AI状态
func (ai *AIBehavior) GetState() string {
	return ai.state
}

// SetState 设置AI状态
func (ai *AIBehavior) SetState(state string) {
	ai.state = state
}

// GetPatrolPath 获取巡逻路径
func (ai *AIBehavior) GetPatrolPath() []object.Vector3 {
	return ai.patrolPath
}

// SetPatrolPath 设置巡逻路径
func (ai *AIBehavior) SetPatrolPath(path []object.Vector3) {
	ai.patrolPath = path
}

// GetCurrentPatrolPoint 获取当前巡逻点
func (ai *AIBehavior) GetCurrentPatrolPoint() int {
	return ai.currentPatrolPoint
}

// SetCurrentPatrolPoint 设置当前巡逻点
func (ai *AIBehavior) SetCurrentPatrolPoint(point int) {
	ai.currentPatrolPoint = point
}

// GetPerceptionRange 获取感知范围
func (ai *AIBehavior) GetPerceptionRange() float32 {
	return ai.perceptionRange
}

// SetPerceptionRange 设置感知范围
func (ai *AIBehavior) SetPerceptionRange(range_ float32) {
	ai.perceptionRange = range_
}

// GetChaseRange 获取追击范围
func (ai *AIBehavior) GetChaseRange() float32 {
	return ai.chaseRange
}

// SetChaseRange 设置追击范围
func (ai *AIBehavior) SetChaseRange(range_ float32) {
	ai.chaseRange = range_
}

// GetRunawayRange 获取逃跑范围
func (ai *AIBehavior) GetRunawayRange() float32 {
	return ai.runawayRange
}

// SetRunawayRange 设置逃跑范围
func (ai *AIBehavior) SetRunawayRange(range_ float32) {
	ai.runawayRange = range_
}

// GetAttackRange 获取攻击范围
func (ai *AIBehavior) GetAttackRange() float32 {
	return 2.5
}

// DropConfig 掉落配置
type DropConfig struct {
	*component.BaseComponent
	dropItems map[int32]float32 // 物品ID -> 掉落概率
	exp       int32             // 经验值
	gold      int32             // 金币
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

// GetDropItems 获取掉落物品
func (dc *DropConfig) GetDropItems() map[int32]float32 {
	return dc.dropItems
}

// SetDropItems 设置掉落物品
func (dc *DropConfig) SetDropItems(items map[int32]float32) {
	dc.dropItems = items
}

// GetExp 获取经验值
func (dc *DropConfig) GetExp() int32 {
	return dc.exp
}

// SetExp 设置经验值
func (dc *DropConfig) SetExp(exp int32) {
	dc.exp = exp
}

// GetGold 获取金币
func (dc *DropConfig) GetGold() int32 {
	return dc.gold
}

// SetGold 设置金币
func (dc *DropConfig) SetGold(gold int32) {
	dc.gold = gold
}

// NewMonster 创建新的怪物对象
func NewMonster(id common.ObjectIdType, name string) *Monster {
	// 创建基础生命对象
	livingObj := object.NewLivingObject(id, name)

	// 创建AI行为组件
	aiBehavior := &AIBehavior{
		BaseComponent:      component.NewBaseComponent("ai"),
		state:              "patrol", // 默认巡逻状态
		patrolPath:         make([]object.Vector3, 0),
		currentPatrolPoint: 0,
		perceptionRange:    10.0,
		chaseRange:         15.0,
		runawayRange:       5.0,
	}

	// 创建掉落配置组件
	dropConfig := &DropConfig{
		BaseComponent: component.NewBaseComponent("drop"),
		dropItems:     make(map[int32]float32),
		exp:           100,
		gold:          50,
	}

	// 创建脚本持有者
	scriptHolder := &zScript.ScriptHolder{}

	// 创建怪物对象
	monster := &Monster{
		LivingObject: livingObj,
		aiBehavior:   aiBehavior,
		dropConfig:   dropConfig,
		scriptHolder: scriptHolder,
	}

	// 添加组件
	monster.AddComponentWithName("ai", aiBehavior)
	monster.AddComponentWithName("drop", dropConfig)

	return monster
}

// GetType 获取怪物类型
func (m *Monster) GetType() int {
	return int(common.GameObjectTypeMonster)
}

// GetAIBehavior 获取AI行为
func (m *Monster) GetAIBehavior() *AIBehavior {
	return m.aiBehavior
}

// GetDropConfig 获取掉落配置
func (m *Monster) GetDropConfig() *DropConfig {
	return m.dropConfig
}

// GetScriptHolder 获取脚本持有者
func (m *Monster) GetScriptHolder() *zScript.ScriptHolder {
	return m.scriptHolder
}

// SetScriptHolder 设置脚本持有者
func (m *Monster) SetScriptHolder(holder *zScript.ScriptHolder) {
	m.scriptHolder = holder
}

// BindScript 绑定脚本
func (m *Monster) BindScript(scriptFilename string) error {
	err := m.scriptHolder.BindScript(scriptFilename)
	if err != nil {
		return err
	}
	m.scriptHolder.SetContext(m)
	return nil
}

// UpdateAI 更新AI
func (m *Monster) UpdateAI(deltaTime int) {
	m.scriptHolder.Update(deltaTime)
}

// InitAI 初始化AI
func (m *Monster) InitAI(scriptFilename string) error {
	if scriptFilename != "" {
		err := m.BindScript(scriptFilename)
		if err != nil {
			return err
		}
	}
	return nil
}
