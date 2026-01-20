package player

import (
	"github.com/pzqf/zEngine/zActor"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

type Service struct {
	zObject.BaseObject
	logger        *zap.Logger
	players       *zMap.Map // key: int64(playerId), value: *Player
	playerActors  *zMap.Map // key: int64(playerId), value: *PlayerActor
	sessionPlayer *zMap.Map // key: int64(sessionId), value: int64(playerId)
}

func NewService() *Service {
	logger := zLog.GetLogger()
	ps := &Service{
		logger:        logger,
		players:       zMap.NewMap(),
		playerActors:  zMap.NewMap(),
		sessionPlayer: zMap.NewMap(),
	}
	ps.SetId("player_service")
	return ps
}

func (ps *Service) Init() error {
	ps.logger.Info("Initializing player service...")
	// 初始化玩家服务相关资源
	return nil
}

func (ps *Service) Close() error {
	ps.logger.Info("Closing player service...")
	// 清理玩家服务相关资源
	ps.players.Clear()

	// 停止并清理所有PlayerActor
	ps.playerActors.Range(func(key, value interface{}) bool {
		playerActor := value.(*PlayerActor)
		playerActor.Stop()
		ps.playerActors.Delete(key)
		return true
	})

	ps.sessionPlayer.Clear()
	return nil
}

func (ps *Service) Serve() {
	// 玩家服务不需要持续运行的协程
}

// CreatePlayerActor 创建玩家Actor
func (ps *Service) CreatePlayerActor(session *zNet.TcpServerSession, playerId int64, name string) (*PlayerActor, error) {
	// 检查玩家是否已存在
	if _, exists := ps.playerActors.Get(playerId); exists {
		return nil, nil // 玩家Actor已存在
	}

	// 创建新玩家Actor
	playerActor := NewPlayerActor(playerId, name, session, ps.logger)

	// 注册到全局Actor系统
	// 注意：需要先在zEngine中初始化全局Actor系统
	// 这里假设已经在其他地方初始化了全局Actor系统实例
	// actorSystem := actor.NewActorSystem()
	// if err := actorSystem.Start(); err != nil {
	//     ps.logger.Error("Failed to start actor system", zap.Error(err))
	//     return nil, err
	// }
	// if err := actorSystem.RegisterActor(playerActor); err != nil {
	//     ps.logger.Error("Failed to register player actor", zap.Int64("playerId", playerId), zap.Error(err))
	//     return nil, err
	// }

	// 存储玩家Actor信息
	ps.playerActors.Store(playerId, playerActor)
	ps.sessionPlayer.Store(int64(session.GetSid()), playerId)

	ps.logger.Info("Created new player actor", zap.Int64("playerId", playerId), zap.String("name", name))
	return playerActor, nil
}

func (ps *Service) CreatePlayer(session *zNet.TcpServerSession, playerId int64, name string) (*Player, error) {
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

func (ps *Service) GetPlayer(playerId int64) *Player {
	if player, exists := ps.players.Get(playerId); exists {
		return player.(*Player)
	}
	return nil
}

func (ps *Service) GetPlayerBySession(sessionId int64) *Player {
	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		return ps.GetPlayer(playerId.(int64))
	}
	return nil
}

// GetPlayerActor 获取玩家Actor
func (ps *Service) GetPlayerActor(playerId int64) *PlayerActor {
	if playerActor, exists := ps.playerActors.Get(playerId); exists {
		return playerActor.(*PlayerActor)
	}
	return nil
}

// GetPlayerActorBySession 根据会话获取玩家Actor
func (ps *Service) GetPlayerActorBySession(sessionId int64) *PlayerActor {
	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		return ps.GetPlayerActor(playerId.(int64))
	}
	return nil
}

func (ps *Service) RemovePlayer(playerId int64) {
	// 检查并移除传统Player
	if player, exists := ps.players.Get(playerId); exists {
		ps.sessionPlayer.Delete(int64(player.(*Player).Session.GetSid()))
		ps.players.Delete(playerId)
		ps.logger.Info("Removed player", zap.Int64("playerId", playerId))
		return
	}

	// 检查并移除PlayerActor
	if playerActor, exists := ps.playerActors.Get(playerId); exists {
		// 从全局Actor系统中注销
		if err := zActor.GetGlobalActorSystem().UnregisterActor(playerId); err != nil {
			ps.logger.Error("Failed to unregister player actor", zap.Int64("playerId", playerId), zap.Error(err))
		}

		// 更新会话映射
		ps.sessionPlayer.Delete(int64(playerActor.(*PlayerActor).GetSession().GetSid()))
		ps.playerActors.Delete(playerId)
		ps.logger.Info("Removed player actor", zap.Int64("playerId", playerId))
	}
}

func (ps *Service) OnSessionClose(sessionId int64) {
	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		// 检查是否为PlayerActor
		playerActor := ps.GetPlayerActor(playerId.(int64))
		if playerActor != nil {
			// 发送断开连接消息给Actor
			disconnectMsg := &PlayerActorDisconnectMessage{
				BaseActorMessage: zActor.BaseActorMessage{ActorID: playerId.(int64)},
			}
			playerActor.SendMessage(disconnectMsg)

			// 移除PlayerActor
			ps.RemovePlayer(playerId.(int64))
			return
		}

		// 检查是否为传统Player
		player := ps.GetPlayer(playerId.(int64))
		if player != nil {
			player.OnDisconnect()
			ps.RemovePlayer(playerId.(int64))
		}
	}
}
