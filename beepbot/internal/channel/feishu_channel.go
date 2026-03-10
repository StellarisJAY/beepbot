package channel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// FeishuChannel 飞书机器人 Channel 实现
type FeishuChannel struct {
	BaseChannel

	appID      string
	appSecret  string
	encryptKey string

	client   *lark.Client
	wsClient *larkws.Client
	cancel   context.CancelFunc

	allowedUsers  map[string]struct{}
	allowedGroups map[string]struct{}
}

// NewFeishuChannel 在 API 模式下通过 channel 配置创建
func NewFeishuChannel(appID, appSecret, encryptKey string, id string, bus *MessageBus, allowedUsers, allowedGroups []string) Channel {
	allowedUsersMap := make(map[string]struct{})
	allowedGroupsMap := make(map[string]struct{})
	for _, u := range allowedUsers {
		allowedUsersMap[u] = struct{}{}
	}
	for _, g := range allowedGroups {
		allowedGroupsMap[g] = struct{}{}
	}

	return &FeishuChannel{
		BaseChannel:   NewBaseChannel(id, bus),
		appID:         appID,
		appSecret:     appSecret,
		encryptKey:    encryptKey,
		allowedUsers:  allowedUsersMap,
		allowedGroups: allowedGroupsMap,
	}
}

// Start 启动飞书 Channel
func (c *FeishuChannel) Start(ctx context.Context) error {
	if c.appID == "" || c.appSecret == "" {
		return errors.New("must provide AppID and AppSecret")
	}

	// 创建 API Client（用于发送消息）
	c.client = lark.NewClient(c.appID, c.appSecret,
		lark.WithLogLevel(larkcore.LogLevelInfo),
	)

	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// 创建事件处理器
	eventHandler := dispatcher.NewEventDispatcher(c.encryptKey, "").
		OnP2MessageReceiveV1(c.handleMessage)

	// 创建 WebSocket 长连接客户端
	c.wsClient = larkws.NewClient(c.appID, c.appSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)

	// 启动长连接
	go func() {
		if err := c.wsClient.Start(ctx); err != nil {
			slog.Error("feishu websocket client error", "error", err)
			c.available = false
		}
	}()

	c.available = true
	slog.Info("feishu bot channel started", "app_id", c.appID)
	return nil
}

// Stop 停止飞书 Channel
func (c *FeishuChannel) Stop() {
	slog.Info("stopping feishu bot channel")
	c.available = false
	if c.cancel != nil {
		c.cancel()
	}
}

// Send 发送消息
func (c *FeishuChannel) Send(ctx context.Context, message OutboundMessage) error {
	if !c.IsAvailable() {
		return errors.New("feishu bot channel is not available")
	}

	switch message.MessageType {
	case TextMessage:
		return c.sendTextMessage(ctx, message)
	case MarkdownMsg:
		return c.sendMarkdownMessage(ctx, message)
	default:
		return c.sendTextMessage(ctx, message)
	}
}

// sendTextMessage 发送文本消息
func (c *FeishuChannel) sendTextMessage(ctx context.Context, message OutboundMessage) error {
	// 构建文本消息内容
	content, err := json.Marshal(map[string]string{"text": message.Content})
	if err != nil {
		return fmt.Errorf("marshal text content failed: %w", err)
	}

	// 构建消息请求
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			ReceiveId(c.getReceiveID(message)).
			Content(string(content)).
			Build()).
		Build()

	// 发送消息
	resp, err := c.client.Im.Message.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("send feishu text message failed: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("send feishu text message failed, code: %d, msg: %s", resp.Code, resp.Msg)
	}
	return nil
}

// sendMarkdownMessage 发送 Markdown 消息（飞书使用 Post 富文本消息）
func (c *FeishuChannel) sendMarkdownMessage(ctx context.Context, message OutboundMessage) error {
	// 飞书的 Post 消息格式
	postContent := map[string]interface{}{
		"zh_cn": map[string]interface{}{
			"title": "",
			"content": [][]map[string]interface{}{
				{
					{"tag": "text", "text": message.Content},
				},
			},
		},
	}
	content, err := json.Marshal(postContent)
	if err != nil {
		return fmt.Errorf("marshal post content failed: %w", err)
	}

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypePost).
			ReceiveId(c.getReceiveID(message)).
			Content(string(content)).
			Build()).
		Build()

	resp, err := c.client.Im.Message.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("send feishu post message failed: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("send feishu post message failed, code: %d, msg: %s", resp.Code, resp.Msg)
	}
	return nil
}

// getReceiveID 获取接收者 ID
func (c *FeishuChannel) getReceiveID(message OutboundMessage) string {
	if message.GroupID != "" {
		return message.GroupID
	}
	return message.UserID
}

// handleMessage 处理接收到的消息
func (c *FeishuChannel) handleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	if event.Event == nil || event.Event.Message == nil {
		return nil
	}

	msg := event.Event.Message
	sender := event.Event.Sender

	// 获取发送者 ID
	var senderID string
	if sender != nil && sender.SenderId != nil {
		if sender.SenderId.OpenId != nil {
			senderID = *sender.SenderId.OpenId
		}
		if senderID == "" && sender.SenderId.UserId != nil {
			senderID = *sender.SenderId.UserId
		}
	}

	// 获取消息内容
	var content string
	if msg.Content != nil {
		content = *msg.Content
	}

	// 获取消息 ID
	var messageID string
	if msg.MessageId != nil {
		messageID = *msg.MessageId
	}

	// 获取聊天 ID
	var chatID string
	if msg.ChatId != nil {
		chatID = *msg.ChatId
	}

	slog.Debug("receive feishu message", "sender", senderID, "chat_id", chatID)

	// 检查是否在允许列表中
	if !c.IsAllowed(senderID) {
		slog.Debug("sender not in allowed list", "sender", senderID)
		return nil
	}

	// 解析文本消息内容
	textContent := parseFeishuTextContent(content)

	// 构建入站消息
	inboundMsg := InboundMessage{
		Channel:   c.ID(),
		UserID:    senderID,
		GroupID:   chatID,
		MessageID: messageID,
		Content:   textContent,
	}

	// 发布到 MessageBus
	if err := c.BaseChannel.HandleMessage(ctx, inboundMsg); err != nil {
		return fmt.Errorf("handle feishu message failed: %w", err)
	}

	return nil
}

// parseFeishuTextContent 解析飞书文本消息内容
func parseFeishuTextContent(content string) string {
	// 飞书文本消息格式: {"text":"消息内容"}
	type TextContent struct {
		Text string `json:"text"`
	}
	var tc TextContent
	if err := json.Unmarshal([]byte(content), &tc); err == nil {
		return tc.Text
	}
	return content
}

// ID 返回 Channel ID
func (c *FeishuChannel) ID() string {
	return c.BaseChannel.ID()
}

// IsAllowed 检查发送者是否在允许列表中
func (c *FeishuChannel) IsAllowed(senderID string) bool {
	// 如果没有配置允许列表，则允许所有人
	if len(c.allowedUsers) == 0 && len(c.allowedGroups) == 0 {
		return true
	}

	_, userAllowed := c.allowedUsers[senderID]
	return userAllowed
}

// IsAvailable 返回 Channel 是否可用
func (c *FeishuChannel) IsAvailable() bool {
	return c.BaseChannel.IsAvailable()
}
