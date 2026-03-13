package session

import (
	"sync"

	"github.com/StellarisJAY/beepbot/internal/types"
)

// InMemorySession 内存会话实现
// 用于子智能体调用，不持久化到数据库
type InMemorySession struct {
	messages []types.Message
	summary  string
	mutex    *sync.RWMutex
}

// NewInMemorySession 创建内存会话
func NewInMemorySession() *InMemorySession {
	return &InMemorySession{
		messages: make([]types.Message, 0),
		mutex:    &sync.RWMutex{},
	}
}

// AppendMessage 添加消息到会话历史
func (s *InMemorySession) AppendMessage(message types.Message) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.messages = append(s.messages, message)
	return false // 内存会话不需要压缩
}

// GetHistory 获取会话历史消息
func (s *InMemorySession) GetHistory() []types.Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	result := make([]types.Message, len(s.messages))
	copy(result, s.messages)
	return result
}

// GetSummary 获取会话摘要
func (s *InMemorySession) GetSummary() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.summary
}

// NeedCompress 内存会话不需要压缩
func (s *InMemorySession) NeedCompress() bool {
	return false
}

// SetSummary 设置会话摘要
func (s *InMemorySession) SetSummary(summary string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.summary = summary
}

// ClearHistory 清空历史消息
func (s *InMemorySession) ClearHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.messages = make([]types.Message, 0)
}

// Compress 内存会话不支持压缩
func (s *InMemorySession) Compress() []types.Message {
	return nil
}

// GetTokenUsage 返回当前 token 用量（内存会话不追踪）
func (s *InMemorySession) GetTokenUsage() int64 {
	return 0
}

// GetMaxTokens 返回 token 上限（内存会话不限制）
func (s *InMemorySession) GetMaxTokens() int64 {
	return 0
}

// GetSessionKey 生成会话 Key（内存会话不需要）
func (s *InMemorySession) GetSessionKey(sessionType types.SessionType, channelID string, chatID string, userID string) string {
	return ""
}

// GetSessionID 返回会话 ID（内存会话返回空）
func (s *InMemorySession) GetSessionID() string {
	return ""
}

// GetCronJobID 返回定时任务 ID（内存会话返回 nil）
func (s *InMemorySession) GetCronJobID() *string {
	return nil
}

// GetIMContext 返回 IM 会话上下文（内存会话返回 nil）
func (s *InMemorySession) GetIMContext() *types.IMSessionContext {
	return nil
}