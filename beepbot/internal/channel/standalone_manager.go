package channel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/StellarisJAY/beepbot/internal/config"
)

// StandaloneChannelManager Standalone 模式的 Channel 管理器
// 从配置文件初始化 Channel
type StandaloneChannelManager struct {
	channels map[string]Channel
	bus      *MessageBus
	config   config.StandaloneConfig
}

// NewStandaloneChannelManager 创建 Standalone 模式的 Channel 管理器
func NewStandaloneChannelManager(config config.StandaloneConfig, bus *MessageBus) *StandaloneChannelManager {
	return &StandaloneChannelManager{
		channels: make(map[string]Channel),
		bus:      bus,
		config:   config,
	}
}

// DispatchOutbound 分发出站消息
func (c *StandaloneChannelManager) DispatchOutbound(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, ok := c.bus.ConsumeOutbound(ctx)
			if !ok {
				continue
			}
			channel := c.channels[msg.Channel]
			if channel == nil || !channel.IsAvailable() {
				slog.Warn("send message to non-exist or not available channel", "channel", msg.Channel)
				continue
			}
			if !channel.IsAllowed(msg.UserID) {
				slog.Warn("send message to non-allowed user", "channel", msg.Channel, "user", msg.UserID)
				continue
			}
			if err := channel.Send(ctx, msg); err != nil {
				slog.Error("send message failed", "channel", msg.Channel, "user", msg.UserID, "error", err)
				continue
			}
		}
	}
}

// InitChannels 从配置初始化 Channel
func (c *StandaloneChannelManager) InitChannels(ctx context.Context, config config.ChannelConfig) error {
	// 注册系统消息通道
	c.channels["system"] = newSystemChannel()
	// qq机器人消息通道
	if config.QQ != nil {
		qqBotChannel := NewQQBotChannelFromConfig(c.config, c.bus)
		c.channels[qqBotChannel.ID()] = qqBotChannel
	}

	for id, channel := range c.channels {
		if err := channel.Start(ctx); err != nil {
			return fmt.Errorf("start channel %s failed, err:%w", id, err)
		}
	}
	return nil
}

// StopAllChannels 停止所有 Channel
func (c *StandaloneChannelManager) StopAllChannels() {
	for id, ch := range c.channels {
		ch.Stop()
		slog.Info("channel stopped", "channel_id", id)
	}
}
