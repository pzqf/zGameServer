package connector

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/config"
	"go.uber.org/zap"
)

// DBConnector 数据库连接器
type DBConnector struct {
	name      string         // 数据库名称
	db        *sql.DB        // 数据库连接
	logger    *zap.Logger    // 日志记录器
	wg        sync.WaitGroup // 等待组，用于优雅关闭
	isRunning bool           // 运行状态
	queryCh   chan *DBQuery  // 查询通道
	capacity  int            // 通道容量
}

// DBQuery 数据库查询请求
type DBQuery struct {
	Query    string
	Args     []interface{}
	Callback func(*sql.Rows, error)
}

// DBExecutor 数据库执行请求
type DBExecutor struct {
	Query    string
	Args     []interface{}
	Callback func(sql.Result, error)
}

// NewDBConnector 创建数据库连接器
func NewDBConnector(name string, capacity int) *DBConnector {
	if capacity <= 0 {
		capacity = 1000
	}
	return &DBConnector{
		name:     name,
		logger:   zLog.GetLogger(),
		queryCh:  make(chan *DBQuery, capacity),
		capacity: capacity,
	}
}

// Init 初始化数据库连接
func (c *DBConnector) Init() error {
	// 获取数据库配置
	dbConfig := config.GetDBConfig(c.name)

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
		return fmt.Errorf("failed to open database connection: %v", err)
	}

	// 配置连接池
	c.db.SetMaxIdleConns(dbConfig.MaxIdle)
	c.db.SetMaxOpenConns(dbConfig.MaxOpen)
	c.db.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := c.db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	c.logger.Info("Database connection established",
		zap.String("host", dbConfig.Host),
		zap.Int("port", dbConfig.Port),
		zap.String("dbname", dbConfig.DBName),
	)

	return nil
}

// Start 启动数据库查询处理协程
func (c *DBConnector) Start() {
	if c.isRunning {
		return
	}

	c.isRunning = true
	c.wg.Add(1)

	go c.queryWorker()
}

// queryWorker 处理数据库查询请求的工作协程
func (c *DBConnector) queryWorker() {
	defer c.wg.Done()

	for query := range c.queryCh {
		rows, err := c.db.Query(query.Query, query.Args...)
		if query.Callback != nil {
			query.Callback(rows, err)
		}
	}
}

// Query 异步执行数据库查询
func (c *DBConnector) Query(query string, args []interface{}, callback func(*sql.Rows, error)) {
	if !c.isRunning {
		c.logger.Error("DBConnector is not running")
		if callback != nil {
			callback(nil, fmt.Errorf("db connector is not running"))
		}
		return
	}

	select {
	case c.queryCh <- &DBQuery{
		Query:    query,
		Args:     args,
		Callback: callback,
	}:
	default:
		c.logger.Error("DB query channel is full")
		if callback != nil {
			callback(nil, fmt.Errorf("db query channel is full"))
		}
	}
}

// Execute 异步执行数据库更新操作
func (c *DBConnector) Execute(query string, args []interface{}, callback func(sql.Result, error)) {
	if !c.isRunning {
		c.logger.Error("DBConnector is not running")
		if callback != nil {
			callback(nil, fmt.Errorf("db connector is not running"))
		}
		return
	}

	// 在单独的协程中执行更新操作
	go func() {
		result, err := c.db.Exec(query, args...)
		if callback != nil {
			callback(result, err)
		}
	}()
}

// Close 关闭数据库连接
func (c *DBConnector) Close() error {
	if !c.isRunning {
		return nil
	}

	// 停止查询协程
	close(c.queryCh)
	c.isRunning = false
	c.wg.Wait()

	// 关闭数据库连接
	if err := c.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %v", err)
	}

	c.logger.Info("Database connection closed")
	return nil
}

// GetDB 获取原始数据库连接（谨慎使用）
func (c *DBConnector) GetDB() *sql.DB {
	return c.db
}
