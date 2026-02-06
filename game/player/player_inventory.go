package player

import (
	"sync/atomic"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zGameServer/event"
	"github.com/pzqf/zGameServer/game/object/component"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// Item 物品结构
// 表示背包中的一个物品实例
type Item struct {
	itemId     int64       // 物品ID（对应配置表）
	itemType   int         // 物品类型（武器/防具/消耗品等）
	itemName   string      // 物品名称
	count      atomic.Int32 // 物品数量（原子操作）
	maxStack   int         // 最大堆叠数量
	bind       bool        // 是否绑定
	quality    int         // 品质等级
	levelReq   int         // 使用等级要求
	properties *zMap.Map   // 物品属性（攻击力、防御力等）
}

// Inventory 背包系统
// 管理玩家的物品存储
type Inventory struct {
	*component.BaseComponent
	playerId int64       // 所属玩家ID
	items    *zMap.Map   // 物品映射表（槽位 -> 物品）
	size     int         // 背包容量（槽位数量）
}

// NewItem 创建新物品
// 参数:
//   - itemId: 物品ID
//   - itemType: 物品类型
//   - itemName: 物品名称
//   - count: 数量
//   - maxStack: 最大堆叠
//   - bind: 是否绑定
//   - quality: 品质
//   - levelReq: 等级要求
//
// 返回:
//   - *Item: 新创建的物品
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
	item.count.Store(int32(count))
	return item
}

// NewInventory 创建背包组件
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - *Inventory: 新创建的背包
func NewInventory(playerId common.PlayerIdType) *Inventory {
	return &Inventory{
		BaseComponent: component.NewBaseComponent("inventory"),
		playerId:      int64(playerId),
		items:         zMap.NewMap(),
		size:          60, // 默认背包大小60格
	}
}

// Init 初始化背包组件
// 返回: 初始化错误
func (inv *Inventory) Init() error {
	zLog.Debug("Initializing inventory", zap.Int64("playerId", inv.playerId))
	return nil
}

// Destroy 销毁背包组件
// 清理所有物品数据
func (inv *Inventory) Destroy() {
	zLog.Debug("Destroying inventory", zap.Int64("playerId", inv.playerId))
	inv.items.Clear()
}

