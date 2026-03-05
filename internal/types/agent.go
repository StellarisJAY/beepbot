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
	ID               string         `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name             string         `json:"name" gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	Description      string         `json:"description" gorm:"column:description;type:text"`
	ProviderID       string         `json:"provider_id" gorm:"column:provider_id;type:varchar(64);not null;index"`
	Model            string         `json:"model" gorm:"column:model;type:varchar(128);not null"`
	SystemPrompt     string         `json:"system_prompt" gorm:"column:system_prompt;type:text"`
	Temperature      float32        `json:"temperature" gorm:"column:temperature;type:real;default:0.7"`
	MaxIterations    int            `json:"max_iterations" gorm:"column:max_iterations;type:integer;default:50"`
	MaxTokens        int64          `json:"max_tokens" gorm:"column:max_tokens;type:integer;default:4096"`
	WorkingDir       string         `json:"working_dir" gorm:"column:working_dir;type:varchar(512);not null"`
	MemoryWindowSize int            `json:"memory_window_size" gorm:"column:memory_window_size;type:integer;default:20"`
	EnableShell      bool           `json:"enable_shell" gorm:"column:enable_shell;type:boolean;default:true"`
	ForbiddenCmds    datatypes.JSON `json:"forbidden_commands" gorm:"column:forbidden_commands;type:jsonb"`
	ShellTimeout     string         `json:"shell_timeout" gorm:"column:shell_timeout;type:varchar(32);default:30s"`
	Status           AgentStatus    `json:"status" gorm:"column:status;type:varchar(16);default:active;index"`
	CreatedAt        time.Time      `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`

	// 关联
	Provider *Provider      `json:"provider,omitempty" gorm:"foreignKey:ProviderID"`
	Channels []AgentChannel `json:"channels,omitempty" gorm:"foreignKey:AgentID"`
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
