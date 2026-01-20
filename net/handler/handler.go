package handler

import (
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/db"
	"github.com/pzqf/zGameServer/game/auction"
	"github.com/pzqf/zGameServer/game/guild"
	"github.com/pzqf/zGameServer/game/maps"
	"github.com/pzqf/zGameServer/game/player"
	"github.com/pzqf/zGameServer/net/router"
)

// Init 初始化所有处理器
func Init(router *router.PacketRouter,
	playerService *player.Service,
	guildService *guild.Service,
	auctionService *auction.Service,
	mapService *maps.Service,
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
