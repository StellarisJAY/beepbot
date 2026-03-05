package tool

import (
	"context"

	"github.com/StellarisJAY/beepbot/internal/channel"
)

type MessageTool struct {
	bus *channel.MessageBus
}

func NewMessageTool(bus *channel.MessageBus) *MessageTool {
	return &MessageTool{
		bus: bus,
	}
}

func (t *MessageTool) Name() string {
	return "send_message"
}

func (t *MessageTool) Description() string {
	return "Send message to user"
}

func (t *MessageTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"content": map[string]any{
				"description": "the text content of the message",
			},
		},
		"required": []string{"content"},
	}
}

func (t *MessageTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	channelName := ctx.Value("channel").(string)
	userID := ctx.Value("userID").(string)
	content := params["content"].(string)
	t.bus.PublishOutbound(ctx, channel.OutboundMessage{
		Channel: channelName,
		UserID:  userID,
		Content: content,
	})
	return "success", nil
}
