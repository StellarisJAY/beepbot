package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID        string    `gorm:"column:id;type:varchar(64);primaryKey"`
	Key       string    `gorm:"column:key;type:varchar(255);uniqueIndex;not null"`
	AgentID   string    `gorm:"column:agent_id;type:varchar(64);not null;index"`
	BotID     string    `gorm:"column:bot_id;type:varchar(64);not null;index"`
	Summary   string    `gorm:"column:summary;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;autoUpdateTime"`
}

// BeforeCreate 在创建记录前自动设置ID和时间
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		id, _ := uuid.NewV7()
		s.ID = id.String()
	}
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	return nil
}

type SessionMessage struct {
	ID           string    `gorm:"column:id;type:varchar(64);primaryKey"`
	SessionID    string    `gorm:"column:session_id;type:varchar(64);not null;index"`
	Role         string    `gorm:"column:role;type:varchar(32);not null"`
	Content      string    `gorm:"column:content;type:text"`
	ToolCallID   string    `gorm:"column:tool_call_id;type:varchar(64)"`
	ToolCalls    string    `gorm:"column:tool_calls;type:jsonb"`
	FinishReason string    `gorm:"column:finish_reason;type:varchar(32)"`
	InputTokens  int64     `gorm:"column:input_tokens;type:bigint;default:0"`
	OutputTokens int64     `gorm:"column:output_tokens;type:bigint;default:0"`
	TotalTokens  int64     `gorm:"column:total_tokens;type:bigint;default:0"`
	InWindow     bool      `gorm:"column:in_window;type:boolean;default:true;index"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime"`
}

// BeforeCreate 在创建记录前自动设置ID和CreatedAt
func (m *SessionMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		id, _ := uuid.NewV7()
		m.ID = id.String()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}
