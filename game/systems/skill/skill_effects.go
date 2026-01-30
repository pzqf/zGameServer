package skill

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
)

// SkillEffectType 技能效果类型
type SkillEffectType string

const (
	SkillEffectTypeDamage    SkillEffectType = "damage"    // 伤害
	SkillEffectTypeHeal      SkillEffectType = "heal"      // 治疗
	SkillEffectTypeBuff      SkillEffectType = "buff"      // 增益
	SkillEffectTypeDebuff    SkillEffectType = "debuff"    // 减益
	SkillEffectTypeStun      SkillEffectType = "stun"      // 眩晕
	SkillEffectTypeKnockback SkillEffectType = "knockback" // 击退
	SkillEffectTypeTeleport  SkillEffectType = "teleport"  // 传送
	SkillEffectTypeSummon    SkillEffectType = "summon"    // 召唤
	SkillEffectTypeArea      SkillEffectType = "area"      // 区域效果
	SkillEffectTypeCombo     SkillEffectType = "combo"     // 连击
)

// SkillEffect 技能效果
type SkillEffect struct {
	EffectID   int32                  // 效果ID
	Type       SkillEffectType        // 效果类型
	Value      float32                // 效果值
	Duration   float32                // 持续时间（秒）
	Range      float32                // 效果范围
	TargetType string                 // 目标类型
	Properties map[string]interface{} // 附加属性
}

// SkillCombo 技能连击
type SkillCombo struct {
	ComboID     uint64    // 连击ID
	OwnerID     uint64    // 所有者ID
	SkillIDs    []int32   // 连击技能序列
	CurrentStep int       // 当前步骤
	LastUsed    time.Time // 最后使用时间
	ExpiryTime  time.Time // 过期时间
	Bonus       float32   // 连击奖励系数
}

// SkillEffectSystem 技能效果系统
type SkillEffectSystem struct {
	mu             sync.RWMutex
	effects        map[uint64][]*SkillEffect // 活跃效果
	effectPool     *zObject.GenericPool      // 效果对象池
	combos         map[uint64]*SkillCombo    // 活跃连击
	comboPool      *zObject.GenericPool      // 连击对象池
	effectsBySkill map[int32][]*SkillEffect  // 技能效果映射
}

// GlobalSkillEffectSystem 全局技能效果系统
var GlobalSkillEffectSystem *SkillEffectSystem

// init 初始化全局技能效果系统
func init() {
	GlobalSkillEffectSystem = &SkillEffectSystem{
		effects:        make(map[uint64][]*SkillEffect),
		effectPool:     zObject.NewGenericPool(func() interface{} { return &SkillEffect{} }, 1000),
		combos:         make(map[uint64]*SkillCombo),
		comboPool:      zObject.NewGenericPool(func() interface{} { return &SkillCombo{} }, 1000),
		effectsBySkill: make(map[int32][]*SkillEffect),
	}
}

// Init 初始化技能效果系统
func (ses *SkillEffectSystem) Init() error {
	// 加载技能效果配置
	if err := ses.loadSkillEffects(); err != nil {
		return err
	}
	return nil
}

// loadSkillEffects 加载技能效果配置
func (ses *SkillEffectSystem) loadSkillEffects() error {
	// 从技能配置中加载效果
	// 这里可以根据实际配置格式加载
	return nil
}

// AddSkillEffect 添加技能效果
func (ses *SkillEffectSystem) AddSkillEffect(ownerID uint64, effect *SkillEffect) {
	if effect == nil {
		return
	}

	ses.mu.Lock()
	defer ses.mu.Unlock()

	// 添加到效果列表
	ses.effects[ownerID] = append(ses.effects[ownerID], effect)

	// 应用效果
	ses.applySkillEffect(ownerID, effect)
}

