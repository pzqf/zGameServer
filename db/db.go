package db

import (
	"sync"

	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/db/connector"
	"github.com/pzqf/zGameServer/db/dao"
	"github.com/pzqf/zGameServer/db/repository"
)

// DBManager 数据库管理器
type DBManager struct {
	connectors          map[string]connector.DBConnector
	CharacterRepository repository.CharacterRepository
	AccountRepository   repository.AccountRepository
	LoginLogRepository  repository.LoginLogRepository
}

var (
	dbManager *DBManager
	dbOnce    sync.Once
)

func GetDBManager() *DBManager {
	return dbManager
}

func InitDBManager() error {
	var err error
	dbOnce.Do(func() {
		dbManager = &DBManager{
			connectors: make(map[string]connector.DBConnector),
		}
		err = dbManager.Init()
	})
	return err
}

// Init 初始化数据库连接和所有Repository
func (manager *DBManager) Init() error {
	// 获取所有数据库配置
	dbConfigs := config.GetAllDBConfigs()

	// 初始化所有数据库连接器
	for dbName, dbConfig := range dbConfigs {
		// 创建数据库连接器
		conn := connector.NewDBConnector(dbName, dbConfig.Driver, 1000)

		// 初始化数据库连接
		conn.Init(dbConfig)

		// 启动数据库连接器
		if err := conn.Start(); err != nil {
			return err
		}

		// 存储连接器
		manager.connectors[dbName] = conn
	}

	// 初始化所有DAO
	var accountDAO *dao.AccountDAO

	// AccountDAO使用account数据库
	if accountConn, ok := manager.connectors["account"]; ok {
		accountDAO = dao.NewAccountDAO(accountConn)
	}

	// 初始化所有Repository
	if accountDAO != nil {
		manager.AccountRepository = repository.NewAccountRepository(accountDAO)
	}

	return nil
}

// Close 关闭所有数据库连接
func (manager *DBManager) Close() error {
	for _, conn := range manager.connectors {
		if err := conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// GetConnector 获取指定名称的数据库连接器
func (manager *DBManager) GetConnector(dbName string) connector.DBConnector {
	return manager.connectors[dbName]
}

// GetAllConnectors 获取所有数据库连接器
func (manager *DBManager) GetAllConnectors() map[string]connector.DBConnector {
	return manager.connectors
}
