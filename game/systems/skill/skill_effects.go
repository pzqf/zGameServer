package skill

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/common"
)

type SkillEffectType string

const (
	SkillEffectTypeDamage    SkillEffectType = "damage"
	SkillEffectTypeHeal      SkillEffectType = "heal"
	SkillEffectTypeBuff      SkillEffectType = "buff"
	SkillEffectTypeDebuff    SkillEffectType = "debuff"
	SkillEffectTypeStun      SkillEffectType = "stun"
	SkillEffectTypeKnockback SkillEffectType = "knockback"
	SkillEffectTypeTeleport  SkillEffectType = "teleport"
	SkillEffectTypeSummon    SkillEffectType = "summon"
	SkillEffectTypeArea      SkillEffectType = "area"
	SkillEffectTypeCombo     SkillEffectType = "combo"
)

type SkillEffect struct {
	EffectID   int32
	Type       SkillEffectType
	Value      float32
	Duration   float32
	Range      float32
	TargetType string
	Properties map[string]interface{}
}

type SkillCombo struct {
	ComboID     common.ComboIdType
	OwnerID     common.ObjectIdType
	SkillIDs    []int32
	CurrentStep int
	LastUsed    time.Time
	ExpiryTime  time.Time
	Bonus       float32
}

type SkillEffectSystem struct {
	mu             sync.RWMutex
	effects        map[common.ObjectIdType][]*SkillEffect
	effectPool     *zObject.GenericPool
	combos         map[common.ComboIdType]*SkillCombo
	comboPool      *zObject.GenericPool
	effectsBySkill map[int32][]*SkillEffect
}

var GlobalSkillEffectSystem *SkillEffectSystem

func init() {
	GlobalSkillEffectSystem = &SkillEffectSystem{
		effects:        make(map[common.ObjectIdType][]*SkillEffect),
		effectPool:     zObject.NewGenericPool(func() interface{} { return &SkillEffect{} }, 1000),
		combos:         make(map[common.ComboIdType]*SkillCombo),
		comboPool:      zObject.NewGenericPool(func() interface{} { return &SkillCombo{} }, 1000),
		effectsBySkill: make(map[int32][]*SkillEffect),
	}
}

func (ses *SkillEffectSystem) Init() error {
	if err := ses.loadSkillEffects(); err != nil {
		return err
	}
	return nil
}

func (ses *SkillEffectSystem) loadSkillEffects() error {
	return nil
}

func (ses *SkillEffectSystem) AddSkillEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
	if effect == nil {
		return
	}

	ses.mu.Lock()
	defer ses.mu.Unlock()

	ses.effects[ownerID] = append(ses.effects[ownerID], effect)

	ses.applySkillEffect(ownerID, effect)
}

func (ses *SkillEffectSystem) applySkillEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
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

func (ses *SkillEffectSystem) applyDamageEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
}

func (ses *SkillEffectSystem) applyHealEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
}

func (ses *SkillEffectSystem) applyBuffEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
}

func (ses *SkillEffectSystem) applyDebuffEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
}

func (ses *SkillEffectSystem) applyStunEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
}

func (ses *SkillEffectSystem) applyKnockbackEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
}

func (ses *SkillEffectSystem) applyTeleportEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
}

func (ses *SkillEffectSystem) applySummonEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
}

func (ses *SkillEffectSystem) applyAreaEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
}

func (ses *SkillEffectSystem) applyComboEffect(ownerID common.ObjectIdType, effect *SkillEffect) {
	ses.processCombo(ownerID, effect)
}

func (ses *SkillEffectSystem) StartCombo(ownerID common.ObjectIdType, skillIDs []int32) common.ComboIdType {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	combo := ses.comboPool.Get().(*SkillCombo)
	comboID, err := common.GenerateComboID()
	if err != nil {
		return 0
	}
	combo.ComboID = comboID
	combo.OwnerID = ownerID
	combo.SkillIDs = skillIDs
	combo.CurrentStep = 0
	combo.LastUsed = time.Now()
	combo.ExpiryTime = time.Now().Add(5 * time.Second)
	combo.Bonus = 1.0

	ses.combos[comboID] = combo
	return comboID
}

func (ses *SkillEffectSystem) ContinueCombo(ownerID common.ObjectIdType, skillID int32) bool {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	var targetCombo *SkillCombo
	for _, combo := range ses.combos {
		if combo.OwnerID == ownerID && time.Now().Before(combo.ExpiryTime) {
			targetCombo = combo
			break
		}
	}

	if targetCombo == nil {
		comboID := ses.StartCombo(ownerID, []int32{skillID})
		combo := ses.combos[comboID]
		combo.CurrentStep = 1
		combo.LastUsed = time.Now()
		combo.ExpiryTime = time.Now().Add(5 * time.Second)
		return true
	}

	if targetCombo.CurrentStep < len(targetCombo.SkillIDs) && targetCombo.SkillIDs[targetCombo.CurrentStep] == skillID {
		targetCombo.CurrentStep++
		targetCombo.LastUsed = time.Now()
		targetCombo.ExpiryTime = time.Now().Add(5 * time.Second)
		targetCombo.Bonus += 0.1

		if targetCombo.CurrentStep >= len(targetCombo.SkillIDs) {
			ses.applyComboBonus(ownerID, targetCombo)
		}

		return true
	}

	ses.StartCombo(ownerID, []int32{skillID})
	return false
}

func (ses *SkillEffectSystem) processCombo(ownerID common.ObjectIdType, effect *SkillEffect) {
	ses.ContinueCombo(ownerID, effect.Properties["skill_id"].(int32))
}

func (ses *SkillEffectSystem) applyComboBonus(ownerID common.ObjectIdType, combo *SkillCombo) {
}

func (ses *SkillEffectSystem) GetSkillEffects(skillID int32) []*SkillEffect {
	ses.mu.RLock()
	defer ses.mu.RUnlock()

	effects, exists := ses.effectsBySkill[skillID]
	if !exists {
		return nil
	}

	effectsCopy := make([]*SkillEffect, len(effects))
	for i, effect := range effects {
		effectsCopy[i] = effect
	}

	return effectsCopy
}

func (ses *SkillEffectSystem) AddSkillEffectToSkill(skillID int32, effect *SkillEffect) {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	ses.effectsBySkill[skillID] = append(ses.effectsBySkill[skillID], effect)
}

func (ses *SkillEffectSystem) Update() {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	currentTime := time.Now()

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

	expiredCombos := make([]common.ComboIdType, 0)
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

func (ses *SkillEffectSystem) CleanupExpiredEffects() {
	ses.Update()
}
