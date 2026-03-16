package react

import (
	"context"
	"log/slog"

	"github.com/StellarisJAY/beepbot/internal/channel"
)

// OutputHook 输出钩子，处理智能体的消息输出
// 用于抽象不同场景下的输出行为（对话场景发送消息，子智能体场景收集结果）
type OutputHook interface {
	// OnError 调用模型失败时调用
	OnError(ctx context.Context, err error)

	// OnResponse 有最终响应内容时调用
	// 当智能体完成对话（没有工具调用）时触发
	OnResponse(ctx context.Context, content string)

	// OnIntermediateContent 有中间内容时调用
	// 当智能体在工具调用前有思考内容时触发
	OnIntermediateContent(ctx context.Context, content string)
}

// BusOutputHook 对话场景的输出钩子
// 通过 MessageBus 发送消息给用户
type BusOutputHook struct {
	bus          *channel.MessageBus
	channelName  string
	userID       string
	groupID      string
	chatID       string
	inboundMsgID string
}

// NewBusOutputHook 创建对话场景的输出钩子
func NewBusOutputHook(
	bus *channel.MessageBus,
	channelName string,
	userID string,
	groupID string,
	chatID string,
	inboundMsgID string,
) *BusOutputHook {
	return &BusOutputHook{
		bus:          bus,
		channelName:  channelName,
		userID:       userID,
		groupID:      groupID,
		chatID:       chatID,
		inboundMsgID: inboundMsgID,
	}
}

func (h *BusOutputHook) OnError(ctx context.Context, err error) {
	h.bus.PublishOutbound(ctx, channel.OutboundMessage{
		Channel:          h.channelName,
		UserID:           h.userID,
		GroupID:          h.groupID,
		ChatID:           h.chatID,
		Content:          "调用模型失败, 请重试...",
		MessageType:      channel.TextMessage,
		InboundMessageID: h.inboundMsgID,
	})
}

func (h *BusOutputHook) OnResponse(ctx context.Context, content string) {
	h.bus.PublishOutbound(ctx, channel.OutboundMessage{
		Channel:          h.channelName,
		UserID:           h.userID,
		GroupID:          h.groupID,
		ChatID:           h.chatID,
		Content:          content,
		MessageType:      channel.MarkdownMsg,
		InboundMessageID: h.inboundMsgID,
	})
}

func (h *BusOutputHook) OnIntermediateContent(ctx context.Context, content string) {
	h.bus.PublishOutbound(ctx, channel.OutboundMessage{
		Channel:          h.channelName,
		UserID:           h.userID,
		GroupID:          h.groupID,
		ChatID:           h.chatID,
		Content:          content,
		MessageType:      channel.MarkdownMsg,
		InboundMessageID: h.inboundMsgID,
	})
}

// CollectorHook sub-agent场景的输出钩子
// 仅收集最终结果，不发送消息
type CollectorHook struct {
	result  string
	err     error
	hasData bool
}

// NewCollectorHook 创建子智能体场景的输出钩子
func NewCollectorHook() *CollectorHook {
	return &CollectorHook{}
}

func (h *CollectorHook) OnError(ctx context.Context, err error) {
	h.err = err
}

func (h *CollectorHook) OnResponse(ctx context.Context, content string) {
	h.result = content
	h.hasData = true
}

func (h *CollectorHook) OnIntermediateContent(ctx context.Context, content string) {
	// 子智能体场景忽略中间内容
	slog.Debug("sub-agent intermediate output", "content", content)
}

// GetResult 获取收集的结果
func (h *CollectorHook) GetResult() string {
	return h.result
}

// GetError 获取错误信息
func (h *CollectorHook) GetError() error {
	return h.err
}

// HasData 检查是否有数据
func (h *CollectorHook) HasData() bool {
	return h.hasData
}
