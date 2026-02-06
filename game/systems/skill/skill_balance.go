package skill

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	gamecommon "github.com/pzqf/zGameServer/game/common"
)

// SkillBalanceConfig 技能平衡配置
// 定义单个技能的各项参数，用于计算伤害、消耗、冷却等
type SkillBalanceConfig struct {
	SkillID         int32              // 技能ID
	BaseDamage      float32            // 基础伤害
	DamageScaling   float32            // 伤害缩放系数（与技能等级相关）
	LevelScaling    float32            // 等级缩放系数（与角色等级相关）
	ManaCost        float32            // 基础魔法消耗
	ManaScaling     float32            // 魔法消耗缩放系数
	Cooldown        float32            // 基础冷却时间（秒）
	CooldownScaling float32            // 冷却时间缩放系数
	Range           float32            // 基础施法范围
	RangeScaling    float32            // 范围缩放系数
	CastTime        float32            // 施法时间（秒）
	Interruptible   bool               // 是否可被打断
	ResourceCost    map[string]float32 // 其他资源消耗（如能量、怒气等）
}

// SkillBalanceComponent 技能平衡组件
// 管理技能的冷却时间、资源消耗和伤害计算
type SkillBalanceComponent struct {
	mu             sync.RWMutex
	balanceConfigs map[int32]*SkillBalanceConfig // 技能平衡配置表（skillID -> config）
	cooldowns      map[int32]*SkillCooldown      // 技能冷却状态（skillID -> cooldown）
	resources      map[string]*SkillResource     // 资源管理（resourceType -> resource）
	configPool     *zObject.GenericPool          // 配置对象池（减少内存分配）
	cooldownPool   *zObject.GenericPool          // 冷却对象池
	resourcePool   *zObject.GenericPool          // 资源对象池
	globalCooldown time.Time                     // 全局冷却结束时间
	owner          gamecommon.IGameObject        // 所属游戏对象
}

// NewSkillBalanceComponent 创建技能平衡组件
// 参数:
//   - owner: 所属游戏对象
//
// 返回:
//   - *SkillBalanceComponent: 新创建的组件
func NewSkillBalanceComponent(owner gamecommon.IGameObject) *SkillBalanceComponent {
	return &SkillBalanceComponent{
		balanceConfigs: make(map[int32]*SkillBalanceConfig),
		cooldowns:      make(map[int32]*SkillCooldown),
		resources:      make(map[string]*SkillResource),
		configPool:     zObject.NewGenericPool(func() interface{} { return &SkillBalanceConfig{} }, 100),
		cooldownPool:   zObject.NewGenericPool(func() interface{} { return &SkillCooldown{} }, 100),
		resourcePool:   zObject.NewGenericPool(func() interface{} { return &SkillResource{} }, 100),
		owner:          owner,
	}
}

// SkillCooldown 技能冷却状态
// 记录单个技能的冷却信息
type SkillCooldown struct {
	SkillID   int32     // 技能ID
	StartTime time.Time // 冷却开始时间
	EndTime   time.Time // 冷却结束时间
	Remaining float32   // 剩余冷却时间（秒）
	Reduction float32   // 冷却缩减比例（0-1）
}

// SkillResource 技能资源
// 管理某种资源的当前值和回复速率
type SkillResource struct {
	ResourceType string  // 资源类型（如"energy"、"rage"）
	CurrentValue float32 // 当前值
	MaxValue     float32 // 最大值
	RegenRate    float32 // 回复速率（每秒）
}

// CalculateDamage 计算技能伤害
// 综合考虑攻击力、智力、技能强度、防御等属性
// 参数:
//   - skillID: 技能ID
//   - target: 目标对象
//   - level: 技能等级
//
// 返回:
//   - float32: 最终伤害值
func (sbc *SkillBalanceComponent) CalculateDamage(skillID int32, target gamecommon.IGameObject, level int32) float32 {
	sbc.mu.RLock()
	config, exists := sbc.balanceConfigs[skillID]
	sbc.mu.RUnlock()

	// 无配置时使用默认公式
	if !exists {
		return 100.0 * float32(level)
	}

	// 获取施法者属性
	attackPower := sbc.getOwnerProperty("attack_power")
	intellect := sbc.getOwnerProperty("intellect")
	skillPower := sbc.getOwnerProperty("skill_power")

	// 获取目标防御
	defense := sbc.getTargetProperty(target, "physical_defense")

	// 伤害计算公式:
	// 最终伤害 = (基础伤害 + 属性加成) * 等级加成 * 技能加成 * 防御减免
	baseDamage := config.BaseDamage
	propertyBonus := attackPower*0.6 + intellect*0.3 + skillPower*1.0
	levelBonus := 1.0 + float32(level-1)*config.LevelScaling
	skillBonus := 1.0 + config.DamageScaling

	// 防御减免公式: 防御 / (防御 + 1000)
	defenseReduction := 1.0 - (defense / (defense + 1000.0))

	finalDamage := (baseDamage + propertyBonus) * levelBonus * skillBonus * defenseReduction

	// 最低伤害保护
	if finalDamage < 1.0 {
		finalDamage = 1.0
	}

	return finalDamage
}

