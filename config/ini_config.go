package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"go.uber.org/zap"
	"gopkg.in/ini.v1"
)

// Config 存储所有配置信息
type Config struct {
	Server      ServerConfig
	HTTP        zNet.HttpConfig // HTTP服务配置
	HTTPEnabled bool            // 是否启用HTTP服务
	Log         zLog.Config     // 日志配置
	Compression CompressionConfig
	DDoS        zNet.DDoSConfig     // 防DDoS攻击配置
	Databases   map[string]DBConfig // 多数据库配置，key为数据库名称
}

// 配置监控器
type ConfigMonitor struct {
	configPath     string
	lastModifyTime time.Time
	mu             sync.Mutex
	stopChan       chan struct{}
	isRunning      bool
}

// 全局配置监控器
var configMonitor *ConfigMonitor
var once sync.Once

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

// HTTPConfigWithEnabled HTTP服务配置和启用状态
type HTTPConfigWithEnabled struct {
	Config  zNet.HttpConfig
	Enabled bool
}

// GetHTTPConfig 获取HTTP服务配置
func GetHTTPConfig() *HTTPConfigWithEnabled {
	if GlobalConfig == nil {
		return &HTTPConfigWithEnabled{
			Config:  zNet.HttpConfig{},
			Enabled: true,
		}
	}
	return &HTTPConfigWithEnabled{
		Config:  GlobalConfig.HTTP,
		Enabled: GlobalConfig.HTTPEnabled,
	}
}

// GetLogConfig 获取日志配置
func GetLogConfig() *zLog.Config {
	if GlobalConfig == nil {
		return &zLog.Config{}
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

	// 解析防DDoS攻击配置
	config.DDoS = zNet.DDoSConfig{
		MaxConnPerIP:      serverSection.Key("max_conn_per_ip").MustInt(10),
		ConnTimeWindow:    serverSection.Key("conn_time_window").MustInt(60),
		MaxPacketsPerIP:   serverSection.Key("max_packets_per_ip").MustInt(100),
		PacketTimeWindow:  serverSection.Key("packet_time_window").MustInt(1),
		MaxBytesPerIP:     serverSection.Key("max_bytes_per_ip").MustInt64(10 * 1024 * 1024),
		TrafficTimeWindow: serverSection.Key("traffic_time_window").MustInt(3600),
		BanDuration:       serverSection.Key("ban_duration").MustInt(24 * 3600),
	}

	// 解析HTTP服务配置
	httpSection := cfg.Section("http")
	config.HTTP = zNet.HttpConfig{
		ListenAddress:     httpSection.Key("listen_address").MustString("0.0.0.0:8080"),
		MaxClientCount:    httpSection.Key("max_client_count").MustInt(10000),
		MaxPacketDataSize: int32(httpSection.Key("max_packet_data_size").MustInt(1024 * 1024)),
	}

	// 解析日志配置
	logSection := cfg.Section("log")
	config.Log = zLog.Config{
		Level:    logSection.Key("level").MustInt(0),
		Console:  logSection.Key("console").MustBool(true),
		Filename: logSection.Key("filename").MustString("./logs/server.log"),
		MaxSize:  logSection.Key("max-size").MustInt(100),
		MaxDays:  logSection.Key("max-days").MustInt(30),
	}

	// 解析HTTP服务启用状态
	config.HTTPEnabled = httpSection.Key("enabled").MustBool(true)

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

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 验证服务器配置
	if c.Server.ListenAddress == "" {
		return fmt.Errorf("server listen_address is required")
	}
	if c.Server.ChanSize <= 0 {
		c.Server.ChanSize = 1024
	}
	if c.Server.MaxClientCount <= 0 {
		c.Server.MaxClientCount = 10000
	}

	// 验证防DDOS攻击配置
	if c.DDoS.MaxConnPerIP <= 0 {
		c.DDoS.MaxConnPerIP = 10
	}
	if c.DDoS.ConnTimeWindow <= 0 {
		c.DDoS.ConnTimeWindow = 60
	}
	if c.DDoS.MaxPacketsPerIP <= 0 {
		c.DDoS.MaxPacketsPerIP = 100
	}
	if c.DDoS.PacketTimeWindow <= 0 {
		c.DDoS.PacketTimeWindow = 1
	}
	if c.DDoS.MaxBytesPerIP <= 0 {
		c.DDoS.MaxBytesPerIP = 10 * 1024 * 1024
	}
	if c.DDoS.TrafficTimeWindow <= 0 {
		c.DDoS.TrafficTimeWindow = 3600
	}
	if c.DDoS.BanDuration <= 0 {
		c.DDoS.BanDuration = 24 * 3600
	}

	// 验证HTTP配置
	if c.HTTP.ListenAddress == "" {
		c.HTTP.ListenAddress = "0.0.0.0:8080"
	}
	if c.HTTP.MaxClientCount <= 0 {
		c.HTTP.MaxClientCount = 10000
	}
	if c.HTTP.MaxPacketDataSize <= 0 {
		c.HTTP.MaxPacketDataSize = 1024 * 1024
	}

	// 验证日志配置
	if c.Log.Filename == "" {
		c.Log.Filename = "./logs/server.log"
	}
	if c.Log.MaxSize <= 0 {
		c.Log.MaxSize = 100
	}
	if c.Log.MaxDays <= 0 {
		c.Log.MaxDays = 30
	}

	// 验证压缩配置
	if c.Compression.Threshold <= 0 {
		c.Compression.Threshold = 1024
	}
	if c.Compression.Level < 1 || c.Compression.Level > 9 {
		c.Compression.Level = 5
	}
	if c.Compression.MinQuality < 0 || c.Compression.MinQuality > 100 {
		c.Compression.MinQuality = 0
	}
	if c.Compression.MaxQuality < 0 || c.Compression.MaxQuality > 100 {
		c.Compression.MaxQuality = 100
	}

	// 验证数据库配置
	for name, dbCfg := range c.Databases {
		if dbCfg.Host == "" {
			return fmt.Errorf("database %s host is required", name)
		}
		if dbCfg.Port <= 0 {
			return fmt.Errorf("database %s port is required", name)
		}
		if dbCfg.User == "" {
			return fmt.Errorf("database %s user is required", name)
		}
		if dbCfg.DBName == "" {
			return fmt.Errorf("database %s dbname is required", name)
		}
		if dbCfg.Driver == "" {
			return fmt.Errorf("database %s driver is required", name)
		}
		if dbCfg.MaxIdle < 0 {
			dbCfg.MaxIdle = 10
		}
		if dbCfg.MaxOpen < 0 {
			dbCfg.MaxOpen = 100
		}
		if dbCfg.MaxIdle > dbCfg.MaxOpen {
			dbCfg.MaxIdle = dbCfg.MaxOpen
		}
	}

	return nil
}

// StartConfigMonitor 启动配置监控
func StartConfigMonitor(configPath string) error {
	once.Do(func() {
		configMonitor = &ConfigMonitor{
			configPath: configPath,
			stopChan:   make(chan struct{}),
		}
	})

	if configMonitor.isRunning {
		return nil
	}

	// 获取初始修改时间
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %v", err)
	}
	configMonitor.lastModifyTime = fileInfo.ModTime()

	// 启动监控协程
	configMonitor.isRunning = true
	go configMonitor.monitor()

	zLog.Info("Config monitor started", zap.String("config_path", configPath))
	return nil
}

