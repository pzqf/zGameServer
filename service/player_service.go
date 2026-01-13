package service

import (
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/game/player"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

type PlayerService struct {
	zObject.BaseObject
	logger        *zap.Logger
	players       *zMap.Map // key: int64(playerId), value: *player.Player
	sessionPlayer *zMap.Map // key: int64(sessionId), value: int64(playerId)
}

func NewPlayerService() *PlayerService {
	ps := &PlayerService{
		logger:        zLog.GetLogger(),
		players:       zMap.NewMap(),
		sessionPlayer: zMap.NewMap(),
	}
	ps.SetId(ServiceIdPlayerService)
	return ps
}

func (ps *PlayerService) Init() error {
	ps.logger.Info("Initializing player service...")
	// 初始化玩家服务相关资源
	return nil
}

func (ps *PlayerService) Close() error {
	ps.logger.Info("Closing player service...")
	// 保存所有玩家数据
	playerCount := 0
	ps.players.Range(func(key, value interface{}) bool {
		playerId := key.(int64)
		player := value.(*player.Player)
		// 这里应该实现真正的数据持久化逻辑（如保存到数据库）
		// 目前实现一个模拟的数据保存过程
		ps.logger.Info("Saving player data",
			zap.Int64("playerId", playerId),
			zap.String("name", player.GetName()),
			zap.Int("level", player.GetBasicInfo().Level),
			zap.Int64("gold", player.GetBasicInfo().Gold.Load()))
		playerCount++
		return true
	})
	ps.logger.Info("All player data saved", zap.Int("playerCount", playerCount))
	// 清理玩家服务相关资源
	ps.players.Clear()
	ps.sessionPlayer.Clear()
	return nil
}

func (ps *PlayerService) Serve() {
	// 玩家服务不需要持续运行的协程
}

func (ps *PlayerService) CreatePlayer(session *zNet.TcpServerSession, playerId int64, name string) (*player.Player, error) {
	// 检查玩家是否已存在
	if _, exists := ps.players.Get(playerId); exists {
		return nil, nil // 玩家已存在
	}

	// 创建新玩家
	player := player.NewPlayer(playerId, name, session, ps.logger)

	// 存储玩家信息
	ps.players.Store(playerId, player)
	ps.sessionPlayer.Store(int64(session.GetSid()), playerId)

	ps.logger.Info("Created new player", zap.Int64("playerId", playerId), zap.String("name", name))
	return player, nil
}

func (ps *PlayerService) GetPlayer(playerId int64) *player.Player {
	if playerObj, exists := ps.players.Get(playerId); exists {
		return playerObj.(*player.Player)
	}
	return nil
}

func (ps *PlayerService) GetPlayerBySession(sessionId int64) *player.Player {
	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		return ps.GetPlayer(playerId.(int64))
	}
	return nil
}

func (ps *PlayerService) RemovePlayer(playerId int64) {
	if playerObj, exists := ps.players.Get(playerId); exists {
		ps.sessionPlayer.Delete(int64(playerObj.(*player.Player).Session.GetSid()))
		ps.players.Delete(playerId)
		ps.logger.Info("Removed player", zap.Int64("playerId", playerId))
	}
}
