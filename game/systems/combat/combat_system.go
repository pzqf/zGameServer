package combat

import (
	"sync"
	"time"

	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/systems/property"
)

// CombatState 战斗状态
// 管理战斗相关的状态信息
type CombatState struct {
	mu            sync.RWMutex        // 读写锁
	targetID      common.ObjectIdType // 当前目标ID
	inCombat      bool                // 是否在战斗中
	lastCombat    time.Time           // 最后战斗时间
	combatTimeout time.Duration       // 战斗超时时间
}

// NewCombatState 创建战斗状态
// 返回: 初始化后的战斗状态实例
func NewCombatState() *CombatState {
	return &CombatState{
		combatTimeout: 30 * time.Second,
	}
}

// StartCombat 开始战斗
// 参数:
//   - targetID: 目标ID
func (state *CombatState) StartCombat(targetID common.ObjectIdType) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.targetID = targetID
	state.inCombat = true
	state.lastCombat = time.Now()
}

// EndCombat 结束战斗
func (state *CombatState) EndCombat() {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.inCombat = false
	state.targetID = 0
}

// IsInCombat 检查是否在战斗中
// 返回: 是否在战斗中
func (state *CombatState) IsInCombat() bool {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.inCombat
}

// IsCombatExpired 检查战斗是否超时
// 返回: 是否已超时
func (state *CombatState) IsCombatExpired() bool {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return time.Since(state.lastCombat) > state.combatTimeout
}

// GetTargetID 获取目标ID
// 返回: 目标ID
func (state *CombatState) GetTargetID() common.ObjectIdType {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.targetID
}

// HateManager 仇恨管理器
// 管理对不同目标的仇恨值
type HateManager struct {
	mu       sync.RWMutex                    // 读写锁
	hateList map[common.ObjectIdType]float32 // 仇恨列表（目标ID -> 仇恨值）
}

// NewHateManager 创建仇恨管理器
// 返回: 初始化后的仇恨管理器实例
func NewHateManager() *HateManager {
	return &HateManager{
		hateList: make(map[common.ObjectIdType]float32),
	}
}

// AddHate 增加仇恨
// 参数:
//   - targetID: 目标ID
//   - amount: 增加的仇恨值
func (hm *HateManager) AddHate(targetID common.ObjectIdType, amount float32) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	currentHate := hm.hateList[targetID]
	hm.hateList[targetID] = currentHate + amount
}

// SubHate 减少仇恨
// 参数:
//   - targetID: 目标ID
//   - amount: 减少的仇恨值
func (hm *HateManager) SubHate(targetID common.ObjectIdType, amount float32) {
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

// ClearHate 清除所有仇恨
func (hm *HateManager) ClearHate() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.hateList = make(map[common.ObjectIdType]float32)
}

// GetHighestHateTarget 获取最高仇恨目标
// 返回: 最高仇恨目标ID
func (hm *HateManager) GetHighestHateTarget() common.ObjectIdType {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	highestHate := float32(0)
	highestTarget := common.ObjectIdType(0)

	for targetID, hate := range hm.hateList {
		if hate > highestHate {
			highestHate = hate
			highestTarget = targetID
		}
	}

	return highestTarget
}

// CombatComponent 战斗组件
// 为游戏对象提供战斗系统功能
type CombatComponent struct {
	mu          sync.RWMutex           // 读写锁
	combatState *CombatState           // 战斗状态
	hateManager *HateManager           // 仇恨管理器
	owner       gamecommon.IGameObject // 所属游戏对象
}

// NewCombatComponent 创建战斗组件
// 参数:
//   - owner: 所属游戏对象
//
// 返回: 战斗组件实例
func NewCombatComponent(owner gamecommon.IGameObject) *CombatComponent {
	return &CombatComponent{
		combatState: NewCombatState(),
		hateManager: NewHateManager(),
		owner:       owner,
	}
}

