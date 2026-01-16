package tables

import (
	"github.com/pzqf/zGameServer/config/models"
)

// Load 加载地图表数据
func (mtl *MapTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "map.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 9,
		TableName:  "maps",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempMaps := make(map[int32]*models.Map)

	err := ReadExcelFile(config, dir, func(row []string) error {
		mapData := &models.Map{
			MapID:         StrToInt32(row[0]),
			Name:          row[1],
			Width:         StrToInt32(row[2]),
			Height:        StrToInt32(row[3]),
			MaxPlayer:     StrToInt32(row[4]),
			MonsterConfig: row[5],
			TerrainData:   row[6],
			RespawnPointX: StrToFloat32(row[7]),
			RespawnPointY: StrToFloat32(row[8]),
		}

		tempMaps[mapData.MapID] = mapData
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		mtl.mu.Lock()
		mtl.maps = tempMaps
		mtl.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (mtl *MapTableLoader) GetTableName() string {
	return "maps"
}

// GetMap 根据ID获取地图
func (mtl *MapTableLoader) GetMap(mapID int32) (*models.Map, bool) {
	mtl.mu.RLock()
	mapData, ok := mtl.maps[mapID]
	mtl.mu.RUnlock()
	return mapData, ok
}

// GetAllMaps 获取所有地图
func (mtl *MapTableLoader) GetAllMaps() map[int32]*models.Map {
	mtl.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	mapsCopy := make(map[int32]*models.Map, len(mtl.maps))
	for id, mapData := range mtl.maps {
		mapsCopy[id] = mapData
	}
	mtl.mu.RUnlock()
	return mapsCopy
}
