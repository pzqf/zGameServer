package tables

import (
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// AITableLoader AI表加载器
type AITableLoader struct {
	mu  sync.RWMutex
	ais map[int32]*models.AI
}

// NewAITableLoader 创建AI表加载器
func NewAITableLoader() *AITableLoader {
	return &AITableLoader{
		ais: make(map[int32]*models.AI),
	}
}

// Load 加载AI表数据
func (atl *AITableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "ai.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 10,
		TableName:  "ais",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempAIs := make(map[int32]*models.AI)

	err := ReadExcelFile(config, dir, func(row []string) error {
		ai := &models.AI{
			AIID:           StrToInt32(row[0]),
			Type:           row[1],
			DetectionRange: StrToFloat32(row[2]),
			AttackRange:    StrToFloat32(row[3]),
			ChaseRange:     StrToFloat32(row[4]),
			FleeHealth:     StrToFloat32(row[5]),
			PatrolPoints:   row[6],
			Behavior:       row[7],
			SkillIDs:       row[8],
		}

		tempAIs[ai.AIID] = ai
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		atl.mu.Lock()
		atl.ais = tempAIs
		atl.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (atl *AITableLoader) GetTableName() string {
	return "ais"
}

// GetAI 根据ID获取AI
func (atl *AITableLoader) GetAI(aiID int32) (*models.AI, bool) {
	atl.mu.RLock()
	ai, ok := atl.ais[aiID]
	atl.mu.RUnlock()
	return ai, ok
}

// GetAllAIs 获取所有AI
func (atl *AITableLoader) GetAllAIs() map[int32]*models.AI {
	atl.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	aisCopy := make(map[int32]*models.AI, len(atl.ais))
	for id, ai := range atl.ais {
		aisCopy[id] = ai
	}
	atl.mu.RUnlock()
	return aisCopy
}
