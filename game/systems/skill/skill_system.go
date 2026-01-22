package skill

import (
	"sync"
	"time"

	"github.com/pzqf/zGameServer/config/tables"
	"github.com/pzqf/zGameServer/game/systems/property"
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
	ownerID   uint64
	skills    map[int32]*Skill
	cooldowns map[int32]time.Time
}

// SkillSystem 技能系统
type SkillSystem struct {
	mu          sync.RWMutex
	skillStates map[uint64]*SkillState
}

// GlobalSkillSystem 全局技能系统实例
var GlobalSkillSystem *SkillSystem

// init 初始化全局技能系统
func init() {
	GlobalSkillSystem = &SkillSystem{
		skillStates: make(map[uint64]*SkillState),
	}
}

// LearnSkill 学习技能
func (ss *SkillSystem) LearnSkill(ownerID uint64, skillID int32) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// 从配置表获取技能信息
	skSkill := tables.GetSkillByID(skillID)
	if skSkill == nil {
		return
	}

	// 确保技能状态存在
	if _, exists := ss.skillStates[ownerID]; !exists {
		ss.skillStates[ownerID] = &SkillState{
			ownerID:   ownerID,
			skills:    make(map[int32]*Skill),
			cooldowns: make(map[int32]time.Time),
		}
	}

	// 检查是否已经学习了该技能
	if _, exists := ss.skillStates[ownerID].skills[skillID]; exists {
		return
	}

	// 创建新技能
	// 转换技能类型为字符串表示
	skillType := "active"
	switch skSkill.Type {
	case 1:
		skillType = "active"
	case 2:
		skillType = "passive"
	case 3:
		skillType = "heal"
	}

	// 判断目标类型
	targetType := "single"
	if skSkill.AreaRadius > 0 {
		targetType = "area"
	}

	sk := &Skill{
		ID:          skillID,
		Level:       1,
		Name:        skSkill.Name,
		Description: skSkill.Description,
		Cooldown:    float32(skSkill.Cooldown),
		Range:       float32(skSkill.Range),
		Damage:      float32(skSkill.Damage),
		ManaCost:    float32(skSkill.ManaCost),
		Type:        skillType,
		TargetType:  targetType,
		LastUsed:    time.Time{},
	}

	ss.skillStates[ownerID].skills[skillID] = sk
}

// UpgradeSkill 升级技能
func (ss *SkillSystem) UpgradeSkill(ownerID uint64, skillID int32) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// 检查技能状态是否存在
	if _, exists := ss.skillStates[ownerID]; !exists {
		return
	}

	// 检查是否已经学习了该技能
	sk, exists := ss.skillStates[ownerID].skills[skillID]
	if !exists {
		return
	}

	// 增加技能等级
	sk.Level++

	// 从配置表获取技能信息，更新技能属性
	skSkill := tables.GetSkillByID(skillID)
	if skSkill != nil {
		// 根据等级调整技能属性
		sk.Damage = float32(skSkill.Damage) * (1 + 0.2*float32(sk.Level-1))
		sk.Cooldown = float32(skSkill.Cooldown)
		sk.Range = float32(skSkill.Range) * (1 + 0.1*float32(sk.Level-1))
	}
}

// CanUseSkill 检查是否可以使用技能
func (ss *SkillSystem) CanUseSkill(ownerID uint64, skillID int32) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	// 检查技能状态是否存在
	if _, exists := ss.skillStates[ownerID]; !exists {
		return false
	}

	// 检查是否已经学习了该技能
	sk, exists := ss.skillStates[ownerID].skills[skillID]
	if !exists {
		return false
	}

	// 检查冷却时间
	if lastUsed, exists := ss.skillStates[ownerID].cooldowns[skillID]; exists {
		elapsed := time.Since(lastUsed).Seconds()
		if elapsed < float64(sk.Cooldown) {
			return false
		}
	}

	// 检查魔法消耗
	mana := property.GlobalPropertySystem.GetProperty(ownerID, "mana")
	if mana < sk.ManaCost {
		return false
	}

	return true
}

// UseSkill 使用技能
func (ss *SkillSystem) UseSkill(ownerID uint64, skillID int32) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// 检查是否可以使用技能
	if _, exists := ss.skillStates[ownerID]; !exists {
		return false
	}

	sk, exists := ss.skillStates[ownerID].skills[skillID]
	if !exists {
		return false
	}

	// 检查冷却时间和魔法消耗
	if !ss.checkSkillConditionsLocked(ownerID, sk) {
		return false
	}

	// 设置冷却时间
	ss.skillStates[ownerID].cooldowns[skillID] = time.Now()
	sk.LastUsed = time.Now()

	// 消耗魔法
	property.GlobalPropertySystem.SubProperty(ownerID, "mana", sk.ManaCost)

	return true
}

// checkSkillConditionsLocked 检查技能使用条件（内部方法，已加锁）
func (ss *SkillSystem) checkSkillConditionsLocked(ownerID uint64, skill *Skill) bool {
	// 检查冷却时间
	if lastUsed, exists := ss.skillStates[ownerID].cooldowns[skill.ID]; exists {
		elapsed := time.Since(lastUsed).Seconds()
		if elapsed < float64(skill.Cooldown) {
			return false
		}
	}

	// 检查魔法消耗
	mana := property.GlobalPropertySystem.GetProperty(ownerID, "mana")
	if mana < skill.ManaCost {
		return false
	}

	return true
}

// GetSkill 获取技能
func (ss *SkillSystem) GetSkill(ownerID uint64, skillID int32) (*Skill, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if _, exists := ss.skillStates[ownerID]; !exists {
		return nil, false
	}

	sk, exists := ss.skillStates[ownerID].skills[skillID]
	return sk, exists
}

// GetAllSkills 获取所有技能
func (ss *SkillSystem) GetAllSkills(ownerID uint64) map[int32]*Skill {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if _, exists := ss.skillStates[ownerID]; !exists {
		return make(map[int32]*Skill)
	}

	result := make(map[int32]*Skill, len(ss.skillStates[ownerID].skills))
	for k, v := range ss.skillStates[ownerID].skills {
		result[k] = v
	}
	return result
}
