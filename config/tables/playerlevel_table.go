package tables

import (
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// PlayerLevelTableLoader 人物等级表加载器
type PlayerLevelTableLoader struct {
	mu           sync.RWMutex
	playerLevels map[int32]*models.PlayerLevel
}

// NewPlayerLevelTableLoader 创建人物等级表加载器
func NewPlayerLevelTableLoader() *PlayerLevelTableLoader {
	return &PlayerLevelTableLoader{
		playerLevels: make(map[int32]*models.PlayerLevel),
	}
}

// Load 加载人物等级表数据
func (plt *PlayerLevelTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "playerlevel.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 8,
		TableName:  "playerLevels",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempPlayerLevels := make(map[int32]*models.PlayerLevel)

	err := ReadExcelFile(config, dir, func(row []string) error {
		level := &models.PlayerLevel{
			LevelID:      StrToInt32(row[0]),
			RequiredExp:  StrToInt64(row[1]),
			HP:           StrToInt32(row[2]),
			MP:           StrToInt32(row[3]),
			Attack:       StrToInt32(row[4]),
			Defense:      StrToInt32(row[5]),
			CriticalRate: StrToFloat32(row[6]),
			SkillPoints:  StrToInt32(row[7]),
		}

		tempPlayerLevels[level.LevelID] = level
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		plt.mu.Lock()
		plt.playerLevels = tempPlayerLevels
		plt.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (plt *PlayerLevelTableLoader) GetTableName() string {
	return "playerLevels"
}

// GetPlayerLevel 根据ID获取等级配置
func (plt *PlayerLevelTableLoader) GetPlayerLevel(levelID int32) (*models.PlayerLevel, bool) {
	plt.mu.RLock()
	level, ok := plt.playerLevels[levelID]
	plt.mu.RUnlock()
	return level, ok
}

// GetAllPlayerLevels 获取所有等级配置
func (plt *PlayerLevelTableLoader) GetAllPlayerLevels() map[int32]*models.PlayerLevel {
	plt.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	playerLevelsCopy := make(map[int32]*models.PlayerLevel, len(plt.playerLevels))
	for id, level := range plt.playerLevels {
		playerLevelsCopy[id] = level
	}
	plt.mu.RUnlock()
	return playerLevelsCopy
}