// applySkillEffect 应用技能效果
func (ses *SkillEffectSystem) applySkillEffect(ownerID uint64, effect *SkillEffect) {
	switch effect.Type {
	case SkillEffectTypeDamage:
		ses.applyDamageEffect(ownerID, effect)
	case SkillEffectTypeHeal:
		ses.applyHealEffect(ownerID, effect)
	case SkillEffectTypeBuff:
		ses.applyBuffEffect(ownerID, effect)
	case SkillEffectTypeDebuff:
		ses.applyDebuffEffect(ownerID, effect)
	case SkillEffectTypeStun:
		ses.applyStunEffect(ownerID, effect)
	case SkillEffectTypeKnockback:
		ses.applyKnockbackEffect(ownerID, effect)
	case SkillEffectTypeTeleport:
		ses.applyTeleportEffect(ownerID, effect)
	case SkillEffectTypeSummon:
		ses.applySummonEffect(ownerID, effect)
	case SkillEffectTypeArea:
		ses.applyAreaEffect(ownerID, effect)
	case SkillEffectTypeCombo:
		ses.applyComboEffect(ownerID, effect)
	}
}

// applyDamageEffect 应用伤害效果
func (ses *SkillEffectSystem) applyDamageEffect(ownerID uint64, effect *SkillEffect) {
	// 应用伤害
	// 这里需要调用战斗系统
}

// applyHealEffect 应用治疗效果
func (ses *SkillEffectSystem) applyHealEffect(ownerID uint64, effect *SkillEffect) {
	// 应用治疗
	// 这里需要调用属性系统
}

// applyBuffEffect 应用增益效果
func (ses *SkillEffectSystem) applyBuffEffect(ownerID uint64, effect *SkillEffect) {
	// 应用增益
	// 这里需要调用Buff系统
}

// applyDebuffEffect 应用减益效果
func (ses *SkillEffectSystem) applyDebuffEffect(ownerID uint64, effect *SkillEffect) {
	// 应用减益
	// 这里需要调用Buff系统
}

// applyStunEffect 应用眩晕效果
func (ses *SkillEffectSystem) applyStunEffect(ownerID uint64, effect *SkillEffect) {
	// 应用眩晕
	// 这里需要调用状态系统
}

// applyKnockbackEffect 应用击退效果
func (ses *SkillEffectSystem) applyKnockbackEffect(ownerID uint64, effect *SkillEffect) {
	// 应用击退
	// 这里需要调用移动系统
}

// applyTeleportEffect 应用传送效果
func (ses *SkillEffectSystem) applyTeleportEffect(ownerID uint64, effect *SkillEffect) {
	// 应用传送
	// 这里需要调用移动系统
}

// applySummonEffect 应用召唤效果
func (ses *SkillEffectSystem) applySummonEffect(ownerID uint64, effect *SkillEffect) {
	// 应用召唤
	// 这里需要调用对象系统
}

// applyAreaEffect 应用区域效果
func (ses *SkillEffectSystem) applyAreaEffect(ownerID uint64, effect *SkillEffect) {
	// 应用区域效果
	// 这里需要调用地图系统
}

// applyComboEffect 应用连击效果
func (ses *SkillEffectSystem) applyComboEffect(ownerID uint64, effect *SkillEffect) {
	// 应用连击
	ses.processCombo(ownerID, effect)
}

// StartCombo 开始连击
func (ses *SkillEffectSystem) StartCombo(ownerID uint64, skillIDs []int32) uint64 {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	combo := ses.comboPool.Get().(*SkillCombo)
	comboID := uint64(time.Now().UnixNano())
	combo.ComboID = comboID
	combo.OwnerID = ownerID
	combo.SkillIDs = skillIDs
	combo.CurrentStep = 0
	combo.LastUsed = time.Now()
	combo.ExpiryTime = time.Now().Add(5 * time.Second) // 5秒过期
	combo.Bonus = 1.0

	ses.combos[comboID] = combo
	return comboID
}

