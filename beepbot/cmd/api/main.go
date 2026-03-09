package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/StellarisJAY/beepbot/internal/api"
	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/database"
	"github.com/StellarisJAY/beepbot/internal/logger"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/types"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "config.json", "Path to the config file")
}

func main() {
	flag.Parse()

	// 加载配置文件
	cfg, err := config.FromConfigFile[config.APIConfig](configFile)
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	// 初始化日志
	logger.InitLogger(cfg.Logging)
	slog.Info("Starting BeepBot API Server...")

	// 获取配置文件所在目录，用于存储加密密钥
	configDir := filepath.Dir(configFile)
	keyFilePath := filepath.Join(configDir, ".encryption_key")

	// 初始化加密器
	encryptionKey, err := crypto.GetOrCreateEncryptionKey(cfg.Encryption.Key, keyFilePath)
	if err != nil {
		panic(fmt.Errorf("failed to initialize encryption: %w", err))
	}
	slog.Info("Encryption initialized", "key_file", keyFilePath)

	encryptor, err := crypto.NewEncryptor(encryptionKey)
	if err != nil {
		panic(fmt.Errorf("failed to create encryptor: %w", err))
	}

	// 初始化数据库
	db, err := database.InitDatabase(cfg.Database, cfg.Logging)
	if err != nil {
		panic(fmt.Errorf("failed to initialize database: %w", err))
	}
	defer func() {
		if err := database.CloseDatabase(db); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}()

	// 初始化仓储层
	providerRepo := repository.NewProviderRepository(db)
	agentRepo := repository.NewAgentRepository(db)
	botRepo := repository.NewBotRepository(db)

	// 初始化服务层
	providerService := service.NewProviderService(providerRepo, encryptor)
	agentService := service.NewAgentService(agentRepo, providerService)
	botService := service.NewBotService(botRepo, agentService)

	// 创建消息总线
	messageBus := channel.NewMessageBus()

	// 创建 ChannelManager
	channelManager := channel.NewChannelManager(messageBus)

	// 将 ChannelManager 注入到 BotService
	botService.SetChannelManager(channelManager)

	// 设置 API 路由
	router := api.SetupRouter(providerService, agentService, botService)

	// 确定 API 端口
	port := cfg.Port
	if port == 0 {
		port = 8080
	}

	// 监听系统信号
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// 启动出站消息分发
	go func() {
		channelManager.DispatchOutbound(ctx)
	}()

	// 启动 AgentManager 消息循环
	agentManager := service.NewAgentManager(*cfg, db, agentRepo, botRepo, providerRepo, messageBus, encryptor)
	go func() {
		agentManager.MessageLoop(ctx)
	}()

	// 加载所有 active 状态的 Bot 并启动 Channel
	activeBots, err := botRepo.GetByStatus(types.BotStatusActive)
	if err != nil {
		slog.Warn("failed to load active bots", "error", err)
	} else {
		slog.Info("loading active bots", "count", len(activeBots))
		errs := channelManager.StartAllActiveChannels(ctx, activeBots)
		for _, e := range errs {
			slog.Error("failed to start channel", "error", e)
		}
	}

	// 启动 API 服务器
	go func() {
		slog.Info("API server starting", "port", port)
		if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
			slog.Error("API server error", "error", err)
			cancel()
		}
	}()

	// 等待退出信号
	select {
	case <-ctx.Done():
		slog.Info("API server stopped")
	case sig := <-signalCh:
		slog.Info("Received signal, shutting down", "signal", sig)
		cancel()
	}

	// 停止所有 Channel
	slog.Info("Stopping all channels...")
	channelManager.Shutdown()

	slog.Info("BeepBot API Server shutdown complete")
}
