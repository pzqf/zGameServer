package handler

import (
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/db"
	"github.com/pzqf/zGameServer/router"
	"github.com/pzqf/zGameServer/service"
)

// Init 初始化所有处理器
func Init(router *router.PacketRouter,
	playerService *service.PlayerService,
	guildService *service.GuildService,
	auctionService *service.AuctionService,
	mapService *service.MapService,
	dbManager *db.DBManager) {

	zLog.Info("Initializing handlers...")

	// 注册玩家处理器
	RegisterPlayerHandlers(router, playerService, dbManager)

	// 注册其他模块的处理器（根据需要添加）
	// RegisterGuildHandlers(router, guildService)
	// RegisterAuctionHandlers(router, auctionService)
	// RegisterMapHandlers(router, mapService)

	zLog.Info("All handlers initialized")
}
