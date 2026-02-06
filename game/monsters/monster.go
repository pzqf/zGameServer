package monster

import (
	"github.com/pzqf/zEngine/zScript"
	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object"
	"github.com/pzqf/zGameServer/game/object/component"
)

// Monster 怪物类
// 游戏中的怪物实体，继承自LivingObject
type Monster struct {
	*object.LivingObject
	aiBehavior   *AIBehavior           // AI行为组件
	dropConfig   *DropConfig           // 掉落配置组件
	scriptHolder *zScript.ScriptHolder // 脚本持有者
}

// AIBehavior 怪物AI行为
// 管理怪物的AI状态和行为参数
type AIBehavior struct {
	*component.BaseComponent
	state              string           // AI状态（patrol/chase/combat/runaway）
	patrolPath         []object.Vector3 // 巡逻路径
	currentPatrolPoint int              // 当前巡逻点索引
	perceptionRange    float32          // 感知范围（发现玩家的距离）
	chaseRange         float32          // 追击范围（最大追击距离）
	runawayRange       float32          // 逃跑范围（超出后返回巡逻）
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
// 返回固定的攻击距离
func (ai *AIBehavior) GetAttackRange() float32 {
	return 2.5
}

// DropConfig 掉落配置
// 管理怪物的掉落物品和奖励
type DropConfig struct {
	*component.BaseComponent
	dropItems map[int32]float32 // 掉落物品映射（物品ID -> 掉落概率）
	exp       int32             // 击杀获得经验
	gold      int32             // 击杀获得金币
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
// 参数:
//   - id: 怪物对象ID
//   - name: 怪物名称
//
// 返回: 新创建的怪物对象
func NewMonster(id common.ObjectIdType, name string) *Monster {
	livingObj := object.NewLivingObject(id, name)

	aiBehavior := &AIBehavior{
		BaseComponent:      component.NewBaseComponent("ai"),
		state:              "patrol",
		patrolPath:         make([]object.Vector3, 0),
		currentPatrolPoint: 0,
		perceptionRange:    10.0,
		chaseRange:         15.0,
		runawayRange:       5.0,
	}

	dropConfig := &DropConfig{
		BaseComponent: component.NewBaseComponent("drop"),
		dropItems:     make(map[int32]float32),
		exp:           100,
		gold:          50,
	}

	scriptHolder := &zScript.ScriptHolder{}

	monster := &Monster{
		LivingObject: livingObj,
		aiBehavior:   aiBehavior,
		dropConfig:   dropConfig,
		scriptHolder: scriptHolder,
	}

	monster.AddComponentWithName("ai", aiBehavior)
	monster.AddComponentWithName("drop", dropConfig)

	return monster
}

// GetType 获取怪物类型
func (m *Monster) GetType() gamecommon.GameObjectType {
	return gamecommon.GameObjectTypeMonster
}

// GetAIBehavior 获取AI行为组件
func (m *Monster) GetAIBehavior() *AIBehavior {
	return m.aiBehavior
}

// GetDropConfig 获取掉落配置组件
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
// 参数:
//   - scriptFilename: 脚本文件名
//
// 返回: 绑定错误
func (m *Monster) BindScript(scriptFilename string) error {
	err := m.scriptHolder.BindScript(scriptFilename)
	if err != nil {
		return err
	}
	m.scriptHolder.SetContext(m)
	return nil
}

// UpdateAI 更新AI
// 参数:
//   - deltaTime: 时间增量（毫秒）
func (m *Monster) UpdateAI(deltaTime int) {
	m.scriptHolder.Update(deltaTime)
}

// InitAI 初始化AI
// 参数:
//   - scriptFilename: 脚本文件名（可选）
//
// 返回: 初始化错误
func (m *Monster) InitAI(scriptFilename string) error {
	if scriptFilename != "" {
		err := m.BindScript(scriptFilename)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetOnDeath 设置死亡回调
func (m *Monster) SetOnDeath(callback func()) {
	m.LivingObject.SetOnDeath(callback)
}
