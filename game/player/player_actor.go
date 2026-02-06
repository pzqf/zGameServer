package player

import (
	"sync/atomic"
	"time"

	"github.com/pzqf/zEngine/zActor"
	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"go.uber.org/zap"
)

const (
	PlayerUpdateInterval   = 10 * time.Millisecond
	PlayerActorMsgChanSize = 100
)

type PlayerActor struct {
	*zActor.BaseActor
	Player  *Player
	stopCh  chan struct{}
	running atomic.Bool
}

func NewPlayerActor(playerID common.PlayerIdType, name string, session *zNet.TcpServerSession) *PlayerActor {
	baseActor := zActor.NewBaseActor(int64(playerID), PlayerActorMsgChanSize)
	player := NewPlayer(playerID, name, session)

	actor := &PlayerActor{
		BaseActor: baseActor,
		Player:    player,
		stopCh:    make(chan struct{}),
	}

	return actor
}

func (pa *PlayerActor) ProcessMessage(msg zActor.ActorMessage) {
	switch typedMsg := msg.(type) {
	case *PlayerActorAttackMessage:
		if pa.Player != nil {
			pa.Player.PublishEvent(zEvent.NewEvent(5, pa.Player, map[string]interface{}{
				"playerId": pa.Player.GetPlayerId(),
				"targetId": typedMsg.TargetID,
			}))
		}
	case *PlayerActorMoveMessage:
		if pa.Player != nil {
			pa.Player.SetPosition(gamecommon.Vector3{X: typedMsg.X, Y: typedMsg.Y, Z: typedMsg.Z})
		}
	case *PlayerActorAddExpMessage:
		if pa.Player != nil {
			pa.Player.AddExp(typedMsg.Exp)
		}
	case *PlayerActorAddGoldMessage:
		if pa.Player != nil {
			pa.Player.AddGold(typedMsg.Gold)
		}
	case *PlayerActorNetworkMessage:
		if pa.Player != nil {
			zLog.Info("Player received network packet",
				zap.Int64("playerId", int64(pa.Player.GetPlayerId())))
		}
	}
}

func (pa *PlayerActor) Run() {
	if !pa.running.CompareAndSwap(false, true) {
		return
	}

	ticker := time.NewTicker(PlayerUpdateInterval)
	defer ticker.Stop()
	defer pa.running.Store(false)

	for {
		select {
		case msg, ok := <-pa.ActorMsgChan:
			if !ok {
				zLog.Debug("PlayerActor message channel closed, exiting",
					zap.Int64("actorId", pa.ID()))
				return
			}
			pa.ProcessMessage(msg)
		case <-ticker.C:
			if pa.Player != nil {
				pa.Player.Update(float64(PlayerUpdateInterval.Milliseconds()))
			}
		case <-pa.stopCh:
			zLog.Debug("PlayerActor stop signal received, exiting",
				zap.Int64("actorId", pa.ID()))
			return
		}
	}
}

// Stop 重写Stop()方法，确保完整的资源清理
func (pa *PlayerActor) Stop() error {
	select {
	case <-pa.stopCh:
	default:
		close(pa.stopCh)
	}

	if pa.Player != nil {
		pa.Player.Logout()

		if session := pa.Player.GetSession(); session != nil {
			session.Close()
		}
	}

	err := pa.BaseActor.Stop()
	if err != nil {
		zLog.Error("Failed to stop base actor", zap.Int64("playerId", pa.ID()), zap.Error(err))
		return err
	}

	if pa.Player != nil {
		zLog.Info("Player actor stopped",
			zap.Int64("playerId", int64(pa.Player.GetPlayerId())),
			zap.Int64("actorId", pa.ID()))
	} else {
		zLog.Info("Player actor stopped",
			zap.Int64("actorId", pa.ID()))
	}

	return nil
}
