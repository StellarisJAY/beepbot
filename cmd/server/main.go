package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/StellarisJAY/beepbot/internal/agent"
	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/db"
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
	config, err := config.FromConfigFile(configFile)
	if err != nil {
		panic(err)
	}

	// 初始化数据库
	if err := db.InitDatabase(*config); err != nil {
		panic(err)
	}

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
	channelManager := channel.NewChannelManager(*config, messageBus)
	if err := channelManager.InitChannels(ctx, config.ChannelConfig); err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}

	// 创建智能体运行器
	agentRun, err := agent.NewAgentRun(*config, messageBus)
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
		agentRun.MessageLoop(ctx)
	}()
	slog.Info("started beepbot...")
	wg.Wait()
	slog.Info("Beepbot shutdown")
}