// StopConfigMonitor 停止配置监控
func StopConfigMonitor() {
	if configMonitor != nil && configMonitor.isRunning {
		close(configMonitor.stopChan)
		configMonitor.isRunning = false
		zLog.Info("Config monitor stopped")
	}
}

// monitor 监控配置文件变化
func (cm *ConfigMonitor) monitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.checkConfigChange()
		case <-cm.stopChan:
			return
		}
	}
}

// checkConfigChange 检查配置文件是否变化
func (cm *ConfigMonitor) checkConfigChange() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 获取当前修改时间
	fileInfo, err := os.Stat(cm.configPath)
	if err != nil {
		zLog.Error("Failed to stat config file: " + err.Error())
		return
	}

	currentModifyTime := fileInfo.ModTime()
	if currentModifyTime.After(cm.lastModifyTime) {
		// 配置文件已修改，重新加载
		zLog.Info("Config file changed, reloading...", zap.String("config_path", cm.configPath))

		// 重新加载配置
		newConfig, err := LoadConfig(cm.configPath)
		if err != nil {
			zLog.Error("Failed to reload config: " + err.Error())
			return
		}

		// 验证配置
		if err := newConfig.Validate(); err != nil {
			zLog.Error("Failed to validate config: " + err.Error())
			return
		}

		// 更新全局配置
		GlobalConfig = newConfig
		cm.lastModifyTime = currentModifyTime

		zLog.Info("Config reloaded successfully")
	}
}

// ReloadConfig 手动重新加载配置
func ReloadConfig() error {
	if configMonitor == nil {
		return fmt.Errorf("config monitor not started")
	}

	configMonitor.mu.Lock()
	defer configMonitor.mu.Unlock()

	// 重新加载配置
	newConfig, err := LoadConfig(configMonitor.configPath)
	if err != nil {
		return err
	}

	// 验证配置
	if err := newConfig.Validate(); err != nil {
		return err
	}

	// 更新全局配置
	GlobalConfig = newConfig

	// 更新修改时间
	fileInfo, err := os.Stat(configMonitor.configPath)
	if err != nil {
		return err
	}
	configMonitor.lastModifyTime = fileInfo.ModTime()

	zLog.Info("Config reloaded successfully")
	return nil
}
