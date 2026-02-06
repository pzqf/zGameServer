package object

import (
	"sync"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"go.uber.org/zap"
)

// ObjectManager 对象管理器
type ObjectManager struct {
	mu         sync.RWMutex
	allObjects map[common.ObjectIdType]gamecommon.IGameObject
	players    map[common.ObjectIdType]gamecommon.IGameObject
	monsters   map[common.ObjectIdType]gamecommon.IGameObject
	npcs       map[common.ObjectIdType]gamecommon.IGameObject
	items      map[common.ObjectIdType]gamecommon.IGameObject
	buildings  map[common.ObjectIdType]gamecommon.IGameObject
}

// objectManagerInstance 全局实例
var objectManagerInstance *ObjectManager

// InitObjectManager 初始化对象管理器
func InitObjectManager() {
	objectManagerInstance = &ObjectManager{
		allObjects: make(map[common.ObjectIdType]gamecommon.IGameObject),
		players:    make(map[common.ObjectIdType]gamecommon.IGameObject),
		monsters:   make(map[common.ObjectIdType]gamecommon.IGameObject),
		npcs:       make(map[common.ObjectIdType]gamecommon.IGameObject),
		items:      make(map[common.ObjectIdType]gamecommon.IGameObject),
		buildings:  make(map[common.ObjectIdType]gamecommon.IGameObject),
	}
}

// GetObjectManager 获取对象管理器
func GetObjectManager() *ObjectManager {
	return objectManagerInstance
}

// AddObject 添加对象到管理器
func (om *ObjectManager) AddObject(object gamecommon.IGameObject) {
	if object == nil {
		return
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	objectID := object.GetID()
	om.allObjects[objectID] = object

	objType := object.GetType()

	if objType == gamecommon.GameObjectTypePlayer {
		om.players[objectID] = object
	} else if objType == gamecommon.GameObjectTypeMonster {
		om.monsters[objectID] = object
	} else if objType == gamecommon.GameObjectTypeNPC {
		om.npcs[objectID] = object
	} else if objType == gamecommon.GameObjectTypeItem {
		om.items[objectID] = object
	} else if objType == gamecommon.GameObjectTypeBuilding {
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
	if objType == gamecommon.GameObjectTypePlayer {
		delete(om.players, objectID)
	} else if objType == gamecommon.GameObjectTypeMonster {
		delete(om.monsters, objectID)
	} else if objType == gamecommon.GameObjectTypeNPC {
		delete(om.npcs, objectID)
	} else if objType == gamecommon.GameObjectTypeItem {
		delete(om.items, objectID)
	} else if objType == gamecommon.GameObjectTypeBuilding {
		delete(om.buildings, objectID)
	}

	zLog.Debug("Object removed from manager", zap.Int64("objectId", int64(objectID)))
}

// GetObject 获取对象
func (om *ObjectManager) GetObject(objectID common.ObjectIdType) gamecommon.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return om.allObjects[objectID]
}

// GetAllObjects 获取所有对象
func (om *ObjectManager) GetAllObjects() []gamecommon.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()

	objects := make([]gamecommon.IGameObject, 0, len(om.allObjects))
	for _, obj := range om.allObjects {
		objects = append(objects, obj)
	}

	return objects
}

// GetObjectsByType 根据类型获取对象
func (om *ObjectManager) GetObjectsByType(objType gamecommon.GameObjectType) []gamecommon.IGameObject {
	om.mu.RLock()
	defer om.mu.RUnlock()

	var targetMap map[common.ObjectIdType]gamecommon.IGameObject
	switch objType {
	case gamecommon.GameObjectTypePlayer:
		targetMap = om.players
	case gamecommon.GameObjectTypeMonster:
		targetMap = om.monsters
	case gamecommon.GameObjectTypeNPC:
		targetMap = om.npcs
	case gamecommon.GameObjectTypeItem:
		targetMap = om.items
	case gamecommon.GameObjectTypeBuilding:
		targetMap = om.buildings
	default:
		return nil
	}

	objects := make([]gamecommon.IGameObject, 0, len(targetMap))
	for _, obj := range targetMap {
		objects = append(objects, obj)
	}

	return objects
}

// ClearAllObjects 清空所有对象
func (om *ObjectManager) ClearAllObjects() {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.allObjects = make(map[common.ObjectIdType]gamecommon.IGameObject)
	om.players = make(map[common.ObjectIdType]gamecommon.IGameObject)
	om.monsters = make(map[common.ObjectIdType]gamecommon.IGameObject)
	om.npcs = make(map[common.ObjectIdType]gamecommon.IGameObject)
	om.items = make(map[common.ObjectIdType]gamecommon.IGameObject)
	om.buildings = make(map[common.ObjectIdType]gamecommon.IGameObject)

	zLog.Info("All objects cleared from manager")
}
