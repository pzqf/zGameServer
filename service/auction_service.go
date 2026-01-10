package service

import (
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
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
	bidId      int64
	playerId   int64
	playerName string
	auctionId  int64
	bidPrice   int64
	bidTime    int64
}

// AuctionItem 拍卖物品
type AuctionItem struct {
	auctionId     int64
	sellerId      int64
	sellerName    string
	itemId        int64
	itemName      string
	itemType      int
	itemCount     int
	auctionType   int
	startingPrice int64
	currentPrice  int64
	buyoutPrice   int64
	bidIncrement  int64
	startTime     int64
	duration      int64 // 拍卖持续时间（毫秒）
	endTime       int64
	status        int
	currentWinner int64
	bids          zMap.Map // key: int64(bidId), value: *AuctionBid
	isSettled     bool
}

// AuctionService 拍卖行服务
type AuctionService struct {
	zObject.BaseObject
	logger          *zap.Logger
	items           *zMap.Map // key: int64(auctionId), value: *AuctionItem
	playerItems     *zMap.Map // key: int64(playerId), value: []int64(auctionId)
	pendingItems    []int64   // 待开始的拍卖物品ID
	activeItems     []int64   // 进行中的拍卖物品ID
	feeRate         float64   // 拍卖手续费率
	minBidIncrement int64     // 最小竞拍加价
}

func NewAuctionService() *AuctionService {
	as := &AuctionService{
		logger:          zLog.GetLogger(),
		items:           zMap.NewMap(),
		playerItems:     zMap.NewMap(),
		pendingItems:    make([]int64, 0),
		activeItems:     make([]int64, 0),
		feeRate:         0.05, // 5%手续费
		minBidIncrement: 10,   // 最小加价10
	}
	as.SetId(ServiceIdAuctionService)
	return as
}

func (as *AuctionService) Init() error {
	as.logger.Info("Initializing auction service...")
	// 初始化拍卖行服务相关资源
	return nil
}

func (as *AuctionService) Close() error {
	as.logger.Info("Closing auction service...")
	// 清理拍卖行服务相关资源
	as.items.Clear()
	as.playerItems.Clear()
	as.pendingItems = make([]int64, 0)
	as.activeItems = make([]int64, 0)
	return nil
}

func (as *AuctionService) Serve() {
	// 拍卖行服务需要持续运行的协程，用于处理拍卖的开始和结束
	go as.auctionTimerLoop()
}

// auctionTimerLoop 拍卖计时器循环
func (as *AuctionService) auctionTimerLoop() {
	// 每500毫秒检查一次拍卖状态
	for {
		select {
		case <-time.After(time.Millisecond * 500):
			currentTime := time.Now().UnixMilli()
			as.checkPendingAuctions(currentTime)
			as.checkActiveAuctions(currentTime)
		}
	}
}

// checkPendingAuctions 检查待开始的拍卖
func (as *AuctionService) checkPendingAuctions(currentTime int64) {
	// 遍历待开始的拍卖列表
	for i := 0; i < len(as.pendingItems); {
		auctionId := as.pendingItems[i]
		itemInterface, exists := as.items.Get(auctionId)
		if !exists {
			// 拍卖不存在，从列表中移除
			as.pendingItems = append(as.pendingItems[:i], as.pendingItems[i+1:]...)
			continue
		}

		item := itemInterface.(*AuctionItem)
		if item.status == AuctionStatusPending && item.startTime <= currentTime {
			// 设置拍卖为进行中
			item.status = AuctionStatusActive
			// 添加到进行中列表
			as.activeItems = append(as.activeItems, auctionId)
			// 从待开始列表中移除
			as.pendingItems = append(as.pendingItems[:i], as.pendingItems[i+1:]...)

			as.logger.Info("Auction started", zap.Int64("auctionId", auctionId), zap.Int64("itemId", item.itemId))
		} else {
			i++
		}
	}
}

