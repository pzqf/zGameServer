package maps

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pzqf/zEngine/zLog"
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
	ObjectId    int64     // 对象ID
	ObjectType  int       // 对象类型
	MapId       int64     // 所属地图ID
	X           float64   // X坐标
	Y           float64   // Y坐标
	Z           float64   // Z坐标
	Orientation float64   // 方向
	MoveSpeed   float64   // 移动速度
	IsMoving    bool      // 是否移动中
	Status      int       // 状态
	Properties  *zMap.Map // 附加属性
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
	objects       *zMap.Map // key: int64(objectId), value: *MapObject
	regions       *zMap.Map // key: int(regionId), value: *MapRegion
	npcs          *zMap.Map // key: int64(npcId), value: *MapObject
	monsters      *zMap.Map // key: int64(monsterId), value: *MapObject
	dropItems     *zMap.Map // key: int64(itemId), value: *MapObject
	isInstance    bool      // 是否为实例地图
	instanceOwner int64     // 实例所有者（如果有）
	maxPlayers    int
	playerCount   int
	logger        *zap.Logger
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
		isInstance:    isInstance,
		instanceOwner: 0,
		maxPlayers:    100,
		playerCount:   0,
		logger:        zLog.GetLogger(),
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
		m.logger.Error("Failed to read map file", zap.String("file_path", filePath), zap.Error(err))
		return err
	}

	// 解析JSON数据
	var mapData MapJSONData
	if err := json.Unmarshal(data, &mapData); err != nil {
		m.logger.Error("Failed to parse map JSON", zap.String("file_path", filePath), zap.Error(err))
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

	m.logger.Info("Map loaded successfully from file", zap.Int64("mapId", m.mapId), zap.String("mapName", m.name), zap.String("file_path", filePath))
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
	// 创建NPC对象
	npcObj := &MapObject{
		ObjectId:   spawnPoint.ID,
		ObjectType: MapObjectTypeNPC,
		MapId:      m.mapId,
		X:          spawnPoint.X,
		Y:          spawnPoint.Y,
		Z:          spawnPoint.Z,
		Properties: zMap.NewMap(),
	}

	// 设置NPC属性
	npcObj.Properties.Store("name", spawnPoint.Name)

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
		monsterId := spawnPoint.ID + int64(i)

		// 创建怪物对象
		monsterObj := &MapObject{
			ObjectId:   monsterId,
			ObjectType: MapObjectTypeMonster,
			MapId:      m.mapId,
			X:          spawnPoint.X,
			Y:          spawnPoint.Y,
			Z:          spawnPoint.Z,
			Properties: zMap.NewMap(),
		}

		// 设置怪物属性
		monsterObj.Properties.Store("name", spawnPoint.Name)

		// 添加到地图
		m.AddObject(monsterObj)
	}
}

// spawnTeleportPoint 生成传送点对象
func (m *Map) spawnTeleportPoint(teleportPoint TeleportPointData) {
	// 创建传送点对象（作为特殊NPC处理）
	teleportObj := &MapObject{
		ObjectId:   int64(teleportPoint.ID),
		ObjectType: MapObjectTypeNPC,
		MapId:      m.mapId,
		X:          teleportPoint.X,
		Y:          teleportPoint.Y,
		Z:          teleportPoint.Z,
		Properties: zMap.NewMap(),
	}

	// 设置传送点属性
	teleportObj.Properties.Store("name", "Teleport Portal")
	teleportObj.Properties.Store("target_map_id", teleportPoint.TargetMapID)
	teleportObj.Properties.Store("target_x", teleportPoint.TargetX)
	teleportObj.Properties.Store("target_y", teleportPoint.TargetY)
	teleportObj.Properties.Store("target_z", teleportPoint.TargetZ)

	// 添加到地图
	m.AddObject(teleportObj)
}

// spawnBuilding 生成建筑对象
func (m *Map) spawnBuilding(building BuildingData) {
	// 创建建筑对象
	buildingObj := &MapObject{
		ObjectId:   building.ID,
		ObjectType: MapObjectTypeBuilding,
		MapId:      m.mapId,
		X:          building.X,
		Y:          building.Y,
		Z:          building.Z,
		Properties: zMap.NewMap(),
	}

	// 设置建筑属性
	buildingObj.Properties.Store("name", building.Name)
	buildingObj.Properties.Store("type", building.Type)
	buildingObj.Properties.Store("width", building.Width)
	buildingObj.Properties.Store("height", building.Height)

	// 添加到地图
	m.AddObject(buildingObj)
}

// AddObject 添加对象到地图
func (m *Map) AddObject(obj *MapObject) {
	// 计算对象所在区域
	regionId := m.getRegionId(obj.X, obj.Y)

	// 获取区域
	regionInterface, exists := m.regions.Get(regionId)
	if !exists {
		m.logger.Warn("Region not found", zap.Int("regionId", regionId), zap.Float64("x", obj.X), zap.Float64("y", obj.Y))
		return
	}
	region := regionInterface.(*MapRegion)

	// 添加对象到地图
	m.objects.Store(obj.ObjectId, obj)
	region.objects.Store(obj.ObjectId, obj)

	// 记录特定类型对象的映射关系
	switch obj.ObjectType {
	case MapObjectTypeNPC:
		m.npcs.Store(obj.ObjectId, obj)
	case MapObjectTypeMonster:
		m.monsters.Store(obj.ObjectId, obj)
	case MapObjectTypeItem:
		m.dropItems.Store(obj.ObjectId, obj)
	}

	m.logger.Debug("Object added to map", zap.Int64("objectId", obj.ObjectId), zap.Int("objectType", obj.ObjectType), zap.Int64("mapId", m.mapId))
}

// RemoveObject 从地图移除对象
func (m *Map) RemoveObject(objectId int64) {
	// 获取对象
	objInterface, exists := m.objects.Get(objectId)
	if !exists {
		return // 对象不存在
	}
	obj := objInterface.(*MapObject)

	// 计算对象所在区域
	regionId := m.getRegionId(obj.X, obj.Y)

	// 获取区域
	regionInterface, exists := m.regions.Get(regionId)
	if exists {
		region := regionInterface.(*MapRegion)
		region.objects.Delete(objectId)
	}

	// 从地图中移除对象
	m.objects.Delete(objectId)

	// 从特定类型映射中移除
	switch obj.ObjectType {
	case MapObjectTypeNPC:
		m.npcs.Delete(objectId)
	case MapObjectTypeMonster:
		m.monsters.Delete(objectId)
	case MapObjectTypeItem:
		m.dropItems.Delete(objectId)
	}

	m.logger.Debug("Object removed from map", zap.Int64("objectId", objectId), zap.Int64("mapId", m.mapId))
}

// GetObject 获取地图对象
func (m *Map) GetObject(objectId int64) (*MapObject, bool) {
	obj, exists := m.objects.Get(objectId)
	if !exists {
		return nil, false
	}
	return obj.(*MapObject), true
}

// GetObjectsInRegion 获取区域内的对象
func (m *Map) GetObjectsInRegion(regionId int) ([]*MapObject, bool) {
	// 获取区域
	regionInterface, exists := m.regions.Get(regionId)
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
func (m *Map) GetObjectsInRange(x, y, radius float64, objectType int) ([]*MapObject, bool) {
	// 收集范围内的对象
	var objects []*MapObject
	m.objects.Range(func(key, value interface{}) bool {
		obj := value.(*MapObject)
		// 检查对象类型
		if objectType != 0 && obj.ObjectType != objectType {
			return true
		}
		// 检查距离
		dx := obj.X - x
		dy := obj.Y - y
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
