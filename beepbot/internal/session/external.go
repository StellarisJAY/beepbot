package session

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// ExternalSession 外部智能体会话，只记录消息，不做上下文计算
type ExternalSession struct {
	sessionID string
	key       string
	agentID   string // 智能体 ID，用于生成会话 Key

	cronJobID *string                 // 定时任务 ID
	imContext *types.IMSessionContext // IM 会话上下文

	repo  repository.SessionRepository
	mutex *sync.RWMutex
}

// NewExternalSession 创建或加载外部会话
func NewExternalSession(
	repo repository.SessionRepository,
	key string,
	agentID string,
	botID string,
	sessionType types.SessionType,
	cronJobID *string,
	imContext *types.IMSessionContext,
) (*ExternalSession, error) {
	// 尝试获取现有会话
	session, err := repo.GetSessionByKey(key)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 创建新会话
	if session == nil {
		session = &types.Session{
			Key:         key,
			AgentID:     agentID,
			BotID:       botID,
			SessionType: sessionType,
			CronJobID:   cronJobID,
			IMContext:   imContext,
		}
		if err := repo.CreateSession(session); err != nil {
			return nil, err
		}
	}

	return &ExternalSession{
		sessionID: session.ID,
		key:       key,
		agentID:   agentID,
		cronJobID: session.CronJobID,
		imContext: session.IMContext,
		repo:      repo,
		mutex:     &sync.RWMutex{},
	}, nil
}

// AppendMessage 添加消息到会话历史
// 返回是否需要压缩上下文
func (s *ExternalSession) AppendMessage(message types.Message) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 转换并保存消息（包含 token 信息）
	sessionMsg := s.convertToSessionMessage(message)
	if err := s.repo.AppendMessage(s.sessionID, sessionMsg); err != nil {
		slog.Error("failed to append message", "sessionID", s.sessionID, "error", err)
	}

	return false
}

// GetHistory 获取会话历史消息
func (s *ExternalSession) GetHistory() []types.Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	messages, err := s.repo.GetMessagesInWindow(s.sessionID, -1)
	if err != nil {
		slog.Error("failed to get messages in window", "sessionID", s.sessionID, "error", err)
		return nil
	}

	result := make([]types.Message, 0, len(messages))
	for _, m := range messages {
		result = append(result, s.convertToMessage(m))
	}
	return result
}

// GetSummary 获取会话摘要
func (s *ExternalSession) GetSummary() string {
	return ""
}

// NeedCompress 返回是否需要压缩上下文
func (s *ExternalSession) NeedCompress() bool {
	return false
}

// SetSummary 设置会话摘要
func (s *ExternalSession) SetSummary(summary string) {}

// ClearHistory
func (s *ExternalSession) ClearHistory() {}

// Compress 压缩历史消息，保留最近的消息
// 使用 compressionKeepSize 作为保留数量
// 返回被移除的消息，用于生成摘要
func (s *ExternalSession) Compress() {}

// GetTokenUsage 返回当前上下文 token 大小
func (s *ExternalSession) GetTokenUsage() int64 {
	return 0
}

// GetMaxTokens 返回 token 上限
func (s *ExternalSession) GetMaxTokens() int64 {
	return 0
}

// GetSessionKey 生成会话 Key
func (s *ExternalSession) GetSessionKey(sessionType types.SessionType, channelID string, chatID string, userID string) string {
	return GetExternalSessionKey(sessionType, s.agentID, channelID, chatID, userID)
}

// GetExternalSessionKey 生成 API 模式的会话 Key
// 包含 sessionType 和 agentID 以隔离不同类型和智能体的会话
// 格式：{sessionType}:{agentID}:{channelID}:{chatID}:{userID}
// - sessionType: 会话类型（chat/cron）
// - chatID: 飞书为 chat_id，QQ 为群ID或空（私聊）
func GetExternalSessionKey(sessionType types.SessionType, agentID string, channelID string, chatID string, userID string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", sessionType, agentID, channelID, chatID, userID)
}

// GetSessionID 返回会话 ID
func (s *ExternalSession) GetSessionID() string {
	return s.sessionID
}

// GetCronJobID 返回定时任务 ID
func (s *ExternalSession) GetCronJobID() *string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cronJobID
}

// GetIMContext 返回 IM 会话上下文
func (s *ExternalSession) GetIMContext() *types.IMSessionContext {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.imContext
}

// convertToSessionMessage 将 types.Message 转换为 types.SessionMessage
func (s *ExternalSession) convertToSessionMessage(msg types.Message) *types.SessionMessage {
	toolCallsJSON, _ := json.Marshal(msg.ToolCalls)

	var inputTokens, outputTokens, totalTokens int64
	if msg.Usage != nil {
		inputTokens = msg.Usage.InputTokens
		outputTokens = msg.Usage.OutputTokens
		totalTokens = msg.Usage.TotalTokens
	}

	return &types.SessionMessage{
		Role:         msg.Role,
		Content:      msg.Content,
		ToolCallID:   msg.ToolCallID,
		ToolCalls:    string(toolCallsJSON),
		FinishReason: msg.FinishReason,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
		InWindow:     true, // 新消息默认在窗口内
	}
}

// convertToMessage 将 types.SessionMessage 转换为 types.Message
func (s *ExternalSession) convertToMessage(msg types.SessionMessage) types.Message {
	var toolCalls []types.ToolCall
	if msg.ToolCalls != "" {
		if err := json.Unmarshal([]byte(msg.ToolCalls), &toolCalls); err != nil {
			slog.Warn("failed to unmarshal tool calls", "error", err)
		}
	}

	var usage *types.TokenUsage
	if msg.TotalTokens > 0 {
		usage = &types.TokenUsage{
			InputTokens:  msg.InputTokens,
			OutputTokens: msg.OutputTokens,
			TotalTokens:  msg.TotalTokens,
		}
	}

	return types.Message{
		Role:         msg.Role,
		Content:      msg.Content,
		ToolCallID:   msg.ToolCallID,
		ToolCalls:    toolCalls,
		FinishReason: msg.FinishReason,
		Usage:        usage,
	}
}
