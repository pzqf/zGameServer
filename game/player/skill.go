package player

import (
	"time"

	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// 技能类型定义
const (
	SkillTypeActive   = 1 // 主动技能
	SkillTypePassive  = 2 // 被动技能
	SkillTypeUltimate = 3 // 终极技能
)

// 技能状态定义
const (
	SkillStatusLocked   = 1 // 未解锁
	SkillStatusUnlocked = 2 // 已解锁
	SkillStatusUpgraded = 3 // 已升级
)

// 技能效果类型定义
const (
	SkillEffectTypeDamage   = 1 // 伤害
	SkillEffectTypeDefense  = 2 // 防御力提升
	SkillEffectTypeHeal     = 3 // 治疗
	SkillEffectTypeSpeed    = 4 // 速度提升
	SkillEffectTypeCritical = 5 // 暴击率提升
	SkillEffectTypeDodge    = 6 // 闪避率提升
)

// 技能目标类型定义
const (
	SkillTargetTypeSelf   = 0 // 自身
	SkillTargetTypeSingle = 1 // 单个目标
	SkillTargetTypeAoE    = 2 // 范围目标
)

// SkillEffect 技能效果
type SkillEffect struct {
	effectType int     // 效果类型
	value      float64 // 效果值
	duration   int     // 持续时间（毫秒）
	rangeValue float64 // 作用范围
	targetType int     // 目标类型
}

// Skill 技能结构
type Skill struct {
	skillId      int64
	name         string
	description  string
	skillType    int
	status       int
	level        int
	requireLevel int
	expCost      int64
	goldCost     int64
	effects      []*SkillEffect
	cooldown     int   // 冷却时间（毫秒）
	lastUseTime  int64 // 上次使用时间
}

// SkillManager 技能管理系统
type SkillManager struct {
	playerId int64
	logger   *zap.Logger
	skills   *zMap.Map // key: int64(skillId), value: *Skill
	maxCount int
}

func NewSkillManager(playerId int64, logger *zap.Logger) *SkillManager {
	return &SkillManager{
		playerId: playerId,
		logger:   logger,
		skills:   zMap.NewMap(),
		maxCount: 50, // 最大技能数量
	}
}

func (sm *SkillManager) Init() {
	// 初始化技能管理系统
	sm.logger.Debug("Initializing skill manager", zap.Int64("playerId", sm.playerId))
	// 为新玩家初始化基础技能
	sm.initBasicSkills()
}

// initBasicSkills 初始化基础技能
func (sm *SkillManager) initBasicSkills() {
	// 创建基础攻击技能
	basicAttack := &Skill{
		skillId:      1001,
		name:         "基础攻击",
		description:  "对单个目标造成伤害",
		skillType:    SkillTypeActive,
		status:       SkillStatusUnlocked,
		level:        1,
		requireLevel: 1,
		expCost:      0,
		goldCost:     0,
		effects: []*SkillEffect{
			{
				effectType: 1, // 伤害
				value:      10.0,
				duration:   0,
				rangeValue: 1.5,
				targetType: 1, // 单个目标
			},
		},
		cooldown:    1000,
		lastUseTime: 0,
	}

	// 创建基础防御技能
	basicDefense := &Skill{
		skillId:      1002,
		name:         "基础防御",
		description:  "提高自身防御力",
		skillType:    SkillTypePassive,
		status:       SkillStatusUnlocked,
		level:        1,
		requireLevel: 1,
		expCost:      0,
		goldCost:     0,
		effects: []*SkillEffect{
			{
				effectType: 2, // 防御力提升
				value:      5.0,
				duration:   0,
				rangeValue: 0,
				targetType: 0, // 自身
			},
		},
		cooldown:    0,
		lastUseTime: 0,
	}

	// 添加基础技能到技能管理器
	sm.skills.Store(basicAttack.skillId, basicAttack)
	sm.skills.Store(basicDefense.skillId, basicDefense)
}

// LearnSkill 学习技能
func (sm *SkillManager) LearnSkill(skillId int64) error {
	// 检查技能是否已存在
	if _, exists := sm.skills.Get(skillId); exists {
		return nil // 技能已学习
	}

	// 检查是否达到最大技能数量
	if sm.skills.Len() >= int64(sm.maxCount) {
		return nil // 已达到最大技能数量
	}

	// TODO: 从技能配置表获取技能信息
	// skill := GetSkillFromConfig(skillId)

	// 假设我们已经获取了技能信息
	// sm.skills.Store(skillId, skill)

	sm.logger.Info("Skill learned", zap.Int64("skillId", skillId), zap.Int64("playerId", sm.playerId))
	return nil
}

// CanUpgradeSkill 检查技能是否可以升级
func (sm *SkillManager) CanUpgradeSkill(skill *Skill) bool {
	// 简单的升级条件检查：技能等级不能超过玩家等级
	// TODO: 实际项目中应该从玩家对象获取等级并检查其他条件（如金币、经验等）
	// playerLevel := sm.GetPlayerLevel()
	// if skill.level >= playerLevel {
	//     return false
	// }

	// 暂时允许升级
	return true
}

// UpdateSkillEffects 更新技能效果
func (sm *SkillManager) UpdateSkillEffects(skill *Skill) {
	// 根据技能等级更新技能效果
	// 简单实现：每个等级效果值增加10%
	for _, effect := range skill.effects {
		effect.value *= 1.1
	}
}

// UpgradeSkill 升级技能
func (sm *SkillManager) UpgradeSkill(skillId int64) error {
	// 获取技能
	skillInterface, exists := sm.skills.Get(skillId)
	if !exists {
		return nil // 技能不存在
	}

	skill := skillInterface.(*Skill)
	if skill.status == SkillStatusLocked {
		return nil // 技能未解锁
	}

	// 检查升级条件
	if !sm.CanUpgradeSkill(skill) {
		return nil // 不满足升级条件
	}

	// 升级技能
	skill.level++
	skill.status = SkillStatusUpgraded

	// 更新技能效果
	sm.UpdateSkillEffects(skill)

	sm.skills.Store(skillId, skill)
	sm.logger.Info("Skill upgraded", zap.Int64("skillId", skillId), zap.Int("level", skill.level), zap.Int64("playerId", sm.playerId))
	return nil
}

// UseSkill 使用技能
func (sm *SkillManager) UseSkill(skillId int64, targetId int64) error {
	// 获取技能
	skillInterface, exists := sm.skills.Get(skillId)
	if !exists {
		return nil // 技能不存在
	}

	skill := skillInterface.(*Skill)
	if skill.status == SkillStatusLocked {
		return nil // 技能未解锁
	}

	// 检查技能冷却时间
	currentTime := time.Now().UnixMilli()
	if currentTime-skill.lastUseTime < int64(skill.cooldown) {
		return nil // 技能正在冷却中
	}

	// TODO: 检查技能使用条件（如能量、魔法值等）
	// if !sm.CanUseSkill(skill) {
	//     return nil
	// }

	// 使用技能
	skill.lastUseTime = currentTime
	// TODO: 应用技能效果

	sm.skills.Store(skillId, skill)
	sm.logger.Info("Skill used", zap.Int64("skillId", skillId), zap.Int64("targetId", targetId), zap.Int64("playerId", sm.playerId))
	return nil
}

// GetSkill 获取技能信息
func (sm *SkillManager) GetSkill(skillId int64) (*Skill, bool) {
	skill, exists := sm.skills.Get(skillId)
	if !exists {
		return nil, false
	}
	return skill.(*Skill), true
}

// GetAllSkills 获取所有技能
func (sm *SkillManager) GetAllSkills() []*Skill {
	var skills []*Skill
	sm.skills.Range(func(key, value interface{}) bool {
		if value != nil {
			skills = append(skills, value.(*Skill))
		}
		return true
	})
	return skills
}

// GetUnlockedSkills 获取已解锁技能
func (sm *SkillManager) GetUnlockedSkills() []*Skill {
	var skills []*Skill
	sm.skills.Range(func(key, value interface{}) bool {
		if value != nil {
			skill := value.(*Skill)
			if skill.status != SkillStatusLocked {
				skills = append(skills, skill)
			}
		}
		return true
	})
	return skills
}

// GetActiveSkills 获取主动技能
func (sm *SkillManager) GetActiveSkills() []*Skill {
	var skills []*Skill
	sm.skills.Range(func(key, value interface{}) bool {
		if value != nil {
			skill := value.(*Skill)
			if skill.skillType == SkillTypeActive && skill.status != SkillStatusLocked {
				skills = append(skills, skill)
			}
		}
		return true
	})
	return skills
}
