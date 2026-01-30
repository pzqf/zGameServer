package skill

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/game/common"
)

// SkillBalanceConfig 技能平衡配置
type SkillBalanceConfig struct {
	SkillID         int32              // 技能ID
	BaseDamage      float32            // 基础伤害
	DamageScaling   float32            // 伤害缩放系数
	LevelScaling    float32            // 等级缩放系数
	ManaCost        float32            // 基础魔法消耗
	ManaScaling     float32            // 魔法消耗缩放系数
	Cooldown        float32            // 基础冷却时间
	CooldownScaling float32            // 冷却时间缩放系数
	Range           float32            // 基础范围
	RangeScaling    float32            // 范围缩放系数
	CastTime        float32            // 施法时间
	Interruptible   bool               // 是否可打断
	ResourceCost    map[string]float32 // 其他资源消耗
}

// SkillBalanceComponent 技能平衡组件
type SkillBalanceComponent struct {
	mu             sync.RWMutex
	balanceConfigs map[int32]*SkillBalanceConfig // 平衡配置
	cooldowns      map[int32]*SkillCooldown      // 技能冷却
	resources      map[string]*SkillResource     // 资源管理
	configPool     *zObject.GenericPool          // 配置对象池
	cooldownPool   *zObject.GenericPool          // 冷却对象池
	resourcePool   *zObject.GenericPool          // 资源对象池
	globalCooldown time.Time                     // 全局冷却
	owner          common.IGameObject
}

func NewSkillBalanceComponent(owner common.IGameObject) *SkillBalanceComponent {
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

// SkillCooldown 技能冷却
type SkillCooldown struct {
	SkillID   int32     // 技能ID
	StartTime time.Time // 开始时间
	EndTime   time.Time // 结束时间
	Remaining float32   // 剩余时间（秒）
	Reduction float32   // 冷却缩减
}

// SkillResource 技能资源
type SkillResource struct {
	ResourceType string  // 资源类型
	CurrentValue float32 // 当前值
	MaxValue     float32 // 最大值
	RegenRate    float32 // 回复速率
}

func (sbc *SkillBalanceComponent) CalculateDamage(skillID int32, target common.IGameObject, level int32) float32 {
	sbc.mu.RLock()
	config, exists := sbc.balanceConfigs[skillID]
	sbc.mu.RUnlock()

	if !exists {
		return 100.0 * float32(level)
	}

	attackPower := sbc.getOwnerProperty("attack_power")
	intellect := sbc.getOwnerProperty("intellect")
	skillPower := sbc.getOwnerProperty("skill_power")

	defense := sbc.getTargetProperty(target, "physical_defense")

	baseDamage := config.BaseDamage
	propertyBonus := attackPower*0.6 + intellect*0.3 + skillPower*1.0
	levelBonus := 1.0 + float32(level-1)*config.LevelScaling
	skillBonus := 1.0 + config.DamageScaling

	defenseReduction := 1.0 - (defense / (defense + 1000.0))

	finalDamage := (baseDamage + propertyBonus) * levelBonus * skillBonus * defenseReduction

	if finalDamage < 1.0 {
		finalDamage = 1.0
	}

	return finalDamage
}

func (sbc *SkillBalanceComponent) CalculateManaCost(skillID int32, level int32) float32 {
	sbc.mu.RLock()
	config, exists := sbc.balanceConfigs[skillID]
	sbc.mu.RUnlock()

	if !exists {
		return 10.0 * float32(level)
	}

	intellect := sbc.getOwnerProperty("intellect")
	manaEfficiency := sbc.getOwnerProperty("mana_efficiency")

	baseMana := config.ManaCost
	intellectBonus := 1.0 - (intellect * 0.005)
	efficiencyBonus := 1.0 - (manaEfficiency * 0.1)
	levelBonus := 1.0 + float32(level-1)*config.ManaScaling

	finalMana := baseMana * intellectBonus * efficiencyBonus * levelBonus

	if finalMana < 1.0 {
		finalMana = 1.0
	}

	return finalMana
}

func (sbc *SkillBalanceComponent) CalculateCooldown(skillID int32) float32 {
	sbc.mu.RLock()
	config, exists := sbc.balanceConfigs[skillID]
	sbc.mu.RUnlock()

	if !exists {
		return 10.0
	}

	haste := sbc.getOwnerProperty("haste")
	cooldownReduction := sbc.getOwnerProperty("cooldown_reduction")

	baseCooldown := config.Cooldown
	hasteReduction := 1.0 - (haste * 0.01)
	cdrReduction := 1.0 - (cooldownReduction * 0.001)
	finalCooldown := baseCooldown * hasteReduction * cdrReduction

	if finalCooldown < 0.5 {
		finalCooldown = 0.5
	}

	return finalCooldown
}

func (sbc *SkillBalanceComponent) CalculateRange(skillID int32) float32 {
	sbc.mu.RLock()
	config, exists := sbc.balanceConfigs[skillID]
	sbc.mu.RUnlock()

	if !exists {
		return 5.0
	}

	rangeBonus := sbc.getOwnerProperty("range_bonus")

	baseRange := config.Range
	finalRange := baseRange + rangeBonus

	return finalRange
}

func (sbc *SkillBalanceComponent) Update(deltaTime float64) {
	sbc.mu.Lock()
	defer sbc.mu.Unlock()

	for skillID, cooldown := range sbc.cooldowns {
		if time.Since(cooldown.EndTime) >= 0 {
			sbc.cooldownPool.Put(cooldown)
			delete(sbc.cooldowns, skillID)
		}
	}
}

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

func (sbc *SkillBalanceComponent) getTargetProperty(target common.IGameObject, name string) float32 {
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

func (sbc *SkillBalanceComponent) getOwnerPropertyComponent() interface{} {
	if sbc.owner == nil {
		return nil
	}
	return sbc.owner.GetComponent("property")
}

func (sbc *SkillBalanceComponent) getTargetPropertyComponent(target common.IGameObject) interface{} {
	if target == nil {
		return nil
	}
	return target.GetComponent("property")
}
