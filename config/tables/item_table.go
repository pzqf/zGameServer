package tables

import (
	"github.com/pzqf/zGameServer/config/models"
)

// Load 加载物品表数据
func (itl *ItemTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "item.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 11,
		TableName:  "items",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempItems := make(map[int32]*models.ItemBase)

	err := ReadExcelFile(config, dir, func(row []string) error {
		item := &models.ItemBase{
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

		tempItems[item.ItemID] = item
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		itl.mu.Lock()
		itl.items = tempItems
		itl.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (itl *ItemTableLoader) GetTableName() string {
	return "items"
}

// GetItem 根据ID获取物品
func (itl *ItemTableLoader) GetItem(itemID int32) (*models.ItemBase, bool) {
	itl.mu.RLock()
	item, ok := itl.items[itemID]
	itl.mu.RUnlock()
	return item, ok
}

// GetAllItems 获取所有物品
func (itl *ItemTableLoader) GetAllItems() map[int32]*models.ItemBase {
	itl.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	itemsCopy := make(map[int32]*models.ItemBase, len(itl.items))
	for id, item := range itl.items {
		itemsCopy[id] = item
	}
	itl.mu.RUnlock()
	return itemsCopy
}
