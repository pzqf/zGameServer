package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// TableLoader 表格配置加载器
type TableLoader struct {
	mu sync.RWMutex

	items  map[int32]*models.Item
	maps   map[int32]*models.Map
	skills map[int32]*models.Skill
	quests map[int32]*models.Quest
}

// GlobalTableLoader 全局表格配置加载器实例
var GlobalTableLoader *TableLoader

// NewTableLoader 创建表格配置加载器
func NewTableLoader() *TableLoader {
	return &TableLoader{
		items:  make(map[int32]*models.Item),
		maps:   make(map[int32]*models.Map),
		skills: make(map[int32]*models.Skill),
		quests: make(map[int32]*models.Quest),
	}
}

// LoadAllTables 加载所有配置表格
func (tl *TableLoader) LoadAllTables() error {
	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	tablesDir := filepath.Join(rootDir, "config", "tables")

	// 加载物品表
	if err := tl.loadItemTable(tablesDir); err != nil {
		return fmt.Errorf("failed to load item table: %w", err)
	}

	// 加载地图表
	if err := tl.loadMapTable(tablesDir); err != nil {
		return fmt.Errorf("failed to load map table: %w", err)
	}

	// 加载技能表
	if err := tl.loadSkillTable(tablesDir); err != nil {
		return fmt.Errorf("failed to load skill table: %w", err)
	}

	// 加载任务表
	if err := tl.loadQuestTable(tablesDir); err != nil {
		return fmt.Errorf("failed to load quest table: %w", err)
	}

	return nil
}

// GetItems 获取物品映射表（内部使用）
func (tl *TableLoader) GetItems() map[int32]*models.Item {
	return tl.items
}

// GetMaps 获取地图映射表（内部使用）
func (tl *TableLoader) GetMaps() map[int32]*models.Map {
	return tl.maps
}

// GetSkills 获取技能映射表（内部使用）
func (tl *TableLoader) GetSkills() map[int32]*models.Skill {
	return tl.skills
}

// GetQuests 获取任务映射表（内部使用）
func (tl *TableLoader) GetQuests() map[int32]*models.Quest {
	return tl.quests
}

// GetMutex 获取读写锁（内部使用）
func (tl *TableLoader) GetMutex() *sync.RWMutex {
	return &tl.mu
}

// loadItemTable 加载物品表
func (tl *TableLoader) loadItemTable(dir string) error {
	config := ExcelConfig{
		FileName:   "item.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 11,
		TableName:  "items",
	}

	return ReadExcelFile(config, dir, func(row []string) error {
		item := &models.Item{
			ItemID:      StrToInt32(row[0]),
			Name:        row[1],
			Type:        StrToInt32(row[2]),
			SubType:     StrToInt32(row[3]),
			Level:       StrToInt32(row[4]),
			Quality:     StrToInt32(row[5]),
			Price:       StrToInt32(row[6]),
			SellPrice:   StrToInt32(row[7]),
			StackLimit:  StrToInt32(row[8]),
			Description: row[9],
			Effects:     row[10],
		}

		tl.mu.Lock()
		tl.items[item.ItemID] = item
		tl.mu.Unlock()
		return nil
	})
}

// loadMapTable 加载地图表
func (tl *TableLoader) loadMapTable(dir string) error {
	config := ExcelConfig{
		FileName:   "map.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 9,
		TableName:  "maps",
	}

	return ReadExcelFile(config, dir, func(row []string) error {
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

		tl.mu.Lock()
		tl.maps[mapData.MapID] = mapData
		tl.mu.Unlock()
		return nil
	})
}

// loadSkillTable 加载技能表
func (tl *TableLoader) loadSkillTable(dir string) error {
	config := ExcelConfig{
		FileName:   "skill.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 11,
		TableName:  "skills",
	}

	return ReadExcelFile(config, dir, func(row []string) error {
		skill := &models.Skill{
			SkillID:     StrToInt32(row[0]),
			Name:        row[1],
			Type:        StrToInt32(row[2]),
			Level:       StrToInt32(row[3]),
			ManaCost:    StrToInt32(row[4]),
			Cooldown:    StrToFloat32(row[5]),
			Damage:      StrToInt32(row[6]),
			Range:       StrToFloat32(row[7]),
			AreaRadius:  StrToFloat32(row[8]),
			Description: row[9],
			Effects:     row[10],
		}

		tl.mu.Lock()
		tl.skills[skill.SkillID] = skill
		tl.mu.Unlock()
		return nil
	})
}

// loadQuestTable 加载任务表
func (tl *TableLoader) loadQuestTable(dir string) error {
	config := ExcelConfig{
		FileName:   "quest.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 9,
		TableName:  "quests",
	}

	return ReadExcelFile(config, dir, func(row []string) error {
		quest := &models.Quest{
			QuestID:     StrToInt32(row[0]),
			Name:        row[1],
			Type:        StrToInt32(row[2]),
			Level:       StrToInt32(row[3]),
			Description: row[4],
			Objectives:  row[5],
			Rewards:     row[6],
			NextQuestID: StrToInt32(row[7]),
			PreQuestID:  StrToInt32(row[8]),
		}

		tl.mu.Lock()
		tl.quests[quest.QuestID] = quest
		tl.mu.Unlock()
		return nil
	})
}
