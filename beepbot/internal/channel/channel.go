package channel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/StellarisJAY/beepbot/internal/config"
)

type Channel interface {
	ID() string
	IsAvailable() bool
	Send(ctx context.Context, message OutboundMessage) error
	IsAllowed(senderID string) bool
	Start(ctx context.Context) error
	Stop()
}

type ChannelManager struct {
	channels map[string]Channel
	bus      *MessageBus
	config   config.StandaloneConfig
}

type BaseChannel struct {
	id        string
	available bool
	bus       *MessageBus
}

func NewChannelManager(config config.StandaloneConfig, bus *MessageBus) *ChannelManager {
	return &ChannelManager{
		channels: make(map[string]Channel),
		bus:      bus,
		config:   config,
	}
}

func (c *ChannelManager) DispatchOutbound(ctx context.Context) {
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

func (c *ChannelManager) InitChannels(ctx context.Context, config config.ChannelConfig) error {
	// 注册系统消息通道
	c.channels["system"] = newSystemChannel()
	// qq机器人消息通道
	if config.QQ != nil {
		qqBotChannel := NewQQBotChannel(c.config, c.bus)
		c.channels[qqBotChannel.ID()] = qqBotChannel
	}

	for id, channel := range c.channels {
		if err := channel.Start(ctx); err != nil {
			return fmt.Errorf("start channel %s failed, err:%w", id, err)
		}
	}
	return nil
}

func NewBaseChannel(id string, bus *MessageBus) BaseChannel {
	return BaseChannel{
		id:        id,
		available: true,
		bus:       bus,
	}
}

func (c *BaseChannel) ID() string {
	return c.id
}

func (c *BaseChannel) IsAvailable() bool {
	return c.available
}

func (c *BaseChannel) HandleMessage(ctx context.Context, msg InboundMessage) error {
	msg.Channel = c.ID()
	c.bus.PublishInbound(ctx, msg)
	return nil
}
