package channel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
	"golang.org/x/oauth2"
)

type QQBotChannel struct {
	BaseChannel

	tokenSource    oauth2.TokenSource
	api            openapi.OpenAPI
	sessionManager botgo.SessionManager

	cancel context.CancelFunc

	appID     string
	appSecret string
}

// NewQQBotChannel 在api模式下通过channel的配置创建, id为botID，也就是channelID
func NewQQBotChannel(appId, appSecret string, id string, bus *MessageBus) Channel {
	channelID := id
	return &QQBotChannel{
		BaseChannel: NewBaseChannel(channelID, bus),
		appID:       appId,
		appSecret:   appSecret,
	}
}

func (c *QQBotChannel) Start(ctx context.Context) error {
	if c.appID == "" || c.appSecret == "" {
		return errors.New("must provide AppID and AppSecret")
	}

	// 创建机器人token
	c.tokenSource = token.NewQQBotTokenSource(&token.QQBotCredentials{
		AppID:     c.appID,
		AppSecret: c.appSecret,
	})

	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// 启动token刷新
	if err := token.StartRefreshAccessToken(ctx, c.tokenSource); err != nil {
		return fmt.Errorf("start qq channel failed, start refresh token error: %w", err)
	}

	c.api = botgo.NewOpenAPI(c.appID, c.tokenSource)

	// 注册消息回调
	intent := event.RegisterHandlers(
		c.CreateC2CMessageHandler(ctx),
		c.CreateGroupMessageHandler(ctx),
	)

	// 创建websocket连接
	ws, err := c.api.WS(ctx, map[string]string{}, "")
	if err != nil {
		return fmt.Errorf("start qq channel failed, create ws error: %w", err)
	}

	// 启动会话管理器，单独创建一个goroutine避免阻塞
	c.sessionManager = botgo.NewSessionManager()
	go func() {
		if err := c.sessionManager.Start(ws, c.tokenSource, &intent); err != nil {
			slog.Error("start qq channel failed, start session error", "error", err)
		}
		c.available = false
	}()
	c.available = true

	slog.Info("qq bot channel started")
	return nil
}

func (c *QQBotChannel) Stop() {
	slog.Info("stopping qq bot channel")
	c.available = false
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *QQBotChannel) Send(ctx context.Context, message OutboundMessage) error {
	if !c.IsAvailable() {
		return errors.New("qq bot channel is not available")
	}
	msgToCreate := &dto.MessageToCreate{
		MsgID: message.InboundMessageID,
	}

	// 替换 . 防止qq把文件名识别成url拦截
	message.Content = strings.ReplaceAll(message.Content, ".", "·")

	switch message.MessageType {
	case MarkdownMsg:
		msgToCreate.MsgType = dto.MarkdownMsg
		msgToCreate.Markdown = &dto.Markdown{
			Content: message.Content,
		}
	case TextMessage:
		msgToCreate.MsgType = dto.TextMsg
		msgToCreate.Content = message.Content
	default:
		msgToCreate.MsgType = dto.TextMsg
		msgToCreate.Content = message.Content
	}

	var err error
	if message.GroupID != "" {
		_, err = c.api.PostGroupMessage(ctx, message.GroupID, msgToCreate)
	} else {
		_, err = c.api.PostC2CMessage(ctx, message.UserID, msgToCreate)
	}
	if err != nil {
		return fmt.Errorf("send qq message failed, post c2c message error: %w", err)
	}
	return nil
}

func (c *QQBotChannel) CreateC2CMessageHandler(ctx context.Context) event.C2CMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSC2CMessageData) error {
		if data.Author == nil || data.Author.ID == "" {
			return errors.New("received message with no sender ID")
		}
		senderID := data.Author.ID
		slog.Debug("receive QQ c2c message", "userID", senderID)
		message := InboundMessage{
			Channel:   c.ID(),
			UserID:    senderID,
			Content:   data.Content,
			MessageID: data.ID,
		}
		if err := c.BaseChannel.HandleMessage(ctx, message); err != nil {
			return fmt.Errorf("handle qq message failed, handle message error: %w", err)
		}
		return nil
	}
}

func (c *QQBotChannel) CreateGroupMessageHandler(ctx context.Context) event.GroupATMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSGroupATMessageData) error {
		senderID := data.Author.ID
		groupID := data.GroupID
		slog.Debug("receive QQ Group @beepbot", "groupID", groupID, "senderID", senderID)
		message := InboundMessage{
			Channel:   c.ID(),
			GroupID:   groupID,
			UserID:    senderID,
			Content:   data.Content,
			MessageID: data.ID,
		}
		if err := c.BaseChannel.HandleMessage(ctx, message); err != nil {
			return fmt.Errorf("handle qq group message failed, error: %w", err)
		}
		return nil
	}
}

func (c *QQBotChannel) ID() string {
	return c.BaseChannel.ID()
}

func (c *QQBotChannel) IsAllowed(senderID string) bool {
	return true
}

func (c *QQBotChannel) IsAvailable() bool {
	return c.BaseChannel.IsAvailable()
}

// CanPushProactively 返回是否支持主动推送
// QQ 机器人不支持主动推送，需要被动回复的 MessageID
func (c *QQBotChannel) CanPushProactively() bool {
	return false
}
