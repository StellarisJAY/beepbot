package types

import (
	"time"
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
	ProviderID        string      `json:"provider_id" gorm:"column:provider_id;type:varchar(64);index"` // 可为空，编辑时配置
	Model             string      `json:"model" gorm:"column:model;type:varchar(128)"`                  // 可为空，编辑时配置
	SystemPrompt      string      `json:"system_prompt" gorm:"column:system_prompt;type:text"`
	Temperature       float32     `json:"temperature" gorm:"column:temperature;type:real;default:0.7"`
	MaxIterations     int         `json:"max_iterations" gorm:"column:max_iterations;type:integer;default:50"`
	MaxOutputTokens   int64       `json:"max_output_tokens" gorm:"column:max_output_tokens;type:integer;default:4096"`
	WorkingDir        string      `json:"working_dir" gorm:"column:working_dir;type:varchar(512)"` // 可为空，自动生成或编辑时配置
	ContextWindowSize int         `json:"context_window_size" gorm:"column:context_window_size;type:integer;default:20"`
	WindowSize        int         `json:"window_size" gorm:"column:window_size;type:integer;default:20"`
	CompressionRatio  float64     `json:"compression_ratio" gorm:"column:compression_ratio;type:real;default:0.7"`
	ContextMaxTokens  int64       `json:"context_max_tokens" gorm:"column:context_max_tokens;type:integer;default:4096"`
	Status            AgentStatus `json:"status" gorm:"column:status;type:varchar(16);default:active;index"`
	CreatedAt         time.Time   `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt         time.Time   `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (Agent) TableName() string {
	return "agents"
}
