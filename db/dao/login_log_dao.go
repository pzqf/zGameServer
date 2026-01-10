package dao

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/db/connector"
	"github.com/pzqf/zGameServer/db/models"
	"go.uber.org/zap"
)

// LoginLogDAO 角色登录/登出日志数据访问对象
type LoginLogDAO struct {
	connector *connector.DBConnector
	logger    *zap.Logger
}

// NewLoginLogDAO 创建角色登录/登出日志DAO实例
func NewLoginLogDAO(dbConnector *connector.DBConnector) *LoginLogDAO {
	return &LoginLogDAO{
		connector: dbConnector,
		logger:    zLog.GetLogger(),
	}
}

// GetLoginLogByCharID 根据角色ID获取登录/登出日志
func (dao *LoginLogDAO) GetLoginLogByCharID(charID int64, callback func(*models.LoginLog, error)) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE char_id = ?", models.LoginLog{}.TableName())

	dao.connector.Query(query, []interface{}{charID}, func(rows *sql.Rows, err error) {
		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer rows.Close()

		var loginLog models.LoginLog
		if rows.Next() {
			if err := rows.Scan(
				&loginLog.LogID,
				&loginLog.CharID,
				&loginLog.CharName,
				&loginLog.OpType,
				&loginLog.CreatedAt,
			); err != nil {
				if callback != nil {
					callback(nil, err)
				}
				return
			}

			if callback != nil {
				callback(&loginLog, nil)
			}
		} else {
			if callback != nil {
				callback(nil, nil) // 未找到日志
			}
		}
	})
}

// CreateLoginLog 创建角色登录/登出日志
func (dao *LoginLogDAO) CreateLoginLog(loginLog *models.LoginLog, callback func(int64, error)) {
	// 生成唯一的log_id
	loginLog.LogID = time.Now().UnixNano() / 1000000

	query := fmt.Sprintf("INSERT INTO %s (log_id, char_id, char_name, op_type, created_at) VALUES (?, ?, ?, ?, ?)", models.LoginLog{}.TableName())

	args := []interface{}{
		loginLog.LogID,
		loginLog.CharID,
		loginLog.CharName,
		loginLog.OpType,
		loginLog.CreatedAt,
	}

	dao.connector.Execute(query, args, func(result sql.Result, err error) {
		if err != nil {
			if callback != nil {
				callback(0, err)
			}
			return
		}

		id, err := result.LastInsertId()
		if callback != nil {
			callback(id, err)
		}
	})
}

// GetLoginLogsByTimeRange 根据时间范围获取登录/登出日志
func (dao *LoginLogDAO) GetLoginLogsByTimeRange(startTime, endTime string, callback func([]*models.LoginLog, error)) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE created_at BETWEEN ? AND ? ORDER BY created_at DESC", models.LoginLog{}.TableName())

	dao.connector.Query(query, []interface{}{startTime, endTime}, func(rows *sql.Rows, err error) {
		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer rows.Close()

		var loginLogs []*models.LoginLog
		for rows.Next() {
			var loginLog models.LoginLog
			if err := rows.Scan(
				&loginLog.CharID,
				&loginLog.CharName,
				&loginLog.OpType,
				&loginLog.CreatedAt,
			); err != nil {
				if callback != nil {
					callback(nil, err)
				}
				return
			}

			loginLogs = append(loginLogs, &loginLog)
		}

		if callback != nil {
			callback(loginLogs, nil)
		}
	})
}

// GetLoginLogsByOpType 根据操作类型获取登录/登出日志
func (dao *LoginLogDAO) GetLoginLogsByOpType(opType int, callback func([]*models.LoginLog, error)) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE op_type = ? ORDER BY created_at DESC", models.LoginLog{}.TableName())

	dao.connector.Query(query, []interface{}{opType}, func(rows *sql.Rows, err error) {
		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer rows.Close()

		var loginLogs []*models.LoginLog
		for rows.Next() {
			var loginLog models.LoginLog
			if err := rows.Scan(
				&loginLog.CharID,
				&loginLog.CharName,
				&loginLog.OpType,
				&loginLog.CreatedAt,
			); err != nil {
				if callback != nil {
					callback(nil, err)
				}
				return
			}

			loginLogs = append(loginLogs, &loginLog)
		}

		if callback != nil {
			callback(loginLogs, nil)
		}
	})
}
