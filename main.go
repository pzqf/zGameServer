package main

import (
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zSignal"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/config/tables"
	"github.com/pzqf/zGameServer/db"
	"github.com/pzqf/zGameServer/game/auction"
	"github.com/pzqf/zGameServer/game/guild"
	"github.com/pzqf/zGameServer/game/maps"
	"github.com/pzqf/zGameServer/game/player"
	"github.com/pzqf/zGameServer/gameserver"
	"github.com/pzqf/zGameServer/metrics"
	"github.com/pzqf/zGameServer/net/handler"
	"github.com/pzqf/zGameServer/net/service"
	"go.uber.org/zap"
)

func main() {
	// 捕获所有panic并记录到日志
	defer func() {
		if r := recover(); r != nil {
			// 创建一个默认的日志配置，以防日志系统尚未初始化
			if zLog.GetLogger() == nil {
				zLog.InitLogger(&zLog.Config{
					Level:    zLog.ErrorLevel,
					Console:  true,
					Filename: "./logs/game_server.log",
					MaxSize:  100,
					MaxDays:  7,
				})
			}

			// 捕获并输出堆栈信息
			stack := make([]byte, 4096)
			stack = stack[:runtime.Stack(stack, false)]

			zLog.Fatal("Server crashed with panic",
				zap.Any("panic", r),
				zap.String("stack", string(stack)),
			)
		}
	}()

	// 加载配置文件
	cfg, err := config.LoadConfig("config.ini")
	if err != nil {
		// 如果配置加载失败，使用默认日志配置
		zLog.Fatal("Failed to load config", zap.Error(err))
	}

	// 初始化日志系统
	if err := zLog.InitLogger(&cfg.Log); err != nil {
		// 如果日志初始化失败，使用默认日志
		zLog.Fatal("Failed to initialize logger", zap.Error(err))
	}
	zLog.Info("Config loaded successfully")

	// 输出配置信息
	zLog.Info("Server starting with config",
		zap.String("listen_address", cfg.Server.ListenAddress),
		zap.Int("chan_size", cfg.Server.ChanSize),
		zap.Int("max_client_count", cfg.Server.MaxClientCount),
		zap.Int("log_level", cfg.Log.Level),
		zap.String("log_path", cfg.Log.Filename),
	)

	// 初始化表格配置加载器
	zLog.Info("Initializing table loader...")
	// 初始化全局表格管理器
	tables.GlobalTableManager = tables.NewTableManager()
	if err := tables.GlobalTableManager.LoadAllTables(); err != nil {
		zLog.Fatal("Failed to load configuration tables", zap.Error(err))
	}

	zLog.Info("Starting MMO Game Server...")

	// 创建游戏服务器
	gameServer := gameserver.NewGameServer()

	// 添加核心网络服务
	tcpService := service.NewTcpService(gameServer.GetPacketRouter())
	if err := gameServer.AddService(tcpService); err != nil {
		zLog.Fatal("Failed to add TCP service", zap.Error(err))
	}

	httpService := service.NewHTTPService()
	if err := gameServer.AddService(httpService); err != nil {
		zLog.Fatal("Failed to add HTTP service", zap.Error(err))
	}

	// 注册玩家系统服务
	playerService := player.NewPlayerService()
	if err := gameServer.AddService(playerService); err != nil {
		zLog.Fatal("Failed to add player service", zap.Error(err))
	}

	// 注册全局系统服务
	guildService := guild.NewGuildService()
	if err := gameServer.AddService(guildService); err != nil {
		zLog.Fatal("Failed to add guild service", zap.Error(err))
	}

	auctionService := auction.NewAuctionService()
	if err := gameServer.AddService(auctionService); err != nil {
		zLog.Fatal("Failed to add auction service", zap.Error(err))
	}

	// 注册地图服务
	mapService := maps.NewMapService()
	if err := gameServer.AddService(mapService); err != nil {
		zLog.Fatal("Failed to add map service", zap.Error(err))
	}

	// 初始化数据库管理器
	dbManager := db.NewDBManager()
	if err := dbManager.Init(); err != nil {
		zLog.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer dbManager.Close()

	// 初始化所有处理器
	handler.Init(gameServer.GetPacketRouter(), playerService, guildService, auctionService, mapService, dbManager)

	// 初始化所有服务
	gameServer.InitServices()

	// 注册基本的Prometheus指标
	registerBasicMetrics()

	// 启动pprof性能分析服务器
	go func() {
		pprofAddr := "localhost:6060"
		zLog.Info("Starting pprof server on " + pprofAddr)
		if err := http.ListenAndServe(pprofAddr, nil); err != nil {
			zLog.Error("Failed to start pprof server", zap.Error(err))
		}
	}()

	// 启动配置监控
	if err := config.StartConfigMonitor("config.ini"); err != nil {
		zLog.Error("Failed to start config monitor", zap.Error(err))
	}

	// 启动游戏服务器
	if err := gameServer.Start(); err != nil {
		zLog.Fatal("Failed to start game server", zap.Error(err))
	}

	zLog.Info("Game Server started successfully!")

	// 设置信号处理
	//quit := make(chan os.Signal, 1)
	//signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 等待信号或服务器关闭
	//<-quit
	zSignal.GracefulExit()

	zLog.Info("Received shutdown signal, stopping server...")

	// 停止配置监控
	config.StopConfigMonitor()

	// 停止游戏服务器
	gameServer.Stop()

	// 等待服务器完全关闭
	gameServer.Wait()
}

// registerBasicMetrics 注册基本的Prometheus指标
func registerBasicMetrics() {
	// 注册服务器启动时间指标
	startTime := time.Now()
	metrics.RegisterGauge("server_start_time", "Server start time in Unix timestamp", nil)
	if gauge := metrics.GetGauge("server_start_time"); gauge != nil {
		gauge.Set(float64(startTime.Unix()))
	}

	// 注册活跃连接数指标
	metrics.RegisterGauge("active_connections", "Number of active connections", nil)

	// 注册总连接数指标
	metrics.RegisterCounter("total_connections", "Total number of connections", nil)

	// 注册丢弃连接数指标
	metrics.RegisterCounter("dropped_connections", "Number of dropped connections", nil)

	// 注册发送字节数指标
	metrics.RegisterCounter("total_bytes_sent", "Total bytes sent", nil)

	// 注册接收字节数指标
	metrics.RegisterCounter("total_bytes_received", "Total bytes received", nil)

	// 注册编码错误数指标
	metrics.RegisterCounter("encoding_errors", "Number of encoding errors", nil)

	// 注册解码错误数指标
	metrics.RegisterCounter("decoding_errors", "Number of decoding errors", nil)

	// 注册压缩错误数指标
	metrics.RegisterCounter("compression_errors", "Number of compression errors", nil)

	// 注册丢弃数据包数指标
	metrics.RegisterCounter("dropped_packets", "Number of dropped packets", nil)

	zLog.Info("Basic Prometheus metrics registered successfully")
}
