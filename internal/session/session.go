package session

import (
	"fmt"
	"time"

	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/types"
)

type Session struct {
	Key       string
	History   []types.Message
	Summary   string
	CreatedAt time.Time
	UpdatedAt time.Time
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
			Key:       key,
			History:   make([]types.Message, 0, m.windowSize),
			Summary:   "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		m.sessions[key] = session
	}
	return session
}

func (s *Session) AppendMessage(message types.Message) {
	s.History = append(s.History, message)
	s.UpdatedAt = time.Now()
	// TODO 维护Summary和窗口大小
}

func (s *Session) GetHistory() []types.Message {
	return s.History
}

func (s *Session) GetSummary() string {
	return s.Summary
}

func GetSessionKey(channelID string, userID string) string {
	return fmt.Sprintf("%s:%s", channelID, userID)
}
