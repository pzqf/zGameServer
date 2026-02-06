package common

import (
	"errors"
)

var (
	globalSnowflake *Snowflake // 全局Snowflake分布式ID生成器实例
	workerID        int64      // 工作节点ID
	datacenterID    int64      // 数据中心ID
)

// InitIDGenerator 初始化全局ID生成器
// wID: 工作节点ID, dcID: 数据中心ID
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

// generateID 生成一个全局唯一的int64类型ID
func generateID() (int64, error) {
	if globalSnowflake == nil {
		return 0, errors.New("ID generator not initialized")
	}
	return globalSnowflake.NextID()
}

// generateInt32ID 生成一个全局唯一的int32类型ID
func generateInt32ID() (int32, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return int32(id & 0x7FFFFFFF), nil
}

// GeneratePlayerID 生成玩家ID
func GeneratePlayerID() (PlayerIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return PlayerIdType(id), nil
}

// GenerateMapID 生成地图ID
func GenerateMapID() (MapIdType, error) {
	id, err := generateInt32ID()
	if err != nil {
		return 0, err
	}
	return MapIdType(id), nil
}

// GenerateObjectID 生成游戏对象ID
func GenerateObjectID() (ObjectIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return ObjectIdType(id), nil
}

// GenerateAccountID 生成账号ID
func GenerateAccountID() (AccountIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return AccountIdType(id), nil
}

// GenerateTeamID 生成队伍ID
func GenerateTeamID() (TeamIdType, error) {
	id, err := generateInt32ID()
	if err != nil {
		return 0, err
	}
	return TeamIdType(id), nil
}

// GenerateComboID 生成连招/组合ID
func GenerateComboID() (ComboIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return ComboIdType(id), nil
}

// GenerateVisualID 生成视觉效果ID
func GenerateVisualID() (VisualIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return VisualIdType(id), nil
}

// GenerateLogID 生成日志ID
func GenerateLogID() (LogIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return LogIdType(id), nil
}

// GenerateItemID 生成道具ID
func GenerateItemID() (ItemIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return ItemIdType(id), nil
}

// GenerateMailID 生成邮件ID
func GenerateMailID() (MailIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return MailIdType(id), nil
}

// GenerateGuildID 生成公会ID
func GenerateGuildID() (GuildIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return GuildIdType(id), nil
}

// GenerateAuctionID 生成拍卖物品ID
func GenerateAuctionID() (AuctionIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return AuctionIdType(id), nil
}

// GenerateBidID 生成竞拍记录ID
func GenerateBidID() (BidIdType, error) {
	id, err := generateID()
	if err != nil {
		return 0, err
	}
	return BidIdType(id), nil
}
