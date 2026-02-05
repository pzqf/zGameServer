package common

import (
	"errors"
)

var (
	globalSnowflake *Snowflake
	workerID        int64
	datacenterID    int64
)

func InitIDGenerator(wID, dcID int64) error {
	sf, err := NewSnowflake(wID, dcID)
	if err != nil {
		return err
	}
	globalSnowflake = sf
	workerID = wID
	datacenterID = dcID
	return nil
}

func generateID() (int64, error) {
	if globalSnowflake == nil {
		return 0, errors.New("ID generator not initialized")
	}
	return globalSnowflake.NextID()
}

// PlayerIdType 玩家ID类型
type PlayerIdType int64

// MapIdType 地图ID类型
type MapIdType int64

// RegionIdType 区域ID类型
type RegionIdType int64

// ObjectIdType 游戏对象ID类型
type ObjectIdType int64

// AccountIdType 账号ID类型
type AccountIdType int64

// CharIdType 角色ID类型
type CharIdType int64

// GroupIdType 队伍ID类型
type GroupIdType int64

// ComboIdType 连击ID类型
type ComboIdType int64

// VisualIdType 视觉效果ID类型
type VisualIdType int64

// LogIdType 日志ID类型
type LogIdType int64

// ItemIdType 物品ID类型
type ItemIdType int64

// MailIdType 邮件ID类型
type MailIdType int64

// SkillIdType 技能ID类型
type SkillIdType int64

// TaskIdType 任务ID类型
type TaskIdType int64

// GuildIdType 公会ID类型
type GuildIdType int64

// ApplyIdType 申请ID类型
type ApplyIdType int64

// AuctionIdType 拍卖ID类型
type AuctionIdType int64

// BidIdType 出价ID类型
type BidIdType int64

// SenderIdType 发送者ID类型
type SenderIdType int64

// ReceiverIdType 接收者ID类型
type ReceiverIdType int64

// LeaderIdType 领导者ID类型
type LeaderIdType int64

// TargetIdType 目标ID类型
type TargetIdType int64

// SellerIdType 卖家ID类型
type SellerIdType int64

// OperatorIdType 操作者ID类型
type OperatorIdType int64

func GeneratePlayerID() (PlayerIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return PlayerIdType(id), nil
}

func GenerateMapID() (MapIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return MapIdType(id), nil
}

func GenerateObjectID() (ObjectIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return ObjectIdType(id), nil
}

func GenerateAccountID() (AccountIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return AccountIdType(id), nil
}

func GenerateCharID() (CharIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return CharIdType(id), nil
}

func GenerateGroupID() (GroupIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return GroupIdType(id), nil
}

func GenerateComboID() (ComboIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return ComboIdType(id), nil
}

func GenerateVisualID() (VisualIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return VisualIdType(id), nil
}

func GenerateLogID() (LogIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return LogIdType(id), nil
}

func GenerateItemID() (ItemIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return ItemIdType(id), nil
}

func GenerateMailID() (MailIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return MailIdType(id), nil
}

func GenerateSkillID() (SkillIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return SkillIdType(id), nil
}

func GenerateTaskID() (TaskIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return TaskIdType(id), nil
}

func GenerateGuildID() (GuildIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return GuildIdType(id), nil
}

func GenerateApplyID() (ApplyIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return ApplyIdType(id), nil
}

func GenerateAuctionID() (AuctionIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return AuctionIdType(id), nil
}

func GenerateBidID() (BidIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return BidIdType(id), nil
}
