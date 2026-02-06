package player

import (
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zGameServer/game/object/component"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// EquipPosType 装备位置类型
type EquipPosType int

// 装备位置定义
// 标识角色身上不同的装备槽位
const (
	EquipPosWeapon   = 1  // 武器（主手武器）
	EquipPosArmor    = 2  // 盔甲（身体护甲）
	EquipPosHelmet   = 3  // 头盔（头部防护）
	EquipPosBoots    = 4  // 靴子（脚部装备）
	EquipPosGloves   = 5  // 手套（手部装备）
	EquipPosNecklace = 6  // 项链（颈部饰品）
	EquipPosRing1    = 7  // 戒指1（手指饰品）
	EquipPosRing2    = 8  // 戒指2（手指饰品）
	EquipPosBelt     = 9  // 腰带（腰部装备）
	EquipPosShoulder = 10 // 肩甲（肩部防护）
)

// Equipment 装备系统
// 管理玩家的装备穿戴和卸下
type Equipment struct {
	*component.BaseComponent
	playerId   common.PlayerIdType                     // 所属玩家ID
	equipments *zMap.TypedMap[EquipPosType, *Item]     // 装备映射表（位置 -> 物品）
}

// NewEquipment 创建装备组件
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - *Equipment: 新创建的装备组件
func NewEquipment(playerId common.PlayerIdType) *Equipment {
	return &Equipment{
		BaseComponent: component.NewBaseComponent("equipment"),
		playerId:      playerId,
		equipments:    zMap.NewTypedMap[EquipPosType, *Item](),
	}
}

// Init 初始化装备组件
// 返回: 初始化错误
func (eq *Equipment) Init() error {
	zLog.Debug("Initializing equipment", zap.Int64("playerId", int64(eq.playerId)))
	return nil
}

// Destroy 销毁装备组件
// 清理所有装备数据
func (eq *Equipment) Destroy() {
	zLog.Debug("Destroying equipment", zap.Int64("playerId", int64(eq.playerId)))
	eq.equipments.Clear()
}

// Equip 装备物品
// 将物品穿戴到指定位置，如果该位置已有装备则返回旧装备
// 参数:
//   - equipPos: 装备位置
//   - item: 要装备的物品
//
// 返回:
//   - *Item: 被替换的旧装备（如果有）
//   - error: 装备错误
func (eq *Equipment) Equip(equipPos EquipPosType, item *Item) (*Item, error) {
	// 检查装备位置是否合法
	if !eq.IsValidEquipPos(equipPos) {
		return nil, nil
	}

	// 检查物品是否可以装备
	if !eq.CanEquip(item, equipPos) {
		return nil, nil
	}

	// 获取当前装备的物品（如果有）
	oldItem, exists := eq.equipments.Load(equipPos)
	var oldItemPtr *Item
	if exists {
		oldItemPtr = oldItem
	}

	// 装备新物品
	eq.equipments.Store(equipPos, item)
	zLog.Info("Item equipped", zap.Int64("playerId", int64(eq.playerId)),
		zap.Int("equipPos", int(equipPos)), zap.Int64("itemId", item.itemId))

	return oldItemPtr, nil
}

// Unequip 卸下装备
// 从指定位置取下装备并返回
// 参数:
//   - equipPos: 装备位置
//
// 返回:
//   - *Item: 卸下的物品
//   - error: 卸下错误
func (eq *Equipment) Unequip(equipPos EquipPosType) (*Item, error) {
	// 检查装备位置是否合法
	if !eq.IsValidEquipPos(equipPos) {
		return nil, nil
	}

	// 获取装备的物品
	item, exists := eq.equipments.Load(equipPos)
	if !exists {
		return nil, nil
	}

	// 卸下装备
	eq.equipments.Delete(equipPos)
	zLog.Info("Item unequipped", zap.Int64("playerId", int64(eq.playerId)),
		zap.Int("equipPos", int(equipPos)))

	return item, nil
}

// GetEquipment 获取指定位置的装备
// 参数:
//   - equipPos: 装备位置
//
// 返回:
//   - *Item: 装备物品
//   - bool: 是否存在
func (eq *Equipment) GetEquipment(equipPos EquipPosType) (*Item, bool) {
	item, exists := eq.equipments.Load(equipPos)
	if !exists {
		return nil, false
	}
	return item, true
}

// GetAllEquipments 获取所有装备
// 返回: 装备映射表
func (eq *Equipment) GetAllEquipments() *zMap.TypedMap[EquipPosType, *Item] {
	return eq.equipments
}

// IsValidEquipPos 检查装备位置是否合法
// 参数:
//   - equipPos: 装备位置
//
// 返回: 是否合法
func (eq *Equipment) IsValidEquipPos(equipPos EquipPosType) bool {
	return equipPos >= EquipPosWeapon && equipPos <= EquipPosShoulder
}

// CanEquip 检查物品是否可以装备到指定位置
// 参数:
//   - item: 要装备的物品
//   - equipPos: 目标位置
//
// 返回: 是否可以装备
func (eq *Equipment) CanEquip(item *Item, equipPos EquipPosType) bool {
	return true
}
