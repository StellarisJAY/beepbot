package session

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// ReActSession ReAct智能体会话实现，基于数据库持久化
type ReActSession struct {
	sessionID           string
	key                 string
	agentID             string                  // 智能体 ID，用于生成会话 Key
	maxTokens           int64                   // 上下文最大token数量
	compressionRatio    float64                 // 触发压缩的窗口阈值
	compressionKeepSize int                     // 压缩后保留的消息条数
	contextTokens       int64                   // 当前上下文 token 大小（最后一次 LLM 调用的 InputTokens）
	needCompress        bool                    // 是否需要压缩标记
	summary             string                  // 会话摘要
	cronJobID           *string                 // 定时任务 ID
	imContext           *types.IMSessionContext // IM 会话上下文

	repo  repository.SessionRepository // 会话数据库
	mutex *sync.RWMutex
}

// NewApiSession 创建或加载 API 会话
func NewApiSession(
	repo repository.SessionRepository,
	key string,
	agentID string,
	botID string,
	sessionType types.SessionType,
	cronJobID *string,
	imContext *types.IMSessionContext,
	maxTokens int64,
	compressionRatio float64,
	compressionKeepSize int,
) (*ReActSession, error) {
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

	return &ReActSession{
		sessionID:           session.ID,
		key:                 key,
		agentID:             agentID,
		maxTokens:           maxTokens,
		compressionRatio:    compressionRatio,
		compressionKeepSize: compressionKeepSize,
		contextTokens:       session.LastContextTokens, // 从数据库加载上下文 token 大小
		summary:             session.Summary,
		cronJobID:           session.CronJobID,
		imContext:           session.IMContext,
		repo:                repo,
		mutex:               &sync.RWMutex{},
	}, nil
}

// AppendMessage 添加消息到会话历史
// 返回是否需要压缩上下文
func (s *ReActSession) AppendMessage(message types.Message) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 转换并保存消息（包含 token 信息）
	sessionMsg := s.convertToSessionMessage(message)
	if err := s.repo.AppendMessage(s.sessionID, sessionMsg); err != nil {
		slog.Error("failed to append message", "sessionID", s.sessionID, "error", err)
	}

	// 更新上下文 token 大小
	// 只有 assistant 消息才有 Usage（来自 LLM API 响应）
	// InputTokens 已经包含了发送给 LLM 的所有内容的 token 数（系统提示词 + 历史消息 + 当前消息）
	if message.Usage != nil && message.Role == types.RoleAssistant {
		s.contextTokens = message.Usage.InputTokens
		// 持久化到数据库
		if err := s.repo.UpdateSessionContextTokens(s.sessionID, s.contextTokens); err != nil {
			slog.Error("failed to update context tokens", "sessionID", s.sessionID, "error", err)
		}
	}

	// 检查是否需要压缩
	if s.maxTokens > 0 && s.contextTokens >= int64(float64(s.maxTokens)*s.compressionRatio) {
		s.needCompress = true
		slog.Debug("session needs compression", "sessionKey", s.key, "contextTokens", s.contextTokens, "maxTokens", s.maxTokens, "ratio", s.compressionRatio)
	}

	return s.needCompress
}

// GetHistory 获取会话历史消息（只返回窗口内消息）
func (s *ReActSession) GetHistory() []types.Message {
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
func (s *ReActSession) GetSummary() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.summary
}

// NeedCompress 返回是否需要压缩上下文
func (s *ReActSession) NeedCompress() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.needCompress
}

// SetSummary 设置会话摘要，并清除压缩标记
func (s *ReActSession) SetSummary(summary string) {
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
func (s *ReActSession) ClearHistory() {
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

	// 清空后重置上下文 token 大小
	s.contextTokens = 0
	if err := s.repo.UpdateSessionContextTokens(s.sessionID, 0); err != nil {
		slog.Error("failed to reset context tokens", "sessionID", s.sessionID, "error", err)
	}
}

// Compress 压缩历史消息，保留最近的消息
// 使用 compressionKeepSize 作为保留数量
// 返回被移除的消息，用于生成摘要
func (s *ReActSession) Compress() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 标记旧消息为窗口外
	if err := s.repo.EvictMessagesInWindow(s.sessionID); err != nil {
		slog.Error("failed to evict old messages during compression", "sessionID", s.sessionID, "error", err)
		return
	}

	// 压缩后重置上下文 token 大小，下次 LLM 调用会更新准确值
	s.contextTokens = 0
	if err := s.repo.UpdateSessionContextTokens(s.sessionID, 0); err != nil {
		slog.Error("failed to reset context tokens after compression", "sessionID", s.sessionID, "error", err)
	}
}

// GetTokenUsage 返回当前上下文 token 大小
func (s *ReActSession) GetTokenUsage() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.contextTokens
}

// GetMaxTokens 返回 token 上限
func (s *ReActSession) GetMaxTokens() int64 {
	return s.maxTokens
}

// GetSessionKey 生成会话 Key
func (s *ReActSession) GetSessionKey(sessionType types.SessionType, channelID string, chatID string, userID string) string {
	return GetApiSessionKey(sessionType, s.agentID, channelID, chatID, userID)
}

// GetApiSessionKey 生成 API 模式的会话 Key
// 包含 sessionType 和 agentID 以隔离不同类型和智能体的会话
// 格式：{sessionType}:{agentID}:{channelID}:{chatID}:{userID}
// - sessionType: 会话类型（chat/cron）
// - chatID: 飞书为 chat_id，QQ 为群ID或空（私聊）
func GetApiSessionKey(sessionType types.SessionType, agentID string, channelID string, chatID string, userID string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", sessionType, agentID, channelID, chatID, userID)
}

// GetSessionID 返回会话 ID
func (s *ReActSession) GetSessionID() string {
	return s.sessionID
}

// GetCronJobID 返回定时任务 ID
func (s *ReActSession) GetCronJobID() *string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cronJobID
}

// GetIMContext 返回 IM 会话上下文
func (s *ReActSession) GetIMContext() *types.IMSessionContext {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.imContext
}

// GetSessionWorkDir 返回会话的工作目录
// 格式：{baseWorkDir}/sessions/{sha256(sessionKey)[:16]}
func GetSessionWorkDir(baseWorkDir string, sessionKey string) string {
	hash := sha256.Sum256([]byte(sessionKey))
	shortHash := hex.EncodeToString(hash[:8]) // 16 字符
	return filepath.Join(baseWorkDir, "sessions", shortHash)
}

// EnsureSessionWorkDir 确保会话工作目录存在，返回完整路径
func EnsureSessionWorkDir(baseWorkDir string, sessionKey string) (string, error) {
	workDir := GetSessionWorkDir(baseWorkDir, sessionKey)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", err
	}
	return workDir, nil
}

// convertToSessionMessage 将 types.Message 转换为 types.SessionMessage
func (s *ReActSession) convertToSessionMessage(msg types.Message) *types.SessionMessage {
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
func (s *ReActSession) convertToMessage(msg types.SessionMessage) types.Message {
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
