package object

import (
	"sync"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/game/common"
	"go.uber.org/zap"
)

// ObjectManager 对象管理器
type ObjectManager struct {
	mu         sync.RWMutex
	allObjects map[uint64]common.IGameObject
	players    map[uint64]common.IGameObject
	monsters   map[uint64]common.IGameObject
	npcs       map[uint64]common.IGameObject
	items      map[uint64]common.IGameObject
	buildings  map[uint64]common.IGameObject
}

// objectManagerInstance 全局实例
var objectManagerInstance *ObjectManager

// InitObjectManager 初始化对象管理器
func InitObjectManager() {
	objectManagerInstance = &ObjectManager{
		allObjects: make(map[uint64]common.IGameObject),
		players:    make(map[uint64]common.IGameObject),
		monsters:   make(map[uint64]common.IGameObject),
		npcs:       make(map[uint64]common.IGameObject),
		items:      make(map[uint64]common.IGameObject),
		buildings:  make(map[uint64]common.IGameObject),
	}
}

// GetObjectManager 获取对象管理器
func GetObjectManager() *ObjectManager {
	return objectManagerInstance
}

// AddObject 添加对象到管理器
func (om *ObjectManager) AddObject(object common.IGameObject) {
	if object == nil {
		return
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	objectID := object.GetID()
	om.allObjects[objectID] = object

	objType := object.GetType()

	if objType == int(common.GameObjectTypePlayer) {
		om.players[objectID] = object
	} else if objType == int(common.GameObjectTypeMonster) {
		om.monsters[objectID] = object
	} else if objType == int(common.GameObjectTypeNPC) {
		om.npcs[objectID] = object
	} else if objType == int(common.GameObjectTypeItem) {
		om.items[objectID] = object
	} else if objType == int(common.GameObjectTypeBuilding) {
		om.buildings[objectID] = object
	}
}

// RemoveObject 移除对象
func (om *ObjectManager) RemoveObject(objectID uint64) {
	om.mu.Lock()
	defer om.mu.Unlock()

	obj, exists := om.allObjects[objectID]
	if !exists {
		return
	}

	delete(om.allObjects, objectID)

	objType := obj.GetType()
	if objType == int(common.GameObjectTypePlayer) {
		delete(om.players, objectID)
	} else if objType == int(common.GameObjectTypeMonster) {
		delete(om.monsters, objectID)
	} else if objType == int(common.GameObjectTypeNPC) {
		delete(om.npcs, objectID)
	} else if objType == int(common.GameObjectTypeItem) {
		delete(om.items, objectID)
	} else if objType == int(common.GameObjectTypeBuilding) {
		delete(om.buildings, objectID)
	}
}

// GetObject 根据ID获取对象
func (om *ObjectManager) GetObject(objectID uint64) common.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.allObjects[objectID]
}

// GetPlayer 获取玩家
func (om *ObjectManager) GetPlayer(playerID uint64) common.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.players[playerID]
}

// GetMonster 获取怪物
func (om *ObjectManager) GetMonster(monsterID uint64) common.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.monsters[monsterID]
}

// GetNPC 获取NPC
func (om *ObjectManager) GetNPC(npcID uint64) common.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.npcs[npcID]
}

// GetAllObjects 获取所有对象
func (om *ObjectManager) GetAllObjects() []common.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()

	objects := make([]common.IGameObject, 0, len(om.allObjects))
	for _, obj := range om.allObjects {
		objects = append(objects, obj)
	}

	return objects
}

// GetAllPlayers 获取所有玩家
func (om *ObjectManager) GetAllPlayers() []common.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()

	players := make([]common.IGameObject, 0, len(om.players))
	for _, player := range om.players {
		players = append(players, player)
	}

	return players
}

// GetAllMonsters 获取所有怪物
func (om *ObjectManager) GetAllMonsters() []common.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()

	monsters := make([]common.IGameObject, 0, len(om.monsters))
	for _, monster := range om.monsters {
		monsters = append(monsters, monster)
	}

	return monsters
}

// Update 更新所有对象
func (om *ObjectManager) Update(deltaTime float64) {
	om.mu.RLock()
	allObjects := make([]common.IGameObject, 0, len(om.allObjects))
	for _, obj := range om.allObjects {
		allObjects = append(allObjects, obj)
	}
	om.mu.RUnlock()

	for _, obj := range allObjects {
		obj.Update(deltaTime)
	}
}

// Shutdown 关闭对象管理器
func (om *ObjectManager) Shutdown() {
	om.mu.Lock()
	defer om.mu.Unlock()

	zLog.Info("Shutting down ObjectManager",
		zap.Int("players", len(om.players)),
		zap.Int("monsters", len(om.monsters)),
		zap.Int("npcs", len(om.npcs)),
		zap.Int("allObjects", len(om.allObjects)))

	for _, obj := range om.allObjects {
		obj.Destroy()
	}

	om.allObjects = nil
	om.players = nil
	om.monsters = nil
	om.npcs = nil
	om.items = nil
	om.buildings = nil
}

// GetObjectCount 获取对象统计
func (om *ObjectManager) GetObjectCount() map[string]int {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return map[string]int{
		"total":     len(om.allObjects),
		"players":   len(om.players),
		"monsters":  len(om.monsters),
		"npcs":      len(om.npcs),
		"items":     len(om.items),
		"buildings": len(om.buildings),
	}
}

// CleanupInactiveObjects 清理不活跃对象
func (om *ObjectManager) CleanupInactiveObjects() {
	om.mu.Lock()
	defer om.mu.Unlock()

	for objectID, obj := range om.allObjects {
		if !obj.IsActive() {
			delete(om.allObjects, objectID)

			objType := obj.GetType()
			if objType == int(common.GameObjectTypePlayer) {
				delete(om.players, objectID)
			} else if objType == int(common.GameObjectTypeMonster) {
				delete(om.monsters, objectID)
			} else if objType == int(common.GameObjectTypeNPC) {
				delete(om.npcs, objectID)
			} else if objType == int(common.GameObjectTypeItem) {
				delete(om.items, objectID)
			} else if objType == int(common.GameObjectTypeBuilding) {
				delete(om.buildings, objectID)
			}
		}
	}
}
