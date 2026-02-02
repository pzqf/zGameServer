package repository

import (
	"fmt"
	"time"

	"github.com/pzqf/zGameServer/db/dao"
	"github.com/pzqf/zGameServer/db/models"
	"github.com/pzqf/zUtil/zCache"
)

// CharacterRepositoryImpl 角色数据仓库实现
type CharacterRepositoryImpl struct {
	characterDAO *dao.CharacterDAO
	cache        zCache.Cache
}

// NewCharacterRepository 创建角色数据仓库实例
func NewCharacterRepository(characterDAO *dao.CharacterDAO) *CharacterRepositoryImpl {
	return &CharacterRepositoryImpl{
		characterDAO: characterDAO,
		cache:        zCache.NewLRUCache(1000, 5*time.Minute), // 1000容量，5分钟过期
	}
}

// GetByIDAsync 根据ID异步获取角色
func (r *CharacterRepositoryImpl) GetByIDAsync(characterID int64, callback func(*models.Character, error)) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("character:%d", characterID)
	if cached, err := r.cache.Get(cacheKey); err == nil {
		if character, ok := cached.(*models.Character); ok {
			if callback != nil {
				callback(character, nil)
			}
			return
		}
	}

	// 缓存未命中，从数据库获取
	r.characterDAO.GetCharacterByID(characterID, func(c *models.Character, err error) {
		if err == nil && c != nil {
			_ = r.cache.Set(cacheKey, c, 5*time.Minute)
		}

		if callback != nil {
			callback(c, err)
		}
	})
}

// GetByAccountIDAsync 根据账号ID异步获取角色列表
func (r *CharacterRepositoryImpl) GetByAccountIDAsync(accountID int64, callback func([]*models.Character, error)) {
	// 直接从数据库获取（不缓存列表数据）
	r.characterDAO.GetCharactersByAccountID(accountID, func(characters []*models.Character, err error) {
		if callback != nil {
			callback(characters, err)
		}
	})
}

// CreateAsync 异步创建角色
func (r *CharacterRepositoryImpl) CreateAsync(character *models.Character, callback func(int64, error)) {
	// 异步执行创建
	r.characterDAO.CreateCharacter(character, func(id int64, err error) {
		if err == nil && id > 0 {
			cacheKey := fmt.Sprintf("character:%d", id)
			_ = r.cache.Set(cacheKey, character, 5*time.Minute)
		}

		if callback != nil {
			callback(id, err)
		}
	})
}

// UpdateAsync 异步更新角色
func (r *CharacterRepositoryImpl) UpdateAsync(character *models.Character, callback func(bool, error)) {
	// 异步执行更新
	r.characterDAO.UpdateCharacter(character, func(updated bool, err error) {
		if err == nil && updated {
			cacheKey := fmt.Sprintf("character:%d", character.CharID)
			_ = r.cache.Set(cacheKey, character, 5*time.Minute)
		}

		if callback != nil {
			callback(updated, err)
		}
	})
}

// DeleteAsync 异步删除角色
func (r *CharacterRepositoryImpl) DeleteAsync(characterID int64, callback func(bool, error)) {
	// 异步执行删除
	r.characterDAO.DeleteCharacter(characterID, func(deleted bool, err error) {
		if err == nil && deleted {
			cacheKey := fmt.Sprintf("character:%d", characterID)
			_ = r.cache.Delete(cacheKey)
		}

		if callback != nil {
			callback(deleted, err)
		}
	})
}

// GetByID 根据ID获取角色（同步方法）
func (r *CharacterRepositoryImpl) GetByID(characterID int64) (*models.Character, error) {
	var result *models.Character
	var resultErr error
	ch := make(chan struct{})
	r.GetByIDAsync(characterID, func(c *models.Character, err error) {
		result = c
		resultErr = err
		close(ch)
	})
	<-ch
	return result, resultErr
}

// GetByAccountID 根据账号ID获取角色列表（同步方法）
func (r *CharacterRepositoryImpl) GetByAccountID(accountID int64) ([]*models.Character, error) {
	var result []*models.Character
	var resultErr error
	ch := make(chan struct{})
	r.GetByAccountIDAsync(accountID, func(characters []*models.Character, err error) {
		result = characters
		resultErr = err
		close(ch)
	})
	<-ch
	return result, resultErr
}

// Create 创建角色（同步方法）
func (r *CharacterRepositoryImpl) Create(character *models.Character) (int64, error) {
	var result int64
	var resultErr error
	ch := make(chan struct{})
	r.CreateAsync(character, func(id int64, err error) {
		result = id
		resultErr = err
		close(ch)
	})
	<-ch
	return result, resultErr
}

// Update 更新角色（同步方法）
func (r *CharacterRepositoryImpl) Update(character *models.Character) (bool, error) {
	var result bool
	var resultErr error
	ch := make(chan struct{})
	r.UpdateAsync(character, func(updated bool, err error) {
		result = updated
		resultErr = err
		close(ch)
	})
	<-ch
	return result, resultErr
}

// Delete 删除角色（同步方法）
func (r *CharacterRepositoryImpl) Delete(characterID int64) (bool, error) {
	var result bool
	var resultErr error
	ch := make(chan struct{})
	r.DeleteAsync(characterID, func(deleted bool, err error) {
		result = deleted
		resultErr = err
		close(ch)
	})
	<-ch
	return result, resultErr
}
