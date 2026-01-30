package property

import (
	"sync"

	"github.com/pzqf/zGameServer/game/common"
)

// PropertyListener 属性变化监听器
type PropertyListener func(owner common.IGameObject, key string, oldValue, newValue float32)

// PropertyState 属性状态
type PropertyState struct {
	mu         sync.RWMutex
	properties map[string]float32
}

func NewPropertyState() *PropertyState {
	return &PropertyState{
		properties: make(map[string]float32),
	}
}

func (state *PropertyState) GetProperty(key string) float32 {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.properties[key]
}

func (state *PropertyState) SetProperty(key string, value float32) float32 {
	state.mu.Lock()
	defer state.mu.Unlock()
	oldValue := state.properties[key]
	state.properties[key] = value
	return oldValue
}

func (state *PropertyState) AddProperty(key string, value float32) {
	state.mu.Lock()
	defer state.mu.Unlock()
	current := state.properties[key]
	state.properties[key] = current + value
}

func (state *PropertyState) SubProperty(key string, value float32) {
	state.mu.Lock()
	defer state.mu.Unlock()
	current := state.properties[key]
	state.properties[key] = current - value
}

func (state *PropertyState) GetAllProperties() map[string]float32 {
	state.mu.RLock()
	defer state.mu.RUnlock()
	result := make(map[string]float32, len(state.properties))
	for k, v := range state.properties {
		result[k] = v
	}
	return result
}

// PropertyComponent 属性组件
type PropertyComponent struct {
	mu            sync.RWMutex
	propertyState *PropertyState
	listeners     map[string][]PropertyListener
	owner         common.IGameObject
}

func NewPropertyComponent(owner common.IGameObject) *PropertyComponent {
	return &PropertyComponent{
		propertyState: NewPropertyState(),
		listeners:     make(map[string][]PropertyListener),
		owner:         owner,
	}
}

func (pc *PropertyComponent) GetProperty(key string) float32 {
	return pc.propertyState.GetProperty(key)
}

func (pc *PropertyComponent) GetPropertyByType(propType PropertyType) float32 {
	return pc.GetProperty(GetPropertyType(propType))
}

func (pc *PropertyComponent) SetProperty(key string, value float32) {
	oldValue := pc.propertyState.SetProperty(key, value)
	pc.triggerPropertyChange(key, oldValue, value)
}

func (pc *PropertyComponent) SetPropertyByType(propType PropertyType, value float32) {
	pc.SetProperty(GetPropertyType(propType), value)
}

func (pc *PropertyComponent) AddProperty(key string, value float32) {
	current := pc.GetProperty(key)
	pc.SetProperty(key, current+value)
}

func (pc *PropertyComponent) AddPropertyByType(propType PropertyType, value float32) {
	pc.AddProperty(GetPropertyType(propType), value)
}

func (pc *PropertyComponent) SubProperty(key string, value float32) {
	current := pc.GetProperty(key)
	pc.SetProperty(key, current-value)
}

func (pc *PropertyComponent) SubPropertyByType(propType PropertyType, value float32) {
	pc.SubProperty(GetPropertyType(propType), value)
}

func (pc *PropertyComponent) GetAllProperties() map[string]float32 {
	return pc.propertyState.GetAllProperties()
}

func (pc *PropertyComponent) AddPropertyListener(key string, listener PropertyListener) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.listeners[key] = append(pc.listeners[key], listener)
}

func (pc *PropertyComponent) AddPropertyListenerByType(propType PropertyType, listener PropertyListener) {
	pc.AddPropertyListener(GetPropertyType(propType), listener)
}

func (pc *PropertyComponent) RemovePropertyListener(key string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	delete(pc.listeners, key)
}

func (pc *PropertyComponent) RemovePropertyListenerByType(propType PropertyType) {
	pc.RemovePropertyListener(GetPropertyType(propType))
}

func (pc *PropertyComponent) Update(deltaTime float64) {
	// 属性组件不需要定期更新
}

func (pc *PropertyComponent) triggerPropertyChange(key string, oldValue, newValue float32) {
	if oldValue == newValue {
		return
	}

	pc.mu.RLock()
	listeners, ok := pc.listeners[key]
	pc.mu.RUnlock()

	if ok {
		for _, listener := range listeners {
			listener(pc.owner, key, oldValue, newValue)
		}
	}
}
