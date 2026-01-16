package tables

import (
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// TableLoaderInterface 表格加载器接口
type TableLoaderInterface interface {
	// Load 加载表格数据
	Load(dir string) error
	// GetTableName 获取表格名称
	GetTableName() string
}

// ItemTableLoader 物品表格加载器
type ItemTableLoader struct {
	mu    sync.RWMutex
	items map[int32]*models.ItemBase
}

// MapTableLoader 地图表格加载器
type MapTableLoader struct {
	mu   sync.RWMutex
	maps map[int32]*models.Map
}

// SkillTableLoader 技能表格加载器
type SkillTableLoader struct {
	mu     sync.RWMutex
	skills map[int32]*models.Skill
}

// QuestTableLoader 任务表格加载器
type QuestTableLoader struct {
	mu     sync.RWMutex
	quests map[int32]*models.Quest
}

// NewItemTableLoader 创建物品表格加载器
func NewItemTableLoader() *ItemTableLoader {
	return &ItemTableLoader{
		items: make(map[int32]*models.ItemBase),
	}
}

// NewMapTableLoader 创建地图表格加载器
func NewMapTableLoader() *MapTableLoader {
	return &MapTableLoader{
		maps: make(map[int32]*models.Map),
	}
}

// NewSkillTableLoader 创建技能表格加载器
func NewSkillTableLoader() *SkillTableLoader {
	return &SkillTableLoader{
		skills: make(map[int32]*models.Skill),
	}
}

// NewQuestTableLoader 创建任务表格加载器
func NewQuestTableLoader() *QuestTableLoader {
	return &QuestTableLoader{
		quests: make(map[int32]*models.Quest),
	}
}
