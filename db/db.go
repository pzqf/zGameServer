package db

import (
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/db/connector"
	"github.com/pzqf/zGameServer/db/dao"
)

// DBManager 数据库管理器
type DBManager struct {
	connectors   map[string]connector.DBConnector
	CharacterDAO *dao.CharacterDAO
	AccountDAO   *dao.AccountDAO
	LoginLogDAO  *dao.LoginLogDAO
}

// NewDBManager 创建数据库管理器实例
func NewDBManager() *DBManager {
	return &DBManager{
		connectors: make(map[string]connector.DBConnector),
	}
}

// Init 初始化数据库连接和所有DAO
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
	// CharacterDAO使用game数据库
	if gameConn, ok := manager.connectors["game"]; ok {
		manager.CharacterDAO = dao.NewCharacterDAO(gameConn)
	}

	// AccountDAO使用account数据库
	if accountConn, ok := manager.connectors["account"]; ok {
		manager.AccountDAO = dao.NewAccountDAO(accountConn)
	}

	// LoginLogDAO使用log数据库
	if logConn, ok := manager.connectors["log"]; ok {
		manager.LoginLogDAO = dao.NewLoginLogDAO(logConn)
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
