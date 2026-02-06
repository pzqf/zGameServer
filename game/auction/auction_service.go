package auction

import (
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// AuctionService 拍卖行服务
// 管理所有拍卖物品的创建、竞拍、结算等功能
type AuctionService struct {
	zService.BaseService
	items           *zMap.TypedShardedMap[common.AuctionIdType, *AuctionItem]          // 拍卖物品映射表（AuctionId -> AuctionItem）
	playerItems     *zMap.TypedShardedMap[common.PlayerIdType, []common.AuctionIdType] // 玩家拍卖物品映射表（PlayerId -> []AuctionId）
	pendingItems    []common.AuctionIdType                                             // 待开始的拍卖列表
	activeItems     []common.AuctionIdType                                             // 进行中的拍卖列表
	feeRate         float64                                                            // 手续费率
	minBidIncrement int64                                                              // 最小加价幅度
}

// NewAuctionService 创建拍卖行服务
// 返回: 新创建的AuctionService实例
func NewAuctionService() *AuctionService {
	as := &AuctionService{
		BaseService:     *zService.NewBaseService(common.ServiceIdAuction),
		items:           zMap.NewTypedShardedMap32[common.AuctionIdType, *AuctionItem](),
		playerItems:     zMap.NewTypedShardedMap32[common.PlayerIdType, []common.AuctionIdType](),
		pendingItems:    make([]common.AuctionIdType, 0),
		activeItems:     make([]common.AuctionIdType, 0),
		feeRate:         0.05, // 5%手续费
		minBidIncrement: 10,   // 最小加价10金币
	}
	return as
}

// Init 初始化拍卖行服务
// 返回: 初始化错误（如果有）
func (as *AuctionService) Init() error {
	as.SetState(zService.ServiceStateInit)
	zLog.Info("Initializing auction service...", zap.String("serviceId", as.ServiceId()))
	return nil
}

// Close 关闭拍卖行服务
// 清理所有拍卖数据
// 返回: 关闭错误（如果有）
func (as *AuctionService) Close() error {
	as.SetState(zService.ServiceStateStopping)
	zLog.Info("Closing auction service...", zap.String("serviceId", as.ServiceId()))
	as.items.Clear()
	as.playerItems.Clear()
	as.pendingItems = make([]common.AuctionIdType, 0)
	as.activeItems = make([]common.AuctionIdType, 0)
	as.SetState(zService.ServiceStateStopped)
	return nil
}

// Serve 启动服务
// 启动拍卖计时器协程
func (as *AuctionService) Serve() {
	as.SetState(zService.ServiceStateRunning)
	go as.auctionTimerLoop()
}

// auctionTimerLoop 拍卖计时器循环
// 每500毫秒检查一次拍卖状态，处理开始和结束
func (as *AuctionService) auctionTimerLoop() {
	for range time.Tick(time.Millisecond * 500) {
		currentTime := time.Now().UnixMilli()
		as.checkPendingAuctions(currentTime)
		as.checkActiveAuctions(currentTime)
	}
}

// checkPendingAuctions 检查待开始的拍卖
// 将到达开始时间的拍卖转为进行中状态
// 参数:
//   - currentTime: 当前时间戳（毫秒）
func (as *AuctionService) checkPendingAuctions(currentTime int64) {
	for i := 0; i < len(as.pendingItems); {
		auctionId := as.pendingItems[i]
		item, exists := as.items.Load(auctionId)
		if !exists {
			as.pendingItems = append(as.pendingItems[:i], as.pendingItems[i+1:]...)
			continue
		}

		// 到达开始时间，转为进行中状态
		if item.Status == AuctionStatusPending && item.StartTime <= currentTime {
			item.Status = AuctionStatusActive
			as.activeItems = append(as.activeItems, auctionId)
			as.pendingItems = append(as.pendingItems[:i], as.pendingItems[i+1:]...)

			zLog.Info("Auction started", zap.Int64("auctionId", int64(auctionId)), zap.Int64("itemId", item.ItemId))
		} else {
			i++
		}
	}
}

// checkActiveAuctions 检查进行中的拍卖
// 将到达结束时间的拍卖转为已完成状态并触发结算
// 参数:
//   - currentTime: 当前时间戳（毫秒）
func (as *AuctionService) checkActiveAuctions(currentTime int64) {
	for i := 0; i < len(as.activeItems); {
		auctionId := as.activeItems[i]
		item, exists := as.items.Load(auctionId)
		if !exists {
			as.activeItems = append(as.activeItems[:i], as.activeItems[i+1:]...)
			continue
		}

		// 到达结束时间，转为已完成状态并结算
		if item.Status == AuctionStatusActive && item.EndTime <= currentTime {
			item.Status = AuctionStatusCompleted
			as.activeItems = append(as.activeItems[:i], as.activeItems[i+1:]...)

			zLog.Info("Auction ended", zap.Int64("auctionId", int64(auctionId)), zap.Int64("itemId", item.ItemId))

			go as.SettleAuction(auctionId)
		} else {
			i++
		}
	}
}

// removeFromPendingItems 从待开始列表中移除拍卖
// 参数:
//   - auctionId: 拍卖ID
func (as *AuctionService) removeFromPendingItems(auctionId common.AuctionIdType) {
	for i, id := range as.pendingItems {
		if id == auctionId {
			as.pendingItems = append(as.pendingItems[:i], as.pendingItems[i+1:]...)
			return
		}
	}
}

// removeFromActiveItems 从进行中列表中移除拍卖
// 参数:
//   - auctionId: 拍卖ID
func (as *AuctionService) removeFromActiveItems(auctionId common.AuctionIdType) {
	for i, id := range as.activeItems {
		if id == auctionId {
			as.activeItems = append(as.activeItems[:i], as.activeItems[i+1:]...)
			return
		}
	}
}

// CreateAuction 创建拍卖
// 参数:
//   - item: 拍卖物品
//
// 返回:
//   - error: 创建错误
func (as *AuctionService) CreateAuction(item *AuctionItem) error {
	// 检查拍卖ID是否已存在
	if _, exists := as.items.Load(common.AuctionIdType(item.AuctionId)); exists {
		return nil
	}

	as.items.Store(common.AuctionIdType(item.AuctionId), item)

	// 添加到卖家的拍卖列表
	sellerId := common.PlayerIdType(item.SellerId)
	if sellerItems, exists := as.playerItems.Load(sellerId); exists {
		sellerItems = append(sellerItems, common.AuctionIdType(item.AuctionId))
		as.playerItems.Store(sellerId, sellerItems)
	} else {
		sellerItems := []common.AuctionIdType{common.AuctionIdType(item.AuctionId)}
		as.playerItems.Store(sellerId, sellerItems)
	}

	zLog.Info("Auction created", zap.Int64("auctionId", item.AuctionId), zap.Int64("sellerId", item.SellerId), zap.Int64("itemId", item.ItemId))
	return nil
}

// PlaceBid 竞拍物品
// 参数:
//   - playerId: 竞拍玩家ID
//   - playerName: 竞拍玩家名称
//   - auctionId: 拍卖ID
//   - bidPrice: 竞拍价格
//
// 返回:
//   - error: 竞拍错误
func (as *AuctionService) PlaceBid(playerId common.PlayerIdType, playerName string, auctionId common.AuctionIdType, bidPrice int64) error {
	item, exists := as.items.Load(auctionId)
	if !exists {
		return nil
	}

	// 检查拍卖状态
	if item.Status != AuctionStatusActive {
		return nil
	}

	// 检查竞拍价格合法性
	if !as.isValidBid(item, bidPrice) {
		return nil
	}

	// 创建竞拍记录
	bid := &AuctionBid{
		BidId:      0,
		PlayerId:   int64(playerId),
		PlayerName: playerName,
		AuctionId:  int64(auctionId),
		BidPrice:   bidPrice,
		BidTime:    0,
	}

	// 更新当前价格和领先者
	item.CurrentPrice = bidPrice
	item.CurrentWinner = int64(playerId)
	item.Bids.Store(bid.BidId, bid)

	zLog.Info("Bid placed", zap.Int64("auctionId", int64(auctionId)), zap.Int64("playerId", int64(playerId)), zap.Int64("bidPrice", bidPrice))
	return nil
}

// BuyoutItem 一口价购买物品
// 参数:
//   - playerId: 购买玩家ID
//   - playerName: 购买玩家名称
//   - auctionId: 拍卖ID
//
// 返回:
//   - error: 购买错误
func (as *AuctionService) BuyoutItem(playerId common.PlayerIdType, playerName string, auctionId common.AuctionIdType) error {
	item, exists := as.items.Load(auctionId)
	if !exists {
		return nil
	}

	// 检查拍卖状态
	if item.Status != AuctionStatusActive {
		return nil
	}

	// 检查拍卖类型是否支持一口价
	if item.AuctionType != AuctionTypeBuy && item.AuctionType != AuctionTypeBoth {
		return nil
	}

	// 检查一口价是否有效
	if item.BuyoutPrice <= 0 {
		return nil
	}

	// 设置当前价格为一口价并标记完成
	item.CurrentPrice = item.BuyoutPrice
	item.CurrentWinner = int64(playerId)
	item.Status = AuctionStatusCompleted

	zLog.Info("Item bought out", zap.Int64("auctionId", int64(auctionId)), zap.Int64("playerId", int64(playerId)), zap.Int64("buyoutPrice", item.BuyoutPrice))
	return nil
}

// CancelAuction 取消拍卖
// 只有卖家可以取消自己的拍卖
// 参数:
//   - auctionId: 拍卖ID
//
// 返回:
//   - error: 取消错误
func (as *AuctionService) CancelAuction(auctionId common.AuctionIdType) error {
	item, exists := as.items.Load(auctionId)
	if !exists {
		return nil
	}

	// 只能取消待开始或进行中的拍卖
	if item.Status != AuctionStatusPending && item.Status != AuctionStatusActive {
		return nil
	}

	item.Status = AuctionStatusCanceled

	zLog.Info("Auction canceled", zap.Int64("auctionId", int64(auctionId)), zap.Int64("sellerId", item.SellerId))
	return nil
}

// SettleAuction 结算拍卖
// 拍卖结束后调用，处理物品和金币的转移
// 参数:
//   - auctionId: 拍卖ID
//
// 返回:
//   - error: 结算错误
func (as *AuctionService) SettleAuction(auctionId common.AuctionIdType) error {
	item, exists := as.items.Load(auctionId)
	if !exists {
		return nil
	}

	// 检查是否已结算
	if item.IsSettled {
		return nil
	}

	item.IsSettled = true

	zLog.Info("Auction settled", zap.Int64("auctionId", int64(auctionId)), zap.Int64("sellerId", item.SellerId), zap.Int64("winnerId", item.CurrentWinner))
	return nil
}

// GetAuctionItem 获取拍卖物品信息
// 参数:
//   - auctionId: 拍卖ID
//
// 返回:
//   - *AuctionItem: 拍卖物品
//   - bool: 是否存在
func (as *AuctionService) GetAuctionItem(auctionId common.AuctionIdType) (*AuctionItem, bool) {
	return as.items.Load(auctionId)
}

// GetPlayerAuctions 获取玩家的拍卖物品
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - []*AuctionItem: 拍卖物品列表
//   - bool: 是否存在
func (as *AuctionService) GetPlayerAuctions(playerId common.PlayerIdType) ([]*AuctionItem, bool) {
	auctionIds, exists := as.playerItems.Load(playerId)
	if !exists {
		return nil, false
	}

	var items []*AuctionItem
	for _, auctionId := range auctionIds {
		if item, exists := as.GetAuctionItem(auctionId); exists {
			items = append(items, item)
		}
	}

	return items, true
}

// isValidBid 检查竞拍价格是否合法
// 参数:
//   - item: 拍卖物品
//   - bidPrice: 竞拍价格
//
// 返回:
//   - bool: 是否合法
func (as *AuctionService) isValidBid(item *AuctionItem, bidPrice int64) bool {
	// 检查竞拍价格是否大于当前价格
	if bidPrice <= item.CurrentPrice {
		return false
	}

	// 检查竞拍价格是否大于等于起拍价
	if bidPrice < item.StartingPrice {
		return false
	}

	// 检查竞拍价格是否大于等于当前价格加上最小加价（物品配置）
	if bidPrice < item.CurrentPrice+item.BidIncrement {
		return false
	}

	// 检查竞拍价格是否大于等于当前价格加上默认最小加价（服务配置）
	if bidPrice < item.CurrentPrice+as.minBidIncrement {
		return false
	}

	return true
}
