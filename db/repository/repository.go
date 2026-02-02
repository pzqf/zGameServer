package repository

import (
	"github.com/pzqf/zGameServer/db/models"
)

// AccountRepository 账号数据仓库接口
type AccountRepository interface {
	// GetByIDAsync 根据ID异步获取账号
	GetByIDAsync(accountID int64, callback func(*models.Account, error))
	// GetByNameAsync 根据名称异步获取账号
	GetByNameAsync(accountName string, callback func(*models.Account, error))
	// CreateAsync 异步创建账号
	CreateAsync(account *models.Account, callback func(int64, error))
	// UpdateAsync 异步更新账号
	UpdateAsync(account *models.Account, callback func(bool, error))
	// DeleteAsync 异步删除账号
	DeleteAsync(accountID int64, callback func(bool, error))
	// UpdateLastLoginAtAsync 异步更新最后登录时间
	UpdateLastLoginAtAsync(accountID int64, lastLoginAt string, callback func(bool, error))

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
	// GetByIDAsync 根据ID异步获取角色
	GetByIDAsync(characterID int64, callback func(*models.Character, error))
	// GetByAccountIDAsync 根据账号ID异步获取角色列表
	GetByAccountIDAsync(accountID int64, callback func([]*models.Character, error))
	// CreateAsync 异步创建角色
	CreateAsync(character *models.Character, callback func(int64, error))
	// UpdateAsync 异步更新角色
	UpdateAsync(character *models.Character, callback func(bool, error))
	// DeleteAsync 异步删除角色
	DeleteAsync(characterID int64, callback func(bool, error))

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
	// CreateAsync 异步创建登录日志
	CreateAsync(loginLog *models.LoginLog, callback func(int64, error))
	// GetByAccountIDAsync 根据账号ID异步获取登录日志列表
	GetByAccountIDAsync(accountID int64, limit int, callback func([]*models.LoginLog, error))
}
