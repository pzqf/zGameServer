package player

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

const (
	PlayerServiceName = "PlayerService"
	MaxPlayerActors   = 10000
)

type PlayerService struct {
	zService.BaseService
	mu            sync.RWMutex
	playerActors  *zMap.Map // key: int64(playerId), value: *PlayerActor
	sessionPlayer *zMap.Map // key: int64(sessionId), value: int64(playerId)
	playerCount   int64
	metrics       *PlayerMetrics
}

type PlayerMetrics struct {
	PlayerCount int64
	OnlineTime  map[int64]time.Time
}

func NewPlayerService() *PlayerService {
	ps := &PlayerService{
		BaseService:   *zService.NewBaseService(common.ServiceIdPlayer),
		playerActors:  zMap.NewMap(),
		sessionPlayer: zMap.NewMap(),
		metrics: &PlayerMetrics{
			OnlineTime: make(map[int64]time.Time),
		},
	}
	return ps
}

func (ps *PlayerService) Init() error {
	ps.SetState(zService.ServiceStateInit)
	zLog.Info("Initializing player service...", zap.String("serviceId", ps.ServiceId()))
	return nil
}

func (ps *PlayerService) Close() error {
	ps.SetState(zService.ServiceStateStopping)
	zLog.Info("Closing player service...", zap.String("serviceId", ps.ServiceId()))

	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.playerActors.Range(func(key, value interface{}) bool {
		playerActor := value.(*PlayerActor)
		playerActor.Stop()
		delete(ps.metrics.OnlineTime, key.(int64))
		ps.playerActors.Delete(key)
		return true
	})

	ps.sessionPlayer.Clear()
	ps.SetState(zService.ServiceStateStopped)
	return nil
}

func (ps *PlayerService) CreatePlayerActor(session *zNet.TcpServerSession, playerId int64, name string) (*PlayerActor, error) {
	if ps.getPlayerCount() >= MaxPlayerActors {
		return nil, errTooManyPlayers
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, exists := ps.playerActors.Get(playerId); exists {
		return nil, nil
	}

	playerActor := NewPlayerActor(playerId, name, session)

	ps.playerActors.Store(playerId, playerActor)
	ps.sessionPlayer.Store(int64(session.GetSid()), playerId)
	ps.metrics.OnlineTime[playerId] = time.Now()
	ps.playerCount++

	zLog.Info("Created new player actor",
		zap.Int64("playerId", playerId),
		zap.String("name", name),
		zap.Int64("totalPlayers", ps.playerCount))

	return playerActor, nil
}

func (ps *PlayerService) Serve() {
	ps.SetState(zService.ServiceStateRunning)
}

func (ps *PlayerService) GetPlayer(playerId int64) *Player {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerActor, exists := ps.playerActors.Get(playerId); exists {
		return playerActor.(*PlayerActor).Player
	}
	return nil
}

func (ps *PlayerService) GetPlayerBySession(sessionId int64) *Player {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		return ps.getPlayerUnsafe(playerId.(int64))
	}
	return nil
}

func (ps *PlayerService) GetPlayerActor(playerId int64) *PlayerActor {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerActor, exists := ps.playerActors.Get(playerId); exists {
		return playerActor.(*PlayerActor)
	}
	return nil
}

func (ps *PlayerService) GetPlayerActorBySession(sessionId int64) *PlayerActor {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		return ps.getPlayerActorUnsafe(playerId.(int64))
	}
	return nil
}

func (ps *PlayerService) RemovePlayer(playerId int64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if playerActor, exists := ps.playerActors.Get(playerId); exists {
		playerActor.(*PlayerActor).Stop()
		ps.playerActors.Delete(playerId)
		delete(ps.metrics.OnlineTime, playerId)
		ps.playerCount--

		if player := playerActor.(*PlayerActor).Player; player != nil {
			if session := player.GetSession(); session != nil {
				ps.sessionPlayer.Delete(int64(session.GetSid()))
			}
		}

		zLog.Info("Removed player actor",
			zap.Int64("playerId", playerId),
			zap.Int64("totalPlayers", ps.playerCount))
	}
}

func (ps *PlayerService) OnSessionClose(sessionId int64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if playerId, exists := ps.sessionPlayer.Get(sessionId); exists {
		if playerActor, exists := ps.playerActors.Get(playerId); exists {
			playerActor.(*PlayerActor).Stop()
			ps.playerActors.Delete(playerId)
			ps.sessionPlayer.Delete(sessionId)
			delete(ps.metrics.OnlineTime, playerId.(int64))
			ps.playerCount--

			zLog.Info("Session closed, removed player actor",
				zap.Int64("sessionId", sessionId),
				zap.Int64("playerId", playerId.(int64)))
		}
	}
}

func (ps *PlayerService) getPlayerCount() int64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.playerCount
}

func (ps *PlayerService) getPlayerUnsafe(playerId int64) *Player {
	if playerActor, exists := ps.playerActors.Get(playerId); exists {
		return playerActor.(*PlayerActor).Player
	}
	return nil
}

func (ps *PlayerService) getPlayerActorUnsafe(playerId int64) *PlayerActor {
	if playerActor, exists := ps.playerActors.Get(playerId); exists {
		return playerActor.(*PlayerActor)
	}
	return nil
}
