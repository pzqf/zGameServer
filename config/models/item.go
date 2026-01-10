package models

// Item 物品配置结构
type Item struct {
	ItemID      int32  `json:"item_id"`
	Name        string `json:"name"`
	Type        int32  `json:"type"`
	SubType     int32  `json:"sub_type"`
	Level       int32  `json:"level"`
	Quality     int32  `json:"quality"`
	Price       int32  `json:"price"`
	SellPrice   int32  `json:"sell_price"`
	StackLimit  int32  `json:"stack_limit"`
	Description string `json:"description"`
	Effects     string `json:"effects"` // JSON格式的效果描述
}
