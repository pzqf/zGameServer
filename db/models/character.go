package models

import (
	"time"
)

// Character 角色模型，映射character表
type Character struct {
	CharID    int64     `db:"char_id"`
	CharName  string    `db:"char_name"`
	AccountID int64     `db:"account_id"`
	Sex       int       `db:"sex"`
	Age       int       `db:"age"`
	Level     int       `db:"level"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TableName 返回表名
func (Character) TableName() string {
	return "`character`"
}
