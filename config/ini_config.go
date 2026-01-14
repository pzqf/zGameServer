package config

import (
	"fmt"
	"os"

	"github.com/pzqf/zEngine/zLog"
	"gopkg.in/ini.v1"
)

// Config 存储所有配置信息
type Config struct {
	Server    ServerConfig
	Log       LogConfig
	Databases map[string]DBConfig // 多数据库配置，key为数据库名称
}

// ServerConfig 服务器网络配置
type ServerConfig struct {
	ListenAddress  string
	ChanSize       int
	MaxClientCount int
	Protocol       string // 协议类型: protobuf, json, xml
	ServerID       int32  // 服务器ID
	ServerName     string // 服务器名称
}

// LogConfig 日志配置
type LogConfig struct {
	Level   string
	Console bool
	Path    string
	MaxSize int
	MaxAge  int
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Charset  string
	MaxIdle  int
	MaxOpen  int
	Driver   string // 数据库驱动类型: mysql, mongo
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// GetConfig 获取全局配置实例
func GetConfig() *Config {
	return GlobalConfig
}

// GetServerConfig 获取服务器配置
func GetServerConfig() *ServerConfig {
	if GlobalConfig == nil {
		return &ServerConfig{}
	}
	return &GlobalConfig.Server
}

// GetLogConfig 获取日志配置
func GetLogConfig() *LogConfig {
	if GlobalConfig == nil {
		return &LogConfig{}
	}
	return &GlobalConfig.Log
}

// GetDBConfig 获取特定名称的数据库配置
func GetDBConfig(name string) *DBConfig {
	if GlobalConfig == nil || GlobalConfig.Databases == nil {
		return &DBConfig{}
	}
	if dbCfg, exists := GlobalConfig.Databases[name]; exists {
		return &dbCfg
	}
	return &DBConfig{}
}

// GetAllDBConfigs 获取所有数据库配置
func GetAllDBConfigs() map[string]DBConfig {
	if GlobalConfig == nil || GlobalConfig.Databases == nil {
		return make(map[string]DBConfig)
	}
	return GlobalConfig.Databases
}

// LoadConfig 从INI文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 如果文件不存在，创建默认配置文件
		if err := createDefaultConfig(filePath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %v", err)
		}
	}

	// 加载INI文件
	cfg, err := ini.Load(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %v", err)
	}

	// 解析配置
	config := &Config{}

	// 解析服务器配置
	config.Server.ListenAddress = cfg.Section("server").Key("listen_address").MustString("0.0.0.0:8888")
	config.Server.ChanSize = cfg.Section("server").Key("chan_size").MustInt(1024)
	config.Server.MaxClientCount = cfg.Section("server").Key("max_client_count").MustInt(10000)
	config.Server.Protocol = cfg.Section("server").Key("protocol").MustString("protobuf")
	config.Server.ServerID = int32(cfg.Section("server").Key("server_id").MustInt(1))
	config.Server.ServerName = cfg.Section("server").Key("server_name").MustString("GameServer")

	// 解析日志配置
	config.Log.Level = cfg.Section("log").Key("level").MustString("info")
	config.Log.Console = cfg.Section("log").Key("console").MustBool(true)
	config.Log.Path = cfg.Section("log").Key("path").MustString("./logs/server.log")
	config.Log.MaxSize = cfg.Section("log").Key("max_size").MustInt(100)
	config.Log.MaxAge = cfg.Section("log").Key("max_age").MustInt(30)

	// 解析所有数据库配置
	config.Databases = make(map[string]DBConfig)
	for _, section := range cfg.Sections() {
		name := section.Name()
		if name != "" && len(name) >= 9 && name[:9] == "database." {
			dbName := name[9:] // 提取数据库名称（如 "game", "account", "log"）
			dbCfg := DBConfig{
				Host:     section.Key("host").MustString("localhost"),
				Port:     section.Key("port").MustInt(3306),
				User:     section.Key("user").MustString("root"),
				Password: section.Key("password").MustString(""),
				DBName:   section.Key("dbname").MustString(dbName),
				Charset:  section.Key("charset").MustString("utf8mb4"),
				MaxIdle:  section.Key("max_idle").MustInt(10),
				MaxOpen:  section.Key("max_open").MustInt(100),
				Driver:   section.Key("driver").MustString("mysql"),
			}
			config.Databases[dbName] = dbCfg
		}
	}

	// 如果没有配置数据库，添加默认的MongoDB配置
	if len(config.Databases) == 0 {
		config.Databases["game"] = DBConfig{
			Host:     "localhost",
			Port:     27017,
			User:     "",
			Password: "",
			DBName:   "game",
			Charset:  "",
			MaxIdle:  10,
			MaxOpen:  100,
			Driver:   "mongo",
		}
	}

	GlobalConfig = config
	return config, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(filePath string) error {
	cfg := ini.Empty()

	// 服务器配置
	serverSection := cfg.Section("server")
	serverSection.Key("listen_address").SetValue("0.0.0.0:8888")
	serverSection.Key("chan_size").SetValue("1024")
	serverSection.Key("max_client_count").SetValue("10000")
	serverSection.Key("protocol").SetValue("protobuf")
	serverSection.Key("server_id").SetValue("1")
	serverSection.Key("server_name").SetValue("GameServer")

	// 日志配置
	logSection := cfg.Section("log")
	logSection.Key("level").SetValue("info")
	logSection.Key("console").SetValue("true")
	logSection.Key("path").SetValue("./logs/server.log")
	logSection.Key("max_size").SetValue("100")
	logSection.Key("max_age").SetValue("30")

	// 游戏数据库配置（MongoDB）
	gameDBSection := cfg.Section("database.game")
	gameDBSection.Key("host").SetValue("192.168.91.128")
	gameDBSection.Key("port").SetValue("27017")
	gameDBSection.Key("user").SetValue("")
	gameDBSection.Key("password").SetValue("")
	gameDBSection.Key("dbname").SetValue("game")
	gameDBSection.Key("charset").SetValue("")
	gameDBSection.Key("max_idle").SetValue("10")
	gameDBSection.Key("max_open").SetValue("100")
	gameDBSection.Key("driver").SetValue("mongo")

	// 账号数据库配置（MongoDB）
	accountDBSection := cfg.Section("database.account")
	accountDBSection.Key("host").SetValue("192.168.91.128")
	accountDBSection.Key("port").SetValue("27017")
	accountDBSection.Key("user").SetValue("")
	accountDBSection.Key("password").SetValue("")
	accountDBSection.Key("dbname").SetValue("account")
	accountDBSection.Key("charset").SetValue("")
	accountDBSection.Key("max_idle").SetValue("10")
	accountDBSection.Key("max_open").SetValue("100")
	accountDBSection.Key("driver").SetValue("mongo")

	// 日志数据库配置（MongoDB）
	logDBSection := cfg.Section("database.log")
	logDBSection.Key("host").SetValue("192.168.91.128")
	logDBSection.Key("port").SetValue("27017")
	logDBSection.Key("user").SetValue("")
	logDBSection.Key("password").SetValue("")
	logDBSection.Key("dbname").SetValue("log")
	logDBSection.Key("charset").SetValue("")
	logDBSection.Key("max_idle").SetValue("10")
	logDBSection.Key("max_open").SetValue("100")
	logDBSection.Key("driver").SetValue("mongo")

	// 保存配置文件
	if err := cfg.SaveTo(filePath); err != nil {
		return err
	}

	fmt.Printf("Created default config file: %s\n", filePath)
	return nil
}

// GetLogLevel 将字符串日志级别转换为zLog.Level
func (lc *LogConfig) GetLogLevel() int {
	switch lc.Level {
	case "debug":
		return zLog.DebugLevel
	case "info":
		return zLog.InfoLevel
	case "warn":
		return zLog.WarnLevel
	case "error":
		return zLog.ErrorLevel
	case "panic":
		return zLog.PanicLevel
	case "fatal":
		return zLog.FatalLevel
	default:
		return zLog.InfoLevel
	}
}

// ToZLogConfig 将LogConfig转换为zLog.Config
func (lc *LogConfig) ToZLogConfig() *zLog.Config {
	return &zLog.Config{
		Level:    lc.GetLogLevel(),
		Console:  lc.Console,
		Filename: lc.Path,
		MaxSize:  lc.MaxSize,
		MaxDays:  lc.MaxAge,
	}
}
