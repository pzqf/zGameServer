package dao

import (
	"database/sql"
	"fmt"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/db/connector"
	"github.com/pzqf/zGameServer/db/models"
	"go.uber.org/zap"
)

// AccountDAO 账号数据访问对象
type AccountDAO struct {
	connector *connector.DBConnector
	logger    *zap.Logger
}

// NewAccountDAO 创建账号DAO实例
func NewAccountDAO(dbConnector *connector.DBConnector) *AccountDAO {
	return &AccountDAO{
		connector: dbConnector,
		logger:    zLog.GetLogger(),
	}
}

// GetAccountByID 根据ID获取账号信息
func (dao *AccountDAO) GetAccountByID(accountID int64, callback func(*models.Account, error)) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE account_id = ?", models.Account{}.TableName())

	dao.connector.Query(query, []interface{}{accountID}, func(rows *sql.Rows, err error) {
		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer rows.Close()

		var account models.Account
		if rows.Next() {
			if err := rows.Scan(
				&account.AccountID,
				&account.AccountName,
				&account.Password,
				&account.Status,
				&account.CreatedAt,
				&account.LastLoginAt,
			); err != nil {
				if callback != nil {
					callback(nil, err)
				}
				return
			}

			if callback != nil {
				callback(&account, nil)
			}
		} else {
			if callback != nil {
				callback(nil, nil) // 未找到账号
			}
		}
	})
}

// GetAccountByName 根据名称获取账号信息
func (dao *AccountDAO) GetAccountByName(accountName string, callback func(*models.Account, error)) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE account_name = ?", models.Account{}.TableName())

	dao.connector.Query(query, []interface{}{accountName}, func(rows *sql.Rows, err error) {
		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer rows.Close()

		var account models.Account
		if rows.Next() {
			if err := rows.Scan(
				&account.AccountID,
				&account.AccountName,
				&account.Password,
				&account.Status,
				&account.CreatedAt,
				&account.LastLoginAt,
			); err != nil {
				if callback != nil {
					callback(nil, err)
				}
				return
			}

			if callback != nil {
				callback(&account, nil)
			}
		} else {
			if callback != nil {
				callback(nil, nil) // 未找到账号
			}
		}
	})
}

// CreateAccount 创建账号
func (dao *AccountDAO) CreateAccount(account *models.Account, callback func(int64, error)) {
	query := fmt.Sprintf("INSERT INTO %s (account_id, account_name, password, status, created_at, last_login_at) VALUES (?, ?, ?, ?, ?, ?)", models.Account{}.TableName())

	args := []interface{}{
		account.AccountID,
		account.AccountName,
		account.Password,
		account.Status,
		account.CreatedAt,
		account.LastLoginAt,
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

// UpdateAccount 更新账号信息
func (dao *AccountDAO) UpdateAccount(account *models.Account, callback func(bool, error)) {
	query := fmt.Sprintf("UPDATE %s SET account_name = ?, password = ?, status = ?, last_login_at = ? WHERE account_id = ?", models.Account{}.TableName())

	args := []interface{}{
		account.AccountName,
		account.Password,
		account.Status,
		account.LastLoginAt,
		account.AccountID,
	}

	dao.connector.Execute(query, args, func(result sql.Result, err error) {
		if err != nil {
			if callback != nil {
				callback(false, err)
			}
			return
		}

		rowsAffected, err := result.RowsAffected()
		if callback != nil {
			callback(rowsAffected > 0, err)
		}
	})
}

// DeleteAccount 删除账号
func (dao *AccountDAO) DeleteAccount(accountID int64, callback func(bool, error)) {
	query := fmt.Sprintf("DELETE FROM %s WHERE account_id = ?", models.Account{}.TableName())

	dao.connector.Execute(query, []interface{}{accountID}, func(result sql.Result, err error) {
		if err != nil {
			if callback != nil {
				callback(false, err)
			}
			return
		}

		rowsAffected, err := result.RowsAffected()
		if callback != nil {
			callback(rowsAffected > 0, err)
		}
	})
}

// UpdateLastLoginAt 更新最后登录时间
func (dao *AccountDAO) UpdateLastLoginAt(accountID int64, lastLoginAt string, callback func(bool, error)) {
	query := fmt.Sprintf("UPDATE %s SET last_login_at = ? WHERE account_id = ?", models.Account{}.TableName())

	dao.connector.Execute(query, []interface{}{lastLoginAt, accountID}, func(result sql.Result, err error) {
		if err != nil {
			if callback != nil {
				callback(false, err)
			}
			return
		}

		rowsAffected, err := result.RowsAffected()
		if callback != nil {
			callback(rowsAffected > 0, err)
		}
	})
}