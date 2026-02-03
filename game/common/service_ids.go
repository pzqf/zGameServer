package common

import "go.uber.org/zap"

// 类型定义

// PlayerIdType 玩家ID类型
type PlayerIdType int64

func ZapPlayerField(playerId PlayerIdType) zap.Field {
	return zap.Int64("playerId", int64(playerId))
}

// MapIdType 地图ID类型
type MapIdType int64

// ObjectIdType 对象ID类型
type ObjectIdType int64

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
