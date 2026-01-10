package models

import (
	"time"
)

// LoginLog 角色登录/登出日志模型，映射login_log表
type LoginLog struct {
	LogID     int64     `db:"log_id"`
	CharID    int64     `db:"char_id"`
	CharName  string    `db:"char_name"`
	OpType    int       `db:"op_type"`
	CreatedAt time.Time `db:"created_at"`
}

// TableName 返回表名
func (LoginLog) TableName() string {
	return "login_log"
}
