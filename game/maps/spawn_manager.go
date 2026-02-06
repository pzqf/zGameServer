package maps

import (
	"math/rand"
	"sync"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zGameServer/config/models"
	"github.com/pzqf/zGameServer/config/tables"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	monster "github.com/pzqf/zGameServer/game/monsters"
	"github.com/pzqf/zGameServer/game/systems/property"
	"go.uber.org/zap"
)

type SpawnInstance struct {
	ObjectID    common.ObjectIdType
	SpawnPoint  *models.SpawnPoint
	LastSpawned time.Time
}

type SpawnManager struct {
	mu            sync.RWMutex
	mapID         common.MapIdType
	spawnPoints   []*models.SpawnPoint
	instances     map[int32][]*SpawnInstance
	lastSpawnTime map[int32]time.Time
	parentMap     *Map
	stopCh        chan struct{}
}

func NewSpawnManager(mapID common.MapIdType, parentMap *Map) *SpawnManager {
	return &SpawnManager{
		mapID:         mapID,
		spawnPoints:   make([]*models.SpawnPoint, 0),
		instances:     make(map[int32][]*SpawnInstance),
		lastSpawnTime: make(map[int32]time.Time),
		parentMap:     parentMap,
		stopCh:        make(chan struct{}),
	}
}

func (sm *SpawnManager) Init(mapConfigID int32) {
	spawnPoints := tables.GlobalTableManager.GetSpawnPointsByMap(mapConfigID)
	if spawnPoints == nil {
		zLog.Debug("No spawn points found for map", zap.Int32("mapID", mapConfigID))
		return
	}

	sm.mu.Lock()
	sm.spawnPoints = spawnPoints
	sm.mu.Unlock()

	zLog.Info("Spawn manager initialized", zap.Int32("mapID", mapConfigID), zap.Int("spawnPointCount", len(spawnPoints)))

	for _, sp := range spawnPoints {
		sm.spawnInitialMonsters(sp)
	}

	go sm.spawnLoop()
}

func (sm *SpawnManager) spawnInitialMonsters(sp *models.SpawnPoint) {
	for i := int32(0); i < sp.MaxCount; i++ {
		sm.spawnMonster(sp)
	}
}

func (sm *SpawnManager) spawnMonster(sp *models.SpawnPoint) *monster.Monster {
	monsterConfig, ok := tables.GlobalTableManager.GetMonsterLoader().GetMonster(sp.MonsterID)
	if !ok {
		zLog.Warn("Monster config not found", zap.Int32("monsterID", sp.MonsterID))
		return nil
	}

	offsetX := (rand.Float32()*2 - 1) * sp.Radius
	offsetZ := (rand.Float32()*2 - 1) * sp.Radius

	objectID, err := common.GenerateObjectID()
	if err != nil {
		zLog.Error("Failed to generate object ID", zap.Error(err))
		return nil
	}

	m := monster.NewMonster(objectID, monsterConfig.Name)

	pos := gamecommon.Vector3{
		X: sp.PosX + offsetX,
		Y: sp.PosY,
		Z: sp.PosZ + offsetZ,
	}
	m.SetPosition(pos)

	m.SetProperty(property.PropertyMaxHP, float64(monsterConfig.HP))
	m.SetProperty(property.PropertyHP, float64(monsterConfig.HP))
	m.SetProperty(property.PropertyMaxMP, float64(monsterConfig.MP))
	m.SetProperty(property.PropertyMP, float64(monsterConfig.MP))
	m.SetProperty(property.PropertyPhysicalAttack, float64(monsterConfig.Attack))
	m.SetProperty(property.PropertyPhysicalDefense, float64(monsterConfig.Defense))
	m.SetProperty(property.PropertyHaste, float64(monsterConfig.Speed))
	m.SetProperty(property.PropertyExp, float64(monsterConfig.Exp))

	m.SetOnDeath(func() {
		sm.onMonsterDeath(sp.SpawnID, objectID)
	})

	sm.parentMap.AddObject(m)

	sm.mu.Lock()
	sm.instances[sp.SpawnID] = append(sm.instances[sp.SpawnID], &SpawnInstance{
		ObjectID:    objectID,
		SpawnPoint:  sp,
		LastSpawned: time.Now(),
	})
	sm.mu.Unlock()

	zLog.Debug("Monster spawned",
		zap.Int64("objectID", int64(objectID)),
		zap.String("name", monsterConfig.Name),
		zap.Int32("spawnID", sp.SpawnID))

	return m
}

func (sm *SpawnManager) onMonsterDeath(spawnID int32, objectID common.ObjectIdType) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.parentMap.RemoveObject(objectID)

	instances := sm.instances[spawnID]
	for i, inst := range instances {
		if inst.ObjectID == objectID {
			sm.instances[spawnID] = append(instances[:i], instances[i+1:]...)
			break
		}
	}

	zLog.Debug("Monster death recorded",
		zap.Int64("objectID", int64(objectID)),
		zap.Int32("spawnID", spawnID))
}

func (sm *SpawnManager) spawnLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.checkAndSpawn()
		case <-sm.stopCh:
			return
		}
	}
}

func (sm *SpawnManager) checkAndSpawn() {
	sm.mu.RLock()
	spawnPoints := make([]*models.SpawnPoint, len(sm.spawnPoints))
	copy(spawnPoints, sm.spawnPoints)
	sm.mu.RUnlock()

	for _, sp := range spawnPoints {
		sm.mu.RLock()
		currentCount := len(sm.instances[sp.SpawnID])
		lastSpawn := sm.lastSpawnTime[sp.SpawnID]
		sm.mu.RUnlock()

		if currentCount >= int(sp.MaxCount) {
			continue
		}

		if time.Since(lastSpawn) < time.Duration(sp.SpawnInterval)*time.Second {
			continue
		}

		sm.mu.Lock()
		sm.lastSpawnTime[sp.SpawnID] = time.Now()
		sm.mu.Unlock()

		sm.spawnMonster(sp)
	}
}

func (sm *SpawnManager) Stop() {
	select {
	case <-sm.stopCh:
	default:
		close(sm.stopCh)
	}
}

func (sm *SpawnManager) GetSpawnedCount(spawnID int32) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.instances[spawnID])
}
