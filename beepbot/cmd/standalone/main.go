package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/StellarisJAY/beepbot/internal/agent"
	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/heartbeat"
	"github.com/StellarisJAY/beepbot/internal/logger"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "config.json", "Path to the config file")
}

func ensureWorkingDir(config *config.StandaloneConfig) error {
	workingDir := config.AgentConfig.WorkingDir
	workingDir = strings.TrimSpace(workingDir)
	if workingDir == "" {
		workingDir = "./beepbot"
	}
	workingDirAbs, err := filepath.Abs(workingDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(workingDirAbs, 0755); err != nil {
		return err
	}
	config.AgentConfig.WorkingDir = workingDirAbs
	slog.Info("Created working directory", "path", workingDirAbs)
	return nil
}

func ensureDataDir(config *config.StandaloneConfig) error {
	dataDir := config.DataDir
	dataDir = strings.TrimSpace(dataDir)
	if dataDir == "" {
		dataDir = "./beepbot"
	}
	dataDirAbs, err := filepath.Abs(dataDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDirAbs, 0755); err != nil {
		return err
	}
	config.DataDir = dataDirAbs
	slog.Info("Beepbot data directory", "path", dataDirAbs)
	return nil
}

func main() {
	flag.Parse()
	// 加载配置文件
	config, err := config.FromConfigFile[config.StandaloneConfig](configFile)
	if err != nil {
		panic(err)
	}

	// 初始化日志
	logger.InitLogger(config.Logging)

	// 创建公共数据目录
	ensureDataDir(config)
	// 创建工作目录
	ensureWorkingDir(config)

	// 监听系统信号
	ctx, cancel := context.WithCancel(context.Background())
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		<-signalCh
		cancel()
	}()

	// 创建消息总线
	messageBus := channel.NewMessageBus()
	// 创建消息源管理器
	channelManager := channel.NewStandaloneChannelManager(*config, messageBus)
	if err := channelManager.InitChannels(ctx, config.ChannelConfig); err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}

	// 创建智能体运行器
	agentRun, err := agent.NewStandaloneRun(*config, messageBus)
	if err != nil {
		panic(err)
	}

	// 分发出站消息
	wg.Add(1)
	go func() {
		defer wg.Done()
		channelManager.DispatchOutbound(ctx)
	}()

	// 启动智能体消息循环
	wg.Add(1)
	go func() {
		defer wg.Done()
		agentRun.StandaloneMessageLoop(ctx)
	}()

	if config.HeartBeat.Enable {
		// 心跳机制
		hb, err := heartbeat.NewHeartBeat(*config, func(ctx context.Context, prompt string) {
			// 触发心跳，向智能体发送一条要求心跳检查的消息
			messageBus.PublishInbound(ctx, channel.InboundMessage{
				Channel:   "system",    // 系统通道
				UserID:    "heartbeat", // 发送者为心跳
				MessageID: "heartbeat",
				Content:   prompt, // 心跳检查提示词
			})
		})
		if err != nil {
			panic(err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			hb.Start(ctx)
		}()
	}

	slog.Info("started beepbot...")
	wg.Wait()
	slog.Info("Beepbot shutdown")
}
