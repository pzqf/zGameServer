package property

import (
	"sync"
)

// PropertyListener 属性变化监听器
type PropertyListener func(ownerID uint64, key string, oldValue, newValue float32)

// PropertySystem 属性系统
type PropertySystem struct {
	mu        sync.RWMutex
	properties map[uint64]map[string]float32 // key: ownerID, value: 属性映射
	listeners  map[string][]PropertyListener // key: 属性名, value: 监听器列表
}

// GlobalPropertySystem 全局属性系统实例
var GlobalPropertySystem *PropertySystem

// init 初始化全局属性系统
func init() {
	GlobalPropertySystem = &PropertySystem{
		properties: make(map[uint64]map[string]float32),
		listeners:  make(map[string][]PropertyListener),
	}
}

// GetProperty 获取属性值
func (ps *PropertySystem) GetProperty(ownerID uint64, key string) float32 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	if ownerProps, exists := ps.properties[ownerID]; exists {
		return ownerProps[key]
	}
	return 0
}

// SetProperty 设置属性值
func (ps *PropertySystem) SetProperty(ownerID uint64, key string, value float32) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	// 确保所有者的属性映射存在
	if _, exists := ps.properties[ownerID]; !exists {
		ps.properties[ownerID] = make(map[string]float32)
	}
	
	oldValue := ps.properties[ownerID][key]
	ps.properties[ownerID][key] = value
	
	// 触发属性变化事件
	if listeners, ok := ps.listeners[key]; ok {
		for _, listener := range listeners {
			listener(ownerID, key, oldValue, value)
		}
	}
}

// AddProperty 增加属性值
func (ps *PropertySystem) AddProperty(ownerID uint64, key string, value float32) {
	current := ps.GetProperty(ownerID, key)
	ps.SetProperty(ownerID, key, current+value)
}

// SubProperty 减少属性值
func (ps *PropertySystem) SubProperty(ownerID uint64, key string, value float32) {
	current := ps.GetProperty(ownerID, key)
	ps.SetProperty(ownerID, key, current-value)
}

// GetAllProperties 获取所有属性
func (ps *PropertySystem) GetAllProperties(ownerID uint64) map[string]float32 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	if ownerProps, exists := ps.properties[ownerID]; exists {
		result := make(map[string]float32, len(ownerProps))
		for k, v := range ownerProps {
			result[k] = v
		}
		return result
	}
	return make(map[string]float32)
}

// AddPropertyListener 添加属性监听器
func (ps *PropertySystem) AddPropertyListener(key string, listener PropertyListener) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	ps.listeners[key] = append(ps.listeners[key], listener)
}

// RemovePropertyListener 移除属性监听器
func (ps *PropertySystem) RemovePropertyListener(key string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	delete(ps.listeners, key)
}
