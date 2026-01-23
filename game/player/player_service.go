package player

import (
	"github.com/pzqf/zEngine/zActor"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/util"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

type PlayerService struct {
	zService.BaseService
	players       *zMap.Map // key: int64(playerId), value: *Player
	playerActors  *zMap.Map // key: int64(playerId), value: *PlayerActor
	sessionPlayer *zMap.Map // key: int64(sessionId), value: int64(playerId)
}

func NewPlayerService() *PlayerService {
	ps := &PlayerService{
		BaseService:   *zService.NewBaseService(util.ServiceIdPlayer),
		players:       zMap.NewMap(),
		playerActors:  zMap.NewMap(),
		sessionPlayer: zMap.NewMap(),
	}
	return ps
}

func (ps *PlayerService) Init() error {
	ps.SetState(zService.ServiceStateInit)
	zLog.Info("Initializing player service...")
	// 初始化玩家服务相关资源
	return nil
}

func (ps *PlayerService) Close() error {
	ps.SetState(zService.ServiceStateStopping)
	zLog.Info("Closing player service...")
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
	ps.SetState(zService.ServiceStateStopped)
	return nil
}

// CreatePlayerActor 创建玩家Actor
func (ps *PlayerService) CreatePlayerActor(session *zNet.TcpServerSession, playerId int64, name string) (*PlayerActor, error) {
	// 检查玩家是否已存在
	if _, exists := ps.playerActors.Get(playerId); exists {
		return nil, nil // 玩家Actor已存在
	}

	// 创建新玩家Actor
	playerActor := NewPlayerActor(playerId, name, session)

	// 注册到全局Actor系统
	// 注意：需要先在zEngine中初始化全局Actor系统
	// 这里假设已经在其他地方初始化了全局Actor系统实例
	// actorSystem := actor.NewActorSystem()
	// if err := actorSystem.Start(); err != nil {
	//     zLog.Error("Failed to start actor system", zap.Error(err))
	//     return nil, err
	// }
	// if err := actorSystem.RegisterActor(playerActor); err != nil {
	//     zLog.Error("Failed to register player actor", zap.Int64("playerId", playerId), zap.Error(err))
	//     return nil, err
	// }

	// 存储玩家Actor信息
	ps.playerActors.Store(playerId, playerActor)
	ps.sessionPlayer.Store(int64(session.GetSid()), playerId)

	zLog.Info("Created new player actor", zap.Int64("playerId", playerId), zap.String("name", name))
	return playerActor, nil
}

func (ps *PlayerService) Serve() {
	ps.SetState(zService.ServiceStateRunning)
	// 玩家服务不需要持续运行的协程
}

func (ps *PlayerService) CreatePlayer(session *zNet.TcpServerSession, playerId int64, name string) (*Player, error) {
	// 检查玩家是否已存在
	if _, exists := ps.players.Get(playerId); exists {
		return nil, nil // 玩家已存在
	}

	// 创建新玩家
	player := NewPlayer(playerId, name, session)

	// 存储玩家信息
	ps.players.Store(playerId, player)
	ps.sessionPlayer.Store(int64(session.GetSid()), playerId)

	zLog.Info("Created new player", zap.Int64("playerId", playerId), zap.String("name", name))
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

// GetPlayerActor 获取玩家Actor
func (ps *PlayerService) GetPlayerActor(playerId int64) *PlayerActor {
	if playerActor, exists := ps.playerActors.Get(playerId); exists {
		return playerActor.(*PlayerActor)
	}
	return nil
}

// GetPlayerActorBySession 根据会话获取玩家Actor
func (ps *PlayerService) GetPlayerActorBySession(sessionId int64) *PlayerActor {
	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		return ps.GetPlayerActor(playerId.(int64))
	}
	return nil
}

func (ps *PlayerService) RemovePlayer(playerId int64) {
	// 检查并移除传统Player
	if player, exists := ps.players.Get(playerId); exists {
		ps.sessionPlayer.Delete(int64(player.(*Player).Session.GetSid()))
		ps.players.Delete(playerId)
		zLog.Info("Removed player", zap.Int64("playerId", playerId))
		return
	}

	// 检查并移除PlayerActor
	if playerActor, exists := ps.playerActors.Get(playerId); exists {
		// 从全局Actor系统中注销
		if err := zActor.GetGlobalActorSystem().UnregisterActor(playerId); err != nil {
			zLog.Error("Failed to unregister player actor", zap.Int64("playerId", playerId), zap.Error(err))
		}

		// 更新会话映射
		ps.sessionPlayer.Delete(int64(playerActor.(*PlayerActor).GetSession().GetSid()))
		ps.playerActors.Delete(playerId)
		zLog.Info("Removed player actor", zap.Int64("playerId", playerId))
	}
}

func (ps *PlayerService) OnSessionClose(sessionId int64) {
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
