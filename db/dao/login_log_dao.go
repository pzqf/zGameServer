package dao

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pzqf/zGameServer/db/connector"
	"github.com/pzqf/zGameServer/db/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// LoginLogDAO 角色登录/登出日志数据访问对象
type LoginLogDAO struct {
	connector connector.DBConnector
}

// NewLoginLogDAO 创建角色登录/登出日志DAO实例
func NewLoginLogDAO(dbConnector connector.DBConnector) *LoginLogDAO {
	return &LoginLogDAO{
		connector: dbConnector,
	}
}

// GetLoginLogByCharID 根据角色ID获取登录/登出日志
func (dao *LoginLogDAO) GetLoginLogByCharID(charID int64, callback func(*models.LoginLog, error)) {
	// 根据数据库驱动类型执行不同的查询操作
	if dao.connector.GetDriver() == "mongo" {
		// MongoDB查询
		collection := dao.connector.GetMongoDB().Collection(models.LoginLog{}.TableName())
		var loginLog models.LoginLog

		// 使用FindOne查询单个文档
		result := collection.FindOne(nil, bson.M{"char_id": charID})
		err := result.Decode(&loginLog)

		if err != nil {
			// 如果是未找到文档的错误，返回nil
			if err.Error() == "mongo: no documents in result" {
				if callback != nil {
					callback(nil, nil) // 未找到日志
				}
				return
			}

			// 其他错误
			if callback != nil {
				callback(nil, err)
			}
			return
		}

		if callback != nil {
			callback(&loginLog, nil)
		}
	} else {
		// MySQL查询
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
}

// CreateLoginLog 创建角色登录/登出日志
func (dao *LoginLogDAO) CreateLoginLog(loginLog *models.LoginLog, callback func(int64, error)) {
	// 生成唯一的log_id
	loginLog.LogID = time.Now().UnixNano() / 1000000

	// 根据数据库驱动类型执行不同的插入操作
	if dao.connector.GetDriver() == "mongo" {
		// MongoDB插入
		collection := dao.connector.GetMongoDB().Collection(models.LoginLog{}.TableName())

		// 使用InsertOne插入文档
		_, err := collection.InsertOne(nil, loginLog)

		if err != nil {
			if callback != nil {
				callback(0, err)
			}
			return
		}

		if callback != nil {
			// MongoDB使用自增ID或ObjectID，但在这个模型中我们使用自定义的log_id
			callback(loginLog.LogID, nil)
		}
	} else {
		// MySQL插入
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
}

// GetLoginLogsByTimeRange 根据时间范围获取登录/登出日志
func (dao *LoginLogDAO) GetLoginLogsByTimeRange(startTime, endTime string, callback func([]*models.LoginLog, error)) {
	// 根据数据库驱动类型执行不同的查询操作
	if dao.connector.GetDriver() == "mongo" {
		// MongoDB查询
		collection := dao.connector.GetMongoDB().Collection(models.LoginLog{}.TableName())

		// 构建查询条件和排序
		filter := bson.M{
			"created_at": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
		}
		sort := bson.M{"created_at": -1}

		// 使用Find查询匹配的文档
		cursor, err := collection.Find(nil, filter, &options.FindOptions{Sort: sort})

		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer cursor.Close(nil)

		var loginLogs []*models.LoginLog
		// 遍历游标，解码文档到模型
		for cursor.Next(nil) {
			var loginLog models.LoginLog
			if err := cursor.Decode(&loginLog); err != nil {
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
	} else {
		// MySQL查询
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

				loginLogs = append(loginLogs, &loginLog)
			}

			if callback != nil {
				callback(loginLogs, nil)
			}
		})
	}
}

// GetLoginLogsByOpType 根据操作类型获取登录/登出日志
func (dao *LoginLogDAO) GetLoginLogsByOpType(opType int, callback func([]*models.LoginLog, error)) {
	// 根据数据库驱动类型执行不同的查询操作
	if dao.connector.GetDriver() == "mongo" {
		// MongoDB查询
		collection := dao.connector.GetMongoDB().Collection(models.LoginLog{}.TableName())

		// 构建查询条件和排序
		filter := bson.M{"op_type": opType}
		sort := bson.M{"created_at": -1}

		// 使用Find查询匹配的文档
		cursor, err := collection.Find(nil, filter, &options.FindOptions{Sort: sort})

		if err != nil {
			if callback != nil {
				callback(nil, err)
			}
			return
		}
		defer cursor.Close(nil)

		var loginLogs []*models.LoginLog
		// 遍历游标，解码文档到模型
		for cursor.Next(nil) {
			var loginLog models.LoginLog
			if err := cursor.Decode(&loginLog); err != nil {
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
	} else {
		// MySQL查询
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

				loginLogs = append(loginLogs, &loginLog)
			}

			if callback != nil {
				callback(loginLogs, nil)
			}
		})
	}
}
