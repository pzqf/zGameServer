package player

import (
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/pzqf/zGameServer/event"
	"github.com/pzqf/zUtil/zMap"
)

// Item 物品结构
type Item struct {
	itemId     int64
	itemType   int
	itemName   string
	count      atomic.Int32
	maxStack   int
	bind       bool
	quality    int
	levelReq   int
	properties *zMap.Map // 物品属性
}

// Inventory 背包系统
type Inventory struct {
	playerId int64
	logger   *zap.Logger
	items    *zMap.Map // key: int(slot), value: *Item
	size     int
}

// NewItem 创建新物品
func NewItem(itemId int64, itemType int, itemName string, count int, maxStack int, bind bool, quality int, levelReq int) *Item {
	item := &Item{
		itemId:     itemId,
		itemType:   itemType,
		itemName:   itemName,
		maxStack:   maxStack,
		bind:       bind,
		quality:    quality,
		levelReq:   levelReq,
		properties: zMap.NewMap(),
	}
	// 初始化原子字段
	item.count.Store(int32(count))
	return item
}

func NewInventory(playerId int64, logger *zap.Logger) *Inventory {
	return &Inventory{
		playerId: playerId,
		logger:   logger,
		items:    zMap.NewMap(),
		size:     60, // 默认背包大小
	}
}

func (inv *Inventory) Init() {
	// 初始化背包
	inv.logger.Debug("Initializing inventory", zap.Int64("playerId", inv.playerId))
}

// AddItem 添加物品到背包
func (inv *Inventory) AddItem(item *Item) (int, error) {
	// 检查物品是否可以堆叠
	if item.count.Load() > 0 && item.maxStack > 1 {
		// 查找可堆叠的物品槽位
		var stackableSlot int
		var availableSpace int

		inv.items.Range(func(key, value interface{}) bool {
			existingItem := value.(*Item)
			if existingItem.itemId == item.itemId && existingItem.bind == item.bind {
				stackableSlot = key.(int)
				availableSpace = existingItem.maxStack - int(existingItem.count.Load())
				return false
			}
			return true
		})

		if stackableSlot != 0 && availableSpace > 0 {
			// 堆叠物品
			existingItemInterface, _ := inv.items.Get(stackableSlot)
			existingItem := existingItemInterface.(*Item)

			if int(item.count.Load()) <= availableSpace {
				// 完全堆叠
				existingItem.count.Add(item.count.Load())
				// 发布物品增加事件
				eventData := &event.PlayerItemEventData{
					PlayerID: inv.playerId,
					ItemID:   item.itemId,
					Count:    int(item.count.Load()),
					Slot:     stackableSlot,
				}
				event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerItemAdd, inv, eventData))
				return stackableSlot, nil
			} else {
				// 部分堆叠，剩余部分寻找新槽位
				existingItem.count.Store(int32(existingItem.maxStack))
				// 发布堆叠物品增加事件
				eventData := &event.PlayerItemEventData{
					PlayerID: inv.playerId,
					ItemID:   item.itemId,
					Count:    availableSpace,
					Slot:     stackableSlot,
				}
				event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerItemAdd, inv, eventData))
				item.count.Add(-int32(availableSpace))
			}
		}
	}

	// 查找空槽位放置剩余物品
	for slot := 1; slot <= inv.size; slot++ {
		if _, exists := inv.items.Get(slot); !exists {
			// 添加到空槽位
			inv.items.Store(slot, item)
			// 发布物品增加事件
			eventData := &event.PlayerItemEventData{
				PlayerID: inv.playerId,
				ItemID:   item.itemId,
				Count:    int(item.count.Load()),
				Slot:     slot,
			}
			event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerItemAdd, inv, eventData))
			return slot, nil
		}
	}

	return 0, nil // 背包已满
}

// RemoveItem 从背包移除物品
func (inv *Inventory) RemoveItem(slot int, count int) error {
	item, exists := inv.items.Get(slot)
	if !exists {
		return nil // 槽位为空
	}

	existingItem := item.(*Item)
	currentCount := int(existingItem.count.Load())
	removeCount := count
	if currentCount <= count {
		removeCount = currentCount
		// 移除整个物品
		inv.items.Delete(slot)
	} else {
		// 减少物品数量
		existingItem.count.Add(-int32(count))
	}

	// 发布物品移除事件
	eventData := &event.PlayerItemEventData{
		PlayerID: inv.playerId,
		ItemID:   existingItem.itemId,
		Count:    -removeCount, // 使用负数表示减少
		Slot:     slot,
	}
	event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerItemRemove, inv, eventData))

	return nil
}

// GetItem 获取背包中的物品
func (inv *Inventory) GetItem(slot int) (*Item, bool) {
	item, exists := inv.items.Get(slot)
	if !exists {
		return nil, false
	}
	return item.(*Item), true
}

// GetAllItems 获取背包中所有物品
func (inv *Inventory) GetAllItems() []*Item {
	var items []*Item
	inv.items.Range(func(key, value interface{}) bool {
		if value != nil {
			items = append(items, value.(*Item))
		}
		return true
	})
	return items
}

