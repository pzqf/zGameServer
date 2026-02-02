package player

import (
	"github.com/pzqf/zEngine/zActor"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/db/models"
)

// PlayerActorMessage 通用玩家消息
type PlayerActorMessage struct {
	zActor.BaseActorMessage
	Type string
	Data map[string]interface{}
}

func NewPlayerActorMessage(actorID int64, msgType string, data map[string]interface{}) *PlayerActorMessage {
	return &PlayerActorMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		Type:             msgType,
		Data:             data,
	}
}

// PlayerActorLoginMessage 登录消息
type PlayerActorLoginMessage struct {
	zActor.BaseActorMessage
	Account *models.Account
}

func NewPlayerActorLoginMessage(actorID int64, account *models.Account) *PlayerActorLoginMessage {
	return &PlayerActorLoginMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		Account:          account,
	}
}

// PlayerActorCharacterSelectMessage 角色选择消息
type PlayerActorCharacterSelectMessage struct {
	zActor.BaseActorMessage
	Character *models.Character
}

func NewPlayerActorCharacterSelectMessage(actorID int64, character *models.Character) *PlayerActorCharacterSelectMessage {
	return &PlayerActorCharacterSelectMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		Character:        character,
	}
}

// PlayerActorCharacterCreateMessage 角色创建消息
type PlayerActorCharacterCreateMessage struct {
	zActor.BaseActorMessage
	Account       *models.Account
	CharacterName string
	Sex           int
	Age           int
}

func NewPlayerActorCharacterCreateMessage(actorID int64, account *models.Account, characterName string, sex, age int) *PlayerActorCharacterCreateMessage {
	return &PlayerActorCharacterCreateMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		Account:          account,
		CharacterName:    characterName,
		Sex:              sex,
		Age:              age,
	}
}

// PlayerActorAttackMessage 攻击消息
type PlayerActorAttackMessage struct {
	zActor.BaseActorMessage
	TargetID int64
}

func NewPlayerActorAttackMessage(actorID, targetID int64) *PlayerActorAttackMessage {
	return &PlayerActorAttackMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		TargetID:         targetID,
	}
}

// PlayerActorMoveMessage 移动消息
type PlayerActorMoveMessage struct {
	zActor.BaseActorMessage
	X, Y, Z float32
}

func NewPlayerActorMoveMessage(actorID int64, x, y, z float32) *PlayerActorMoveMessage {
	return &PlayerActorMoveMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		X:                x,
		Y:                y,
		Z:                z,
	}
}

// PlayerActorUseItemMessage 使用物品消息
type PlayerActorUseItemMessage struct {
	zActor.BaseActorMessage
	ItemID int64
	Slot   int
}

func NewPlayerActorUseItemMessage(actorID int64, itemID int64, slot int) *PlayerActorUseItemMessage {
	return &PlayerActorUseItemMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		ItemID:           itemID,
		Slot:             slot,
	}
}

// PlayerActorUseSkillMessage 使用技能消息
type PlayerActorUseSkillMessage struct {
	zActor.BaseActorMessage
	SkillID  int64
	TargetID int64
}

func NewPlayerActorUseSkillMessage(actorID int64, skillID int64, targetID int64) *PlayerActorUseSkillMessage {
	return &PlayerActorUseSkillMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		SkillID:          skillID,
		TargetID:         targetID,
	}
}

// PlayerActorAddExpMessage 增加经验消息
type PlayerActorAddExpMessage struct {
	zActor.BaseActorMessage
	Exp int64
}

func NewPlayerActorAddExpMessage(actorID, exp int64) *PlayerActorAddExpMessage {
	return &PlayerActorAddExpMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		Exp:              exp,
	}
}

// PlayerActorAddGoldMessage 增加金币消息
type PlayerActorAddGoldMessage struct {
	zActor.BaseActorMessage
	Gold int64
}

func NewPlayerActorAddGoldMessage(actorID, gold int64) *PlayerActorAddGoldMessage {
	return &PlayerActorAddGoldMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		Gold:             gold,
	}
}

// PlayerActorUpdateMessage 更新消息
type PlayerActorUpdateMessage struct {
	zActor.BaseActorMessage
	DeltaTime float32
}

func NewPlayerActorUpdateMessage(actorID int64, deltaTime float32) *PlayerActorUpdateMessage {
	return &PlayerActorUpdateMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		DeltaTime:        deltaTime,
	}
}

// PlayerActorNetworkMessage 网络消息
type PlayerActorNetworkMessage struct {
	zActor.BaseActorMessage
	Packet *zNet.NetPacket
}

func NewPlayerActorNetworkMessage(actorID int64, packet *zNet.NetPacket) *PlayerActorNetworkMessage {
	return &PlayerActorNetworkMessage{
		BaseActorMessage: zActor.BaseActorMessage{ActorID: actorID},
		Packet:           packet,
	}
}
