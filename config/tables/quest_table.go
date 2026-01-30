package tables

import (
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// QuestTableLoader 任务表格加载器
type QuestTableLoader struct {
	mu     sync.RWMutex
	quests map[int32]*models.Quest
}

// NewQuestTableLoader 创建任务表格加载器
func NewQuestTableLoader() *QuestTableLoader {
	return &QuestTableLoader{
		quests: make(map[int32]*models.Quest),
	}
}

// Load 加载任务表数据
func (qtl *QuestTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "quest.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 9,
		TableName:  "quests",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempQuests := make(map[int32]*models.Quest)

	err := ReadExcelFile(config, dir, func(row []string) error {
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

		tempQuests[quest.QuestID] = quest
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		qtl.mu.Lock()
		qtl.quests = tempQuests
		qtl.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (qtl *QuestTableLoader) GetTableName() string {
	return "quests"
}

// GetQuest 根据ID获取任务
func (qtl *QuestTableLoader) GetQuest(questID int32) (*models.Quest, bool) {
	qtl.mu.RLock()
	quest, ok := qtl.quests[questID]
	qtl.mu.RUnlock()
	return quest, ok
}

// GetAllQuests 获取所有任务
func (qtl *QuestTableLoader) GetAllQuests() map[int32]*models.Quest {
	qtl.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	questsCopy := make(map[int32]*models.Quest, len(qtl.quests))
	for id, quest := range qtl.quests {
		questsCopy[id] = quest
	}
	qtl.mu.RUnlock()
	return questsCopy
}
