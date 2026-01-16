package config

import (
	"fmt"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zUtil/zConfig"
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
	// 使用zConfig加载INI文件
	cfg := zConfig.NewConfig()
	if err := cfg.LoadINI(filePath); err != nil {
		return nil, fmt.Errorf("failed to load config file: %v", err)
	}

	// 解析配置到结构体
	config := &Config{}
	if err := cfg.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// 处理数据库配置
	if config.Databases == nil {
		config.Databases = make(map[string]DBConfig)
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
