package main

// 测试注释1: 验证我可以修改代码
// 测试注释2: 添加多行注释
// 测试注释3: 确认工具链工作正常
// 测试注释4: 系统运行正常

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

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
	"github.com/pzqf/zGameServer/util"
	"go.uber.org/zap"
)

func main() {
	defer util.Recover(func(recover interface{}, stack string) {
		if zLog.GetLogger() != nil {
			zLog.Fatal("Server crashed with panic",
				zap.Any("panic", recover),
				zap.String("stack", stack),
			)
		} else {
			fmt.Println("Server crashed with panic", recover, stack)
		}

	})

	_, err := config.LoadConfig("config.ini")
	if err != nil {
		return
	}
	if err := zLog.InitLogger(config.GetLogConfig()); err != nil {
		zLog.Fatal("Failed to initialize logger", zap.Error(err))
	}
	zLog.Info("Config loaded successfully")

	zLog.Info("Server starting with config",
		zap.String("listen_address", config.GetServerConfig().ListenAddress),
		zap.Int("chan_size", config.GetServerConfig().ChanSize),
		zap.Int("max_client_count", config.GetServerConfig().MaxClientCount),
		zap.Int("log_level", config.GetConfig().Log.Level),
		zap.String("log_path", config.GetConfig().Log.Filename),
	)

	zLog.Info("Initializing table loader...")
	tables.GlobalTableManager = tables.NewTableManager()
	if err := tables.GlobalTableManager.LoadAllTables(); err != nil {
		zLog.Fatal("Failed to load configuration tables", zap.Error(err))
	}

	zLog.Info("Starting MMO Game Server...")

	gameServer := gameserver.NewGameServer()

	dbManager, err := setupServices(gameServer)
	if err != nil {
		zLog.Fatal("Failed to setup services", zap.Error(err))
	}
	defer dbManager.Close()

	metrics.RegisterBasicMetrics()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保退出时调用cancel()
	go func() {
		pprofAddr := "localhost:6060"
		zLog.Info("Starting pprof server on " + pprofAddr)
		server := &http.Server{
			Addr: pprofAddr,
		}

		// 在context取消时关闭服务器
		go func() {
			<-ctx.Done()
			if err := server.Shutdown(context.Background()); err != nil {
				zLog.Error("Failed to shut down pprof server", zap.Error(err))
			}
		}()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zLog.Error("Failed to start pprof server", zap.Error(err))
		}
	}()

	if err := config.StartConfigMonitor("config.ini"); err != nil {
		zLog.Error("Failed to start config monitor", zap.Error(err))
	}

	if err := gameServer.Start(); err != nil {
		zLog.Fatal("Failed to start game server", zap.Error(err))
	}

	zLog.Info("Game Server started successfully!")

	zSignal.GracefulExit()
	cancel()

	zLog.Info("Received shutdown signal, stopping server...")

	config.StopConfigMonitor()

	gameServer.Stop()

	gameServer.Wait()
}

func setupServices(gameServer *gameserver.GameServer) (*db.DBManager, error) {
	tcpService := service.NewTcpService(gameServer.GetPacketRouter())
	if err := gameServer.AddService(tcpService); err != nil {
		return nil, fmt.Errorf("failed to add TCP service: %w", err)
	}

	httpService := service.NewHTTPService()
	if err := gameServer.AddService(httpService); err != nil {
		return nil, fmt.Errorf("failed to add HTTP service: %w", err)
	}

	playerService := player.NewPlayerService()
	if err := gameServer.AddService(playerService); err != nil {
		return nil, fmt.Errorf("failed to add player service: %w", err)
	}

	guildService := guild.NewGuildService()
	if err := gameServer.AddService(guildService); err != nil {
		return nil, fmt.Errorf("failed to add guild service: %w", err)
	}

	auctionService := auction.NewAuctionService()
	if err := gameServer.AddService(auctionService); err != nil {
		return nil, fmt.Errorf("failed to add auction service: %w", err)
	}

	mapService := maps.NewMapService()
	if err := gameServer.AddService(mapService); err != nil {
		return nil, fmt.Errorf("failed to add map service: %w", err)
	}

	dbManager := db.NewDBManager()
	if err := dbManager.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	handler.Init(gameServer.GetPacketRouter(), playerService, guildService, auctionService, mapService, dbManager)

	gameServer.InitServices()

	return dbManager, nil
}
