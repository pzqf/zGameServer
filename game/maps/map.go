package maps

import (
	"sync"
	"time"

	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zGameServer/config/models"
	gamecommon "github.com/pzqf/zGameServer/game/common"
)

// Region 地图区域
// 用于空间分区，管理区域内的游戏对象
type Region struct {
	mu       sync.RWMutex                                   // 读写锁
	regionID common.RegionIdType                            // 区域ID
	objects  map[common.ObjectIdType]gamecommon.IGameObject // 区域内的游戏对象
}

// AddObject 添加游戏对象到区域
func (r *Region) AddObject(object gamecommon.IGameObject) {
	if object == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.objects[object.GetID()] = object
}

// Map 游戏地图
// 管理地图中的所有游戏对象、区域、刷新点等
type Map struct {
	mu             sync.RWMutex                                   // 读写锁
	mapID          common.MapIdType                               // 地图ID
	mapConfigID    int32                                          // 地图配置ID
	name           string                                         // 地图名称
	width          float32                                        // 地图宽度
	height         float32                                        // 地图高度
	regionSize     float32                                        // 区域大小
	objects        map[common.ObjectIdType]gamecommon.IGameObject // 所有游戏对象
	regions        map[common.RegionIdType]*Region                // 区域映射表
	spawnPoints    []*models.MapSpawnPoint                        // 刷新点列表
	teleportPoints []*models.MapTeleportPoint                     // 传送点列表
	buildings      []*models.MapBuilding                          // 建筑列表
	events         []*models.MapEvent                             // 事件列表
	resources      []*models.MapResource                          // 资源点列表
	players        map[common.PlayerIdType]bool                   // 在线玩家
	spawnManager   *SpawnManager                                  // 刷新管理器
	createdAt      time.Time                                      // 创建时间
}

// NewMap 创建新地图
// 参数:
//   - mapID: 地图ID
//   - mapConfigID: 地图配置ID
//   - name: 地图名称
//   - width: 地图宽度
//   - height: 地图高度
//
// 返回: 新创建的地图对象
func NewMap(mapID common.MapIdType, mapConfigID int32, name string, width, height float32) *Map {
	m := &Map{
		mapID:          mapID,
		mapConfigID:    mapConfigID,
		name:           name,
		width:          width,
		height:         height,
		regionSize:     50,
		objects:        make(map[common.ObjectIdType]gamecommon.IGameObject),
		regions:        make(map[common.RegionIdType]*Region),
		spawnPoints:    make([]*models.MapSpawnPoint, 0),
		teleportPoints: make([]*models.MapTeleportPoint, 0),
		buildings:      make([]*models.MapBuilding, 0),
		events:         make([]*models.MapEvent, 0),
		resources:      make([]*models.MapResource, 0),
		players:        make(map[common.PlayerIdType]bool),
		createdAt:      time.Now(),
	}

	m.spawnManager = NewSpawnManager(mapID, m)
	return m
}

// InitSpawnSystem 初始化刷怪系统
func (m *Map) InitSpawnSystem() {
	if m.spawnManager != nil {
		m.spawnManager.Init(m.mapConfigID)
	}
}

// GetSpawnManager 获取刷新管理器
func (m *Map) GetSpawnManager() *SpawnManager {
	return m.spawnManager
}

// GetID 获取地图ID
func (m *Map) GetID() common.MapIdType {
	return m.mapID
}

// GetName 获取地图名称
func (m *Map) GetName() string {
	return m.name
}

// GetObjectsInRange 获取指定范围内的游戏对象
// 参数:
//   - center: 中心坐标
//   - radius: 半径
//
// 返回: 游戏对象列表
func (m *Map) GetObjectsInRange(center gamecommon.Vector3, radius float32) []gamecommon.IGameObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	objects := make([]gamecommon.IGameObject, 0)

	for _, obj := range m.objects {
		distance := obj.GetPosition().DistanceTo(center)
		if distance <= radius*radius {
			objects = append(objects, obj)
		}
	}

	return objects
}

// AddObject 添加游戏对象到地图
func (m *Map) AddObject(object gamecommon.IGameObject) {
	if object == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	objectID := object.GetID()
	m.objects[objectID] = object

	// 添加到对应的区域
	regionID := m.getRegionID(object.GetPosition())
	if _, exists := m.regions[regionID]; !exists {
		m.regions[regionID] = &Region{
			regionID: regionID,
			objects:  make(map[common.ObjectIdType]gamecommon.IGameObject),
		}
	}

	m.regions[regionID].AddObject(object)
}

// RemoveObject 从地图移除游戏对象
func (m *Map) RemoveObject(objectID common.ObjectIdType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.objects, objectID)

	// 从区域中移除
	for regionID, region := range m.regions {
		if _, exists := region.objects[objectID]; exists {
			delete(m.regions[regionID].objects, objectID)
			break
		}
	}
}

// MoveObject 移动游戏对象
// 参数:
//   - object: 游戏对象
//   - targetPos: 目标位置
//
// 返回: 移动错误
func (m *Map) MoveObject(object gamecommon.IGameObject, targetPos gamecommon.Vector3) error {
	oldPos := object.GetPosition()
	oldRegionID := m.getRegionID(oldPos)
	newRegionID := m.getRegionID(targetPos)

	// 同一区域内移动
	if oldRegionID == newRegionID {
		object.SetPosition(targetPos)
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 从旧区域移除
	if _, exists := m.regions[oldRegionID]; exists {
		delete(m.regions[oldRegionID].objects, object.GetID())
	}

	// 添加到新区域
	if _, exists := m.regions[newRegionID]; !exists {
		m.regions[newRegionID] = &Region{
			regionID: newRegionID,
			objects:  make(map[common.ObjectIdType]gamecommon.IGameObject),
		}
	}

	m.regions[newRegionID].AddObject(object)
	object.SetPosition(targetPos)

	return nil
}

// TeleportObject 传送游戏对象
// 参数:
//   - object: 游戏对象
//   - targetPos: 目标位置
//
// 返回: 传送错误
func (m *Map) TeleportObject(object gamecommon.IGameObject, targetPos gamecommon.Vector3) error {
	object.SetPosition(targetPos)
	return nil
}

// getRegionID 根据坐标计算区域ID
// 使用简单的网格分区算法
func (m *Map) getRegionID(pos gamecommon.Vector3) common.RegionIdType {
	if m.regionSize <= 0 {
		return 0
	}

	xRegion := uint64(pos.X / m.regionSize)
	yRegion := uint64(pos.Y / m.regionSize)

	return common.RegionIdType(xRegion*1000000 + yRegion)
}

// GetSize 获取地图尺寸
// 返回: 宽度和高度
func (m *Map) GetSize() (float32, float32) {
	return m.width, m.height
}

// GetObjectsByType 获取指定类型的游戏对象
// 参数:
//   - objectType: 对象类型
//
// 返回: 游戏对象列表
func (m *Map) GetObjectsByType(objectType gamecommon.GameObjectType) []gamecommon.IGameObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	objects := make([]gamecommon.IGameObject, 0)

	for _, obj := range m.objects {
		if obj.GetType() == objectType {
			objects = append(objects, obj)
		}
	}

	return objects
}
