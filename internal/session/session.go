package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/types"
)

type Session struct {
	Key        string
	history    []types.Message
	Summary    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	windowSize int
	mutex      *sync.RWMutex // 读写锁，防止可能的并发问题
}

type SessionManager struct {
	sessions   map[string]*Session
	windowSize int
}

func NewSessionManager(config config.Config) *SessionManager {
	windowSize := config.MemoryConfig.WindowSize
	return &SessionManager{
		sessions:   make(map[string]*Session),
		windowSize: windowSize,
	}
}

func (m *SessionManager) GetOrCreateSession(key string) *Session {
	session, ok := m.sessions[key]
	if !ok {
		session = &Session{
			Key:        key,
			history:    make([]types.Message, 0, m.windowSize),
			Summary:    "",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			windowSize: m.windowSize,
			mutex:      &sync.RWMutex{},
		}
		m.sessions[key] = session
	}
	return session
}

func (s *Session) AppendMessage(message types.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(s.history) == s.windowSize {
		// FIFO 丢弃第一条记录
		s.evictMessage()
	}
	s.history = append(s.history, message)
	s.UpdatedAt = time.Now()
}

func (s *Session) evictMessage() {
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
		for i := 1; i < len(s.history); i++ {
			msg := s.history[i]
			if msg.Role == types.RoleTool && toolCallIDs[msg.ToolCallID] {
				deleteCount++
				delete(toolCallIDs, msg.ToolCallID)
			} else {
				break // 遇到非匹配消息，停止
			}
		}
		s.history = s.history[deleteCount:]
	} else {
		// 普通消息，直接删除第一条
		s.history = s.history[1:]
	}
}

func (s *Session) GetHistory() []types.Message {
	return s.history
}

func (s *Session) GetSummary() string {
	return s.Summary
}

func GetSessionKey(channelID string, groupID string, userID string) string {
	return fmt.Sprintf("%s:%s:%s", channelID, groupID, userID)
}
