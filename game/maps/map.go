package maps

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/game/common"
	monster "github.com/pzqf/zGameServer/game/monsters"
	"github.com/pzqf/zGameServer/game/npc"
	"github.com/pzqf/zGameServer/game/object"
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

// MapRegion 地图区域
type MapRegion struct {
	regionId   int
	minX       float64
	minY       float64
	maxX       float64
	maxY       float64
	objects    *zMap.Map // key: int64(objectId), value: object.IGameObject
	regionName string
}

// Map 地图结构
type Map struct {
	zObject.BaseObject
	mapId         int64
	name          string
	mapType       int
	width         float64
	height        float64
	regionSize    float64
	tileWidth     float64
	tileHeight    float64
	tileMap       [][]int   // 地形数据
	objects       *zMap.Map // key: int64(objectId), value: object.IGameObject
	regions       *zMap.Map // key: int(regionId), value: *MapRegion
	npcs          *zMap.Map // key: int64(npcId), value: object.IGameObject
	monsters      *zMap.Map // key: int64(monsterId), value: object.IGameObject
	dropItems     *zMap.Map // key: int64(itemId), value: object.IGameObject
	players       *zMap.Map // key: int64(playerId), value: object.IGameObject
	isInstance    bool      // 是否为实例地图
	instanceOwner int64     // 实例所有者（如果有）
	maxPlayers    int
	playerCount   int
}

// 地图JSON数据结构体
type MapJSONData struct {
	MapId          int64               `json:"map_id"`
	Name           string              `json:"name"`
	MapType        int                 `json:"map_type"`
	Width          float64             `json:"width"`
	Height         float64             `json:"height"`
	RegionSize     float64             `json:"region_size"`
	TileWidth      float64             `json:"tile_width"`
	TileHeight     float64             `json:"tile_height"`
	IsInstance     bool                `json:"is_instance"`
	MaxPlayers     int                 `json:"max_players"`
	TileMap        [][]int             `json:"tile_map"`
	SpawnPoints    []SpawnPointData    `json:"spawn_points"`
	TeleportPoints []TeleportPointData `json:"teleport_points"`
	Buildings      []BuildingData      `json:"buildings,omitempty"`
}

// 刷新点数据
type SpawnPointData struct {
	Type  string  `json:"type"`
	ID    int64   `json:"id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Name  string  `json:"name"`
	Count int     `json:"count,omitempty"` // 怪物数量
}

// 传送点数据
type TeleportPointData struct {
	ID          int64   `json:"id"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Z           float64 `json:"z"`
	TargetMapID int64   `json:"target_map_id"`
	TargetX     float64 `json:"target_x"`
	TargetY     float64 `json:"target_y"`
	TargetZ     float64 `json:"target_z"`
}