// AddItem 添加物品到背包
// 优先尝试堆叠到已有物品，否则放入空槽位
// 参数:
//   - item: 要添加的物品
//
// 返回:
//   - int: 放置的槽位（0表示失败）
//   - error: 添加错误
func (inv *Inventory) AddItem(item *Item) (int, error) {
	// 检查物品是否可以堆叠
	if item.count.Load() > 0 && item.maxStack > 1 {
		var stackableSlot int
		var availableSpace int

		// 查找可堆叠的物品槽位
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
			existingItemInterface, _ := inv.items.Load(stackableSlot)
			existingItem := existingItemInterface.(*Item)

			if int(item.count.Load()) <= availableSpace {
				// 完全堆叠
				existingItem.count.Add(item.count.Load())
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
		if _, exists := inv.items.Load(slot); !exists {
			inv.items.Store(slot, item)
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
// 参数:
//   - slot: 槽位
//   - count: 移除数量
//
// 返回:
//   - error: 移除错误
func (inv *Inventory) RemoveItem(slot int, count int) error {
	item, exists := inv.items.Load(slot)
	if !exists {
		return nil
	}

	existingItem := item.(*Item)
	currentCount := int(existingItem.count.Load())
	removeCount := count
	if currentCount <= count {
		removeCount = currentCount
		inv.items.Delete(slot)
	} else {
		existingItem.count.Add(-int32(count))
	}

	eventData := &event.PlayerItemEventData{
		PlayerID: inv.playerId,
		ItemID:   existingItem.itemId,
		Count:    -removeCount,
		Slot:     slot,
	}
	event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerItemRemove, inv, eventData))

	return nil
}

// GetItem 获取背包中的物品
// 参数:
//   - slot: 槽位
//
// 返回:
//   - *Item: 物品
//   - bool: 是否存在
func (inv *Inventory) GetItem(slot int) (*Item, bool) {
	item, exists := inv.items.Load(slot)
	if !exists {
		return nil, false
	}
	return item.(*Item), true
}

// GetAllItems 获取背包中所有物品
// 返回: 物品列表
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
// 参数:
//   - size: 新的背包大小
//
// 返回: 是否成功扩展
func (inv *Inventory) Expand(size int) bool {
	if size <= inv.size {
		return false
	}
	inv.size = size
	return true
}

// MoveItem 移动物品
// 参数:
//   - fromSlot: 源槽位
//   - toSlot: 目标槽位
//   - count: 移动数量
//
// 返回: 是否成功移动
func (inv *Inventory) MoveItem(fromSlot int, toSlot int, count int) bool {
	fromItemInterface, exists := inv.items.Load(fromSlot)
	if !exists {
		return false
	}
	fromItem := fromItemInterface.(*Item)

	if count <= 0 || count > int(fromItem.count.Load()) {
		return false
	}

	toItemInterface, exists := inv.items.Load(toSlot)
	if exists {
		toItem := toItemInterface.(*Item)
		if toItem.itemId != fromItem.itemId || toItem.bind != fromItem.bind {
			return false
		}

		availableSpace := toItem.maxStack - int(toItem.count.Load())
		if availableSpace < count {
			return false
		}

		toItem.count.Add(int32(count))

		fromItem.count.Add(-int32(count))
		if fromItem.count.Load() <= 0 {
			inv.items.Delete(fromSlot)
		}
	} else {
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

		inv.items.Store(toSlot, newItem)

		fromItem.count.Add(-int32(count))
		if fromItem.count.Load() <= 0 {
			inv.items.Delete(fromSlot)
		}
	}

	return true
}

// UseItem 使用物品
// 参数:
//   - slot: 槽位
//   - playerLevel: 玩家等级（用于检查等级要求）
//
// 返回: 是否成功使用
func (inv *Inventory) UseItem(slot int, playerLevel int) bool {
	itemInterface, exists := inv.items.Load(slot)
	if !exists {
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

	if playerLevel < item.levelReq {
		eventData := &event.PlayerUseItemEventData{
			PlayerID: inv.playerId,
			ItemID:   item.itemId,
			Slot:     slot,
			Result:   false,
		}
		event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerUseItem, inv, eventData))
		return false
	}

	zLog.Debug("Using item", zap.Int64("itemId", item.itemId), zap.String("itemName", item.itemName), zap.Int64("playerId", inv.playerId))

	if item.count.Add(-1) <= 0 {
		inv.items.Delete(slot)
	}

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
// 重新排列物品顺序
func (inv *Inventory) Sort() {
	var items []*Item
	inv.items.Range(func(key, value interface{}) bool {
		items = append(items, value.(*Item))
		return true
	})

	inv.items.Clear()

	for i, item := range items {
		inv.items.Store(i+1, item)
	}
}

// GetItemCount 获取物品数量
// 参数:
//   - itemId: 物品ID
//
// 返回: 该物品的总数量
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
// 参数:
//   - count: 需要的数量
//   - maxStack: 物品最大堆叠数
//
// 返回: 是否有足够空间
func (inv *Inventory) HasSpace(count int, maxStack int) bool {
	emptySlots := 0
	availableStackSpace := 0

	inv.items.Range(func(key, value interface{}) bool {
		item := value.(*Item)
		if item.maxStack > 1 {
			availableStackSpace += item.maxStack - int(item.count.Load())
		}
		return true
	})

	for slot := 1; slot <= inv.size; slot++ {
		if _, exists := inv.items.Load(slot); !exists {
			emptySlots++
		}
	}

	if maxStack == 1 {
		return emptySlots >= count
	} else {
		if availableStackSpace >= count {
			return true
		}

		count -= availableStackSpace
		return emptySlots >= (count+maxStack-1)/maxStack
	}
}
