package player

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// PlayerService 玩家服务
// 管理所有在线玩家Actor，提供玩家创建、查找、移除等功能
type PlayerService struct {
	zService.BaseService
	mu            sync.RWMutex                                                         // 读写锁
	playerActors  *zMap.TypedShardedMap[common.PlayerIdType, *PlayerActor]             // 玩家Actor映射表（PlayerId -> PlayerActor）
	sessionPlayer *zMap.TypedShardedMap[zNet.SessionIdType, common.PlayerIdType]       // 会话玩家映射表（SessionId -> PlayerId）
	playerCount   int64                                                                // 当前在线玩家数
	metrics       *PlayerMetrics                                                       // 性能指标统计
}

// PlayerMetrics 玩家统计指标
// 用于监控和统计玩家在线情况
type PlayerMetrics struct {
	PlayerCount int64                              // 玩家数量
	OnlineTime  map[common.PlayerIdType]time.Time  // 玩家上线时间记录
}

// NewPlayerService 创建玩家服务
// 返回: 新创建的PlayerService实例
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

// Init 初始化玩家服务
// 返回: 初始化错误（如果有）
func (ps *PlayerService) Init() error {
	ps.SetState(zService.ServiceStateInit)
	zLog.Info("Initializing player service...", zap.String("serviceId", ps.ServiceId()))
	return nil
}

// Close 关闭玩家服务
// 停止所有玩家Actor，清理资源
// 返回: 关闭错误（如果有）
func (ps *PlayerService) Close() error {
	ps.SetState(zService.ServiceStateStopping)
	zLog.Info("Closing player service...", zap.String("serviceId", ps.ServiceId()))

	ps.mu.Lock()
	defer ps.mu.Unlock()

	// 停止所有玩家Actor
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

// CreatePlayerActor 创建玩家Actor
// 参数:
//   - session: 网络会话
//   - playerId: 玩家ID
//   - name: 玩家名称
//
// 返回:
//   - *PlayerActor: 新创建的玩家Actor
//   - error: 创建错误（玩家数已满或玩家已存在）
func (ps *PlayerService) CreatePlayerActor(session *zNet.TcpServerSession, playerId common.PlayerIdType, name string) (*PlayerActor, error) {
	// 检查服务器人数上限
	maxPlayers := config.GetServerConfig().MaxClientCount
	if ps.getPlayerCount() >= int64(maxPlayers) {
		return nil, errTooManyPlayers
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	// 检查玩家是否已在线
	if _, exists := ps.playerActors.Load(playerId); exists {
		return nil, nil
	}

	// 创建新的玩家Actor
	playerActor := NewPlayerActor(playerId, name, session)

	// 注册到映射表
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

// Serve 启动服务
// 将服务状态设置为Running
func (ps *PlayerService) Serve() {
	ps.SetState(zService.ServiceStateRunning)
}

// GetPlayer 获取玩家对象
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - *Player: 玩家对象（如果存在），否则返回nil
func (ps *PlayerService) GetPlayer(playerId common.PlayerIdType) *Player {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		return playerActor.Player
	}
	return nil
}

// GetPlayerBySession 通过会话ID获取玩家对象
// 参数:
//   - sessionId: 会话ID
//
// 返回:
//   - *Player: 玩家对象（如果存在），否则返回nil
func (ps *PlayerService) GetPlayerBySession(sessionId zNet.SessionIdType) *Player {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerId, exists := ps.sessionPlayer.Load(sessionId); exists {
		return ps.getPlayerUnsafe(playerId)
	}
	return nil
}

// GetPlayerActor 获取玩家Actor
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - *PlayerActor: 玩家Actor（如果存在），否则返回nil
func (ps *PlayerService) GetPlayerActor(playerId common.PlayerIdType) *PlayerActor {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		return playerActor
	}
	return nil
}

// GetPlayerActorBySession 通过会话ID获取玩家Actor
// 参数:
//   - sessionId: 会话ID
//
// 返回:
//   - *PlayerActor: 玩家Actor（如果存在），否则返回nil
func (ps *PlayerService) GetPlayerActorBySession(sessionId zNet.SessionIdType) *PlayerActor {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if playerId, exists := ps.sessionPlayer.Load(sessionId); exists {
		return ps.getPlayerActorUnsafe(playerId)
	}
	return nil
}

// RemovePlayer 移除玩家
// 停止玩家Actor，清理映射表
// 参数:
//   - playerId: 玩家ID
func (ps *PlayerService) RemovePlayer(playerId common.PlayerIdType) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		playerActor.Stop()
		ps.playerActors.Delete(playerId)
		delete(ps.metrics.OnlineTime, playerId)
		ps.playerCount--

		// 清理会话映射
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

// OnSessionClose 会话关闭处理
// 当客户端断开连接时调用，清理相关玩家数据
// 参数:
//   - sessionId: 关闭的会话ID
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

// getPlayerCount 获取当前在线玩家数量
// 返回: 在线玩家数
func (ps *PlayerService) getPlayerCount() int64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.playerCount
}

// getPlayerUnsafe 不安全方式获取玩家对象
// 注意: 调用前必须持有锁
func (ps *PlayerService) getPlayerUnsafe(playerId common.PlayerIdType) *Player {
	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		return playerActor.Player
	}
	return nil
}

// getPlayerActorUnsafe 不安全方式获取玩家Actor
// 注意: 调用前必须持有锁
func (ps *PlayerService) getPlayerActorUnsafe(playerId common.PlayerIdType) *PlayerActor {
	if playerActor, exists := ps.playerActors.Load(playerId); exists {
		return playerActor
	}
	return nil
}
