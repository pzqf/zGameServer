package skill

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/config/models"
	"github.com/pzqf/zGameServer/config/tables"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/systems/property"
)

// Skill 技能数据结构
// 存储技能的运行时数据
type Skill struct {
	ID          int32     // 技能ID
	Level       int32     // 技能等级
	Name        string    // 技能名称
	Description string    // 技能描述
	Cooldown    float32   // 冷却时间（秒）
	Range       float32   // 技能范围
	Damage      float32   // 技能伤害
	ManaCost    float32   // 魔法消耗
	Type        string    // 技能类型：active主动、passive被动、heal治疗
	TargetType  string    // 目标类型：single单体、area群体、self自身
	LastUsed    time.Time // 最后使用时间
}

// SkillState 技能状态
// 管理技能列表和冷却状态
type SkillState struct {
	mu        sync.RWMutex         // 读写锁
	skills    map[int32]*Skill     // 已学习技能映射
	cooldowns map[int32]time.Time  // 技能冷却时间映射
	skillPool *zObject.GenericPool // 技能对象池
}

// NewSkillState 创建技能状态
// 返回: 初始化后的技能状态实例
func NewSkillState() *SkillState {
	return &SkillState{
		skills:    make(map[int32]*Skill),
		cooldowns: make(map[int32]time.Time),
		skillPool: zObject.NewGenericPool(func() interface{} { return &Skill{} }, 100),
	}
}

// AddSkill 添加技能
// 从配置加载技能数据并添加到技能列表
// 参数:
//   - skillID: 技能ID
//   - config: 技能配置
func (state *SkillState) AddSkill(skillID int32, config *models.Skill) {
	if config == nil {
		return
	}

	skillType := "active"
	switch config.Type {
	case 1:
		skillType = "active"
	case 2:
		skillType = "passive"
	case 3:
		skillType = "heal"
	}

	targetType := "single"
	if config.AreaRadius > 0 {
		targetType = "area"
	}

	newSkill := state.skillPool.Get().(*Skill)
	newSkill.ID = skillID
	newSkill.Level = 1
	newSkill.Name = config.Name
	newSkill.Description = config.Description
	newSkill.Cooldown = float32(config.Cooldown)
	newSkill.Range = float32(config.Range)
	newSkill.Damage = float32(config.Damage)
	newSkill.ManaCost = float32(config.ManaCost)
	newSkill.Type = skillType
	newSkill.TargetType = targetType
	newSkill.LastUsed = time.Time{}

	state.mu.Lock()
	defer state.mu.Unlock()
	state.skills[skillID] = newSkill
}

// RemoveSkill 移除技能
// 参数:
//   - skillID: 技能ID
func (state *SkillState) RemoveSkill(skillID int32) {
	state.mu.Lock()
	defer state.mu.Unlock()
	if skill, exists := state.skills[skillID]; exists {
		state.skillPool.Put(skill)
		delete(state.skills, skillID)
	}
	delete(state.cooldowns, skillID)
}

// GetSkill 获取技能
// 参数:
//   - skillID: 技能ID
//
// 返回: 技能和是否存在
func (state *SkillState) GetSkill(skillID int32) (*Skill, bool) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	skill, exists := state.skills[skillID]
	return skill, exists
}

// GetAllSkills 获取所有技能
// 返回: 技能映射副本
func (state *SkillState) GetAllSkills() map[int32]*Skill {
	state.mu.RLock()
	defer state.mu.RUnlock()
	result := make(map[int32]*Skill, len(state.skills))
	for k, v := range state.skills {
		result[k] = v
	}
	return result
}

// CanUseSkill 检查是否可以使用技能
// 验证冷却时间和魔法值是否足够
// 参数:
//   - skillID: 技能ID
//   - mana: 当前魔法值
//
// 返回: 是否可使用
func (state *SkillState) CanUseSkill(skillID int32, mana float32) bool {
	state.mu.RLock()
	defer state.mu.RUnlock()

	skill, exists := state.skills[skillID]
	if !exists {
		return false
	}

	if lastUsed, exists := state.cooldowns[skillID]; exists {
		elapsed := time.Since(lastUsed).Seconds()
		if elapsed < float64(skill.Cooldown) {
			return false
		}
	}

	if mana < skill.ManaCost {
		return false
	}

	return true
}

// UseSkill 使用技能
// 更新冷却时间和最后使用时间
// 参数:
//   - skillID: 技能ID
//
// 返回: 是否成功使用
func (state *SkillState) UseSkill(skillID int32) bool {
	state.mu.Lock()
	defer state.mu.Unlock()

	skill, exists := state.skills[skillID]
	if !exists {
		return false
	}

	if lastUsed, exists := state.cooldowns[skillID]; exists {
		elapsed := time.Since(lastUsed).Seconds()
		if elapsed < float64(skill.Cooldown) {
			return false
		}
	}

	state.cooldowns[skillID] = time.Now()
	skill.LastUsed = time.Now()
	return true
}

