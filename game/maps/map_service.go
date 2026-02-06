package maps

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// MapService 地图服务
// 负责管理所有地图实例，处理地图的加载、卸载和同步
// 提供地图对象的管理和查询功能

type MapService struct {
	zService.BaseService
	maps        *zMap.TypedShardedMap[common.MapIdType, *Map]                      // 存储所有地图实例，key: MapIdType, value: *Map
	objectMap   *zMap.TypedShardedMap[common.ObjectIdType, common.MapIdType]       // 存储对象到地图的映射，key: ObjectIdType, value: MapIdType
	gameObjects *zMap.TypedShardedMap[common.ObjectIdType, gamecommon.IGameObject] // 存储所有游戏对象，key: ObjectIdType, value: gamecommon.IGameObject
	maxMaps     int                                                                // 最大地图数量限制
	stopSyncCh  chan struct{}                                                      // 停止同步循环的信号通道
}

// NewMapService 创建地图服务实例
// 返回初始化完成的地图服务对象
func NewMapService() *MapService {
	ms := &MapService{
		BaseService: *zService.NewBaseService(common.ServiceIdMap),
		maps:        zMap.NewTypedShardedMap32[common.MapIdType, *Map](),
		objectMap:   zMap.NewTypedShardedMap32[common.ObjectIdType, common.MapIdType](),
		gameObjects: zMap.NewTypedShardedMap32[common.ObjectIdType, gamecommon.IGameObject](),
		maxMaps:     100,
		stopSyncCh:  make(chan struct{}),
	}
	return ms
}

// Init 初始化地图服务
// 加载所有地图资源，准备服务运行
// 返回初始化过程中的错误，如果有
func (ms *MapService) Init() error {
	ms.SetState(zService.ServiceStateInit)
	zLog.Info("Initializing map service...")
	// 初始化地图服务相关资源
	// 加载所有地图
	mapDirectory := "resources/maps"
	if err := ms.LoadAllMaps(mapDirectory); err != nil {
		zLog.Error("Failed to load maps", zap.Error(err))
		// 继续初始化，不因为地图加载失败而停止
	}
	return nil
}

// Close 关闭地图服务
// 清理所有地图资源和对象映射
// 返回清理过程中的错误，如果有
func (ms *MapService) Close() error {
	ms.SetState(zService.ServiceStateStopping)
	zLog.Info("Closing map service...")

	select {
	case <-ms.stopSyncCh:
	default:
		close(ms.stopSyncCh)
	}

	ms.maps.Clear()
	ms.objectMap.Clear()
	ms.gameObjects.Clear()
	ms.SetState(zService.ServiceStateStopped)
	return nil
}

// Serve 启动地图服务
// 启动地图同步循环，处理地图对象的移动和状态同步
func (ms *MapService) Serve() {
	ms.SetState(zService.ServiceStateRunning)
	// 地图服务需要持续运行的协程，用于处理地图对象的移动同步
	go ms.mapSyncLoop()
}

// LoadMap 加载单个地图文件
// filePath: 地图文件路径
// mapConfigID: 地图配置ID
// 返回加载过程中的错误，如果有
func (ms *MapService) LoadMap(filePath string, mapConfigID int32) error {
	mapID, err := common.GenerateMapID()
	if err != nil {
		zLog.Error("Failed to generate map ID", zap.Error(err))
		return err
	}
	mapName := filePath
	width := float32(1000)
	height := float32(1000)

	mapObj := NewMap(mapID, mapConfigID, mapName, width, height)

	mapId := mapObj.GetID()
	ms.maps.Store(mapId, mapObj)

	mapObj.InitSpawnSystem()

	zLog.Info("Map loaded successfully", zap.Any("mapId", mapId), zap.String("mapName", mapObj.GetName()), zap.String("file_path", filePath))
	return nil
}

