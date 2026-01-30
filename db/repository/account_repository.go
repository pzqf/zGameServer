package repository

import (
	"fmt"
	"time"

	"github.com/pzqf/zGameServer/db/dao"
	"github.com/pzqf/zGameServer/db/models"
	"github.com/pzqf/zUtil/zCache"
)

// AccountRepositoryImpl 账号数据仓库实现
type AccountRepositoryImpl struct {
	accountDAO *dao.AccountDAO
	cache      zCache.Cache
}

// NewAccountRepository 创建账号数据仓库实例
func NewAccountRepository(accountDAO *dao.AccountDAO) *AccountRepositoryImpl {
	return &AccountRepositoryImpl{
		accountDAO: accountDAO,
		cache:      zCache.NewLRUCache(1000, 5*time.Minute), // 1000容量，5分钟过期
	}
}

// GetByID 根据ID获取账号
func (r *AccountRepositoryImpl) GetByID(accountID int64) (*models.Account, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("account:%d", accountID)
	if cached, err := r.cache.Get(cacheKey); err == nil {
		if account, ok := cached.(*models.Account); ok {
			return account, nil
		}
	}

	// 缓存未命中，从数据库获取
	var account *models.Account
	var daoErr error

	// 使用通道同步获取结果
	ch := make(chan struct{})
	r.accountDAO.GetAccountByID(accountID, func(a *models.Account, err error) {
		account = a
		daoErr = err
		close(ch)
	})

	<-ch

	// 缓存结果
	if daoErr == nil && account != nil {
		_ = r.cache.Set(cacheKey, account, 5*time.Minute)
	}

	return account, daoErr
}

// GetByName 根据名称获取账号
func (r *AccountRepositoryImpl) GetByName(accountName string) (*models.Account, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("account:name:%s", accountName)
	if cached, err := r.cache.Get(cacheKey); err == nil {
		if account, ok := cached.(*models.Account); ok {
			return account, nil
		}
	}

	// 缓存未命中，从数据库获取
	var account *models.Account
	var daoErr error

	// 使用通道同步获取结果
	ch := make(chan struct{})
	r.accountDAO.GetAccountByName(accountName, func(a *models.Account, err error) {
		account = a
		daoErr = err
		close(ch)
	})

	<-ch

	// 缓存结果
	if daoErr == nil && account != nil {
		_ = r.cache.Set(cacheKey, account, 5*time.Minute)
		// 同时缓存ID索引
		idCacheKey := fmt.Sprintf("account:%d", account.AccountID)
		_ = r.cache.Set(idCacheKey, account, 5*time.Minute)
	}

	return account, daoErr
}

// Create 创建账号
func (r *AccountRepositoryImpl) Create(account *models.Account) (int64, error) {
	// 使用通道同步获取结果
	var result int64
	var daoErr error

	ch := make(chan struct{})
	r.accountDAO.CreateAccount(account, func(id int64, err error) {
		result = id
		daoErr = err
		close(ch)
	})

	<-ch

	// 缓存结果
	if daoErr == nil && result > 0 {
		cacheKey := fmt.Sprintf("account:%d", result)
		_ = r.cache.Set(cacheKey, account, 5*time.Minute)
		nameCacheKey := fmt.Sprintf("account:name:%s", account.AccountName)
		_ = r.cache.Set(nameCacheKey, account, 5*time.Minute)
	}

	return result, daoErr
}

// Update 更新账号
func (r *AccountRepositoryImpl) Update(account *models.Account) (bool, error) {
	// 使用通道同步获取结果
	var result bool
	var daoErr error

	ch := make(chan struct{})
	r.accountDAO.UpdateAccount(account, func(updated bool, err error) {
		result = updated
		daoErr = err
		close(ch)
	})

	<-ch

	// 更新缓存
	if daoErr == nil && result {
		cacheKey := fmt.Sprintf("account:%d", account.AccountID)
		_ = r.cache.Set(cacheKey, account, 5*time.Minute)
		nameCacheKey := fmt.Sprintf("account:name:%s", account.AccountName)
		_ = r.cache.Set(nameCacheKey, account, 5*time.Minute)
	}

	return result, daoErr
}

// Delete 删除账号
func (r *AccountRepositoryImpl) Delete(accountID int64) (bool, error) {
	// 使用通道同步获取结果
	var result bool
	var daoErr error

	ch := make(chan struct{})
	r.accountDAO.DeleteAccount(accountID, func(deleted bool, err error) {
		result = deleted
		daoErr = err
		close(ch)
	})

	<-ch

	// 清理缓存
	if daoErr == nil && result {
		cacheKey := fmt.Sprintf("account:%d", accountID)
		_ = r.cache.Delete(cacheKey)
		// 注意：这里无法清理name索引的缓存，因为不知道账号名称
		// 可以考虑在账号对象中保存名称，或者使用单独的映射
	}

	return result, daoErr
}

// UpdateLastLoginAt 更新最后登录时间
func (r *AccountRepositoryImpl) UpdateLastLoginAt(accountID int64, lastLoginAt string) (bool, error) {
	// 使用通道同步获取结果
	var result bool
	var daoErr error

	ch := make(chan struct{})
	r.accountDAO.UpdateLastLoginAt(accountID, lastLoginAt, func(updated bool, err error) {
		result = updated
		daoErr = err
		close(ch)
	})

	<-ch

	// 清理缓存，下次获取时会重新加载
	if daoErr == nil && result {
		cacheKey := fmt.Sprintf("account:%d", accountID)
		_ = r.cache.Delete(cacheKey)
	}

	return result, daoErr
}