// 建筑数据
type BuildingData struct {
	ID     int64   `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Type   string  `json:"type"`
	Name   string  `json:"name"`
}

// NewMap 创建新地图
func NewMap(mapId int64, name string, mapType int, width, height, regionSize, tileWidth, tileHeight float64, isInstance bool) *Map {
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
		players:       zMap.NewMap(),
		isInstance:    isInstance,
		instanceOwner: 0,
		maxPlayers:    100,
		playerCount:   0,
	}

	// 初始化地图区域
	mapObj.initRegions()

	return mapObj
}

// LoadFromFile 从文件加载地图
func (m *Map) LoadFromFile(filePath string) error {
	// 读取地图文件
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		zLog.Error("Failed to read map file", zap.String("file_path", filePath), zap.Error(err))
		return err
	}

	// 解析JSON数据
	var mapData MapJSONData
	if err := json.Unmarshal(data, &mapData); err != nil {
		zLog.Error("Failed to parse map JSON", zap.String("file_path", filePath), zap.Error(err))
		return err
	}

	// 设置地图属性
	m.mapId = mapData.MapId
	m.name = mapData.Name
	m.mapType = mapData.MapType
	m.width = mapData.Width
	m.height = mapData.Height
	m.regionSize = mapData.RegionSize
	m.tileWidth = mapData.TileWidth
	m.tileHeight = mapData.TileHeight
	m.isInstance = mapData.IsInstance
	m.maxPlayers = mapData.MaxPlayers
	m.tileMap = mapData.TileMap

	// 重新初始化区域（如果地图尺寸改变）
	m.initRegions()

	// 加载地图对象
	m.loadObjects(mapData)

	m.SetId(m.mapId)

	zLog.Info("Map loaded successfully from file", zap.Int64("mapId", m.mapId), zap.String("mapName", m.name), zap.String("file_path", filePath))
	return nil
}

// initRegions 初始化地图区域
func (m *Map) initRegions() {
	// 计算区域数量
	regionCountX := int(m.width / m.regionSize)
	regionCountY := int(m.height / m.regionSize)

	// 创建区域
	for x := 0; x < regionCountX; x++ {
		for y := 0; y < regionCountY; y++ {
			regionId := y*regionCountX + x
			region := &MapRegion{
				regionId:   regionId,
				minX:       float64(x) * m.regionSize,
				minY:       float64(y) * m.regionSize,
				maxX:       float64(x+1) * m.regionSize,
				maxY:       float64(y+1) * m.regionSize,
				objects:    zMap.NewMap(),
				regionName: "Region_" + string(rune(regionId)),
			}
			m.regions.Store(regionId, region)
		}
	}
}

// loadObjects 加载地图中的各种对象
func (m *Map) loadObjects(mapData MapJSONData) {
	// 加载刷新点对象
	for _, spawnPoint := range mapData.SpawnPoints {
		// 根据类型创建不同的对象
		switch spawnPoint.Type {
		case "npc":
			m.spawnNPC(spawnPoint)
		case "monster":
			m.spawnMonsters(spawnPoint)
		}
	}

	// 加载传送点对象
	for _, teleportPoint := range mapData.TeleportPoints {
		m.spawnTeleportPoint(teleportPoint)
	}

	// 加载建筑对象
	for _, building := range mapData.Buildings {
		m.spawnBuilding(building)
	}
}

// spawnNPC 生成NPC对象
func (m *Map) spawnNPC(spawnPoint SpawnPointData) {
	// 创建实际的NPC实例
	npcObj := npc.NewNPC(uint64(spawnPoint.ID), spawnPoint.Name, npc.NPCTypeCommon)

	// 设置NPC位置
	position := common.NewVector3(
		float32(spawnPoint.X),
		float32(spawnPoint.Y),
		float32(spawnPoint.Z),
	)
	npcObj.SetPosition(position)

	// 添加到地图
	m.AddObject(npcObj)
}

// spawnMonsters 生成怪物对象
func (m *Map) spawnMonsters(spawnPoint SpawnPointData) {
	count := spawnPoint.Count
	if count <= 0 {
		count = 1
	}

	// 生成指定数量的怪物
	for i := 0; i < count; i++ {
		// 为每个怪物生成唯一ID
		monsterId := uint64(spawnPoint.ID + int64(i))

		// 创建实际的Monster实例
		monsterObj := monster.NewMonster(monsterId, spawnPoint.Name)

		// 设置怪物位置
		position := common.NewVector3(
			float32(spawnPoint.X),
			float32(spawnPoint.Y),
			float32(spawnPoint.Z),
		)
		monsterObj.SetPosition(position)

		// 添加到地图
		m.AddObject(monsterObj)
	}
}

// spawnTeleportPoint 生成传送点对象
func (m *Map) spawnTeleportPoint(teleportPoint TeleportPointData) {
	// 创建传送点作为特殊NPC处理
	npcId := uint64(teleportPoint.ID)
	teleportNPC := npc.NewNPC(npcId, "Teleport Portal", npc.NPCTypeCommon)

	// 设置传送点位置
	position := common.NewVector3(
		float32(teleportPoint.X),
		float32(teleportPoint.Y),
		float32(teleportPoint.Z),
	)
	teleportNPC.SetPosition(position)

	// 为传送点添加传送属性（在实际应用中，这应该通过NPC的组件系统实现）
	// 这里简化处理，直接使用NPC的扩展数据

	// 添加到地图
	m.AddObject(teleportNPC)
}

// spawnBuilding 生成建筑对象
func (m *Map) spawnBuilding(building BuildingData) {
	// 创建建筑对象作为基础游戏对象
	buildingObj := object.NewGameObjectWithType(uint64(building.ID), building.Name, object.GameObjectTypeBuilding)

	// 设置建筑位置
	position := common.NewVector3(
		float32(building.X),
		float32(building.Y),
		float32(building.Z),
	)
	buildingObj.SetPosition(position)

	// 添加到地图
	m.AddObject(buildingObj)
}

// AddObject 添加对象到地图
func (m *Map) AddObject(obj common.IGameObject) {
	// 获取对象位置
	pos := obj.GetPosition()
	x, y := float64(pos.X), float64(pos.Y)

	// 计算对象所在区域
	regionId := m.getRegionId(x, y)

	// 获取区域
	regionInterface, exists := m.regions.Get(regionId)
	if !exists {
		zLog.Warn("Region not found", zap.Int("regionId", regionId), zap.Float64("x", x), zap.Float64("y", y))
		return
	}
	region := regionInterface.(*MapRegion)

	// 获取对象ID
	objectId := int64(obj.GetID())

	// 添加对象到地图
	m.objects.Store(objectId, obj)
	region.objects.Store(objectId, obj)

	// 根据对象类型将对象添加到相应的集合
	// 使用GetType方法获取对象类型，更加统一和可靠
	objectType := obj.GetType()

	switch objectType {
	case object.GameObjectTypeNPC:
		m.npcs.Store(objectId, obj)
		zLog.Debug("NPC added to map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	case object.GameObjectTypeMonster:
		m.monsters.Store(objectId, obj)
		zLog.Debug("Monster added to map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	case object.GameObjectTypePlayer:
		m.players.Store(objectId, obj)
		m.playerCount++
		zLog.Debug("Player added to map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	case object.GameObjectTypeItem:
		m.dropItems.Store(objectId, obj)
		zLog.Debug("Item added to map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	default:
		zLog.Debug("Object added to map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	}
}

// RemoveObject 从地图移除对象
func (m *Map) RemoveObject(objectId int64) {
	// 获取对象
	objInterface, exists := m.objects.Get(objectId)
	if !exists {
		return // 对象不存在
	}
	obj := objInterface.(common.IGameObject)

	// 获取对象位置
	pos := obj.GetPosition()
	x, y := float64(pos.X), float64(pos.Y)

	// 计算对象所在区域
	regionId := m.getRegionId(x, y)

	// 获取区域
	regionInterface, exists := m.regions.Get(regionId)
	if exists {
		region := regionInterface.(*MapRegion)
		region.objects.Delete(objectId)
	}

	// 从地图中移除对象
	m.objects.Delete(objectId)

	// 根据对象类型将对象从相应的集合中移除
	// 使用GetType方法获取对象类型，更加统一和可靠
	objectType := obj.GetType()

	switch objectType {
	case object.GameObjectTypeNPC:
		m.npcs.Delete(objectId)
		zLog.Debug("NPC removed from map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	case object.GameObjectTypeMonster:
		m.monsters.Delete(objectId)
		zLog.Debug("Monster removed from map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	case object.GameObjectTypePlayer:
		m.players.Delete(objectId)
		m.playerCount--
		zLog.Debug("Player removed from map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	case object.GameObjectTypeItem:
		m.dropItems.Delete(objectId)
		zLog.Debug("Item removed from map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	default:
		zLog.Debug("Object removed from map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
	}
}

// GetObject 获取地图对象
func (m *Map) GetObject(objectId int64) (common.IGameObject, bool) {
	obj, exists := m.objects.Get(objectId)
	if !exists {
		return nil, false
	}
	return obj.(common.IGameObject), true
}

// GetObjectsInRegion 获取区域内的对象
func (m *Map) GetObjectsInRegion(regionId int) ([]common.IGameObject, bool) {
	// 获取区域
	regionInterface, exists := m.regions.Get(regionId)
	if !exists {
		return nil, false
	}
	region := regionInterface.(*MapRegion)

	// 收集区域内的对象
	var objects []common.IGameObject
	region.objects.Range(func(key, value interface{}) bool {
		if value != nil {
			objects = append(objects, value.(common.IGameObject))
		}
		return true
	})

	return objects, true
}

// GetObjectsInRange 获取范围内的对象
func (m *Map) GetObjectsInRange(x, y, radius float64, objectType int) ([]common.IGameObject, bool) {
	// 收集范围内的对象
	var objects []common.IGameObject
	m.objects.Range(func(key, value interface{}) bool {
		obj := value.(common.IGameObject)

		// 获取对象位置
		pos := obj.GetPosition()
		objX, objY := float64(pos.X), float64(pos.Y)

		// 检查对象类型
		if objectType != 0 && obj.GetType() != objectType {
			return true
		}

		// 检查距离
		dx := objX - x
		dy := objY - y
		distance := dx*dx + dy*dy
		if distance <= radius*radius {
			objects = append(objects, obj)
		}
		return true
	})

	return objects, true
}

// getRegionId 获取区域ID
func (m *Map) getRegionId(x, y float64) int {
	regionX := int(x / m.regionSize)
	regionY := int(y / m.regionSize)
	regionCountX := int(m.width / m.regionSize)
	return regionY*regionCountX + regionX
}

// GetID 获取地图ID
func (m *Map) GetMapID() int64 {
	return m.mapId
}

// GetName 获取地图名称
func (m *Map) GetName() string {
	return m.name
}

// GetType 获取地图类型
func (m *Map) GetType() int {
	return m.mapType
}

// GetWidth 获取地图宽度
func (m *Map) GetWidth() float64 {
	return m.width
}

// GetHeight 获取地图高度
func (m *Map) GetHeight() float64 {
	return m.height
}

// IsInstance 检查是否为实例地图
func (m *Map) IsInstance() bool {
	return m.isInstance
}

// GetMaxPlayers 获取最大玩家数量
func (m *Map) GetMaxPlayers() int {
	return m.maxPlayers
}

// GetPlayerCount 获取当前玩家数量
func (m *Map) GetPlayerCount() int {
	return m.playerCount
}
