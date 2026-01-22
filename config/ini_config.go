package config

import (
	"fmt"

	"github.com/pzqf/zEngine/zLog"
	"gopkg.in/ini.v1"
)

// Config 存储所有配置信息
type Config struct {
	Server      ServerConfig
	HTTP        HTTPConfig
	Log         LogConfig
	Compression CompressionConfig
	Databases   map[string]DBConfig // 多数据库配置，key为数据库名称
}

// CompressionConfig 压缩配置
type CompressionConfig struct {
	Enabled    bool // 是否启用压缩
	Threshold  int  // 压缩阈值（字节）
	Level      int  // 压缩级别（1-9，1最快，9压缩率最高）
	MinQuality int  // 最低网络质量（0-100）
	MaxQuality int  // 最高网络质量（0-100）
}

// ServerConfig 服务器网络配置
type ServerConfig struct {
	ListenAddress  string // 监听地址
	ChanSize       int    // 通道大小
	MaxClientCount int    // 最大客户端数量
	Protocol       string // 协议类型: protobuf, json, xml
	ServerID       int32  // 服务器ID
	ServerName     string // 服务器名称
}

// HTTPConfig HTTP服务配置
type HTTPConfig struct {
	ListenAddress string // 监听地址
	Enabled       bool   // 是否启用HTTP服务
}

// LogConfig 日志配置
type LogConfig struct {
	Level   string // 日志级别
	Console bool   // 是否输出到控制台
	Path    string // 日志文件路径
	MaxSize int    // 日志文件最大大小（MB）
	MaxAge  int    // 日志文件最大保存天数
}

// DBConfig 数据库配置
type DBConfig struct {
	Host           string // 数据库主机
	Port           int    // 数据库端口
	User           string // 数据库用户名
	Password       string // 数据库密码
	DBName         string // 数据库名称
	Charset        string // 字符集
	MaxIdle        int    // 最大空闲连接数
	MaxOpen        int    // 最大打开连接数
	Driver         string // 数据库驱动类型: mysql, mongo
	URI            string // 数据库连接URI（用于MongoDB等支持URI的数据库）
	MaxPoolSize    int    // 连接池最大连接数（MongoDB）
	MinPoolSize    int    // 连接池最小连接数（MongoDB）
	ConnectTimeout int    // 连接超时时间（秒，MongoDB）
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

// GetHTTPConfig 获取HTTP服务配置
func GetHTTPConfig() *HTTPConfig {
	if GlobalConfig == nil {
		return &HTTPConfig{}
	}
	return &GlobalConfig.HTTP
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

// GetCompressionConfig 获取压缩配置
func GetCompressionConfig() *CompressionConfig {
	if GlobalConfig == nil {
		return &CompressionConfig{
			Enabled:    true,
			Threshold:  1024,
			Level:      5,
			MinQuality: 0,
			MaxQuality: 100,
		}
	}
	return &GlobalConfig.Compression
}

// LoadConfig 从INI文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 直接使用ini库加载配置文件
	cfg, err := ini.Load(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %v", err)
	}

	// 创建配置实例
	config := &Config{
		Databases: make(map[string]DBConfig),
	}

	// 解析服务器配置
	serverSection := cfg.Section("server")
	config.Server = ServerConfig{
		ListenAddress:  serverSection.Key("listen_address").MustString("0.0.0.0:8888"),
		ChanSize:       serverSection.Key("chan_size").MustInt(1024),
		MaxClientCount: serverSection.Key("max_client_count").MustInt(10000),
		Protocol:       serverSection.Key("protocol").MustString("protobuf"),
		ServerID:       int32(serverSection.Key("server_id").MustInt(1)),
		ServerName:     serverSection.Key("server_name").MustString("GameServer"),
	}

	// 解析HTTP服务配置
	httpSection := cfg.Section("http")
	config.HTTP = HTTPConfig{
		ListenAddress: httpSection.Key("listen_address").MustString("0.0.0.0:8080"),
		Enabled:       httpSection.Key("enabled").MustBool(true),
	}

	// 解析日志配置
	logSection := cfg.Section("log")
	config.Log = LogConfig{
		Level:   logSection.Key("level").MustString("info"),
		Console: logSection.Key("console").MustBool(true),
		Path:    logSection.Key("path").MustString("./logs/server.log"),
		MaxSize: logSection.Key("max_size").MustInt(100),
		MaxAge:  logSection.Key("max_age").MustInt(30),
	}

	// 解析压缩配置
	compressionSection := cfg.Section("compression")
	config.Compression = CompressionConfig{
		Enabled:    compressionSection.Key("enabled").MustBool(true),
		Threshold:  compressionSection.Key("threshold").MustInt(1024),
		Level:      compressionSection.Key("level").MustInt(5),
		MinQuality: compressionSection.Key("min_quality").MustInt(0),
		MaxQuality: compressionSection.Key("max_quality").MustInt(100),
	}

	// 解析数据库配置
	for _, section := range cfg.Sections() {
		name := section.Name()
		if len(name) >= 9 && name[:9] == "database." {
			// 提取数据库名称
			dbName := name[9:]

			// 解析数据库配置
			dbCfg := DBConfig{
				Host:           section.Key("host").MustString("localhost"),
				Port:           section.Key("port").MustInt(3306),
				User:           section.Key("user").MustString("root"),
				Password:       section.Key("password").MustString(""),
				DBName:         section.Key("dbname").MustString(dbName),
				Charset:        section.Key("charset").MustString("utf8mb4"),
				MaxIdle:        section.Key("max_idle").MustInt(10),
				MaxOpen:        section.Key("max_open").MustInt(100),
				Driver:         section.Key("driver").MustString("mysql"),
				MaxPoolSize:    section.Key("max_pool_size").MustInt(100),
				MinPoolSize:    section.Key("min_pool_size").MustInt(10),
				ConnectTimeout: section.Key("connect_timeout").MustInt(30),
			}

			// 添加到数据库配置map
			config.Databases[dbName] = dbCfg
		}
	}

	// 如果没有配置数据库，添加默认的MongoDB配置
	if len(config.Databases) == 0 {
		config.Databases["game"] = DBConfig{
			Host:           "localhost",
			Port:           27017,
			User:           "",
			Password:       "",
			DBName:         "game",
			Charset:        "",
			MaxIdle:        10,
			MaxOpen:        100,
			Driver:         "mongo",
			URI:            "mongodb://localhost:27017/game",
			MaxPoolSize:    100,
			MinPoolSize:    10,
			ConnectTimeout: 30,
		}
	}

	// 设置全局配置实例
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

// ToLogConfig 将LogConfig转换为zLog.Config
func (lc *LogConfig) ToLogConfig() *zLog.Config {
	return &zLog.Config{
		Level:    lc.GetLogLevel(),
		Console:  lc.Console,
		Filename: lc.Path,
		MaxSize:  lc.MaxSize,
		MaxDays:  lc.MaxAge,
	}
}
