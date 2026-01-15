package event

import (
	"github.com/pzqf/zEngine/zEvent"
)

// 游戏业务相关的事件类型定义
const (
	// 玩家相关事件
	EventPlayerExpAdd       zEvent.EventType = 1001 // 玩家获得经验
	EventPlayerGoldAdd      zEvent.EventType = 1002 // 玩家获得金币
	EventPlayerGoldSub      zEvent.EventType = 1003 // 玩家消耗金币
	EventPlayerItemAdd      zEvent.EventType = 1004 // 玩家获得物品
	EventPlayerItemRemove   zEvent.EventType = 1005 // 玩家失去物品
	EventPlayerLogin        zEvent.EventType = 1007 // 玩家登录
	EventPlayerLogout       zEvent.EventType = 1008 // 玩家登出
	EventPlayerUseItem      zEvent.EventType = 1009 // 玩家使用物品
	EventPlayerMailReceived zEvent.EventType = 1012 // 玩家收到邮件
	EventPlayerMailClaimed  zEvent.EventType = 1013 // 玩家领取邮件附件
)

// PlayerExpEventData 玩家经验变化事件数据
type PlayerExpEventData struct {
	PlayerID int64
	Exp      int64
}

// PlayerGoldEventData 玩家金币变化事件数据
type PlayerGoldEventData struct {
	PlayerID int64
	Gold     int64
}

// PlayerItemEventData 玩家物品变化事件数据
type PlayerItemEventData struct {
	PlayerID int64
	ItemID   int64
	Count    int
	Slot     int
}

// PlayerUseItemEventData 玩家使用物品事件数据
type PlayerUseItemEventData struct {
	PlayerID int64
	ItemID   int64
	Slot     int
	Result   bool
}

// PlayerMailEventData 玩家邮件事件数据
type PlayerMailEventData struct {
	PlayerID int64
	MailID   int64
}

// NewEvent 创建游戏事件
func NewEvent(eventType zEvent.EventType, source interface{}, data interface{}) *zEvent.Event {
	return zEvent.NewEvent(eventType, source, data)
}

// GetGlobalEventBus 获取全局事件总线实例
func GetGlobalEventBus() *zEvent.EventBus {
	return zEvent.GetGlobalEventBus()
}

// 全局Actor系统通过zEngine获取
