package tables

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TableManager 表格配置管理器
type TableManager struct {
	mu                sync.RWMutex
	itemLoader        *ItemTableLoader
	mapLoader         *MapTableLoader
	skillLoader       *SkillTableLoader
	questLoader       *QuestTableLoader
	playerLevelLoader *PlayerLevelTableLoader
	monsterLoader     *MonsterTableLoader
	shopLoader        *ShopTableLoader
	guildLoader       *GuildTableLoader
	petLoader         *PetTableLoader
	buffLoader        *BuffTableLoader
	aiLoader          *AITableLoader
	loaders           []TableLoaderInterface
	initialized       bool
}

// GlobalTableManager 全局表格配置管理器实例
var GlobalTableManager *TableManager

// NewTableManager 创建表格配置管理器
func NewTableManager() *TableManager {
	itemLoader := NewItemTableLoader()
	mapLoader := NewMapTableLoader()
	skillLoader := NewSkillTableLoader()
	questLoader := NewQuestTableLoader()
	playerLevelLoader := NewPlayerLevelTableLoader()
	monsterLoader := NewMonsterTableLoader()
	shopLoader := NewShopTableLoader()
	guildLoader := NewGuildTableLoader()
	petLoader := NewPetTableLoader()
	buffLoader := NewBuffTableLoader()
	aiLoader := NewAITableLoader()

	return &TableManager{
		itemLoader:        itemLoader,
		mapLoader:         mapLoader,
		skillLoader:       skillLoader,
		questLoader:       questLoader,
		playerLevelLoader: playerLevelLoader,
		monsterLoader:     monsterLoader,
		shopLoader:        shopLoader,
		guildLoader:       guildLoader,
		petLoader:         petLoader,
		buffLoader:        buffLoader,
		aiLoader:          aiLoader,
		loaders: []TableLoaderInterface{
			itemLoader,
			mapLoader,
			skillLoader,
			questLoader,
			playerLevelLoader,
			monsterLoader,
			shopLoader,
			guildLoader,
			petLoader,
			buffLoader,
			aiLoader,
		},
		initialized: false,
	}
}

// LoadAllTables 加载所有配置表格
func (tm *TableManager) LoadAllTables() error {
	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	tablesDir := filepath.Join(rootDir, "resources", "excel_tables")

	for _, loader := range tm.loaders {
		if err := loader.Load(tablesDir); err != nil {
			return fmt.Errorf("failed to load table %T: %w", loader, err)
		}
	}

	tm.initialized = true
	return nil
}

// GetItemLoader 获取物品表格加载器
func (tm *TableManager) GetItemLoader() *ItemTableLoader {
	return tm.itemLoader
}

// GetMapLoader 获取地图表格加载器
func (tm *TableManager) GetMapLoader() *MapTableLoader {
	return tm.mapLoader
}

// GetSkillLoader 获取技能表格加载器
func (tm *TableManager) GetSkillLoader() *SkillTableLoader {
	return tm.skillLoader
}

// GetQuestLoader 获取任务表格加载器
func (tm *TableManager) GetQuestLoader() *QuestTableLoader {
	return tm.questLoader
}

// GetPlayerLevelLoader 获取人物等级表格加载器
func (tm *TableManager) GetPlayerLevelLoader() *PlayerLevelTableLoader {
	return tm.playerLevelLoader
}

// GetMonsterLoader 获取怪物表格加载器
func (tm *TableManager) GetMonsterLoader() *MonsterTableLoader {
	return tm.monsterLoader
}

// GetShopLoader 获取商店表格加载器
func (tm *TableManager) GetShopLoader() *ShopTableLoader {
	return tm.shopLoader
}

// GetGuildLoader 获取公会表格加载器
func (tm *TableManager) GetGuildLoader() *GuildTableLoader {
	return tm.guildLoader
}

// GetPetLoader 获取宠物表格加载器
func (tm *TableManager) GetPetLoader() *PetTableLoader {
	return tm.petLoader
}

// GetBuffLoader 获取buff表格加载器
func (tm *TableManager) GetBuffLoader() *BuffTableLoader {
	return tm.buffLoader
}

// GetAILoader 获取AI表格加载器
func (tm *TableManager) GetAILoader() *AITableLoader {
	return tm.aiLoader
}

// IsInitialized 检查表格是否已经初始化
func (tm *TableManager) IsInitialized() bool {
	return tm.initialized
}
