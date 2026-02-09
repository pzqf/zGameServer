package models

import (
	"time"
)

type Player struct {
	PlayerID   int64     `db:"player_id" bson:"player_id"`
	PlayerName string    `db:"player_name" bson:"player_name"`
	AccountID  int64     `db:"account_id" bson:"account_id"`
	Sex        int       `db:"sex" bson:"sex"`
	Age        int       `db:"age" bson:"age"`
	Level      int       `db:"level" bson:"level"`
	CreatedAt  time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" bson:"updated_at"`
}

func (Player) TableName() string {
	return "players"
}
