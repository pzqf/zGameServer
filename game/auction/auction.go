package auction

import (
	"github.com/pzqf/zUtil/zMap"
)

// 拍卖类型定义
const (
	AuctionTypeBid  = 1 // 竞拍
	AuctionTypeBuy  = 2 // 一口价
	AuctionTypeBoth = 3 // 竞拍+一口价
)

// 拍卖状态定义
const (
	AuctionStatusPending   = 1 // 待开始
	AuctionStatusActive    = 2 // 进行中
	AuctionStatusCompleted = 3 // 已结束
	AuctionStatusCanceled  = 4 // 已取消
)

// AuctionBid 竞拍记录
type AuctionBid struct {
	BidId      int64
	PlayerId   int64
	PlayerName string
	AuctionId  int64
	BidPrice   int64
	BidTime    int64
}

// AuctionItem 拍卖物品
type AuctionItem struct {
	AuctionId     int64
	SellerId      int64
	SellerName    string
	ItemId        int64
	ItemName      string
	ItemType      int
	ItemCount     int
	AuctionType   int
	StartingPrice int64
	CurrentPrice  int64
	BuyoutPrice   int64
	BidIncrement  int64
	StartTime     int64
	Duration      int64 // 拍卖持续时间（毫秒）
	EndTime       int64
	Status        int
	CurrentWinner int64
	Bids          zMap.Map // key: int64(bidId), value: *AuctionBid
	IsSettled     bool
}
