package channel

import "context"

type MessageType byte

const (
	TextMessage MessageType = iota + 1
	MarkdownMsg
)

type InboundMessage struct {
	Channel    string
	UserID     string
	SessionKey string
	Content    string
}

type OutboundMessage struct {
	Channel     string
	UserID      string
	Content     string
	File        string
	MessageType MessageType
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
