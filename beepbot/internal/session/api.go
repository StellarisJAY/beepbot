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

// ApiSession API 模式会话实现，基于数据库持久化
type ApiSession struct {
	sessionID string
	key       string
	agentID   string // 智能体 ID，用于生成会话 Key

	windowSize          int
	maxTokens           int64
	compressionRatio    float64
	compressionKeepSize int

	tokenUsed    int64
	needCompress bool
	summary      string

	repo  repository.SessionRepository
	mutex *sync.RWMutex
}

// NewApiSession 创建或加载 API 会话
func NewApiSession(
	repo repository.SessionRepository,
	key string,
	agentID string,
	botID string,
	windowSize int,
	maxTokens int64,
	compressionRatio float64,
	compressionKeepSize int,
) (*ApiSession, error) {
	// 尝试获取现有会话
	session, err := repo.GetSessionByKey(key)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 创建新会话
	if session == nil {
		session = &types.Session{
			Key:     key,
			AgentID: agentID,
			BotID:   botID,
		}
		if err := repo.CreateSession(session); err != nil {
			return nil, err
		}
	}

	// 从数据库聚合计算窗口内 token 用量
	tokenUsed, err := repo.GetTokenUsageInWindow(session.ID)
	if err != nil {
		slog.Warn("failed to get token usage", "sessionID", session.ID, "error", err)
		tokenUsed = 0
	}

	return &ApiSession{
		sessionID:           session.ID,
		key:                 key,
		agentID:             agentID,
		windowSize:          windowSize,
		maxTokens:           maxTokens,
		compressionRatio:    compressionRatio,
		compressionKeepSize: compressionKeepSize,
		tokenUsed:           tokenUsed,
		summary:             session.Summary,
		repo:                repo,
		mutex:               &sync.RWMutex{},
	}, nil
}

// AppendMessage 添加消息到会话历史
// 返回是否需要压缩上下文
func (s *ApiSession) AppendMessage(message types.Message) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查是否需要淘汰消息（只统计窗口内消息）
	count, err := s.repo.CountMessagesInWindow(s.sessionID)
	if err != nil {
		slog.Error("failed to count messages in window", "sessionID", s.sessionID, "error", err)
	} else if count >= int64(s.windowSize) {
		if err := s.evictMessage(); err != nil {
			slog.Error("failed to evict message", "sessionID", s.sessionID, "error", err)
		}
	}

	// 转换并保存消息（包含 token 信息）
	sessionMsg := s.convertToSessionMessage(message)
	if err := s.repo.AppendMessage(s.sessionID, sessionMsg); err != nil {
		slog.Error("failed to append message", "sessionID", s.sessionID, "error", err)
	}

	// 更新内存中的 token 用量
	if message.Usage != nil {
		s.tokenUsed += message.Usage.InputTokens + message.Usage.OutputTokens
	}

	// 检查是否需要压缩
	if s.maxTokens > 0 && s.tokenUsed >= int64(float64(s.maxTokens)*s.compressionRatio) {
		s.needCompress = true
		slog.Debug("session needs compression", "sessionKey", s.key, "tokenUsed", s.tokenUsed, "maxTokens", s.maxTokens, "ratio", s.compressionRatio)
	}

	return s.needCompress
}

// evictMessage 淘汰最早的消息（标记为窗口外）
// 如果最早的消息是包含 tool_calls 的 assistant 消息，同时标记关联的 tool 结果消息
func (s *ApiSession) evictMessage() error {
	// 获取窗口内最早的消息
	messages, err := s.repo.GetOldestMessagesInWindow(s.sessionID, 1)
	if err != nil || len(messages) == 0 {
		return err
	}

	firstMsg := messages[0]

	// 解析 ToolCalls
	var toolCalls []types.ToolCall
	if firstMsg.Role == types.RoleAssistant && firstMsg.ToolCalls != "" {
		if err := json.Unmarshal([]byte(firstMsg.ToolCalls), &toolCalls); err != nil {
			slog.Warn("failed to unmarshal tool calls", "error", err)
		}
	}

	if len(toolCalls) > 0 {
		// 收集 tool_call_id
		toolCallIDs := make(map[string]bool)
		for _, tc := range toolCalls {
			toolCallIDs[tc.ID] = true
		}

		// 获取窗口内所有消息来查找关联的 tool 结果
		allMessages, err := s.repo.GetMessagesInWindow(s.sessionID, 0)
		if err != nil {
			return err
		}

		// 找到需要标记为窗口外的消息 ID
		var messageIDsToEvict []string
		messageIDsToEvict = append(messageIDsToEvict, firstMsg.ID)

		// 从第二条消息开始查找关联的 tool 结果
		for i := 1; i < len(allMessages); i++ {
			msg := allMessages[i]
			if msg.Role == types.RoleTool && toolCallIDs[msg.ToolCallID] {
				messageIDsToEvict = append(messageIDsToEvict, msg.ID)
				delete(toolCallIDs, msg.ToolCallID)
			} else {
				break // 遇到非匹配消息，停止
			}
		}

		// 标记消息为窗口外
		return s.repo.EvictMessages(s.sessionID, messageIDsToEvict)
	}

	// 普通消息，直接标记最早的一条为窗口外
	return s.repo.EvictOldestMessagesInWindow(s.sessionID, 1)
}

