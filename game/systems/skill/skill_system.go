package skill

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/config/models"
	"github.com/pzqf/zGameServer/config/tables"
	"github.com/pzqf/zGameServer/game/common"
)

// Skill 技能数据结构
type Skill struct {
	ID          int32     // 技能ID
	Level       int32     // 技能等级
	Name        string    // 技能名称
	Description string    // 技能描述
	Cooldown    float32   // 冷却时间（秒）
	Range       float32   // 技能范围
	Damage      float32   // 技能伤害
	ManaCost    float32   // 魔法消耗
	Type        string    // 技能类型：主动、被动
	TargetType  string    // 目标类型：单体、群体、自身
	LastUsed    time.Time // 最后使用时间
}

// SkillState 技能状态
type SkillState struct {
	mu        sync.RWMutex
	skills    map[int32]*Skill
	cooldowns map[int32]time.Time
	skillPool *zObject.GenericPool
}

func NewSkillState() *SkillState {
	return &SkillState{
		skills:    make(map[int32]*Skill),
		cooldowns: make(map[int32]time.Time),
		skillPool: zObject.NewGenericPool(func() interface{} { return &Skill{} }, 100),
	}
}

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

func (state *SkillState) RemoveSkill(skillID int32) {
	state.mu.Lock()
	defer state.mu.Unlock()
	if skill, exists := state.skills[skillID]; exists {
		state.skillPool.Put(skill)
		delete(state.skills, skillID)
	}
	delete(state.cooldowns, skillID)
}

func (state *SkillState) GetSkill(skillID int32) (*Skill, bool) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	skill, exists := state.skills[skillID]
	return skill, exists
}

func (state *SkillState) GetAllSkills() map[int32]*Skill {
	state.mu.RLock()
	defer state.mu.RUnlock()
	result := make(map[int32]*Skill, len(state.skills))
	for k, v := range state.skills {
		result[k] = v
	}
	return result
}

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
type SkillComponent struct {
	mu         sync.RWMutex
	skillState *SkillState
	owner      common.IGameObject
}

func NewSkillComponent(owner common.IGameObject) *SkillComponent {
	return &SkillComponent{
		skillState: NewSkillState(),
		owner:      owner,
	}
}

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

func (sc *SkillComponent) UpgradeSkill(skillID int32) {
	sc.skillState.UpgradeSkill(skillID)
}

func (sc *SkillComponent) CanUseSkill(skillID int32) bool {
	mana := sc.getOwnerMana()
	return sc.skillState.CanUseSkill(skillID, mana)
}

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

func (sc *SkillComponent) GetSkill(skillID int32) (*Skill, bool) {
	return sc.skillState.GetSkill(skillID)
}

func (sc *SkillComponent) GetAllSkills() map[int32]*Skill {
	return sc.skillState.GetAllSkills()
}

func (sc *SkillComponent) Update(deltaTime float64) {
	// 技能组件不需要定期更新
}

func (sc *SkillComponent) getOwnerMana() float32 {
	if sc.owner == nil {
		return 0
	}
	if propertyComponent := sc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface{ GetProperty(name string) float32 }); ok {
			return prop.GetProperty("mana")
		}
	}
	return 0
}

func (sc *SkillComponent) subtractMana(amount float32) {
	if sc.owner == nil {
		return
	}
	if propertyComponent := sc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface {
			SubProperty(name string, value float32)
		}); ok {
			prop.SubProperty("mana", amount)
		}
	}
}

func (sc *SkillComponent) getOwnerPropertyComponent() interface{} {
	if sc.owner == nil {
		return nil
	}
	return sc.owner.GetComponent("property")
}
