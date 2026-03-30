package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
	cronscheduler "github.com/StellarisJAY/beepbot/internal/cron"
	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/database"
	"github.com/StellarisJAY/beepbot/internal/logger"
	"github.com/StellarisJAY/beepbot/internal/mcp"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "config.json", "Path to the config file")
}

// getOrCreateJWTSecret 获取或创建 JWT 密钥
func getOrCreateJWTSecret(configSecret, keyFilePath string) (string, error) {
	// 如果配置中提供了密钥，直接使用
	if configSecret != "" {
		return configSecret, nil
	}

	// 尝试从文件读取
	data, err := os.ReadFile(keyFilePath)
	if err == nil {
		return string(data), nil
	}

	// 文件不存在，生成新的密钥
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}
	secret := base64.StdEncoding.EncodeToString(bytes)

	// 保存到文件
	if err := os.WriteFile(keyFilePath, []byte(secret), 0600); err != nil {
		return "", fmt.Errorf("failed to save JWT secret: %w", err)
	}

	return secret, nil
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

	keyFilePath := filepath.Join(".keys", ".encryption_key")

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
	sessionRepo := repository.NewSessionRepository(db)
	cronRepo := repository.NewCronRepository(db)
	skillRepo := repository.NewSkillRepository(db)
	mcpRepo := repository.NewMCPServerRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	teamRepo := repository.NewAgentTeamRepository(db)

	// 初始化 JWT 密钥
	jwtKeyFilePath := filepath.Join(".keys", ".jwt_secret")
	jwtSecret, err := getOrCreateJWTSecret(cfg.JWT.Secret, jwtKeyFilePath)
	if err != nil {
		panic(fmt.Errorf("failed to initialize JWT secret: %w", err))
	}
	slog.Info("JWT initialized", "key_file", jwtKeyFilePath)

	// 初始化认证服务
	authService := service.NewAuthService(adminRepo, jwtSecret)

	// 创建默认管理员
	if err := authService.CreateDefaultAdmin(); err != nil {
		slog.Warn("failed to create default admin", "error", err)
	}

	// 初始化服务层
	providerService := service.NewProviderService(providerRepo, encryptor)
	agentService := service.NewAgentService(agentRepo, skillRepo, providerService)
	botService := service.NewBotService(botRepo, agentService)
	sessionService := service.NewSessionService(sessionRepo, botRepo)
	cronService := service.NewCronService(cronRepo, agentService)
	skillService := service.NewSkillService(skillRepo, agentRepo, cfg.DataDir)
	teamService := service.NewTeamService(teamRepo, agentRepo, db)

	// 初始化 MCP 管理器和服务
	mcpManager := mcp.NewManager(mcpRepo)
	mcpService := service.NewMCPService(mcpRepo, mcpManager)

	// 确保技能目录存在
	skillsDir := filepath.Join(cfg.DataDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		slog.Warn("failed to create skills directory", "error", err)
	}

	// 创建消息总线
	messageBus := channel.NewMessageBus()

	// 创建 ChannelManager
	channelManager := channel.NewChannelManager(messageBus)

	// 将 ChannelManager 注入到 BotService
	botService.SetChannelManager(channelManager)

	// 创建 CronToolDeps 用于定时任务工具
	cronDeps := &tool.CronToolDeps{
		CronRepo:   cronRepo,
		Scheduler:  nil, // 稍后设置
		Validator:  tool.DefaultCronValidator(),
		AgentCheck: func(agentID string) bool { return true },
	}

	// 创建 ChatHandler
	chatHandler := api.NewChatHandler(
		agentService,
		sessionService,
		agentRepo,
		providerRepo,
		sessionRepo,
		skillRepo,
		encryptor,
		*cfg,
		cronDeps,
		mcpManager,
		messageBus,
	)

	// 设置 API 路由
	router := api.SetupRouter(providerService, agentService, botService, sessionService, cronService, skillService, mcpService, authService, teamService, chatHandler)

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

	// 初始化并启动定时任务调度器（需要在 AgentManager 之前初始化）
	var cronScheduler *cronscheduler.Scheduler
	cronScheduler = cronscheduler.NewScheduler(cronRepo, messageBus, channelManager)
	if err := cronScheduler.Start(ctx); err != nil {
		slog.Error("Failed to start cron scheduler", "error", err)
	} else {
		slog.Info("Cron scheduler started")
		// 将 scheduler 注入到 CronService
		cronService.SetScheduler(cronScheduler)
	}
	defer func() {
		if cronScheduler != nil {
			cronScheduler.Stop()
			slog.Info("Cron scheduler stopped")
		}
	}()

	// 更新 CronToolDeps 的 Scheduler
	cronDeps.Scheduler = cronScheduler

	// 启动 AgentManager 消息循环
	agentManager := service.NewAgentManager(*cfg, db, agentRepo, botRepo, providerRepo, messageBus, encryptor, cronDeps, mcpManager)
	go func() {
		agentManager.MessageLoop(ctx)
	}()

	// 启动所有活跃的 MCP 服务器
	if err := mcpService.StartAllActiveServers(); err != nil {
		slog.Warn("some MCP servers failed to start", "error", err)
	}

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

	// 停止所有 MCP 服务器
	slog.Info("Stopping all MCP servers...")
	mcpService.StopAllServers()

	slog.Info("BeepBot API Server shutdown complete")
}
