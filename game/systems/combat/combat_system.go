package combat

import (
	"sync"
	"time"

	"github.com/pzqf/zGameServer/game/common"
)

// CombatState 战斗状态
type CombatState struct {
	mu            sync.RWMutex
	targetID      uint64
	inCombat      bool
	lastCombat    time.Time
	combatTimeout time.Duration
}

func NewCombatState() *CombatState {
	return &CombatState{
		combatTimeout: 30 * time.Second,
	}
}

func (state *CombatState) StartCombat(targetID uint64) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.targetID = targetID
	state.inCombat = true
	state.lastCombat = time.Now()
}

func (state *CombatState) EndCombat() {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.inCombat = false
	state.targetID = 0
}

func (state *CombatState) IsInCombat() bool {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.inCombat
}

func (state *CombatState) IsCombatExpired() bool {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return time.Since(state.lastCombat) > state.combatTimeout
}

func (state *CombatState) GetTargetID() uint64 {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.targetID
}

// HateManager 仇恨管理
type HateManager struct {
	mu       sync.RWMutex
	hateList map[uint64]float32 // key: targetID, value: hateValue
}

func NewHateManager() *HateManager {
	return &HateManager{
		hateList: make(map[uint64]float32),
	}
}

func (hm *HateManager) AddHate(targetID uint64, amount float32) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	currentHate := hm.hateList[targetID]
	hm.hateList[targetID] = currentHate + amount
}

func (hm *HateManager) SubHate(targetID uint64, amount float32) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	currentHate, exists := hm.hateList[targetID]
	if !exists {
		return
	}

	newHate := currentHate - amount
	if newHate <= 0 {
		delete(hm.hateList, targetID)
	} else {
		hm.hateList[targetID] = newHate
	}
}

func (hm *HateManager) ClearHate() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.hateList = make(map[uint64]float32)
}

func (hm *HateManager) GetHighestHateTarget() uint64 {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	highestHate := float32(0)
	highestTarget := uint64(0)

	for targetID, hate := range hm.hateList {
		if hate > highestHate {
			highestHate = hate
			highestTarget = targetID
		}
	}

	return highestTarget
}

// CombatComponent 战斗组件
type CombatComponent struct {
	mu          sync.RWMutex
	combatState *CombatState
	hateManager *HateManager
	owner       common.IGameObject
}

func NewCombatComponent(owner common.IGameObject) *CombatComponent {
	return &CombatComponent{
		combatState: NewCombatState(),
		hateManager: NewHateManager(),
		owner:       owner,
	}
}

func (cc *CombatComponent) StartCombat(targetID uint64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.combatState.StartCombat(targetID)
	cc.hateManager.AddHate(targetID, 100)
}

func (cc *CombatComponent) EndCombat() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.combatState.EndCombat()
	cc.hateManager.ClearHate()
}

func (cc *CombatComponent) IsInCombat() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.combatState.IsInCombat() && !cc.combatState.IsCombatExpired()
}

func (cc *CombatComponent) GetCurrentTarget() uint64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.combatState.GetTargetID()
}

func (cc *CombatComponent) AddHate(targetID uint64, amount float32) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.hateManager.AddHate(targetID, amount)
}

func (cc *CombatComponent) SubHate(targetID uint64, amount float32) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.hateManager.SubHate(targetID, amount)
}

func (cc *CombatComponent) GetHighestHateTarget() uint64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.hateManager.GetHighestHateTarget()
}

func (cc *CombatComponent) GetAttack() float32 {
	if cc.owner == nil {
		return 0
	}
	if propertyComponent := cc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface{ GetProperty(name string) float32 }); ok {
			return prop.GetProperty("attack")
		}
	}
	return 0
}

func (cc *CombatComponent) GetDefense() float32 {
	if cc.owner == nil {
		return 0
	}
	if propertyComponent := cc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface{ GetProperty(name string) float32 }); ok {
			return prop.GetProperty("defense")
		}
	}
	return 0
}

func (cc *CombatComponent) CalculateDamage(target common.IGameObject) float32 {
	if target == nil || cc.owner == nil {
		return 0
	}

	attack := cc.GetAttack()
	defense := cc.getDefenseFromTarget(target)

	damageValue := float64(attack) - float64(defense)*0.5
	if damageValue < 1.0 {
		damageValue = 1.0
	}

	return float32(damageValue)
}

func (cc *CombatComponent) Update(deltaTime float64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.combatState.IsCombatExpired() && cc.combatState.IsInCombat() {
		cc.combatState.EndCombat()
		cc.hateManager.ClearHate()
	}
}

func (cc *CombatComponent) getOwnerPropertyComponent() interface{} {
	if cc.owner == nil {
		return nil
	}
	return cc.owner.GetComponent("property")
}

func (cc *CombatComponent) getDefenseFromTarget(target common.IGameObject) float32 {
	if target == nil {
		return 0
	}
	if propertyComponent := cc.getTargetPropertyComponent(target); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface{ GetProperty(name string) float32 }); ok {
			return prop.GetProperty("defense")
		}
	}
	return 0
}

func (cc *CombatComponent) getTargetPropertyComponent(target common.IGameObject) interface{} {
	if target == nil {
		return nil
	}
	return target.GetComponent("property")
}
