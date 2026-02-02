package player

import (
	"time"

	"github.com/pzqf/zEngine/zActor"
	"github.com/pzqf/zEngine/zEvent"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/game/common"
	"go.uber.org/zap"
)

const (
	PlayerUpdateInterval   = 10 * time.Millisecond
	PlayerActorMsgChanSize = 100
)

type PlayerActor struct {
	*zActor.BaseActor
	Player *Player
	Update chan struct{}
}

func NewPlayerActor(playerID int64, name string, session *zNet.TcpServerSession) *PlayerActor {
	baseActor := zActor.NewBaseActor(playerID, PlayerActorMsgChanSize)
	player := NewPlayer(playerID, name, session)

	actor := &PlayerActor{
		BaseActor: baseActor,
		Player:    player,
		Update:    make(chan struct{}, 1),
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
			pa.Player.SetPosition(common.Vector3{X: typedMsg.X, Y: typedMsg.Y, Z: typedMsg.Z})
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
				zap.Int64("playerId", pa.Player.GetPlayerId()))
		}
	}
}

func (pa *PlayerActor) Run() {
	ticker := time.NewTicker(PlayerUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case msg := <-pa.ActorMsgChan:
			pa.ProcessMessage(msg)
		case <-ticker.C:
			pa.Player.Update(float64(PlayerUpdateInterval.Milliseconds()))
		}
	}
}