// checkActiveAuctions 检查进行中的拍卖
func (as *AuctionService) checkActiveAuctions(currentTime int64) {
	// 遍历进行中的拍卖列表
	for i := 0; i < len(as.activeItems); {
		auctionId := as.activeItems[i]
		itemInterface, exists := as.items.Get(auctionId)
		if !exists {
			// 拍卖不存在，从列表中移除
			as.activeItems = append(as.activeItems[:i], as.activeItems[i+1:]...)
			continue
		}

		item := itemInterface.(*AuctionItem)
		if item.status == AuctionStatusActive && item.endTime <= currentTime {
			// 设置拍卖为已结束
			item.status = AuctionStatusCompleted
			// 从进行中列表中移除
			as.activeItems = append(as.activeItems[:i], as.activeItems[i+1:]...)

			as.logger.Info("Auction ended", zap.Int64("auctionId", auctionId), zap.Int64("itemId", item.itemId))

			// 结算拍卖
			go as.SettleAuction(auctionId)
		} else {
			i++
		}
	}
}

// removeFromPendingItems 从待开始列表中移除拍卖
func (as *AuctionService) removeFromPendingItems(auctionId int64) {
	for i, id := range as.pendingItems {
		if id == auctionId {
			as.pendingItems = append(as.pendingItems[:i], as.pendingItems[i+1:]...)
			return
		}
	}
}

// removeFromActiveItems 从进行中列表中移除拍卖
func (as *AuctionService) removeFromActiveItems(auctionId int64) {
	for i, id := range as.activeItems {
		if id == auctionId {
			as.activeItems = append(as.activeItems[:i], as.activeItems[i+1:]...)
			return
		}
	}
}

// CreateAuction 创建拍卖
func (as *AuctionService) CreateAuction(item *AuctionItem) error {
	// 检查拍卖物品是否已存在
	if _, exists := as.items.Get(item.auctionId); exists {
		return nil // 拍卖物品已存在
	}

	// 检查卖家是否有足够的物品
	// TODO: 实现物品检查逻辑

	// 存储拍卖物品
	as.items.Store(item.auctionId, item)

	// 添加到卖家的拍卖物品列表
	if sellerItemsInterface, exists := as.playerItems.Get(item.sellerId); exists {
		sellerItems := sellerItemsInterface.([]int64)
		sellerItems = append(sellerItems, item.auctionId)
		as.playerItems.Store(item.sellerId, sellerItems)
	} else {
		sellerItems := []int64{item.auctionId}
		as.playerItems.Store(item.sellerId, sellerItems)
	}

	// 根据拍卖开始时间添加到待开始或进行中列表
	// currentTime := time.Now().UnixMilli()
	// if item.startTime > currentTime {
	//     as.pendingItems = append(as.pendingItems, item.auctionId)
	// } else if item.endTime > currentTime {
	//     item.status = AuctionStatusActive
	//     as.activeItems = append(as.activeItems, item.auctionId)
	// } else {
	//     item.status = AuctionStatusCompleted
	// }

	as.logger.Info("Auction created", zap.Int64("auctionId", item.auctionId), zap.Int64("sellerId", item.sellerId), zap.Int64("itemId", item.itemId))
	return nil
}

// PlaceBid 竞拍物品
func (as *AuctionService) PlaceBid(playerId int64, playerName string, auctionId int64, bidPrice int64) error {
	// 获取拍卖物品
	itemInterface, exists := as.items.Get(auctionId)
	if !exists {
		return nil // 拍卖物品不存在
	}
	item := itemInterface.(*AuctionItem)

	// 检查拍卖状态
	if item.status != AuctionStatusActive {
		return nil // 拍卖未进行中
	}

	// 检查竞拍价格是否合法
	if !as.isValidBid(item, bidPrice) {
		return nil // 竞拍价格不合法
	}

	// 检查玩家是否有足够的金币
	// TODO: 实现金币检查逻辑

	// 创建竞拍记录
	bid := &AuctionBid{
		bidId:      0, // 应该生成唯一ID
		playerId:   playerId,
		playerName: playerName,
		auctionId:  auctionId,
		bidPrice:   bidPrice,
		bidTime:    0, // 应该设置为当前时间
	}

	// 更新拍卖物品信息
	item.currentPrice = bidPrice
	item.currentWinner = playerId
	item.bids.Store(bid.bidId, bid)

	as.logger.Info("Bid placed", zap.Int64("auctionId", auctionId), zap.Int64("playerId", playerId), zap.Int64("bidPrice", bidPrice))
	return nil
}

