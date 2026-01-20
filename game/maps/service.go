package maps

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

type Service struct {
	zObject.BaseObject
	logger     *zap.Logger
	maps       *zMap.Map // key: int64(mapId), value: *Map
	objectMap  *zMap.Map // key: int64(objectId), value: int64(mapId)
	npcMap     *zMap.Map // key: int64(npcId), value: *MapObject
	monsterMap *zMap.Map // key: int64(monsterId), value: *MapObject
	playerMap  *zMap.Map // key: int64(playerId), value: *MapObject
	maxMaps    int
}

func NewService() *Service {
	ms := &Service{
		logger:     zLog.GetLogger(),
		maps:       zMap.NewMap(),
		objectMap:  zMap.NewMap(),
		npcMap:     zMap.NewMap(),
		monsterMap: zMap.NewMap(),
		playerMap:  zMap.NewMap(),
		maxMaps:    100,
	}
	ms.SetId("map_service")
	return ms
}

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

func (ms *Service) Close() error {
	ms.logger.Info("Closing map service...")
	// 清理地图服务相关资源
	ms.maps.Clear()
	ms.objectMap.Clear()
	ms.npcMap.Clear()
	ms.monsterMap.Clear()
	ms.playerMap.Clear()
	return nil
}

func (ms *Service) Serve() {
	// 地图服务需要持续运行的协程，用于处理地图对象的移动同步
	go ms.mapSyncLoop()
}

// LoadMap 加载单个地图文件
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
func (ms *Service) mapSyncLoop() {
	// 每100毫秒同步一次地图对象
	for range time.Tick(time.Millisecond * 1000) {
		ms.syncMapObjects()
	}
}

// syncMapObjects 同步地图对象
func (ms *Service) syncMapObjects() {
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