// CalculateManaCost 计算魔法消耗
// 参数:
//   - skillID: 技能ID
//   - level: 技能等级
//
// 返回:
//   - float32: 魔法消耗值
func (sbc *SkillBalanceComponent) CalculateManaCost(skillID int32, level int32) float32 {
	sbc.mu.RLock()
	config, exists := sbc.balanceConfigs[skillID]
	sbc.mu.RUnlock()

	// 无配置时使用默认公式
	if !exists {
		return 10.0 * float32(level)
	}

	// 获取施法者属性
	intellect := sbc.getOwnerProperty("intellect")
	manaEfficiency := sbc.getOwnerProperty("mana_efficiency")

	// 魔法消耗公式:
	// 最终消耗 = 基础消耗 * 智力减免 * 效率减免 * 等级加成
	baseMana := config.ManaCost
	intellectBonus := 1.0 - (intellect * 0.005)
	efficiencyBonus := 1.0 - (manaEfficiency * 0.1)
	levelBonus := 1.0 + float32(level-1)*config.ManaScaling

	finalMana := baseMana * intellectBonus * efficiencyBonus * levelBonus

	// 最低消耗保护
	if finalMana < 1.0 {
		finalMana = 1.0
	}

	return finalMana
}

// CalculateCooldown 计算冷却时间
// 参数:
//   - skillID: 技能ID
//
// 返回:
//   - float32: 冷却时间（秒）
func (sbc *SkillBalanceComponent) CalculateCooldown(skillID int32) float32 {
	sbc.mu.RLock()
	config, exists := sbc.balanceConfigs[skillID]
	sbc.mu.RUnlock()

	// 无配置时使用默认值
	if !exists {
		return 10.0
	}

	// 获取冷却缩减属性
	haste := sbc.getOwnerProperty("haste")
	cooldownReduction := sbc.getOwnerProperty("cooldown_reduction")

	// 冷却时间公式:
	// 最终冷却 = 基础冷却 * 急速减免 * CDR减免
	baseCooldown := config.Cooldown
	hasteReduction := 1.0 - (haste * 0.01)
	cdrReduction := 1.0 - (cooldownReduction * 0.001)
	finalCooldown := baseCooldown * hasteReduction * cdrReduction

	// 最低冷却保护（0.5秒）
	if finalCooldown < 0.5 {
		finalCooldown = 0.5
	}

	return finalCooldown
}

// CalculateRange 计算施法范围
// 参数:
//   - skillID: 技能ID
//
// 返回:
//   - float32: 施法范围
func (sbc *SkillBalanceComponent) CalculateRange(skillID int32) float32 {
	sbc.mu.RLock()
	config, exists := sbc.balanceConfigs[skillID]
	sbc.mu.RUnlock()

	// 无配置时使用默认值
	if !exists {
		return 5.0
	}

	rangeBonus := sbc.getOwnerProperty("range_bonus")

	baseRange := config.Range
	finalRange := baseRange + rangeBonus

	return finalRange
}

// Update 更新冷却状态
// 每帧调用，清理已完成的冷却
func (sbc *SkillBalanceComponent) Update(deltaTime float64) {
	sbc.mu.Lock()
	defer sbc.mu.Unlock()

	// 清理已完成的冷却
	for skillID, cooldown := range sbc.cooldowns {
		if time.Since(cooldown.EndTime) >= 0 {
			sbc.cooldownPool.Put(cooldown)
			delete(sbc.cooldowns, skillID)
		}
	}
}

// getOwnerProperty 获取所属对象的属性值
func (sbc *SkillBalanceComponent) getOwnerProperty(name string) float32 {
	if sbc.owner == nil {
		return 0
	}
	if propertyComponent := sbc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface{ GetProperty(name string) float32 }); ok {
			return prop.GetProperty(name)
		}
	}
	return 0
}

// getTargetProperty 获取目标对象的属性值
func (sbc *SkillBalanceComponent) getTargetProperty(target gamecommon.IGameObject, name string) float32 {
	if target == nil {
		return 0
	}
	if propertyComponent := sbc.getTargetPropertyComponent(target); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface{ GetProperty(name string) float32 }); ok {
			return prop.GetProperty(name)
		}
	}
	return 0
}

// getOwnerPropertyComponent 获取所属对象的属性组件
func (sbc *SkillBalanceComponent) getOwnerPropertyComponent() interface{} {
	if sbc.owner == nil {
		return nil
	}
	return sbc.owner.GetComponent("property")
}

// getTargetPropertyComponent 获取目标对象的属性组件
func (sbc *SkillBalanceComponent) getTargetPropertyComponent(target gamecommon.IGameObject) interface{} {
	if target == nil {
		return nil
	}
	return target.GetComponent("property")
}
