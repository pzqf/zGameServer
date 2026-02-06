package auction

import (
	"github.com/pzqf/zUtil/zMap"
)

// 拍卖类型定义
const (
	AuctionTypeBid  = 1 // 竞拍（仅竞价）
	AuctionTypeBuy  = 2 // 一口价（仅直接购买）
	AuctionTypeBoth = 3 // 竞拍+一口价（两种方式都支持）
)

// 拍卖状态定义
const (
	AuctionStatusPending   = 1 // 待开始（拍卖未到开始时间）
	AuctionStatusActive    = 2 // 进行中（拍卖正在进行）
	AuctionStatusCompleted = 3 // 已结束（拍卖正常结束）
	AuctionStatusCanceled  = 4 // 已取消（卖家取消拍卖）
)

// AuctionBid 竞拍记录
// 记录玩家每次竞拍的详细信息
type AuctionBid struct {
	BidId      int64  // 竞拍记录ID
	PlayerId   int64  // 竞拍玩家ID
	PlayerName string // 竞拍玩家名称
	AuctionId  int64  // 所属拍卖ID
	BidPrice   int64  // 竞拍价格
	BidTime    int64  // 竞拍时间戳（毫秒）
}

// AuctionItem 拍卖物品
// 表示拍卖行中的一个拍卖条目
type AuctionItem struct {
	AuctionId     int64            // 拍卖ID（全局唯一）
	SellerId      int64            // 卖家玩家ID
	SellerName    string           // 卖家玩家名称
	ItemId        int64            // 物品ID
	ItemName      string           // 物品名称
	ItemType      int              // 物品类型
	ItemCount     int              // 物品数量
	AuctionType   int              // 拍卖类型（AuctionTypeBid/Buy/Both）
	StartingPrice int64            // 起拍价格
	CurrentPrice  int64            // 当前最高出价
	BuyoutPrice   int64            // 一口价（0表示不支持一口价）
	BidIncrement  int64            // 最小加价幅度
	StartTime     int64            // 拍卖开始时间戳（毫秒）
	Duration      int64            // 拍卖持续时间（毫秒）
	EndTime       int64            // 拍卖结束时间戳（毫秒）
	Status        int              // 拍卖状态（AuctionStatus*）
	CurrentWinner int64            // 当前领先者ID（最高出价者）
	Bids          *zMap.ShardedMap // 竞拍记录映射表（bidId -> *AuctionBid）
	IsSettled     bool             // 是否已结算
}
