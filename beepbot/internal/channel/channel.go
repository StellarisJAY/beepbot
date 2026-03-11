package channel

import (
	"context"
)

// Channel 接口定义了消息渠道的基本操作
type Channel interface {
	ID() string
	IsAvailable() bool
	Send(ctx context.Context, message OutboundMessage) error
	IsAllowed(senderID string) bool
	Start(ctx context.Context) error
	Stop()
	// CanPushProactively 返回是否支持主动推送消息（无需 InboundMessageID）
	// QQ 机器人不支持主动推送，需要被动回复的 MessageID
	// 飞书机器人支持主动推送，可以通过 chat_id 发送
	CanPushProactively() bool
}

// BaseChannel 基础 Channel 实现
type BaseChannel struct {
	id        string
	available bool
	bus       *MessageBus
}

// NewBaseChannel 创建基础 Channel
func NewBaseChannel(id string, bus *MessageBus) BaseChannel {
	return BaseChannel{
		id:        id,
		available: true,
		bus:       bus,
	}
}

// ID 返回 Channel ID
func (c *BaseChannel) ID() string {
	return c.id
}

// IsAvailable 返回 Channel 是否可用
func (c *BaseChannel) IsAvailable() bool {
	return c.available
}

// HandleMessage 处理入站消息，发布到 MessageBus
func (c *BaseChannel) HandleMessage(ctx context.Context, msg InboundMessage) error {
	msg.Channel = c.ID()
	c.bus.PublishInbound(ctx, msg)
	return nil
}
