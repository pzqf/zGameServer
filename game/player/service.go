package player

import (
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

type PlayerService struct {
	zObject.BaseObject
	logger        *zap.Logger
	players       *zMap.Map // key: int64(playerId), value: *Player
	sessionPlayer *zMap.Map // key: int64(sessionId), value: int64(playerId)
}

func NewPlayerService(logger *zap.Logger) *PlayerService {
	ps := &PlayerService{
		logger:        logger,
		players:       zMap.NewMap(),
		sessionPlayer: zMap.NewMap(),
	}
	ps.BaseObject.Id = "PlayerService"
	return ps
}

func (ps *PlayerService) Init() error {
	ps.logger.Info("Initializing player service...")
	// 初始化玩家服务相关资源
	return nil
}

func (ps *PlayerService) Close() error {
	ps.logger.Info("Closing player service...")
	// 清理玩家服务相关资源
	ps.players.Clear()
	ps.sessionPlayer.Clear()
	return nil
}

func (ps *PlayerService) Serve() {
	// 玩家服务不需要持续运行的协程
}

func (ps *PlayerService) CreatePlayer(session *zNet.TcpServerSession, playerId int64, name string) (*Player, error) {
	// 检查玩家是否已存在
	if _, exists := ps.players.Get(playerId); exists {
		return nil, nil // 玩家已存在
	}

	// 创建新玩家
	player := NewPlayer(playerId, name, session, ps.logger)

	// 存储玩家信息
	ps.players.Store(playerId, player)
	ps.sessionPlayer.Store(int64(session.GetSid()), playerId)

	ps.logger.Info("Created new player", zap.Int64("playerId", playerId), zap.String("name", name))
	return player, nil
}

func (ps *PlayerService) GetPlayer(playerId int64) *Player {
	if player, exists := ps.players.Get(playerId); exists {
		return player.(*Player)
	}
	return nil
}

func (ps *PlayerService) GetPlayerBySession(sessionId int64) *Player {
	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		return ps.GetPlayer(playerId.(int64))
	}
	return nil
}

func (ps *PlayerService) RemovePlayer(playerId int64) {
	if player, exists := ps.players.Get(playerId); exists {
		ps.sessionPlayer.Delete(int64(player.(*Player).Session.GetSid()))
		ps.players.Delete(playerId)
		ps.logger.Info("Removed player", zap.Int64("playerId", playerId))
	}
}

func (ps *PlayerService) OnSessionClose(sessionId int64) {
	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		player := ps.GetPlayer(playerId.(int64))
		if player != nil {
			player.OnDisconnect()
			ps.RemovePlayer(playerId.(int64))
		}
	}
}
