package types

import (
	"time"

	"gorm.io/datatypes"
)

// AgentStatus 智能体状态
type AgentStatus string

const (
	AgentStatusActive   AgentStatus = "active"
	AgentStatusInactive AgentStatus = "inactive"
)

// Agent 智能体配置
type Agent struct {
	ID                string      `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name              string      `json:"name" gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	Description       string      `json:"description" gorm:"column:description;type:text"`
	ProviderID        string      `json:"provider_id" gorm:"column:provider_id;type:varchar(64);not null;index"`
	Model             string      `json:"model" gorm:"column:model;type:varchar(128);not null"`
	SystemPrompt      string      `json:"system_prompt" gorm:"column:system_prompt;type:text"`
	Temperature       float32     `json:"temperature" gorm:"column:temperature;type:real;default:0.7"`
	MaxIterations     int         `json:"max_iterations" gorm:"column:max_iterations;type:integer;default:50"`
	MaxOutputTokens   int64       `json:"max_output_tokens" gorm:"column:max_output_tokens;type:integer;default:4096"`
	WorkingDir        string      `json:"working_dir" gorm:"column:working_dir;type:varchar(512);not null"`
	ContextWindowSize int         `json:"context_window_size" gorm:"column:context_window_size;type:integer;default:20"`
	WindowSize        int         `json:"window_size" gorm:"column:window_size;type:integer;default:20"`
	CompressionRatio  float64     `json:"compression_ratio" gorm:"column:compression_ratio;type:real;default:0.7"`
	ContextMaxTokens  int64       `json:"context_max_tokens" gorm:"column:context_max_tokens;type:integer;default:4096"`
	Status            AgentStatus `json:"status" gorm:"column:status;type:varchar(16);default:active;index"`
	CreatedAt         time.Time   `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt         time.Time   `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`

	// 关联
	Provider *Provider `json:"provider,omitempty" gorm:"foreignKey:ProviderID"`
}

// TableName 指定表名
func (Agent) TableName() string {
	return "agents"
}

// AgentChannel 智能体渠道绑定
type AgentChannel struct {
	ID                string         `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	AgentID           string         `json:"agent_id" gorm:"column:agent_id;type:varchar(64);not null;index"`
	ChannelType       string         `json:"channel_type" gorm:"column:channel_type;type:varchar(32);not null"`
	ChannelIdentifier string         `json:"channel_identifier" gorm:"column:channel_identifier;type:varchar(128)"`
	Config            datatypes.JSON `json:"config" gorm:"column:config;type:jsonb"`
	CreatedAt         time.Time      `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`

	// 关联
	Agent *Agent `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

// TableName 指定表名
func (AgentChannel) TableName() string {
	return "agent_channels"
}
