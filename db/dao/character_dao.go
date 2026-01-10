package dao

import (
	"database/sql"
	"fmt"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/db/connector"
	"github.com/pzqf/zGameServer/db/models"
	"go.uber.org/zap"
)

// CharacterDAO 角色数据访问对象
type CharacterDAO struct {
	connector *connector.DBConnector
	logger    *zap.Logger
}

// NewCharacterDAO 创建角色DAO实例
func NewCharacterDAO(dbConnector *connector.DBConnector) *CharacterDAO {
	return &CharacterDAO{
		connector: dbConnector,
		logger:    zLog.GetLogger(),
	}
}

// GetCharacterByID 根据ID获取角色信息
func (dao *CharacterDAO) GetCharacterByID(charID int64, callback func(*models.Character, error)) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE char_id = ?", models.Character{}.TableName())

	dao.connector.Query(query, []interface{}{charID}, func(rows *sql.Rows, err error) {
		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer rows.Close()

		var char models.Character
		if rows.Next() {
			if err := rows.Scan(
				&char.CharID,
				&char.CharName,
				&char.AccountID,
				&char.Sex,
				&char.Age,
				&char.Level,
				&char.CreatedAt,
				&char.UpdatedAt,
			); err != nil {
				if callback != nil {
					callback(nil, err)
				}
				return
			}

			if callback != nil {
				callback(&char, nil)
			}
		} else {
			if callback != nil {
				callback(nil, nil) // 未找到角色
			}
		}
	})
}

// CreateCharacter 创建角色
func (dao *CharacterDAO) CreateCharacter(char *models.Character, callback func(int64, error)) {
	query := fmt.Sprintf("INSERT INTO %s (char_id, account_id, char_name, sex, age, level, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", models.Character{}.TableName())

	args := []interface{}{
		char.CharID,
		char.AccountID,
		char.CharName,
		char.Sex,
		char.Age,
		char.Level,
		char.CreatedAt,
		char.UpdatedAt,
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

// UpdateCharacter 更新角色信息
func (dao *CharacterDAO) UpdateCharacter(char *models.Character, callback func(bool, error)) {
	query := fmt.Sprintf("UPDATE %s SET char_name = ?, sex = ?, age = ?, level = ?, updated_at = ? WHERE char_id = ?", models.Character{}.TableName())

	args := []interface{}{
		char.CharName,
		char.Sex,
		char.Age,
		char.Level,
		char.UpdatedAt,
		char.CharID,
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

// DeleteCharacter 删除角色
func (dao *CharacterDAO) DeleteCharacter(charID int64, callback func(bool, error)) {
	query := fmt.Sprintf("DELETE FROM %s WHERE char_id = ?", models.Character{}.TableName())

	dao.connector.Execute(query, []interface{}{charID}, func(result sql.Result, err error) {
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

// GetAllCharacters 获取所有角色（用于管理面板或测试）
func (dao *CharacterDAO) GetAllCharacters(callback func([]*models.Character, error)) {
	query := fmt.Sprintf("SELECT * FROM %s", models.Character{}.TableName())

	dao.connector.Query(query, nil, func(rows *sql.Rows, err error) {
		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer rows.Close()

		var characters []*models.Character
		for rows.Next() {
			var char models.Character
			if err := rows.Scan(
				&char.CharID,
				&char.CharName,
				&char.AccountID,
				&char.Sex,
				&char.Age,
				&char.Level,
				&char.CreatedAt,
				&char.UpdatedAt,
			); err != nil {
				if callback != nil {
					callback(nil, err)
				}
				return
			}

			characters = append(characters, &char)
		}

		if callback != nil {
			callback(characters, nil)
		}
	})
}

// GetCharactersByAccountID 根据账号ID获取所有角色
func (dao *CharacterDAO) GetCharactersByAccountID(accountID int64, callback func([]*models.Character, error)) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE account_id = ?", models.Character{}.TableName())

	dao.connector.Query(query, []interface{}{accountID}, func(rows *sql.Rows, err error) {
		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer rows.Close()

		var characters []*models.Character
		for rows.Next() {
			var char models.Character
			if err := rows.Scan(
				&char.CharID,
				&char.CharName,
				&char.AccountID,
				&char.Sex,
				&char.Age,
				&char.Level,
				&char.CreatedAt,
				&char.UpdatedAt,
			); err != nil {
				if callback != nil {
					callback(nil, err)
				}
				return
			}

			characters = append(characters, &char)
		}

		if callback != nil {
			callback(characters, nil)
		}
	})
}

// GetCharacterByName 根据名称获取角色
func (dao *CharacterDAO) GetCharacterByName(name string, callback func(*models.Character, error)) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE char_name = ?", models.Character{}.TableName())

	dao.connector.Query(query, []interface{}{name}, func(rows *sql.Rows, err error) {
		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer rows.Close()

		var char models.Character
		if rows.Next() {
			if err := rows.Scan(
				&char.CharID,
				&char.CharName,
				&char.AccountID,
				&char.Sex,
				&char.Age,
				&char.Level,
				&char.CreatedAt,
				&char.UpdatedAt,
			); err != nil {
				if callback != nil {
					callback(nil, err)
				}
				return
			}

			if callback != nil {
				callback(&char, nil)
			}
		} else {
			if callback != nil {
				callback(nil, nil) // 未找到角色
			}
		}
	})
}
