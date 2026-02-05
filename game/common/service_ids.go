package common

// ServiceId 服务ID常量定义
const (
	// 网络服务
	ServiceIdTcpServer  = "tcp_server"
	ServiceIdHttpServer = "http_server"

	// 游戏服务
	ServiceIdPlayer  = "player_service"
	ServiceIdGuild   = "guild_service"
	ServiceIdAuction = "auction_service"
	ServiceIdMap     = "map_service"

	// 其他服务
	ServiceIdDBManager = "db_manager"
	ServiceIdConfig    = "config_service"
)