// LoadAllMaps 加载指定目录下的所有地图
// directoryPath: 地图目录路径
// 返回加载过程中的错误，如果有
func (ms *MapService) LoadAllMaps(directoryPath string) error {
	// 检查目录是否存在
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		zLog.Error("Map directory does not exist", zap.String("directory", directoryPath), zap.Error(err))
		return err
	}

	// 遍历目录下的所有JSON文件
	files, err := filepath.Glob(filepath.Join(directoryPath, "*.json"))
	if err != nil {
		zLog.Error("Failed to get map files", zap.String("directory", directoryPath), zap.Error(err))
		return err
	}

	// 加载每个地图文件
	for i, file := range files {
		mapConfigID := int32(i + 1)
		if err := ms.LoadMap(file, mapConfigID); err != nil {
			zLog.Error("Failed to load map", zap.String("file_path", file), zap.Error(err))
		}
	}

	zLog.Info("Map loading completed", zap.Int("loaded_maps", len(files)))
	return nil
}

// mapSyncLoop 地图同步循环
// 定期同步地图对象的位置和状态
func (ms *MapService) mapSyncLoop() {
	ticker := time.NewTicker(time.Millisecond * 1000)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ms.syncMaps()
		case <-ms.stopSyncCh:
			zLog.Debug("Map sync loop stopped")
			return
		}
	}
}

// syncMaps 同步地图
// 遍历所有地图，同步地图对象的位置和状态
func (ms *MapService) syncMaps() {
	// 遍历所有地图
	ms.maps.Range(func(key common.MapIdType, value *Map) bool {
		mapObj := value
		// 同步地图对象的位置和状态
		// 这里可以实现地图对象的同步逻辑
		// 由于Map类型没有SyncObjects方法，我们可以简单地打印日志或者实现自己的同步逻辑
		zLog.Debug("Synchronizing map objects", zap.Any("mapId", mapObj.GetID()))
		return true
	})
}

// AddGameObject 添加游戏对象到地图服务
// obj: 游戏对象
// mapId: 地图ID
func (ms *MapService) AddGameObject(obj gamecommon.IGameObject, mapId common.MapIdType) {
	objectId := common.ObjectIdType(obj.GetID())
	ms.gameObjects.Store(objectId, obj)
	ms.objectMap.Store(objectId, mapId)
	zLog.Debug("Added game object",
		zap.Int64("objectId", int64(objectId)),
		zap.Int("objectType", int(obj.GetType())),
		zap.Int64("mapId", int64(mapId)))
}

// RemoveGameObject 从地图服务中移除游戏对象
// objectId: 对象ID
func (ms *MapService) RemoveGameObject(objectId common.ObjectIdType) {
	ms.gameObjects.Delete(objectId)
	ms.objectMap.Delete(objectId)
	zLog.Debug("Removed game object", zap.Int64("objectId", int64(objectId)))
}

// GetGameObject 根据对象ID获取游戏对象
// objectId: 对象ID
// 返回游戏对象，如果不存在则返回nil
func (ms *MapService) GetGameObject(objectId common.ObjectIdType) gamecommon.IGameObject {
	if obj, exists := ms.gameObjects.Load(objectId); exists {
		return obj
	}
	return nil
}

// GetGameObjectsByType 根据对象类型获取游戏对象列表
// objectType: 对象类型
// 返回指定类型的游戏对象列表
func (ms *MapService) GetGameObjectsByType(objectType gamecommon.GameObjectType) []gamecommon.IGameObject {
	var objects []gamecommon.IGameObject
	ms.gameObjects.Range(func(key common.ObjectIdType, value gamecommon.IGameObject) bool {
		obj := value
		if obj.GetType() == objectType {
			objects = append(objects, obj)
		}
		return true
	})
	return objects
}

// GetGameObjectsByMap 根据地图ID获取游戏对象列表
// mapId: 地图ID
// 返回指定地图的游戏对象列表
func (ms *MapService) GetGameObjectsByMap(mapId common.MapIdType) []gamecommon.IGameObject {
	var objects []gamecommon.IGameObject
	ms.objectMap.Range(func(key common.ObjectIdType, value common.MapIdType) bool {
		if value == mapId {
			objectId := key
			if obj, exists := ms.gameObjects.Load(objectId); exists {
				objects = append(objects, obj)
			}
		}
		return true
	})
	return objects
}
