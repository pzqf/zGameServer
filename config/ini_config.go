package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zUtil/zConfig"
	"go.uber.org/zap"
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
	Pprof       PprofConfig         // pprof性能分析配置
}

// PprofConfig pprof性能分析配置
type PprofConfig struct {
	Enabled       bool   // 是否启用pprof功能
	ListenAddress string // pprof监听地址，格式为IP:端口
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
var (
	configMonitor *ConfigMonitor
	monitorOnce   sync.Once
)

// 全局配置实例
var (
	GlobalConfig *Config
	configOnce   sync.Once
)

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
	ListenAddress     string // 监听地址
	ChanSize          int    // 通道大小
	MaxClientCount    int    // 最大客户端数量
	Protocol          string // 协议类型: protobuf, json, xml
	ServerID          int32  // 服务器ID
	ServerName        string // 服务器名称
	HeartbeatDuration int    // 心跳时长（秒），0表示禁用心跳
	WorkerID          int64  // Snowflake工作机器ID
	DatacenterID      int64  // Snowflake数据中心ID
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

// GetConfig 获取全局配置实例
func GetConfig() *Config {
	return GlobalConfig
}

// InitConfig 初始化配置
func InitConfig(filePath string) error {
	var err error
	configOnce.Do(func() {
		GlobalConfig, err = LoadConfig(filePath)
		if err == nil {
			err = GlobalConfig.Validate()
		}
	})
	return err
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

// GetPprofConfig 获取pprof配置
func GetPprofConfig() *PprofConfig {
	if GlobalConfig == nil {
		return &PprofConfig{
			Enabled:       false,
			ListenAddress: "localhost:6060",
		}
	}
	return &GlobalConfig.Pprof
}

// LoadConfig 从INI文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 使用zConfig加载配置文件
	zcfg := zConfig.NewConfig()
	if err := zcfg.LoadINI(filePath); err != nil {
		return nil, fmt.Errorf("failed to load config file: %v", err)
	}

	// 创建配置实例
	config := &Config{
		Databases: make(map[string]DBConfig),
	}

	// 解析服务器配置
	config.Server = ServerConfig{
		ListenAddress:     getConfigString(zcfg, "server.listen_address", "0.0.0.0:8888"),
		ChanSize:          getConfigInt(zcfg, "server.chan_size", 1024),
		MaxClientCount:    getConfigInt(zcfg, "server.max_client_count", 10000),
		Protocol:          getConfigString(zcfg, "server.protocol", "protobuf"),
		ServerID:          int32(getConfigInt(zcfg, "server.server_id", 1)),
		ServerName:        getConfigString(zcfg, "server.server_name", "GameServer"),
		HeartbeatDuration: getConfigInt(zcfg, "server.heartbeat_duration", 0),
		WorkerID:          int64(getConfigInt(zcfg, "server.worker_id", 1)),
		DatacenterID:      int64(getConfigInt(zcfg, "server.datacenter_id", 1)),
	}

	// 解析防DDoS攻击配置
	config.DDoS = zNet.DDoSConfig{
		MaxConnPerIP:      getConfigInt(zcfg, "ddos.max_conn_per_ip", 10),
		ConnTimeWindow:    getConfigInt(zcfg, "ddos.conn_time_window", 60),
		MaxPacketsPerIP:   getConfigInt(zcfg, "ddos.max_packets_per_ip", 100),
		PacketTimeWindow:  getConfigInt(zcfg, "ddos.packet_time_window", 1),
		MaxBytesPerIP:     int64(getConfigInt(zcfg, "ddos.max_bytes_per_ip", 10*1024*1024)),
		TrafficTimeWindow: getConfigInt(zcfg, "ddos.traffic_time_window", 3600),
		BanDuration:       getConfigInt(zcfg, "ddos.ban_duration", 24*3600),
	}

	// 解析HTTP服务配置
	config.HTTP = zNet.HttpConfig{
		ListenAddress:     getConfigString(zcfg, "http.listen_address", "0.0.0.0:8080"),
		MaxClientCount:    getConfigInt(zcfg, "http.max_client_count", 10000),
		MaxPacketDataSize: int32(getConfigInt(zcfg, "http.max_packet_data_size", 1024*1024)),
	}

	// 解析日志配置
	config.Log = zLog.Config{
		Level:              getConfigInt(zcfg, "log.level", 0),
		Console:            getConfigBool(zcfg, "log.console", true),
		Filename:           getConfigString(zcfg, "log.filename", "./logs/server.log"),
		MaxSize:            getConfigInt(zcfg, "log.max-size", 100),
		MaxDays:            getConfigInt(zcfg, "log.max-days", 30),
		MaxBackups:         getConfigInt(zcfg, "log.max-backups", 5),
		Compress:           getConfigBool(zcfg, "log.compress", true),
		ShowCaller:         getConfigBool(zcfg, "log.show-caller", true),
		Stacktrace:         getConfigInt(zcfg, "log.stacktrace", 2),
		Sampling:           getConfigBool(zcfg, "log.sampling", true),
		SamplingInitial:    getConfigInt(zcfg, "log.sampling-initial", 100),
		SamplingThereafter: getConfigInt(zcfg, "log.sampling-thereafter", 10),
		Async:              getConfigBool(zcfg, "log.async", false),
		AsyncBufferSize:    getConfigInt(zcfg, "log.async-buffer-size", 1024),
		AsyncFlushInterval: getConfigInt(zcfg, "log.async-flush-interval", 100),
	}

	// 解析HTTP服务启用状态
	config.HTTPEnabled = getConfigBool(zcfg, "http.enabled", true)

	// 解析压缩配置
	config.Compression = CompressionConfig{
		Enabled:    getConfigBool(zcfg, "net_compression.enabled", true),
		Threshold:  getConfigInt(zcfg, "net_compression.threshold", 1024),
		Level:      getConfigInt(zcfg, "net_compression.level", 5),
		MinQuality: getConfigInt(zcfg, "net_compression.min_quality", 0),
		MaxQuality: getConfigInt(zcfg, "net_compression.max_quality", 100),
	}

	// 解析数据库配置
	// 这里简化处理，实际项目中可能需要更复杂的逻辑
	config.Databases["game"] = DBConfig{
		Host:           getConfigString(zcfg, "database.game.host", "localhost"),
		Port:           getConfigInt(zcfg, "database.game.port", 27017),
		User:           getConfigString(zcfg, "database.game.user", ""),
		Password:       getConfigString(zcfg, "database.game.password", ""),
		DBName:         getConfigString(zcfg, "database.game.dbname", "game"),
		Charset:        getConfigString(zcfg, "database.game.charset", ""),
		MaxIdle:        getConfigInt(zcfg, "database.game.max_idle", 10),
		MaxOpen:        getConfigInt(zcfg, "database.game.max_open", 100),
		Driver:         getConfigString(zcfg, "database.game.driver", "mongo"),
		URI:            getConfigString(zcfg, "database.game.uri", "mongodb://localhost:27017/game"),
		MaxPoolSize:    getConfigInt(zcfg, "database.game.max_pool_size", 100),
		MinPoolSize:    getConfigInt(zcfg, "database.game.min_pool_size", 10),
		ConnectTimeout: getConfigInt(zcfg, "database.game.connect_timeout", 30),
	}

	config.Databases["account"] = DBConfig{
		Host:           getConfigString(zcfg, "database.account.host", "localhost"),
		Port:           getConfigInt(zcfg, "database.account.port", 27017),
		User:           getConfigString(zcfg, "database.account.user", ""),
		Password:       getConfigString(zcfg, "database.account.password", ""),
		DBName:         getConfigString(zcfg, "database.account.dbname", "account"),
		Charset:        getConfigString(zcfg, "database.account.charset", ""),
		MaxIdle:        getConfigInt(zcfg, "database.account.max_idle", 10),
		MaxOpen:        getConfigInt(zcfg, "database.account.max_open", 100),
		Driver:         getConfigString(zcfg, "database.account.driver", "mongo"),
		URI:            getConfigString(zcfg, "database.account.uri", "mongodb://localhost:27017/account"),
		MaxPoolSize:    getConfigInt(zcfg, "database.account.max_pool_size", 100),
		MinPoolSize:    getConfigInt(zcfg, "database.account.min_pool_size", 10),
		ConnectTimeout: getConfigInt(zcfg, "database.account.connect_timeout", 30),
	}

	config.Databases["log"] = DBConfig{
		Host:           getConfigString(zcfg, "database.log.host", "localhost"),
		Port:           getConfigInt(zcfg, "database.log.port", 27017),
		User:           getConfigString(zcfg, "database.log.user", ""),
		Password:       getConfigString(zcfg, "database.log.password", ""),
		DBName:         getConfigString(zcfg, "database.log.dbname", "log"),
		Charset:        getConfigString(zcfg, "database.log.charset", ""),
		MaxIdle:        getConfigInt(zcfg, "database.log.max_idle", 10),
		MaxOpen:        getConfigInt(zcfg, "database.log.max_open", 100),
		Driver:         getConfigString(zcfg, "database.log.driver", "mongo"),
		URI:            getConfigString(zcfg, "database.log.uri", "mongodb://localhost:27017/log"),
		MaxPoolSize:    getConfigInt(zcfg, "database.log.max_pool_size", 100),
		MinPoolSize:    getConfigInt(zcfg, "database.log.min_pool_size", 10),
		ConnectTimeout: getConfigInt(zcfg, "database.log.connect_timeout", 30),
	}

	// 解析pprof配置
	config.Pprof = PprofConfig{
		Enabled:       getConfigBool(zcfg, "pprof.enabled", false),
		ListenAddress: getConfigString(zcfg, "pprof.listen_address", "localhost:6060"),
	}

	// 设置全局配置实例
	GlobalConfig = config
	return config, nil
}

// 辅助函数：获取字符串配置
func getConfigString(cfg *zConfig.Config, key string, defaultValue string) string {
	if value, err := cfg.GetString(key); err == nil {
		return value
	}
	return defaultValue
}

// 辅助函数：获取整数配置
func getConfigInt(cfg *zConfig.Config, key string, defaultValue int) int {
	if value, err := cfg.GetInt(key); err == nil {
		return value
	}
	return defaultValue
}

// 辅助函数：获取布尔配置
func getConfigBool(cfg *zConfig.Config, key string, defaultValue bool) bool {
	if value, err := cfg.GetBool(key); err == nil {
		return value
	}
	return defaultValue
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
	// 验证心跳配置（0表示禁用，10-300秒之间较为合理）
	if c.Server.HeartbeatDuration < 0 {
		c.Server.HeartbeatDuration = 0
	} else if c.Server.HeartbeatDuration > 0 && c.Server.HeartbeatDuration < 10 {
		c.Server.HeartbeatDuration = 10
	} else if c.Server.HeartbeatDuration > 300 {
		c.Server.HeartbeatDuration = 300
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

	// 验证pprof配置
	if c.Pprof.ListenAddress == "" {
		c.Pprof.ListenAddress = "localhost:6060"
	}

	return nil
}

// StartConfigMonitor 启动配置监控
func StartConfigMonitor(configPath string) error {
	monitorOnce.Do(func() {
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
