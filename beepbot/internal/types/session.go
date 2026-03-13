package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionType 会话类型
type SessionType string

const (
	// SessionTypeChat 聊天会话（默认）
	SessionTypeChat SessionType = "chat"
	// SessionTypeCron 定时任务会话
	SessionTypeCron SessionType = "cron"
)

// IMSessionContext IM 会话上下文信息，用于存储和调试
type IMSessionContext struct {
	UserID  string `json:"user_id,omitempty"`  // 用户 ID
	GroupID string `json:"group_id,omitempty"` // 群 ID
	ChatID  string `json:"chat_id,omitempty"`  // 会话 ID（飞书）
}

// Value 实现 driver.Valuer 接口，用于 GORM 写入数据库
func (c *IMSessionContext) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口，用于 GORM 从数据库读取
func (c *IMSessionContext) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan IMSessionContext: expected []byte")
	}
	return json.Unmarshal(bytes, c)
}

type Session struct {
	ID                string            `gorm:"column:id;type:varchar(64);primaryKey"`
	Key               string            `gorm:"column:key;type:varchar(255);uniqueIndex;not null"`
	AgentID           string            `gorm:"column:agent_id;type:varchar(64);not null;index"`
	BotID             string            `gorm:"column:bot_id;type:varchar(64);not null;index"`
	SessionType       SessionType       `gorm:"column:session_type;type:varchar(32);not null;default:'chat';index"` // 会话类型：chat/cron
	CronJobID         *string           `gorm:"column:cron_job_id;type:varchar(64);index"`                          // 定时任务 ID（仅定时任务会话）
	IMContext         *IMSessionContext `gorm:"column:im_context;type:jsonb"`                                        // IM 会话上下文，方便调试
	Summary           string            `gorm:"column:summary;type:text"`
	LastContextTokens int64             `gorm:"column:last_context_tokens;type:bigint;default:0"` // 最后一次 LLM 调用的上下文 token 大小
	CreatedAt         time.Time         `gorm:"column:created_at;type:timestamptz;autoCreateTime"`
	UpdatedAt         time.Time         `gorm:"column:updated_at;type:timestamptz;autoUpdateTime"`
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

// SessionQuery 会话查询参数
type SessionQuery struct {
	SessionType SessionType // 会话类型筛选
	Platform    BotPlatform // 平台筛选（需要关联 Bot 表查询）
}
