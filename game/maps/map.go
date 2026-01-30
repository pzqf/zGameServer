package maps

import (
	"sync"
	"time"

	"github.com/pzqf/zGameServer/config/models"
	"github.com/pzqf/zGameServer/game/common"
)

type Region struct {
	mu       sync.RWMutex
	regionID uint64
	objects  map[uint64]common.IGameObject
}

func (r *Region) AddObject(object common.IGameObject) {
	if object == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.objects[object.GetID()] = object
}

type Map struct {
	mu             sync.RWMutex
	mapID          uint64
	name           string
	width          float32
	height         float32
	regionSize     float32
	objects        map[uint64]common.IGameObject
	regions        map[uint64]*Region
	spawnPoints    []*models.MapSpawnPoint
	teleportPoints []*models.MapTeleportPoint
	buildings      []*models.MapBuilding
	events         []*models.MapEvent
	resources      []*models.MapResource
	players        map[uint64]bool
	createdAt      time.Time
}

func NewMap(mapID uint64, name string, width, height float32) *Map {
	return &Map{
		mapID:          mapID,
		name:           name,
		width:          width,
		height:         height,
		regionSize:     50,
		objects:        make(map[uint64]common.IGameObject),
		regions:        make(map[uint64]*Region),
		spawnPoints:    make([]*models.MapSpawnPoint, 0),
		teleportPoints: make([]*models.MapTeleportPoint, 0),
		buildings:      make([]*models.MapBuilding, 0),
		events:         make([]*models.MapEvent, 0),
		resources:      make([]*models.MapResource, 0),
		players:        make(map[uint64]bool),
		createdAt:      time.Now(),
	}
}

func (m *Map) GetID() uint64 {
	return m.mapID
}

func (m *Map) GetName() string {
	return m.name
}

func (m *Map) GetObjectsInRange(center common.Vector3, radius float32) []common.IGameObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	objects := make([]common.IGameObject, 0)

	for _, obj := range m.objects {
		distance := obj.GetPosition().DistanceTo(center)
		if distance <= radius*radius {
			objects = append(objects, obj)
		}
	}

	return objects
}

func (m *Map) AddObject(object common.IGameObject) {
	if object == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	objectID := object.GetID()
	m.objects[objectID] = object

	regionID := m.getRegionID(object.GetPosition())
	if _, exists := m.regions[regionID]; !exists {
		m.regions[regionID] = &Region{
			regionID: regionID,
			objects:  make(map[uint64]common.IGameObject),
		}
	}

	m.regions[regionID].AddObject(object)
}

func (m *Map) RemoveObject(objectID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.objects, objectID)

	for regionID, region := range m.regions {
		if _, exists := region.objects[objectID]; exists {
			delete(m.regions[regionID].objects, objectID)
			break
		}
	}
}

func (m *Map) MoveObject(object common.IGameObject, targetPos common.Vector3) error {
	oldPos := object.GetPosition()
	oldRegionID := m.getRegionID(oldPos)
	newRegionID := m.getRegionID(targetPos)

	if oldRegionID == newRegionID {
		object.SetPosition(targetPos)
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.regions[oldRegionID]; exists {
		delete(m.regions[oldRegionID].objects, object.GetID())
	}

	if _, exists := m.regions[newRegionID]; !exists {
		m.regions[newRegionID] = &Region{
			regionID: newRegionID,
			objects:  make(map[uint64]common.IGameObject),
		}
	}

	m.regions[newRegionID].AddObject(object)
	object.SetPosition(targetPos)

	return nil
}

func (m *Map) TeleportObject(object common.IGameObject, targetPos common.Vector3) error {
	object.SetPosition(targetPos)
	return nil
}

func (m *Map) getRegionID(pos common.Vector3) uint64 {
	if m.regionSize <= 0 {
		return 0
	}

	xRegion := uint64(pos.X / m.regionSize)
	yRegion := uint64(pos.Y / m.regionSize)

	return xRegion*1000000 + yRegion
}

func (m *Map) GetSize() (float32, float32) {
	return m.width, m.height
}

func (m *Map) GetObjectsByType(objectType common.GameObjectType) []common.IGameObject {
	m.mu.RLock()
	defer m.mu.Unlock()

	objects := make([]common.IGameObject, 0)

	for _, obj := range m.objects {
		if obj.GetType() == int(objectType) {
			objects = append(objects, obj)
		}
	}

	return objects
}