// Expand 扩展背包
func (inv *Inventory) Expand(size int) bool {
	if size <= inv.size {
		return false
	}
	inv.size = size
	return true
}

// MoveItem 移动物品
func (inv *Inventory) MoveItem(fromSlot int, toSlot int, count int) bool {
	// 检查源槽位是否有物品
	fromItemInterface, exists := inv.items.Get(fromSlot)
	if !exists {
		return false
	}
	fromItem := fromItemInterface.(*Item)

	// 检查数量是否合法
	if count <= 0 || count > int(fromItem.count.Load()) {
		return false
	}

	// 检查目标槽位
	toItemInterface, exists := inv.items.Get(toSlot)
	if exists {
		toItem := toItemInterface.(*Item)
		// 检查是否可以堆叠
		if toItem.itemId != fromItem.itemId || toItem.bind != fromItem.bind {
			return false
		}

		// 检查堆叠空间
		availableSpace := toItem.maxStack - int(toItem.count.Load())
		if availableSpace < count {
			return false
		}

		// 堆叠物品
		toItem.count.Add(int32(count))

		// 更新源槽位
		fromItem.count.Add(-int32(count))
		if fromItem.count.Load() <= 0 {
			inv.items.Delete(fromSlot)
		}
	} else {
		// 创建新物品
		newProperties := zMap.NewMap()
		fromItem.properties.Range(func(key, value interface{}) bool {
			newProperties.Store(key, value)
			return true
		})
		newItem := &Item{
			itemId:     fromItem.itemId,
			itemType:   fromItem.itemType,
			itemName:   fromItem.itemName,
			maxStack:   fromItem.maxStack,
			bind:       fromItem.bind,
			quality:    fromItem.quality,
			levelReq:   fromItem.levelReq,
			properties: newProperties,
		}
		newItem.count.Store(int32(count))

		// 放置到目标槽位
		inv.items.Store(toSlot, newItem)

		// 更新源槽位
		fromItem.count.Add(-int32(count))
		if fromItem.count.Load() <= 0 {
			inv.items.Delete(fromSlot)
		}
	}

	return true
}

// UseItem 使用物品
func (inv *Inventory) UseItem(slot int, playerLevel int) bool {
	// 检查槽位是否有物品
	itemInterface, exists := inv.items.Get(slot)
	if !exists {
		// 发布物品使用失败事件
		eventData := &event.PlayerUseItemEventData{
			PlayerID: inv.playerId,
			ItemID:   0,
			Slot:     slot,
			Result:   false,
		}
		event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerUseItem, inv, eventData))
		return false
	}
	item := itemInterface.(*Item)

	// 检查等级要求
	if playerLevel < item.levelReq {
		// 发布物品使用失败事件
		eventData := &event.PlayerUseItemEventData{
			PlayerID: inv.playerId,
			ItemID:   item.itemId,
			Slot:     slot,
			Result:   false,
		}
		event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerUseItem, inv, eventData))
		return false
	}

	// TODO: 实现物品使用逻辑
	inv.logger.Debug("Using item", zap.Int64("itemId", item.itemId), zap.String("itemName", item.itemName), zap.Int64("playerId", inv.playerId))

	// 减少物品数量
	if item.count.Add(-1) <= 0 {
		inv.items.Delete(slot)
	}

	// 发布物品使用成功事件
	eventData := &event.PlayerUseItemEventData{
		PlayerID: inv.playerId,
		ItemID:   item.itemId,
		Slot:     slot,
		Result:   true,
	}
	event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerUseItem, inv, eventData))

	return true
}

// Sort 整理背包
func (inv *Inventory) Sort() {
	// 收集所有物品
	var items []*Item
	inv.items.Range(func(key, value interface{}) bool {
		items = append(items, value.(*Item))
		return true
	})

	// 清空背包
	inv.items.Clear()

	// 重新排列物品（按物品ID和绑定状态）
	// TODO: 实现更复杂的排序逻辑
	for i, item := range items {
		inv.items.Store(i+1, item)
	}
}

// GetItemCount 获取物品数量
func (inv *Inventory) GetItemCount(itemId int64) int {
	count := 0
	inv.items.Range(func(key, value interface{}) bool {
		item := value.(*Item)
		if item.itemId == itemId {
			count += int(item.count.Load())
		}
		return true
	})
	return count
}

// HasSpace 检查背包是否有空间
func (inv *Inventory) HasSpace(count int, maxStack int) bool {
	// 计算可用空间
	emptySlots := 0
	availableStackSpace := 0

	inv.items.Range(func(key, value interface{}) bool {
		item := value.(*Item)
		if item.maxStack > 1 {
			availableStackSpace += item.maxStack - int(item.count.Load())
		}
		return true
	})

	// 计算空槽位数量
	for slot := 1; slot <= inv.size; slot++ {
		if _, exists := inv.items.Get(slot); !exists {
			emptySlots++
		}
	}

	// 检查是否有足够空间
	if maxStack == 1 {
		// 不可堆叠物品
		return emptySlots >= count
	} else {
		// 可堆叠物品
		if availableStackSpace >= count {
			return true
		}

		count -= availableStackSpace
		return emptySlots >= (count+maxStack-1)/maxStack
	}
}
