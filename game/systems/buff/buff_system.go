package buff

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/config/tables"
	"github.com/pzqf/zGameServer/game/common"
)

// BuffType buff类型
const (
	BuffTypePositive = "positive" // 增益效果
	BuffTypeNegative = "negative" // 减益效果
	BuffTypeNeutral  = "neutral"  // 中性效果
)

// Buff buff数据结构
type Buff struct {
	ID          int32     // Buff ID
	Name        string    // Buff名称
	Description string    // Buff描述
	Type        string    // Buff类型：增益、减益、中性
	Duration    float32   // 持续时间（秒）
	Value       float32   // Buff值
	Property    string    // 影响的属性类型
	IsPermanent bool      // 是否永久
	StartTime   time.Time // 开始时间
	EndTime     time.Time // 结束时间
}

// BuffState buff状态
type BuffState struct {
	buffs map[int32]*Buff
}

func NewBuffState() *BuffState {
	return &BuffState{
		buffs: make(map[int32]*Buff),
	}
}

func (state *BuffState) AddBuff(buff *Buff) {
	if buff == nil {
		return
	}
	state.buffs[buff.ID] = buff
}

func (state *BuffState) RemoveBuff(buffID int32) {
	delete(state.buffs, buffID)
}

func (state *BuffState) GetBuffs() map[int32]*Buff {
	result := make(map[int32]*Buff, len(state.buffs))
	for k, v := range state.buffs {
		result[k] = v
	}
	return result
}

func (state *BuffState) UpdateExpiredBuffs() []*Buff {
	currentTime := time.Now()
	expiredBuffs := make([]*Buff, 0)

	for buffID, buff := range state.buffs {
		if buff.IsPermanent {
			continue
		}

		if currentTime.After(buff.EndTime) {
			expiredBuffs = append(expiredBuffs, buff)
			delete(state.buffs, buffID)
		}
	}

	return expiredBuffs
}

// BuffComponent buff组件
type BuffComponent struct {
	mu        sync.RWMutex
	buffState *BuffState
	buffPool  *zObject.GenericPool
	owner     common.IGameObject
}

func NewBuffComponent(owner common.IGameObject) *BuffComponent {
	return &BuffComponent{
		buffState: NewBuffState(),
		buffPool:  zObject.NewGenericPool(func() interface{} { return &Buff{} }, 100),
		owner:     owner,
	}
}

func (bc *BuffComponent) AddBuff(buffID int32, name, description, buffType string, duration, value float32, property string, isPermanent bool) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	existingBuff, exists := bc.buffState.buffs[buffID]
	if exists {
		existingBuff.StartTime = time.Now()
		if !isPermanent {
			existingBuff.EndTime = time.Now().Add(time.Duration(duration) * time.Second)
		}
		bc.applyBuffEffect(existingBuff)
		return
	}

	newBuff := bc.buffPool.Get().(*Buff)
	newBuff.ID = buffID
	newBuff.Name = name
	newBuff.Description = description
	newBuff.Type = buffType
	newBuff.Duration = duration
	newBuff.Value = value
	newBuff.Property = property
	newBuff.IsPermanent = isPermanent
	newBuff.StartTime = time.Now()
	if !isPermanent {
		newBuff.EndTime = time.Now().Add(time.Duration(duration) * time.Second)
	}

	bc.buffState.AddBuff(newBuff)
	bc.applyBuffEffect(newBuff)
}

func (bc *BuffComponent) AddBuffFromConfig(buffID int32) {
	buffConfig := tables.GetBuffByID(buffID)
	if buffConfig == nil {
		return
	}

	bc.AddBuff(
		buffConfig.BuffID,
		buffConfig.Name,
		buffConfig.Description,
		buffConfig.Type,
		float32(buffConfig.Duration),
		float32(buffConfig.Value),
		string(buffConfig.Property),
		buffConfig.IsPermanent,
	)
}

func (bc *BuffComponent) RemoveBuff(buffID int32) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	buff, exists := bc.buffState.buffs[buffID]
	if !exists {
		return
	}

	bc.removeBuffEffect(buff)
	bc.buffState.RemoveBuff(buffID)
	bc.buffPool.Put(buff)
}

func (bc *BuffComponent) GetBuffs() map[int32]*Buff {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.buffState.GetBuffs()
}

func (bc *BuffComponent) Update(deltaTime float64) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	expiredBuffs := bc.buffState.UpdateExpiredBuffs()
	for _, buff := range expiredBuffs {
		bc.removeBuffEffect(buff)
		bc.buffPool.Put(buff)
	}
}

func (bc *BuffComponent) applyBuffEffect(buff *Buff) {
	if buff.Property == "" || bc.owner == nil {
		return
	}

	if buff.Type == BuffTypePositive {
		bc.addProperty(buff.Property, buff.Value)
	} else if buff.Type == BuffTypeNegative {
		bc.subProperty(buff.Property, buff.Value)
	}
}

func (bc *BuffComponent) removeBuffEffect(buff *Buff) {
	if buff.Property == "" || bc.owner == nil {
		return
	}

	if buff.Type == BuffTypePositive {
		bc.subProperty(buff.Property, buff.Value)
	} else if buff.Type == BuffTypeNegative {
		bc.addProperty(buff.Property, buff.Value)
	}
}

func (bc *BuffComponent) addProperty(property string, value float32) {
	if propertyComponent := bc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface {
			AddPropertyByType(key string, value float32)
		}); ok {
			prop.AddPropertyByType(property, value)
		}
	}
}

func (bc *BuffComponent) subProperty(property string, value float32) {
	if propertyComponent := bc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface {
			SubPropertyByType(key string, value float32)
		}); ok {
			prop.SubPropertyByType(property, value)
		}
	}
}

func (bc *BuffComponent) getOwnerPropertyComponent() interface{} {
	if bc.owner == nil {
		return nil
	}
	return bc.owner.GetComponent("property")
}
