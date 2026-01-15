package global

import (
	"strconv"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// 地图类型定义
const (
	MapTypeWorld   = 1 // 世界地图
	MapTypeCity    = 2 // 城市地图
	MapTypeDungeon = 3 // 副本地图
	MapTypeBattle  = 4 // 战场地图
	MapTypeGuild   = 5 // 公会地图
)

// 地图对象类型定义
const (
	MapObjectTypePlayer   = 1 // 玩家
	MapObjectTypeNPC      = 2 // NPC
	MapObjectTypeMonster  = 3 // 怪物
	MapObjectTypeItem     = 4 // 物品
	MapObjectTypePet      = 5 // 宠物
	MapObjectTypeVehicle  = 6 // 载具
	MapObjectTypeBuilding = 7 // 建筑
)

// MapObject 地图对象
type MapObject struct {
	objectId    int64
	objectType  int
	mapId       int64
	x           float64
	y           float64
	z           float64
	orientation float64
	moveSpeed   float64
	isMoving    bool
	status      int
	properties  *zMap.Map // 附加属性
}

// MapRegion 地图区域
type MapRegion struct {
	regionId   int
	minX       float64
	minY       float64
	maxX       float64
	maxY       float64
	objects    *zMap.Map // key: int64(objectId), value: *MapObject
	regionName string
}

// Map 地图结构
type Map struct {
	mapId         int64
	name          string
	mapType       int
	width         float64
	height        float64
	regionSize    float64
	tileWidth     float64
	tileHeight    float64
	tileMap       [][]int   // 地形数据
	objects       *zMap.Map // key: int64(objectId), value: *MapObject
	regions       *zMap.Map // key: int(regionId), value: *MapRegion
	npcs          *zMap.Map // key: int64(npcId), value: *MapObject
	monsters      *zMap.Map // key: int64(monsterId), value: *MapObject
	dropItems     *zMap.Map // key: int64(itemId), value: *MapObject
	isInstance    bool      // 是否为实例地图
	instanceOwner int64     // 实例所有者（如果有）
	maxPlayers    int
	playerCount   int
}

// MapService 地图服务
type MapService struct {
	zObject.BaseObject
	logger     *zap.Logger
	maps       *zMap.Map // key: int64(mapId), value: *Map
	objectMap  *zMap.Map // key: int64(objectId), value: int64(mapId)
	npcMap     *zMap.Map // key: int64(npcId), value: *MapObject
	monsterMap *zMap.Map // key: int64(monsterId), value: *MapObject
	playerMap  *zMap.Map // key: int64(playerId), value: *MapObject
	maxMaps    int
}

func NewMapService(logger *zap.Logger) *MapService {
	ms := &MapService{
		logger:     logger,
		maps:       zMap.NewMap(),
		objectMap:  zMap.NewMap(),
		npcMap:     zMap.NewMap(),
		monsterMap: zMap.NewMap(),
		playerMap:  zMap.NewMap(),
		maxMaps:    100,
	}
	ms.BaseObject.Id = "MapService"
	return ms
}

func (ms *MapService) Init() error {
	ms.logger.Info("Initializing map service...")
	// 初始化地图服务相关资源
	// 加载地图配置
	// ms.loadMapConfig()
	return nil
}

func (ms *MapService) Close() error {
	ms.logger.Info("Closing map service...")
	// 清理地图服务相关资源
	ms.maps.Clear()
	ms.objectMap.Clear()
	ms.npcMap.Clear()
	ms.monsterMap.Clear()
	ms.playerMap.Clear()
	return nil
}

func (ms *MapService) Serve() {
	// 地图服务需要持续运行的协程，用于处理地图对象的移动同步
	go ms.mapSyncLoop()
}

// mapSyncLoop 地图同步循环
func (ms *MapService) mapSyncLoop() {
	// 每100毫秒同步一次地图对象
	for {
		select {
		case <-time.After(time.Millisecond * 100):
			ms.syncMapObjects()
		}
	}
}

// syncMapObjects 同步地图对象
func (ms *MapService) syncMapObjects() {
	// 遍历所有地图
	ms.maps.Range(func(key, value interface{}) bool {
		mapObj := value.(*Map)
		ms.syncMapObjectsByMap(mapObj)
		return true
	})
}

// syncMapObjectsByMap 同步指定地图的对象
func (ms *MapService) syncMapObjectsByMap(mapObj *Map) {
	// 收集需要同步的对象
	var objectsToSync []*MapObject

	// 遍历地图中的所有对象
	mapObj.objects.Range(func(key, value interface{}) bool {
		obj := value.(*MapObject)
		// 只同步移动中的对象或状态变化的对象
		if obj.isMoving || obj.status != 0 {
			objectsToSync = append(objectsToSync, obj)
			// 重置状态标记
			if !obj.isMoving {
				obj.status = 0
			}
		}
		return true
	})

	// 如果有对象需要同步，发送同步消息
	if len(objectsToSync) > 0 {
		ms.sendMapObjectSync(mapObj, objectsToSync)
	}
}

// sendMapObjectSync 发送地图对象同步消息
func (ms *MapService) sendMapObjectSync(mapObj *Map, objects []*MapObject) {
	// TODO: 实现发送同步消息的逻辑
	// 这里应该调用网络层发送消息给地图上的所有玩家
	// 或者只发送给相关区域内的玩家
	ms.logger.Debug("Syncing map objects", zap.Int64("mapId", mapObj.mapId), zap.Int("objectCount", len(objects)))
}

// CreateMap 创建地图
func (ms *MapService) CreateMap(mapId int64, name string, mapType int, width, height, regionSize float64, tileWidth, tileHeight float64, isInstance bool) (*Map, error) {
	// 检查地图是否已存在
	if _, exists := ms.maps.Get(mapId); exists {
		return nil, nil // 地图已存在
	}

	// 检查是否达到最大地图数量
	if ms.maps.Len() >= int64(ms.maxMaps) {
		return nil, nil // 已达到最大地图数量
	}

	// 创建新地图
	mapObj := &Map{
		mapId:         mapId,
		name:          name,
		mapType:       mapType,
		width:         width,
		height:        height,
		regionSize:    regionSize,
		tileWidth:     tileWidth,
		tileHeight:    tileHeight,
		tileMap:       make([][]int, 0),
		objects:       zMap.NewMap(),
		regions:       zMap.NewMap(),
		npcs:          zMap.NewMap(),
		monsters:      zMap.NewMap(),
		dropItems:     zMap.NewMap(),
		isInstance:    isInstance,
		instanceOwner: 0,
		maxPlayers:    100,
		playerCount:   0,
	}

	// 初始化地图区域
	ms.initMapRegions(mapObj)

	// 存储地图
	ms.maps.Store(mapId, mapObj)

	ms.logger.Info("Map created", zap.Int64("mapId", mapId), zap.String("mapName", name), zap.Int("mapType", mapType))
	return mapObj, nil
}

// initMapRegions 初始化地图区域
func (ms *MapService) initMapRegions(mapObj *Map) {
	// 计算区域数量
	regionCountX := int(mapObj.width / mapObj.regionSize)
	regionCountY := int(mapObj.height / mapObj.regionSize)

	// 创建区域
	for x := 0; x < regionCountX; x++ {
		for y := 0; y < regionCountY; y++ {
			regionId := y*regionCountX + x
			region := &MapRegion{
				regionId:   regionId,
				minX:       float64(x) * mapObj.regionSize,
				minY:       float64(y) * mapObj.regionSize,
				maxX:       float64(x+1) * mapObj.regionSize,
				maxY:       float64(y+1) * mapObj.regionSize,
				objects:    zMap.NewMap(),
				regionName: "Region_" + strconv.Itoa(regionId),
			}
			mapObj.regions.Store(regionId, region)
		}
	}
}

// AddObject 添加对象到地图
func (ms *MapService) AddObject(mapId int64, obj *MapObject) error {
	// 获取地图
	mapObjInterface, exists := ms.maps.Get(mapId)
	if !exists {
		return nil // 地图不存在
	}
	mapObj := mapObjInterface.(*Map)

	// 检查实例地图玩家数量限制
	if mapObj.isInstance && obj.objectType == MapObjectTypePlayer {
		if mapObj.playerCount >= mapObj.maxPlayers {
			return nil // 实例地图已满
		}
		mapObj.playerCount++
	}

	// 计算对象所在区域
	regionId := ms.getRegionId(mapObj, obj.x, obj.y)

	// 获取区域
	regionInterface, exists := mapObj.regions.Get(regionId)
	if !exists {
		return nil // 区域不存在
	}
	region := regionInterface.(*MapRegion)

	// 添加对象到地图
	mapObj.objects.Store(obj.objectId, obj)
	region.objects.Store(obj.objectId, obj)

	// 记录对象与地图的映射关系
	ms.objectMap.Store(obj.objectId, mapId)

	// 记录特定类型对象的映射关系
	switch obj.objectType {
	case MapObjectTypePlayer:
		ms.playerMap.Store(obj.objectId, obj)
	case MapObjectTypeNPC:
		ms.npcMap.Store(obj.objectId, obj)
	case MapObjectTypeMonster:
		ms.monsterMap.Store(obj.objectId, obj)
	}

	ms.logger.Info("Object added to map", zap.Int64("objectId", obj.objectId), zap.Int("objectType", obj.objectType), zap.Int64("mapId", mapId))
	return nil
}

// RemoveObject 从地图移除对象
func (ms *MapService) RemoveObject(objectId int64) error {
	// 获取对象所在地图
	mapIdInterface, exists := ms.objectMap.Get(objectId)
	if !exists {
		return nil // 对象不存在
	}
	mapId := mapIdInterface.(int64)

	// 获取地图
	mapObjInterface, exists := ms.maps.Get(mapId)
	if !exists {
		return nil // 地图不存在
	}
	mapObj := mapObjInterface.(*Map)

	// 获取对象
	objInterface, exists := mapObj.objects.Get(objectId)
	if !exists {
		return nil // 对象不在地图上
	}
	obj := objInterface.(*MapObject)

	// 计算对象所在区域
	regionId := ms.getRegionId(mapObj, obj.x, obj.y)

	// 获取区域
	regionInterface, exists := mapObj.regions.Get(regionId)
	if !exists {
		return nil // 区域不存在
	}
	region := regionInterface.(*MapRegion)

	// 从区域和地图中移除对象
	region.objects.Delete(objectId)
	mapObj.objects.Delete(objectId)

	// 减少实例地图玩家数量
	if mapObj.isInstance && obj.objectType == MapObjectTypePlayer {
		if mapObj.playerCount > 0 {
			mapObj.playerCount--
		}
	}

	// 移除对象与地图的映射关系
	ms.objectMap.Delete(objectId)

	// 移除特定类型对象的映射关系
	switch obj.objectType {
	case MapObjectTypePlayer:
		ms.playerMap.Delete(objectId)
	case MapObjectTypeNPC:
		ms.npcMap.Delete(objectId)
	case MapObjectTypeMonster:
		ms.monsterMap.Delete(objectId)
	}

	ms.logger.Info("Object removed from map", zap.Int64("objectId", objectId), zap.Int64("mapId", mapId))
	return nil
}

// MoveObject 移动地图对象
func (ms *MapService) MoveObject(objectId int64, x, y, z float64) error {
	// 获取对象所在地图
	mapIdInterface, exists := ms.objectMap.Get(objectId)
	if !exists {
		return nil // 对象不存在
	}
	mapId := mapIdInterface.(int64)

	// 获取地图
	mapObjInterface, exists := ms.maps.Get(mapId)
	if !exists {
		return nil // 地图不存在
	}
	mapObj := mapObjInterface.(*Map)

	// 获取对象
	objInterface, exists := mapObj.objects.Get(objectId)
	if !exists {
		return nil // 对象不在地图上
	}
	obj := objInterface.(*MapObject)

	// 计算原区域和新区域
	oldRegionId := ms.getRegionId(mapObj, obj.x, obj.y)
	newRegionId := ms.getRegionId(mapObj, x, y)

	// 更新对象位置
	obj.x = x
	obj.y = y
	obj.z = z

	// 如果对象跨区域，需要更新区域信息
	if oldRegionId != newRegionId {
		// 获取原区域
		oldRegionInterface, exists := mapObj.regions.Get(oldRegionId)
		if exists {
			oldRegion := oldRegionInterface.(*MapRegion)
			oldRegion.objects.Delete(objectId)
		}

		// 获取新区域
		newRegionInterface, exists := mapObj.regions.Get(newRegionId)
		if exists {
			newRegion := newRegionInterface.(*MapRegion)
			newRegion.objects.Store(objectId, obj)
		}
	}

	ms.logger.Debug("Object moved", zap.Int64("objectId", objectId), zap.Float64("x", x), zap.Float64("y", y), zap.Float64("z", z))
	return nil
}

// GetPath 查找路径
func (ms *MapService) GetPath(mapId int64, startX, startY, endX, endY float64) ([][]float64, error) {
	// 获取地图
	_, exists := ms.maps.Get(mapId)
	if !exists {
		return nil, nil // 地图不存在
	}

	// 简化实现：返回直线路径
	worldPath := make([][]float64, 2)
	worldPath[0] = []float64{startX, startY}
	worldPath[1] = []float64{endX, endY}

	return worldPath, nil
}

// GetMap 获取地图信息
func (ms *MapService) GetMap(mapId int64) (*Map, bool) {
	mapObj, exists := ms.maps.Get(mapId)
	if !exists {
		return nil, false
	}
	return mapObj.(*Map), true
}

// GetMapObject 获取地图对象
func (ms *MapService) GetMapObject(objectId int64) (*MapObject, bool) {
	// 获取对象所在地图
	mapIdInterface, exists := ms.objectMap.Get(objectId)
	if !exists {
		return nil, false
	}
	mapId := mapIdInterface.(int64)

	// 获取地图
	mapObjInterface, exists := ms.maps.Get(mapId)
	if !exists {
		return nil, false
	}
	mapObj := mapObjInterface.(*Map)

	// 获取对象
	obj, exists := mapObj.objects.Get(objectId)
	if !exists {
		return nil, false
	}

	return obj.(*MapObject), true
}

// GetObjectsInRegion 获取区域内的对象
func (ms *MapService) GetObjectsInRegion(mapId int64, regionId int) ([]*MapObject, bool) {
	// 获取地图
	mapObjInterface, exists := ms.maps.Get(mapId)
	if !exists {
		return nil, false
	}
	mapObj := mapObjInterface.(*Map)

	// 获取区域
	regionInterface, exists := mapObj.regions.Get(regionId)
	if !exists {
		return nil, false
	}
	region := regionInterface.(*MapRegion)

	// 收集区域内的对象
	var objects []*MapObject
	region.objects.Range(func(key, value interface{}) bool {
		if value != nil {
			objects = append(objects, value.(*MapObject))
		}
		return true
	})

	return objects, true
}

// GetObjectsInRange 获取范围内的对象
func (ms *MapService) GetObjectsInRange(mapId int64, x, y, radius float64, objectType int) ([]*MapObject, bool) {
	// 获取地图
	mapObjInterface, exists := ms.maps.Get(mapId)
	if !exists {
		return nil, false
	}
	mapObj := mapObjInterface.(*Map)

	// 收集范围内的对象
	var objects []*MapObject
	mapObj.objects.Range(func(key, value interface{}) bool {
		obj := value.(*MapObject)
		// 检查对象类型
		if objectType != 0 && obj.objectType != objectType {
			return true
		}
		// 检查距离
		dx := obj.x - x
		dy := obj.y - y
		distance := dx*dx + dy*dy
		if distance <= radius*radius {
			objects = append(objects, obj)
		}
		return true
	})

	return objects, true
}

// getRegionId 获取区域ID
func (ms *MapService) getRegionId(mapObj *Map, x, y float64) int {
	regionX := int(x / mapObj.regionSize)
	regionY := int(y / mapObj.regionSize)
	regionCountX := int(mapObj.width / mapObj.regionSize)
	return regionY*regionCountX + regionX
}
