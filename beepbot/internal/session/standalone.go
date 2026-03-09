package session

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/types"
)

type StandaloneSession struct {
	Key        string
	history    []types.Message
	Summary    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	windowSize int
	mutex      *sync.RWMutex // 读写锁，防止可能的并发问题

	tokenUsed         int64   // 上下文token大小
	maxTokens         int64   // 上下文token上限
	compressionRatio  float64 // 压缩触发阈值比例
	needCompress      bool    // 标记需要压缩
	compressionPrompt string  // 压缩提示词，用于让 LLM 生成摘要
}

type StandaloneSessionManager struct {
	sessions         map[string]*StandaloneSession
	windowSize       int
	maxTokens        int64
	compressionRatio float64
}

func NewStandaloneSessionManager(config config.StandaloneConfig) *StandaloneSessionManager {
	windowSize := config.MemoryConfig.WindowSize
	windowSize = max(defaultWindowSize, windowSize)

	compressionRatio := config.MemoryConfig.CompressionRatio
	if compressionRatio <= 0 || compressionRatio >= 1 {
		compressionRatio = defaultCompressionRatio
	}
	maxTokens := config.MemoryConfig.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}
	return &StandaloneSessionManager{
		sessions:         make(map[string]*StandaloneSession),
		windowSize:       windowSize,
		maxTokens:        maxTokens,
		compressionRatio: compressionRatio,
	}
}

func (m *StandaloneSessionManager) GetOrCreateSession(key string) *StandaloneSession {
	session, ok := m.sessions[key]
	if !ok {
		session = &StandaloneSession{
			Key:              key,
			history:          make([]types.Message, 0, m.windowSize),
			Summary:          "",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			windowSize:       m.windowSize,
			mutex:            &sync.RWMutex{},
			maxTokens:        m.maxTokens,
			compressionRatio: m.compressionRatio,
		}
		m.sessions[key] = session
	}
	return session
}

// AppendMessage 添加消息到会话历史，并更新 token 用量
// 返回是否需要压缩上下文
func (s *StandaloneSession) AppendMessage(message types.Message) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.history) == s.windowSize {
		// FIFO 丢弃第一条记录
		s.evictMessage()
	}
	s.history = append(s.history, message)
	s.UpdatedAt = time.Now()

	// 更新 token 用量
	if message.Usage != nil {
		s.tokenUsed += message.Usage.InputTokens + message.Usage.OutputTokens
	}

	// 检查是否需要压缩
	if s.maxTokens > 0 && s.tokenUsed >= int64(float64(s.maxTokens)*s.compressionRatio) {
		s.needCompress = true
		slog.Debug("session needs compression", "sessionKey", s.Key, "tokenUsed", s.tokenUsed, "maxTokens", s.maxTokens, "ratio", s.compressionRatio)
	}

	return s.needCompress
}

func (s *StandaloneSession) evictMessage() {
	if len(s.history) == 0 {
		return
	}

	firstMsg := s.history[0]
	// 如果第一条是 assistant 消息且包含 tool_calls
	// 需要同时删除所有对应的 tool 结果消息
	if firstMsg.Role == types.RoleAssistant && len(firstMsg.ToolCalls) > 0 {
		// 收集所有 tool_call_id
		toolCallIDs := make(map[string]bool)
		for _, tc := range firstMsg.ToolCalls {
			toolCallIDs[tc.ID] = true
		}

		// 找到需要删除的边界
		// 删除 assistant 消息 + 所有紧接着的匹配 tool 结果
		deleteCount := 1
		// 减少第一条消息的 token 用量
		s.reduceTokenUsage(&firstMsg)

		for i := 1; i < len(s.history); i++ {
			msg := s.history[i]
			if msg.Role == types.RoleTool && toolCallIDs[msg.ToolCallID] {
				// 减少被删除的 tool 消息的 token 用量
				s.reduceTokenUsage(&msg)
				deleteCount++
				delete(toolCallIDs, msg.ToolCallID)
			} else {
				break // 遇到非匹配消息，停止
			}
		}
		s.history = s.history[deleteCount:]
	} else {
		// 普通消息，直接删除第一条
		// 减少被删除消息的 token 用量
		s.reduceTokenUsage(&firstMsg)
		s.history = s.history[1:]
	}
}

// reduceTokenUsage 减少消息对应的 token 用量
func (s *StandaloneSession) reduceTokenUsage(msg *types.Message) {
	if msg.Usage != nil {
		s.tokenUsed = min(0, s.tokenUsed-msg.Usage.InputTokens-msg.Usage.OutputTokens)
	}
}

func (s *StandaloneSession) GetHistory() []types.Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.history
}

func (s *StandaloneSession) GetSummary() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Summary
}

// NeedCompress 返回是否需要压缩上下文
func (s *StandaloneSession) NeedCompress() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.needCompress
}

// SetSummary 设置会话摘要，并清除压缩标记
func (s *StandaloneSession) SetSummary(summary string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Summary = summary
	s.needCompress = false
}

// ClearHistory 清空历史消息，用于压缩后重置
func (s *StandaloneSession) ClearHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.history = make([]types.Message, 0, s.windowSize)
	s.tokenUsed = 0
}

// Compress 压缩历史消息，保留最近的消息
// keepCount: 保留最近的消息数量
// 返回被移除的消息，用于生成摘要
func (s *StandaloneSession) Compress(keepCount int) []types.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.history) <= keepCount {
		return nil
	}

	// 保留最近的消息
	removed := s.history[:len(s.history)-keepCount]
	s.history = s.history[len(s.history)-keepCount:]

	// 重新计算 token 用量（简化处理，重置为 0）
	s.tokenUsed = 0
	s.needCompress = false

	slog.Info("session compressed", "sessionKey", s.Key, "removedCount", len(removed), "remainingCount", len(s.history))

	return removed
}

// GetTokenUsage 返回当前 token 用量
func (s *StandaloneSession) GetTokenUsage() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.tokenUsed
}

// GetMaxTokens 返回 token 上限
func (s *StandaloneSession) GetMaxTokens() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.maxTokens
}

// GetSessionKey 生成会话 Key
// Standalone模式下只有一个全局智能体，所以sessionKey中没有智能体id
func (s *StandaloneSession) GetSessionKey(channelID string, groupID string, userID string) string {
	return GetSessionKeyForStandalone(channelID, groupID, userID)
}

func GetSessionKeyForStandalone(channelID string, groupID string, userID string) string {
	return fmt.Sprintf("%s:%s:%s", channelID, groupID, userID)
}
