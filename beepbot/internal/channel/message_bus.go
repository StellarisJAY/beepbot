package channel

import "context"

type MessageType byte

const (
	TextMessage MessageType = iota + 1
	MarkdownMsg
)

type InboundMessage struct {
	Channel    string
	UserID     string // 发送者ID（QQ用户ID / 飞书 open_id）
	GroupID    string // 群聊ID（仅群聊时有值，用于@回复等场景）
	ChatID     string // 会话ID（飞书 chat_id，用于主动推送）
	MessageID  string
	SessionKey string
	Content    string
	AgentID    string // 定时任务直接指定智能体ID，机器人消息为空
}

type OutboundMessage struct {
	Channel          string
	UserID           string // 目标用户ID
	GroupID          string // 目标群ID（群聊时）
	ChatID           string // 会话ID（飞书主动推送时使用）
	Content          string
	File             string
	InboundMessageID string
	MessageType      MessageType
	Iteration        int
}

type MessageBus struct {
	inputChan  chan InboundMessage
	outputChan chan OutboundMessage
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		inputChan:  make(chan InboundMessage, 1),
		outputChan: make(chan OutboundMessage, 1),
	}
}

func (mb *MessageBus) PublishInbound(ctx context.Context, msg InboundMessage) {
	select {
	case <-ctx.Done():
		return
	case mb.inputChan <- msg:
	}
}

func (mb *MessageBus) ConsumeInbound(ctx context.Context) (InboundMessage, bool) {
	select {
	case <-ctx.Done():
		return InboundMessage{}, false
	case msg := <-mb.inputChan:
		return msg, true
	}
}

func (mb *MessageBus) PublishOutbound(ctx context.Context, msg OutboundMessage) {
	select {
	case <-ctx.Done():
		return
	case mb.outputChan <- msg:
	}
}

func (mb *MessageBus) ConsumeOutbound(ctx context.Context) (OutboundMessage, bool) {
	select {
	case <-ctx.Done():
		return OutboundMessage{}, false
	case msg := <-mb.outputChan:
		return msg, true
	}
}
