package combat

import (
	"sync"
	"time"

	"github.com/pzqf/zGameServer/game/systems/property"
)

// CombatState 战斗状态
type CombatState struct {
	ownerID       uint64
	targetID      uint64
	inCombat      bool
	lastCombat    time.Time
	combatTimeout time.Duration
}

// CombatSystem 战斗系统
type CombatSystem struct {
	mu           sync.RWMutex
	combatStates map[uint64]*CombatState
	hateList     map[uint64]map[uint64]float32 // key: attackerID, value: map[targetID]hateValue
}

// GlobalCombatSystem 全局战斗系统实例
var GlobalCombatSystem *CombatSystem

// init 初始化全局战斗系统
func init() {
	GlobalCombatSystem = &CombatSystem{
		combatStates: make(map[uint64]*CombatState),
		hateList:     make(map[uint64]map[uint64]float32),
	}
}

// GetAttack 获取攻击力
func (cs *CombatSystem) GetAttack(ownerID uint64) float32 {
	return property.GlobalPropertySystem.GetProperty(ownerID, "attack")
}

// GetDefense 获取防御力
func (cs *CombatSystem) GetDefense(ownerID uint64) float32 {
	return property.GlobalPropertySystem.GetProperty(ownerID, "defense")
}

// StartCombat 开始战斗
func (cs *CombatSystem) StartCombat(attackerID, targetID uint64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// 设置攻击者的战斗状态
	if _, exists := cs.combatStates[attackerID]; !exists {
		cs.combatStates[attackerID] = &CombatState{
			ownerID:       attackerID,
			combatTimeout: 30 * time.Second,
		}
	}

	cs.combatStates[attackerID].targetID = targetID
	cs.combatStates[attackerID].inCombat = true
	cs.combatStates[attackerID].lastCombat = time.Now()

	// 设置目标的战斗状态
	if _, exists := cs.combatStates[targetID]; !exists {
		cs.combatStates[targetID] = &CombatState{
			ownerID:       targetID,
			combatTimeout: 30 * time.Second,
		}
	}

	cs.combatStates[targetID].targetID = attackerID
	cs.combatStates[targetID].inCombat = true
	cs.combatStates[targetID].lastCombat = time.Now()

	// 增加仇恨
	cs.addHate(attackerID, targetID, 100) // 基础仇恨值
}

// EndCombat 结束战斗
func (cs *CombatSystem) EndCombat(ownerID uint64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if state, exists := cs.combatStates[ownerID]; exists {
		state.inCombat = false
		state.targetID = 0
	}

	// 清除仇恨
	delete(cs.hateList, ownerID)
}

// IsInCombat 检查是否在战斗中
func (cs *CombatSystem) IsInCombat(ownerID uint64) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if state, exists := cs.combatStates[ownerID]; exists {
		// 检查战斗超时
		if time.Since(state.lastCombat) > state.combatTimeout {
			state.inCombat = false
			state.targetID = 0
			return false
		}
		return state.inCombat
	}
	return false
}

// GetCurrentTarget 获取当前目标
func (cs *CombatSystem) GetCurrentTarget(ownerID uint64) uint64 {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if state, exists := cs.combatStates[ownerID]; exists && state.inCombat {
		return state.targetID
	}
	return 0
}

// AddHate 增加仇恨
func (cs *CombatSystem) AddHate(attackerID, targetID uint64, amount float32) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.addHate(attackerID, targetID, amount)
}

// addHate 内部方法：增加仇恨
func (cs *CombatSystem) addHate(attackerID, targetID uint64, amount float32) {
	if _, exists := cs.hateList[targetID]; !exists {
		cs.hateList[targetID] = make(map[uint64]float32)
	}

	currentHate := cs.hateList[targetID][attackerID]
	cs.hateList[targetID][attackerID] = currentHate + amount
}

// SubHate 减少仇恨
func (cs *CombatSystem) SubHate(attackerID, targetID uint64, amount float32) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if _, exists := cs.hateList[targetID]; !exists {
		return
	}

	currentHate, exists := cs.hateList[targetID][attackerID]
	if !exists {
		return
	}

	newHate := currentHate - amount
	if newHate <= 0 {
		delete(cs.hateList[targetID], attackerID)
		// 如果仇恨列表为空，删除该目标的仇恨记录
		if len(cs.hateList[targetID]) == 0 {
			delete(cs.hateList, targetID)
		}
	} else {
		cs.hateList[targetID][attackerID] = newHate
	}
}

// ClearHate 清除仇恨
func (cs *CombatSystem) ClearHate(ownerID uint64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.hateList, ownerID)
}

// CalculateDamage 计算伤害
func (cs *CombatSystem) CalculateDamage(attackerID, targetID uint64) float32 {
	attack := cs.GetAttack(attackerID)
	defense := cs.GetDefense(targetID)

	// 伤害公式：攻击 - 防御 * 0.5，最小1点伤害
	damageValue := float64(attack) - float64(defense)*0.5
	if damageValue < 1.0 {
		damageValue = 1.0
	}

	return float32(damageValue)
}