// UpgradeSkill 升级技能
// 提升技能等级并更新伤害和范围属性
// 参数:
//   - skillID: 技能ID
func (state *SkillState) UpgradeSkill(skillID int32) {
	state.mu.Lock()
	defer state.mu.Unlock()

	skill, exists := state.skills[skillID]
	if !exists {
		return
	}

	skill.Level++

	if config := tables.GetSkillByID(skillID); config != nil {
		skill.Damage = float32(config.Damage) * (1 + 0.2*float32(skill.Level-1))
		skill.Range = float32(config.Range) * (1 + 0.1*float32(skill.Level-1))
	}
}

// SkillComponent 技能组件
// 为游戏对象提供技能系统功能
type SkillComponent struct {
	mu         sync.RWMutex           // 读写锁
	skillState *SkillState            // 技能状态
	owner      gamecommon.IGameObject // 所属游戏对象
}

// NewSkillComponent 创建技能组件
// 参数:
//   - owner: 所属游戏对象
//
// 返回: 技能组件实例
func NewSkillComponent(owner gamecommon.IGameObject) *SkillComponent {
	return &SkillComponent{
		skillState: NewSkillState(),
		owner:      owner,
	}
}

// LearnSkill 学习技能
// 参数:
//   - skillID: 技能ID
func (sc *SkillComponent) LearnSkill(skillID int32) {
	if sc.owner == nil {
		return
	}

	skillConfig := tables.GetSkillByID(skillID)
	if skillConfig == nil {
		return
	}

	sc.skillState.AddSkill(skillID, skillConfig)
}

// UpgradeSkill 升级技能
// 参数:
//   - skillID: 技能ID
func (sc *SkillComponent) UpgradeSkill(skillID int32) {
	sc.skillState.UpgradeSkill(skillID)
}

// CanUseSkill 检查是否可以使用技能
// 参数:
//   - skillID: 技能ID
//
// 返回: 是否可使用
func (sc *SkillComponent) CanUseSkill(skillID int32) bool {
	mana := sc.getOwnerMana()
	return sc.skillState.CanUseSkill(skillID, mana)
}

// UseSkill 使用技能
// 参数:
//   - skillID: 技能ID
//
// 返回: 是否成功使用
func (sc *SkillComponent) UseSkill(skillID int32) bool {
	if !sc.CanUseSkill(skillID) {
		return false
	}

	skill, exists := sc.skillState.GetSkill(skillID)
	if !exists || skill == nil {
		return false
	}

	if sc.skillState.UseSkill(skillID) {
		sc.subtractMana(skill.ManaCost)
		return true
	}

	return false
}

// GetSkill 获取技能
// 参数:
//   - skillID: 技能ID
//
// 返回: 技能和是否存在
func (sc *SkillComponent) GetSkill(skillID int32) (*Skill, bool) {
	return sc.skillState.GetSkill(skillID)
}

// GetAllSkills 获取所有技能
// 返回: 技能映射副本
func (sc *SkillComponent) GetAllSkills() map[int32]*Skill {
	return sc.skillState.GetAllSkills()
}

// Update 更新技能组件
// 参数:
//   - deltaTime: 时间增量
func (sc *SkillComponent) Update(deltaTime float64) {
	// 技能组件不需要定期更新
}

// getOwnerMana 获取所有者魔法值
// 返回: 魔法值
func (sc *SkillComponent) getOwnerMana() float32 {
	if sc.owner == nil {
		return 0
	}
	if propertyComponent := sc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface {
			GetPropertyByType(propType property.PropertyType) float32
		}); ok {
			return prop.GetPropertyByType(property.PropertyMP)
		}
	}
	return 0
}

// subtractMana 扣除魔法值
// 参数:
//   - amount: 扣除数量
func (sc *SkillComponent) subtractMana(amount float32) {
	if sc.owner == nil {
		return
	}
	if propertyComponent := sc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface {
			SubPropertyByType(propType property.PropertyType, value float32)
		}); ok {
			prop.SubPropertyByType(property.PropertyMP, amount)
		}
	}
}

// getOwnerPropertyComponent 获取所有者属性组件
// 返回: 属性组件接口
func (sc *SkillComponent) getOwnerPropertyComponent() interface{} {
	if sc.owner == nil {
		return nil
	}
	return sc.owner.GetComponent("property")
}
