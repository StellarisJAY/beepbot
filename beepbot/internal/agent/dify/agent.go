package dify

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/types"
)

type DifyRunner struct {
	baseURL      string
	apiKey       string
	responseMode string

	bus     *channel.MessageBus
	timeout time.Duration
}

func NewDifyRunner(baseURL, apiKey, responseMode string, bus *channel.MessageBus) *DifyRunner {
	timeout := 120 * time.Second
	return &DifyRunner{
		baseURL:      baseURL,
		apiKey:       apiKey,
		responseMode: responseMode,
		bus:          bus,
		timeout:      timeout,
	}
}

func (d *DifyRunner) RunWithMessage(ctx context.Context, sess session.Session, message channel.InboundMessage) error {
	client := http.DefaultClient
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	path, _ := url.JoinPath(d.baseURL, "chat-messages")

	requestBody := map[string]any{
		"query":         message.Content,
		"inputs":        map[string]any{},
		"response_mode": d.responseMode,
		// "conversation_id": sess.GetSessionID(), // TODO save conversation_id in session
		"user": "beepbot",
	}
	body, _ := json.Marshal(requestBody)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Bearer "+d.apiKey)
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	var responseBody map[string]any
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return err
	}
	slog.Info("response", "resp", string(data))
	responseMessage := channel.OutboundMessage{
		Channel:          message.Channel,
		ChatID:           message.ChatID,
		InboundMessageID: message.MessageID,
		Content:          "",
		GroupID:          message.GroupID,
		UserID:           message.UserID,
	}

	if answer, ok := responseBody["answer"].(string); ok {
		responseMessage.Content = answer
	}

	sess.AppendMessage(types.Message{
		Content: responseMessage.Content,
		Role:    types.RoleAssistant,
	})

	d.bus.PublishOutbound(ctx, responseMessage)
	return nil
}