// BuyoutItem 一口价购买物品
func (as *AuctionService) BuyoutItem(playerId int64, playerName string, auctionId int64) error {
	// 获取拍卖物品
	itemInterface, exists := as.items.Get(auctionId)
	if !exists {
		return nil // 拍卖物品不存在
	}
	item := itemInterface.(*AuctionItem)

	// 检查拍卖状态
	if item.status != AuctionStatusActive {
		return nil // 拍卖未进行中
	}

	// 检查是否支持一口价
	if item.auctionType != AuctionTypeBuy && item.auctionType != AuctionTypeBoth {
		return nil // 不支持一口价
	}

	// 检查一口价是否合法
	if item.buyoutPrice <= 0 {
		return nil // 一口价未设置
	}

	// 检查玩家是否有足够的金币
	// TODO: 实现金币检查逻辑

	// 更新拍卖物品信息
	item.currentPrice = item.buyoutPrice
	item.currentWinner = playerId
	item.status = AuctionStatusCompleted

	// 从进行中列表中移除
	// as.removeFromActiveItems(auctionId)

	as.logger.Info("Item bought out", zap.Int64("auctionId", auctionId), zap.Int64("playerId", playerId), zap.Int64("buyoutPrice", item.buyoutPrice))
	return nil
}

// CancelAuction 取消拍卖
func (as *AuctionService) CancelAuction(auctionId int64) error {
	// 获取拍卖物品
	itemInterface, exists := as.items.Get(auctionId)
	if !exists {
		return nil // 拍卖物品不存在
	}
	item := itemInterface.(*AuctionItem)

	// 检查拍卖状态
	if item.status != AuctionStatusPending && item.status != AuctionStatusActive {
		return nil // 拍卖已结束或已取消
	}

	// 更新拍卖物品状态
	item.status = AuctionStatusCanceled

	// 从待开始或进行中列表中移除
	// if item.status == AuctionStatusPending {
	//     as.removeFromPendingItems(auctionId)
	// } else if item.status == AuctionStatusActive {
	//     as.removeFromActiveItems(auctionId)
	// }

	as.logger.Info("Auction canceled", zap.Int64("auctionId", auctionId), zap.Int64("sellerId", item.sellerId))
	return nil
}

// SettleAuction 结算拍卖
func (as *AuctionService) SettleAuction(auctionId int64) error {
	// 获取拍卖物品
	itemInterface, exists := as.items.Get(auctionId)
	if !exists {
		return nil // 拍卖物品不存在
	}
	item := itemInterface.(*AuctionItem)

	// 检查拍卖是否已结算
	if item.isSettled {
		return nil // 拍卖已结算
	}

	// TODO: 实现拍卖结算逻辑
	// 1. 向卖家支付销售金额（扣除手续费）
	// 2. 向买家发放物品
	// 3. 如果没有买家，向卖家返还物品

	// 标记拍卖已结算
	item.isSettled = true

	as.logger.Info("Auction settled", zap.Int64("auctionId", auctionId), zap.Int64("sellerId", item.sellerId), zap.Int64("winnerId", item.currentWinner))
	return nil
}

// GetAuctionItem 获取拍卖物品信息
func (as *AuctionService) GetAuctionItem(auctionId int64) (*AuctionItem, bool) {
	item, exists := as.items.Get(auctionId)
	if !exists {
		return nil, false
	}
	return item.(*AuctionItem), true
}

// GetPlayerAuctions 获取玩家的拍卖物品
func (as *AuctionService) GetPlayerAuctions(playerId int64) ([]*AuctionItem, bool) {
	auctionIdsInterface, exists := as.playerItems.Get(playerId)
	if !exists {
		return nil, false
	}
	auctionIds := auctionIdsInterface.([]int64)

	var items []*AuctionItem
	for _, auctionId := range auctionIds {
		if item, exists := as.GetAuctionItem(auctionId); exists {
			items = append(items, item)
		}
	}

	return items, true
}

// isValidBid 检查竞拍价格是否合法
func (as *AuctionService) isValidBid(item *AuctionItem, bidPrice int64) bool {
	// 检查竞拍价格是否大于当前价格
	if bidPrice <= item.currentPrice {
		return false
	}

	// 检查竞拍价格是否大于等于起拍价
	if bidPrice < item.startingPrice {
		return false
	}

	// 检查竞拍价格是否大于等于当前价格加上最小加价
	if bidPrice < item.currentPrice+item.bidIncrement {
		return false
	}

	// 检查竞拍价格是否大于等于当前价格加上默认最小加价
	if bidPrice < item.currentPrice+as.minBidIncrement {
		return false
	}

	return true
}