// StartCombat 开始战斗
// 参数:
//   - targetID: 目标ID
func (cc *CombatComponent) StartCombat(targetID common.ObjectIdType) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.combatState.StartCombat(targetID)
	cc.hateManager.AddHate(targetID, 100)
}

// EndCombat 结束战斗
func (cc *CombatComponent) EndCombat() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.combatState.EndCombat()
	cc.hateManager.ClearHate()
}

// IsInCombat 检查是否在战斗中
// 返回: 是否在战斗中
func (cc *CombatComponent) IsInCombat() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.combatState.IsInCombat() && !cc.combatState.IsCombatExpired()
}

// GetCurrentTarget 获取当前目标
// 返回: 当前目标ID
func (cc *CombatComponent) GetCurrentTarget() common.ObjectIdType {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.combatState.GetTargetID()
}

// AddHate 增加仇恨
// 参数:
//   - targetID: 目标ID
//   - amount: 增加的仇恨值
func (cc *CombatComponent) AddHate(targetID common.ObjectIdType, amount float32) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.hateManager.AddHate(targetID, amount)
}

// SubHate 减少仇恨
// 参数:
//   - targetID: 目标ID
//   - amount: 减少的仇恨值
func (cc *CombatComponent) SubHate(targetID common.ObjectIdType, amount float32) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.hateManager.SubHate(targetID, amount)
}

// GetHighestHateTarget 获取最高仇恨目标
// 返回: 最高仇恨目标ID
func (cc *CombatComponent) GetHighestHateTarget() common.ObjectIdType {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.hateManager.GetHighestHateTarget()
}

// GetAttack 获取攻击力
// 返回: 攻击力数值
func (cc *CombatComponent) GetAttack() float32 {
	if cc.owner == nil {
		return 0
	}
	if propertyComponent := cc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface {
			GetPropertyByType(propType property.PropertyType) float32
		}); ok {
			return prop.GetPropertyByType(property.PropertyPhysicalAttack)
		}
	}
	return 0
}

// GetDefense 获取防御力
// 返回: 防御力数值
func (cc *CombatComponent) GetDefense() float32 {
	if cc.owner == nil {
		return 0
	}
	if propertyComponent := cc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface {
			GetPropertyByType(propType property.PropertyType) float32
		}); ok {
			return prop.GetPropertyByType(property.PropertyPhysicalDefense)
		}
	}
	return 0
}

// CalculateDamage 计算伤害
// 使用攻击力和目标防御力计算实际伤害
// 参数:
//   - target: 目标对象
//
// 返回: 伤害值
func (cc *CombatComponent) CalculateDamage(target gamecommon.IGameObject) float32 {
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

// Update 更新战斗组件
// 检查战斗超时并清理状态
// 参数:
//   - deltaTime: 时间增量
func (cc *CombatComponent) Update(deltaTime float64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.combatState.IsCombatExpired() && cc.combatState.IsInCombat() {
		cc.combatState.EndCombat()
		cc.hateManager.ClearHate()
	}
}

// getOwnerPropertyComponent 获取所有者属性组件
// 返回: 属性组件接口
func (cc *CombatComponent) getOwnerPropertyComponent() interface{} {
	if cc.owner == nil {
		return nil
	}
	return cc.owner.GetComponent("property")
}

// getDefenseFromTarget 获取目标防御力
// 参数:
//   - target: 目标对象
//
// 返回: 防御力数值
func (cc *CombatComponent) getDefenseFromTarget(target gamecommon.IGameObject) float32 {
	if target == nil {
		return 0
	}
	if propertyComponent := cc.getTargetPropertyComponent(target); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface {
			GetPropertyByType(propType property.PropertyType) float32
		}); ok {
			return prop.GetPropertyByType(property.PropertyPhysicalDefense)
		}
	}
	return 0
}

// getTargetPropertyComponent 获取目标属性组件
// 参数:
//   - target: 目标对象
//
// 返回: 属性组件接口
func (cc *CombatComponent) getTargetPropertyComponent(target gamecommon.IGameObject) interface{} {
	if target == nil {
		return nil
	}
	return target.GetComponent("property")
}