// GetHistory 获取会话历史消息（只返回窗口内消息）
func (s *ApiSession) GetHistory() []types.Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	messages, err := s.repo.GetMessagesInWindow(s.sessionID, s.windowSize)
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
func (s *ApiSession) GetSummary() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.summary
}

// NeedCompress 返回是否需要压缩上下文
func (s *ApiSession) NeedCompress() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.needCompress
}

// SetSummary 设置会话摘要，并清除压缩标记
func (s *ApiSession) SetSummary(summary string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.summary = summary
	s.needCompress = false

	// 只更新 summary 字段，避免覆盖其他字段
	if err := s.repo.UpdateSessionSummary(s.sessionID, summary); err != nil {
		slog.Error("failed to update session summary", "sessionID", s.sessionID, "error", err)
	}
}

// ClearHistory 清空历史消息（标记所有窗口内消息为窗口外）
func (s *ApiSession) ClearHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取所有窗口内消息并标记为窗口外
	messages, err := s.repo.GetMessagesInWindow(s.sessionID, 0)
	if err != nil {
		slog.Error("failed to get messages for clearing", "sessionID", s.sessionID, "error", err)
		return
	}

	if len(messages) > 0 {
		messageIDs := make([]string, len(messages))
		for i, m := range messages {
			messageIDs[i] = m.ID
		}
		if err := s.repo.EvictMessages(s.sessionID, messageIDs); err != nil {
			slog.Error("failed to evict messages", "sessionID", s.sessionID, "error", err)
		}
	}

	// 清空后重新计算 token 用量（应为 0）
	s.tokenUsed, _ = s.repo.GetTokenUsageInWindow(s.sessionID)
}

// Compress 压缩历史消息，保留最近的消息
// 使用 compressionKeepSize 作为保留数量
// 返回被移除的消息，用于生成摘要
func (s *ApiSession) Compress() []types.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取窗口内所有消息
	messages, err := s.repo.GetMessagesInWindow(s.sessionID, 0)
	if err != nil {
		slog.Error("failed to get messages for compression", "sessionID", s.sessionID, "error", err)
		return nil
	}

	keepCount := s.compressionKeepSize
	if keepCount <= 0 {
		keepCount = 5 // 默认保留 5 条
	}

	if len(messages) <= keepCount {
		return nil
	}

	// 计算需要淘汰的消息数量
	evictCount := len(messages) - keepCount

	// 标记旧消息为窗口外
	if err := s.repo.EvictOldestMessagesInWindow(s.sessionID, evictCount); err != nil {
		slog.Error("failed to evict old messages during compression", "sessionID", s.sessionID, "error", err)
		return nil
	}

	// 重新从数据库聚合计算窗口内 token 用量
	s.tokenUsed, _ = s.repo.GetTokenUsageInWindow(s.sessionID)

	// 返回被淘汰的消息（用于生成摘要）
	removed := make([]types.Message, 0, evictCount)
	for i := 0; i < evictCount; i++ {
		removed = append(removed, s.convertToMessage(messages[i]))
	}
	return removed
}

// GetTokenUsage 返回当前 token 用量
func (s *ApiSession) GetTokenUsage() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.tokenUsed
}

// GetMaxTokens 返回 token 上限
func (s *ApiSession) GetMaxTokens() int64 {
	return s.maxTokens
}

// GetSessionKey 生成会话 Key
func (s *ApiSession) GetSessionKey(channelID string, groupID string, userID string) string {
	return GetApiSessionKey(s.agentID, channelID, groupID, userID)
}

// GetApiSessionKey 生成 API 模式的会话 Key
// 包含 agentID 以隔离不同智能体的会话
func GetApiSessionKey(agentID string, channelID string, groupID string, userID string) string {
	return fmt.Sprintf("%s:%s:%s:%s", agentID, channelID, groupID, userID)
}

// convertToSessionMessage 将 types.Message 转换为 types.SessionMessage
func (s *ApiSession) convertToSessionMessage(msg types.Message) *types.SessionMessage {
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
func (s *ApiSession) convertToMessage(msg types.SessionMessage) types.Message {
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
