package player

import (
	"go.uber.org/zap"

	"github.com/pzqf/zGameServer/game/object/component"
	"github.com/pzqf/zUtil/zMap"
)

// 装备位置定义
const (
	EquipPosWeapon   = 1  // 武器
	EquipPosArmor    = 2  // 盔甲
	EquipPosHelmet   = 3  // 头盔
	EquipPosBoots    = 4  // 靴子
	EquipPosGloves   = 5  // 手套
	EquipPosNecklace = 6  // 项链
	EquipPosRing1    = 7  // 戒指1
	EquipPosRing2    = 8  // 戒指2
	EquipPosBelt     = 9  // 腰带
	EquipPosShoulder = 10 // 肩甲
)

// Equipment 装备系统
type Equipment struct {
	*component.BaseComponent
	playerId   int64
	logger     *zap.Logger
	equipments *zMap.Map // key: int(equipPos), value: *Item
}

func NewEquipment(playerId int64, logger *zap.Logger) *Equipment {
	return &Equipment{
		BaseComponent: component.NewBaseComponent("equipment"),
		playerId:      playerId,
		logger:        logger,
		equipments:    zMap.NewMap(),
	}
}

func (eq *Equipment) Init() error {
	// 初始化装备系统
	eq.logger.Debug("Initializing equipment", zap.Int64("playerId", eq.playerId))
	return nil
}

// Destroy 销毁装备组件
func (eq *Equipment) Destroy() {
	// 清理装备资源
	eq.logger.Debug("Destroying equipment", zap.Int64("playerId", eq.playerId))
	eq.equipments.Clear()
}

// Equip 装备物品
func (eq *Equipment) Equip(equipPos int, item *Item) (*Item, error) {
	// 检查装备位置是否合法
	if !eq.IsValidEquipPos(equipPos) {
		return nil, nil // 无效的装备位置
	}

	// 检查物品是否可以装备
	if !eq.CanEquip(item, equipPos) {
		return nil, nil // 物品不能装备到该位置
	}

	// 获取当前装备的物品（如果有）
	oldItem, exists := eq.equipments.Get(equipPos)
	var oldItemPtr *Item
	if exists {
		oldItemPtr = oldItem.(*Item)
	}

	// 装备新物品
	eq.equipments.Store(equipPos, item)
	eq.logger.Info("Item equipped", zap.Int64("playerId", eq.playerId), zap.Int("equipPos", equipPos), zap.Int64("itemId", item.itemId))

	return oldItemPtr, nil
}

// Unequip 卸下装备
func (eq *Equipment) Unequip(equipPos int) (*Item, error) {
	// 检查装备位置是否合法
	if !eq.IsValidEquipPos(equipPos) {
		return nil, nil // 无效的装备位置
	}

	// 获取装备的物品
	item, exists := eq.equipments.Get(equipPos)
	if !exists {
		return nil, nil // 该位置没有装备物品
	}

	// 卸下装备
	eq.equipments.Delete(equipPos)
	eq.logger.Info("Item unequipped", zap.Int64("playerId", eq.playerId), zap.Int("equipPos", equipPos))

	return item.(*Item), nil
}

// GetEquipment 获取指定位置的装备
func (eq *Equipment) GetEquipment(equipPos int) (*Item, bool) {
	item, exists := eq.equipments.Get(equipPos)
	if !exists {
		return nil, false
	}
	return item.(*Item), true
}

// GetAllEquipments 获取所有装备
func (eq *Equipment) GetAllEquipments() *zMap.Map {
	return eq.equipments
}

// IsValidEquipPos 检查装备位置是否合法
func (eq *Equipment) IsValidEquipPos(equipPos int) bool {
	return equipPos >= EquipPosWeapon && equipPos <= EquipPosShoulder
}

// CanEquip 检查物品是否可以装备到指定位置
func (eq *Equipment) CanEquip(item *Item, equipPos int) bool {
	// 这里应该有更复杂的检查逻辑，如物品类型、玩家等级等
	// 简化实现，假设所有物品都可以装备到任何位置
	return true
}
