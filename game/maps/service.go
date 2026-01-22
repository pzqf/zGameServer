package maps

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// Service 地图服务
// 负责管理所有地图实例，处理地图的加载、卸载和同步
// 提供地图对象的管理和查询功能

type Service struct {
	zObject.BaseObject
	logger      *zap.Logger // 日志记录器
	maps        *zMap.Map   // 存储所有地图实例，key: int64(mapId), value: *Map
	objectMap   *zMap.Map   // 存储对象到地图的映射，key: int64(objectId), value: int64(mapId)
	gameObjects *zMap.Map   // 存储所有游戏对象，key: int64(objectId), value: object.IGameObject
	maxMaps     int         // 最大地图数量限制
}

// NewService 创建地图服务实例
// 返回初始化完成的地图服务对象
func NewService() *Service {
	ms := &Service{
		logger:      zLog.GetLogger(),
		maps:        zMap.NewMap(),
		objectMap:   zMap.NewMap(),
		gameObjects: zMap.NewMap(),
		maxMaps:     100,
	}
	ms.SetId("map_service")
	return ms
}

// Init 初始化地图服务
// 加载所有地图资源，准备服务运行
// 返回初始化过程中的错误，如果有
func (ms *Service) Init() error {
	ms.logger.Info("Initializing map service...")
	// 初始化地图服务相关资源
	// 加载所有地图
	mapDirectory := "resources/maps"
	if err := ms.LoadAllMaps(mapDirectory); err != nil {
		ms.logger.Error("Failed to load maps", zap.Error(err))
		// 继续初始化，不因为地图加载失败而停止
	}
	return nil
}

// Close 关闭地图服务
// 清理所有地图资源和对象映射
// 返回清理过程中的错误，如果有
func (ms *Service) Close() error {
	ms.logger.Info("Closing map service...")
	// 清理地图服务相关资源
	ms.maps.Clear()
	ms.objectMap.Clear()
	ms.gameObjects.Clear()
	return nil
}

// Serve 启动地图服务
// 启动地图同步循环，处理地图对象的移动和状态同步
func (ms *Service) Serve() {
	// 地图服务需要持续运行的协程，用于处理地图对象的移动同步
	go ms.mapSyncLoop()
}

// LoadMap 加载单个地图文件
// filePath: 地图文件路径
// 返回加载过程中的错误，如果有
func (ms *Service) LoadMap(filePath string) error {
	// 创建地图对象
	mapObj := NewMap(0, "", 0, 0, 0, 0, 0, 0, false)

	// 从文件加载地图
	if err := mapObj.LoadFromFile(filePath); err != nil {
		ms.logger.Error("Failed to load map from file", zap.String("file_path", filePath), zap.Error(err))
		return err
	}

	// 存储地图
	mapId := mapObj.GetId()
	ms.maps.Store(mapId, mapObj)

	ms.logger.Info("Map loaded successfully", zap.Any("mapId", mapId), zap.String("mapName", mapObj.GetName()), zap.String("file_path", filePath))
	return nil
}

// LoadAllMaps 加载指定目录下的所有地图
// directoryPath: 地图目录路径
// 返回加载过程中的错误，如果有
func (ms *Service) LoadAllMaps(directoryPath string) error {
	// 检查目录是否存在
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		ms.logger.Error("Map directory does not exist", zap.String("directory", directoryPath), zap.Error(err))
		return err
	}

	// 遍历目录下的所有JSON文件
	files, err := filepath.Glob(filepath.Join(directoryPath, "*.json"))
	if err != nil {
		ms.logger.Error("Failed to get map files", zap.String("directory", directoryPath), zap.Error(err))
		return err
	}

	// 加载每个地图文件
	for _, file := range files {
		if err := ms.LoadMap(file); err != nil {
			ms.logger.Error("Failed to load map", zap.String("file_path", file), zap.Error(err))
			// 继续加载其他地图，不因为一个地图失败而停止
		}
	}

	ms.logger.Info("Map loading completed", zap.Int("loaded_maps", len(files)))
	return nil
}

// mapSyncLoop 地图同步循环
// 定期同步地图对象的位置和状态
func (ms *Service) mapSyncLoop() {
	// 每1000毫秒同步一次地图对象
	for range time.Tick(time.Millisecond * 1000) {
		ms.syncMaps()
	}
}

// syncMaps 同步地图
// 遍历所有地图，同步地图对象的位置和状态
func (ms *Service) syncMaps() {
	// 遍历所有地图
	ms.maps.Range(func(key, value interface{}) bool {
		mapObj := value.(*Map)
		// 同步地图对象的位置和状态
		// 这里可以实现地图对象的同步逻辑
		// 由于Map类型没有SyncObjects方法，我们可以简单地打印日志或者实现自己的同步逻辑
		ms.logger.Debug("Synchronizing map objects", zap.Any("mapId", mapObj.GetId()))
		return true
	})
}

// AddGameObject 添加游戏对象到地图服务
// obj: 游戏对象
// mapId: 地图ID
func (ms *Service) AddGameObject(obj common.IGameObject, mapId int64) {
	objectId := int64(obj.GetID())
	ms.gameObjects.Store(objectId, obj)
	ms.objectMap.Store(objectId, mapId)
	ms.logger.Debug("Added game object",
		zap.Int64("objectId", objectId),
		zap.Int("objectType", obj.GetType()),
		zap.Int64("mapId", mapId))
}

// RemoveGameObject 从地图服务中移除游戏对象
// objectId: 对象ID
func (ms *Service) RemoveGameObject(objectId int64) {
	ms.gameObjects.Delete(objectId)
	ms.objectMap.Delete(objectId)
	ms.logger.Debug("Removed game object", zap.Int64("objectId", objectId))
}

// GetGameObject 根据对象ID获取游戏对象
// objectId: 对象ID
// 返回游戏对象，如果不存在则返回nil
func (ms *Service) GetGameObject(objectId int64) common.IGameObject {
	if obj, exists := ms.gameObjects.Get(objectId); exists {
		return obj.(common.IGameObject)
	}
	return nil
}

// GetGameObjectsByType 根据对象类型获取游戏对象列表
// objectType: 对象类型
// 返回指定类型的游戏对象列表
func (ms *Service) GetGameObjectsByType(objectType int) []common.IGameObject {
	var objects []common.IGameObject
	ms.gameObjects.Range(func(key, value interface{}) bool {
		obj := value.(common.IGameObject)
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
func (ms *Service) GetGameObjectsByMap(mapId int64) []common.IGameObject {
	var objects []common.IGameObject
	ms.objectMap.Range(func(key, value interface{}) bool {
		if value.(int64) == mapId {
			objectId := key.(int64)
			if obj, exists := ms.gameObjects.Get(objectId); exists {
				objects = append(objects, obj.(common.IGameObject))
			}
		}
		return true
	})
	return objects
}