// ContinueCombo 继续连击
func (ses *SkillEffectSystem) ContinueCombo(ownerID uint64, skillID int32) bool {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	// 查找活跃的连击
	var targetCombo *SkillCombo
	for _, combo := range ses.combos {
		if combo.OwnerID == ownerID && time.Now().Before(combo.ExpiryTime) {
			targetCombo = combo
			break
		}
	}

	if targetCombo == nil {
		// 没有活跃连击，开始新连击
		comboID := ses.StartCombo(ownerID, []int32{skillID})
		combo := ses.combos[comboID]
		combo.CurrentStep = 1
		combo.LastUsed = time.Now()
		combo.ExpiryTime = time.Now().Add(5 * time.Second)
		return true
	}

	// 检查是否是连击的下一个技能
	if targetCombo.CurrentStep < len(targetCombo.SkillIDs) && targetCombo.SkillIDs[targetCombo.CurrentStep] == skillID {
		// 继续连击
		targetCombo.CurrentStep++
		targetCombo.LastUsed = time.Now()
		targetCombo.ExpiryTime = time.Now().Add(5 * time.Second)
		targetCombo.Bonus += 0.1 // 每次连击增加10%伤害

		// 检查是否完成连击
		if targetCombo.CurrentStep >= len(targetCombo.SkillIDs) {
			// 连击完成，应用奖励
			ses.applyComboBonus(ownerID, targetCombo)
		}

		return true
	}

	// 不是连击的下一个技能，开始新连击
	ses.StartCombo(ownerID, []int32{skillID})
	return false
}

// processCombo 处理连击
func (ses *SkillEffectSystem) processCombo(ownerID uint64, effect *SkillEffect) {
	ses.ContinueCombo(ownerID, effect.Properties["skill_id"].(int32))
}

// applyComboBonus 应用连击奖励
func (ses *SkillEffectSystem) applyComboBonus(ownerID uint64, combo *SkillCombo) {
	// 应用连击奖励
	// 这里可以增加伤害、减少冷却等
}

// GetSkillEffects 获取技能效果
func (ses *SkillEffectSystem) GetSkillEffects(skillID int32) []*SkillEffect {
	ses.mu.RLock()
	defer ses.mu.RUnlock()

	effects, exists := ses.effectsBySkill[skillID]
	if !exists {
		return nil
	}

	// 创建副本
	effectsCopy := make([]*SkillEffect, len(effects))
	for i, effect := range effects {
		effectsCopy[i] = effect
	}

	return effectsCopy
}

// AddSkillEffectToSkill 为技能添加效果
func (ses *SkillEffectSystem) AddSkillEffectToSkill(skillID int32, effect *SkillEffect) {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	ses.effectsBySkill[skillID] = append(ses.effectsBySkill[skillID], effect)
}

// Update 更新技能效果系统
func (ses *SkillEffectSystem) Update() {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	currentTime := time.Now()

	// 清理过期的效果
	for ownerID, effects := range ses.effects {
		validEffects := make([]*SkillEffect, 0)
		for _, effect := range effects {
			if effect.Duration <= 0 || currentTime.Sub(time.Time{}).Seconds() < float64(effect.Duration) {
				validEffects = append(validEffects, effect)
			} else {
				ses.effectPool.Put(effect)
			}
		}

		if len(validEffects) > 0 {
			ses.effects[ownerID] = validEffects
		} else {
			delete(ses.effects, ownerID)
		}
	}

	// 清理过期的连击
	expiredCombos := make([]uint64, 0)
	for comboID, combo := range ses.combos {
		if currentTime.After(combo.ExpiryTime) {
			expiredCombos = append(expiredCombos, comboID)
		}
	}

	for _, comboID := range expiredCombos {
		combo := ses.combos[comboID]
		ses.comboPool.Put(combo)
		delete(ses.combos, comboID)
	}
}

// CleanupExpiredEffects 清理过期效果
func (ses *SkillEffectSystem) CleanupExpiredEffects() {
	ses.Update()
}
