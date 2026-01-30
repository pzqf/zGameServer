package tables

import (
	"strconv"
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// StrToBool 字符串转布尔值
func StrToBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}

// BuffTableLoader buff表加载器
type BuffTableLoader struct {
	mu    sync.RWMutex
	buffs map[int32]*models.Buff
}

// NewBuffTableLoader 创建buff表加载器
func NewBuffTableLoader() *BuffTableLoader {
	return &BuffTableLoader{
		buffs: make(map[int32]*models.Buff),
	}
}

// Load 加载buff表数据
func (btl *BuffTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "buff.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 9,
		TableName:  "buffs",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempBuffs := make(map[int32]*models.Buff)

	err := ReadExcelFile(config, dir, func(row []string) error {
		buff := &models.Buff{
			BuffID:      StrToInt32(row[0]),
			Name:        row[1],
			Description: row[2],
			Type:        row[3],
			Duration:    StrToInt32(row[4]),
			Value:       StrToInt32(row[5]),
			Property:    row[6],
			IsPermanent: StrToBool(row[7]),
		}

		tempBuffs[buff.BuffID] = buff
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		btl.mu.Lock()
		btl.buffs = tempBuffs
		btl.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (btl *BuffTableLoader) GetTableName() string {
	return "buffs"
}

// GetBuff 根据ID获取buff
func (btl *BuffTableLoader) GetBuff(buffID int32) (*models.Buff, bool) {
	btl.mu.RLock()
	buff, ok := btl.buffs[buffID]
	btl.mu.RUnlock()
	return buff, ok
}

// GetAllBuffs 获取所有buff
func (btl *BuffTableLoader) GetAllBuffs() map[int32]*models.Buff {
	btl.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	buffsCopy := make(map[int32]*models.Buff, len(btl.buffs))
	for id, buff := range btl.buffs {
		buffsCopy[id] = buff
	}
	btl.mu.RUnlock()
	return buffsCopy
}
