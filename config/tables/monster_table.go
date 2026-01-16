package tables

import (
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// MonsterTableLoader 怪物表加载器
type MonsterTableLoader struct {
	mu       sync.RWMutex
	monsters map[int32]*models.Monster
}

// NewMonsterTableLoader 创建怪物表加载器
func NewMonsterTableLoader() *MonsterTableLoader {
	return &MonsterTableLoader{
		monsters: make(map[int32]*models.Monster),
	}
}

// Load 加载怪物表数据
func (mtl *MonsterTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "monster.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 11,
		TableName:  "monsters",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempMonsters := make(map[int32]*models.Monster)

	err := ReadExcelFile(config, dir, func(row []string) error {
		monster := &models.Monster{
			MonsterID:    StrToInt32(row[0]),
			Name:         row[1],
			Level:        StrToInt32(row[2]),
			HP:           StrToInt32(row[3]),
			MP:           StrToInt32(row[4]),
			Attack:       StrToInt32(row[5]),
			Defense:      StrToInt32(row[6]),
			Speed:        StrToInt32(row[7]),
			Exp:          StrToInt32(row[8]),
			DropItemRate: StrToFloat32(row[9]),
			DropItems:    row[10],
		}

		tempMonsters[monster.MonsterID] = monster
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		mtl.mu.Lock()
		mtl.monsters = tempMonsters
		mtl.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (mtl *MonsterTableLoader) GetTableName() string {
	return "monsters"
}

// GetMonster 根据ID获取怪物配置
func (mtl *MonsterTableLoader) GetMonster(monsterID int32) (*models.Monster, bool) {
	mtl.mu.RLock()
	monster, ok := mtl.monsters[monsterID]
	mtl.mu.RUnlock()
	return monster, ok
}

// GetAllMonsters 获取所有怪物配置
func (mtl *MonsterTableLoader) GetAllMonsters() map[int32]*models.Monster {
	mtl.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	monstersCopy := make(map[int32]*models.Monster, len(mtl.monsters))
	for id, monster := range mtl.monsters {
		monstersCopy[id] = monster
	}
	mtl.mu.RUnlock()
	return monstersCopy
}
