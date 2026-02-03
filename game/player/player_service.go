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
	playerActors  *zMap.TypedShardedMap[common.PlayerIdType, *PlayerActor]       // key: PlayerIdType, value: *PlayerActor
	sessionPlayer *zMap.TypedShardedMap[zNet.SessionIdType, common.PlayerIdType] // key: SessionIdType, value: PlayerIdType
	playerCount   int64
	metrics       *PlayerMetrics
}

type PlayerMetrics struct {
	PlayerCount int64
	OnlineTime  map[common.PlayerIdType]time.Time
}

func NewPlayerService() *PlayerService {
	ps := &PlayerService{
		BaseService:   *zService.NewBaseService(common.ServiceIdPlayer),
		playerActors:  zMap.NewTypedShardedMap32[common.PlayerIdType, *PlayerActor](),
		sessionPlayer: zMap.NewTypedShardedMap32[zNet.SessionIdType, common.PlayerIdType](),
		metrics: &PlayerMetrics{
			OnlineTime: make(map[common.PlayerIdType]time.Time),
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

	ps.playerActors.Range(func(key common.PlayerIdType, value *PlayerActor) bool {
		playerActor := value
		playerActor.Stop()
		delete(ps.metrics.OnlineTime, key)
		ps.playerActors.Delete(key)
		return true
	})

	ps.sessionPlayer.Clear()
	ps.SetState(zService.ServiceStateStopped)
	return nil
}

func (ps *PlayerService) CreatePlayerActor(session *zNet.TcpServerSession, playerId common.PlayerIdType, name string) (*PlayerActor, error) {
	if ps.getPlayerCount() >= MaxPlayerActors {
		return nil, errTooManyPlayers
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, exists := ps.playerActors.Load(playerId); exists {
		return nil, nil
	}

	playerActor := NewPlayerActor(playerId, name, session)

	ps.playerActors.Store(playerId, playerActor)
	ps.sessionPlayer.Store(session.GetSid(), playerId)
	ps.metrics.OnlineTime[playerId] = time.Now()
	ps.playerCount++

	zLog.Info("Created new player actor",
		zap.Int64("playerId", int64(playerId)),
		zap.String("name", name),
		zap.Int64("totalPlayers", ps.playerCount))

	return playerActor, nil
}

func (ps *PlayerService) Serve() {
	ps.SetState(zService.ServiceStateRunning)
}

func (ps *PlayerService) GetPlayer(playerId common.PlayerIdType) *Player {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		return playerActor.Player
	}
	return nil
}

func (ps *PlayerService) GetPlayerBySession(sessionId zNet.SessionIdType) *Player {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerId, exists := ps.sessionPlayer.Load(sessionId); exists {
		return ps.getPlayerUnsafe(playerId)
	}
	return nil
}

func (ps *PlayerService) GetPlayerActor(playerId common.PlayerIdType) *PlayerActor {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		return playerActor
	}
	return nil
}

func (ps *PlayerService) GetPlayerActorBySession(sessionId zNet.SessionIdType) *PlayerActor {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerId, exists := ps.sessionPlayer.Load(sessionId); exists {
		return ps.getPlayerActorUnsafe(playerId)
	}
	return nil
}

func (ps *PlayerService) RemovePlayer(playerId common.PlayerIdType) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		playerActor.Stop()
		ps.playerActors.Delete(playerId)
		delete(ps.metrics.OnlineTime, playerId)
		ps.playerCount--

		if player := playerActor.Player; player != nil {
			if session := player.GetSession(); session != nil {
				ps.sessionPlayer.Delete(session.GetSid())
			}
		}

		zLog.Info("Removed player actor",
			zap.Int64("playerId", int64(playerId)),
			zap.Int64("totalPlayers", ps.playerCount))
	}
}

func (ps *PlayerService) OnSessionClose(sessionId zNet.SessionIdType) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if playerId, exists := ps.sessionPlayer.Load(sessionId); exists {
		if playerActor, exists := ps.playerActors.Load(playerId); exists {
			playerActor.Stop()
			ps.playerActors.Delete(playerId)
			ps.sessionPlayer.Delete(sessionId)
			delete(ps.metrics.OnlineTime, playerId)
			ps.playerCount--

			zLog.Info("Session closed, removed player actor",
				zap.Uint64("sessionId", uint64(sessionId)),
				zap.Int64("playerId", int64(playerId)))
		}
	}
}

func (ps *PlayerService) getPlayerCount() int64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.playerCount
}

func (ps *PlayerService) getPlayerUnsafe(playerId common.PlayerIdType) *Player {
	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		return playerActor.Player
	}
	return nil
}

func (ps *PlayerService) getPlayerActorUnsafe(playerId common.PlayerIdType) *PlayerActor {
	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		return playerActor
	}
	return nil
}
