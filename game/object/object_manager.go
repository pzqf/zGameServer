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
	allObjects map[common.ObjectIdType]common.IGameObject
	players    map[common.ObjectIdType]common.IGameObject
	monsters   map[common.ObjectIdType]common.IGameObject
	npcs       map[common.ObjectIdType]common.IGameObject
	items      map[common.ObjectIdType]common.IGameObject
	buildings  map[common.ObjectIdType]common.IGameObject
}

// objectManagerInstance 全局实例
var objectManagerInstance *ObjectManager

// InitObjectManager 初始化对象管理器
func InitObjectManager() {
	objectManagerInstance = &ObjectManager{
		allObjects: make(map[common.ObjectIdType]common.IGameObject),
		players:    make(map[common.ObjectIdType]common.IGameObject),
		monsters:   make(map[common.ObjectIdType]common.IGameObject),
		npcs:       make(map[common.ObjectIdType]common.IGameObject),
		items:      make(map[common.ObjectIdType]common.IGameObject),
		buildings:  make(map[common.ObjectIdType]common.IGameObject),
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

	if objType == common.GameObjectTypePlayer {
		om.players[objectID] = object
	} else if objType == common.GameObjectTypeMonster {
		om.monsters[objectID] = object
	} else if objType == common.GameObjectTypeNPC {
		om.npcs[objectID] = object
	} else if objType == common.GameObjectTypeItem {
		om.items[objectID] = object
	} else if objType == common.GameObjectTypeBuilding {
		om.buildings[objectID] = object
	}
}

// RemoveObject 移除对象
func (om *ObjectManager) RemoveObject(objectID common.ObjectIdType) {
	om.mu.Lock()
	defer om.mu.Unlock()

	obj, exists := om.allObjects[objectID]
	if !exists {
		return
	}

	delete(om.allObjects, objectID)

	objType := obj.GetType()
	if objType == common.GameObjectTypePlayer {
		delete(om.players, objectID)
	} else if objType == common.GameObjectTypeMonster {
		delete(om.monsters, objectID)
	} else if objType == common.GameObjectTypeNPC {
		delete(om.npcs, objectID)
	} else if objType == common.GameObjectTypeItem {
		delete(om.items, objectID)
	} else if objType == common.GameObjectTypeBuilding {
		delete(om.buildings, objectID)
	}

	zLog.Debug("Object removed from manager", zap.Int64("objectId", int64(objectID)))
}

// GetObject 获取对象
func (om *ObjectManager) GetObject(objectID common.ObjectIdType) common.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return om.allObjects[objectID]
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

// GetObjectsByType 根据类型获取对象
func (om *ObjectManager) GetObjectsByType(objType common.GameObjectType) []common.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()

	var targetMap map[common.ObjectIdType]common.IGameObject
	switch objType {
	case common.GameObjectTypePlayer:
		targetMap = om.players
	case common.GameObjectTypeMonster:
		targetMap = om.monsters
	case common.GameObjectTypeNPC:
		targetMap = om.npcs
	case common.GameObjectTypeItem:
		targetMap = om.items
	case common.GameObjectTypeBuilding:
		targetMap = om.buildings
	default:
		return nil
	}

	objects := make([]common.IGameObject, 0, len(targetMap))
	for _, obj := range targetMap {
		objects = append(objects, obj)
	}

	return objects
}

// ClearAllObjects 清空所有对象
func (om *ObjectManager) ClearAllObjects() {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.allObjects = make(map[common.ObjectIdType]common.IGameObject)
	om.players = make(map[common.ObjectIdType]common.IGameObject)
	om.monsters = make(map[common.ObjectIdType]common.IGameObject)
	om.npcs = make(map[common.ObjectIdType]common.IGameObject)
	om.items = make(map[common.ObjectIdType]common.IGameObject)
	om.buildings = make(map[common.ObjectIdType]common.IGameObject)

	zLog.Info("All objects cleared from manager")
}
