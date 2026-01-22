package connector

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// MySQLConnector MySQL数据库连接器实现
type MySQLConnector struct {
	BaseConnector
	db        *sql.DB        // MySQL数据库连接
	wg        sync.WaitGroup // 等待组，用于优雅关闭
	isRunning bool           // 运行状态
	queryCh   chan *DBQuery  // 查询通道
	capacity  int            // 通道容量
}

// NewMySQLConnector 创建MySQL数据库连接器
func NewMySQLConnector(name string, capacity int) *MySQLConnector {
	if capacity <= 0 {
		capacity = 1000
	}
	return &MySQLConnector{
		BaseConnector: BaseConnector{
			name:   name,
			driver: "mysql",
		},
		queryCh:  make(chan *DBQuery, capacity),
		capacity: capacity,
	}
}

// Init 初始化MySQL数据库连接
func (c *MySQLConnector) Init(dbConfig config.DBConfig) {
	c.dbConfig = dbConfig
	c.driver = dbConfig.Driver

	// 构建DSN字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
		dbConfig.Charset,
	)

	// 打开数据库连接
	var err error
	c.db, err = sql.Open("mysql", dsn)
	if err != nil {
		zLog.Error("Failed to open MySQL connection", zap.Error(err))
		return
	}

	// 配置连接池
	c.db.SetMaxIdleConns(dbConfig.MaxIdle)
	c.db.SetMaxOpenConns(dbConfig.MaxOpen)
	c.db.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := c.db.Ping(); err != nil {
		zLog.Error("Failed to ping MySQL database", zap.Error(err))
		return
	}

	zLog.Info("MySQL connection established",
		zap.String("host", dbConfig.Host),
		zap.Int("port", dbConfig.Port),
		zap.String("dbname", dbConfig.DBName),
	)
}

// Start 启动MySQL数据库连接和查询处理协程
func (c *MySQLConnector) Start() error {
	if c.isRunning {
		return nil
	}

	c.isRunning = true

	// 启动查询处理协程
	c.wg.Add(1)
	go c.queryWorker()

	return nil
}

// queryWorker 处理MySQL数据库查询请求的工作协程
func (c *MySQLConnector) queryWorker() {
	defer c.wg.Done()

	for query := range c.queryCh {
		rows, err := c.db.Query(query.Query, query.Args...)
		if query.Callback != nil {
			query.Callback(rows, err)
		}
	}
}

// Query 异步执行MySQL数据库查询
func (c *MySQLConnector) Query(sql string, args []interface{}, callback func(*sql.Rows, error)) {
	if !c.isRunning {
		zLog.Error("MySQLConnector is not running")
		if callback != nil {
			callback(nil, fmt.Errorf("mysql connector is not running"))
		}
		return
	}

	// 发送查询请求到通道
	select {
	case c.queryCh <- &DBQuery{
		Query:    sql,
		Args:     args,
		Callback: callback,
	}:
	default:
		zLog.Error("MySQL query channel is full")
		if callback != nil {
			callback(nil, fmt.Errorf("mysql query channel is full"))
		}
	}
}

// Execute 异步执行MySQL数据库执行操作（插入、更新、删除等）
func (c *MySQLConnector) Execute(sql string, args []interface{}, callback func(sql.Result, error)) {
	if !c.isRunning {
		zLog.Error("MySQLConnector is not running")
		if callback != nil {
			callback(nil, fmt.Errorf("mysql connector is not running"))
		}
		return
	}

	// 异步执行查询
	go func() {
		result, err := c.db.Exec(sql, args...)
		if callback != nil {
			callback(result, err)
		}
	}()
}

// Close 关闭MySQL数据库连接
func (c *MySQLConnector) Close() error {
	if !c.isRunning {
		return nil
	}

	c.isRunning = false

	// 停止查询协程
	close(c.queryCh)
	c.wg.Wait()

	// 关闭数据库连接
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			return fmt.Errorf("failed to close MySQL connection: %v", err)
		}
	}

	zLog.Info("MySQL connection closed")
	return nil
}

// GetDriver 获取当前数据库驱动类型
func (c *MySQLConnector) GetDriver() string {
	return c.driver
}

// GetMongoClient 获取MongoDB客户端（MySQL实现中不支持）
func (c *MySQLConnector) GetMongoClient() *mongo.Client {
	zLog.Warn("GetMongoClient called on MySQLConnector")
	return nil
}

// GetMongoDB 获取MongoDB数据库（MySQL实现中不支持）
func (c *MySQLConnector) GetMongoDB() *mongo.Database {
	zLog.Warn("GetMongoDB called on MySQLConnector")
	return nil
}
