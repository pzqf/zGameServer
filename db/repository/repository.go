package repository

import (
	"github.com/pzqf/zGameServer/db/models"
)

// AccountRepository 账号数据仓库接口
type AccountRepository interface {
	// GetByID 根据ID获取账号
	GetByID(accountID int64) (*models.Account, error)
	// GetByName 根据名称获取账号
	GetByName(accountName string) (*models.Account, error)
	// Create 创建账号
	Create(account *models.Account) (int64, error)
	// Update 更新账号
	Update(account *models.Account) (bool, error)
	// Delete 删除账号
	Delete(accountID int64) (bool, error)
	// UpdateLastLoginAt 更新最后登录时间
	UpdateLastLoginAt(accountID int64, lastLoginAt string) (bool, error)
}

// CharacterRepository 角色数据仓库接口
type CharacterRepository interface {
	// GetByID 根据ID获取角色
	GetByID(characterID int64) (*models.Character, error)
	// GetByAccountID 根据账号ID获取角色列表
	GetByAccountID(accountID int64) ([]*models.Character, error)
	// Create 创建角色
	Create(character *models.Character) (int64, error)
	// Update 更新角色
	Update(character *models.Character) (bool, error)
	// Delete 删除角色
	Delete(characterID int64) (bool, error)
}

// LoginLogRepository 登录日志数据仓库接口
type LoginLogRepository interface {
	// Create 创建登录日志
	Create(loginLog *models.LoginLog) (int64, error)
	// GetByAccountID 根据账号ID获取登录日志列表
	GetByAccountID(accountID int64, limit int) ([]*models.LoginLog, error)
}
