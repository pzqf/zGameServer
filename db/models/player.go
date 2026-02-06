package models

import (
	"time"
)

type Player struct {
	PlayerID   int64     `db:"player_id"`
	PlayerName string    `db:"player_name"`
	AccountID  int64     `db:"account_id"`
	Sex        int       `db:"sex"`
	Age        int       `db:"age"`
	Level      int       `db:"level"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func (Player) TableName() string {
	return "`player`"
}
