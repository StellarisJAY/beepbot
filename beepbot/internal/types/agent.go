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
	ID                  string      `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name                string      `json:"name" gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	Description         string      `json:"description" gorm:"column:description;type:text"`
	ProviderID          string      `json:"provider_id" gorm:"column:provider_id;type:varchar(64);index"` // 可为空，编辑时配置
	Model               string      `json:"model" gorm:"column:model;type:varchar(128)"`                  // 可为空，编辑时配置
	SystemPrompt        string      `json:"system_prompt" gorm:"column:system_prompt;type:text"`
	Temperature         float32     `json:"temperature" gorm:"column:temperature;type:real;default:0.7"`
	MaxIterations       int         `json:"max_iterations" gorm:"column:max_iterations;type:integer;default:50"`
	MaxOutputTokens     int64       `json:"max_output_tokens" gorm:"column:max_output_tokens;type:integer;default:4096"`
	WorkingDir          string      `json:"working_dir" gorm:"column:working_dir;type:varchar(512)"` // 可为空，自动生成或编辑时配置
	CompressionRatio    float64     `json:"compression_ratio" gorm:"column:compression_ratio;type:real;default:0.7"`
	CompressionKeepSize int         `json:"compression_keep_size" gorm:"column:compression_keep_size;type:integer;default:5"`
	ContextMaxTokens    int64       `json:"context_max_tokens" gorm:"column:context_max_tokens;type:integer;default:4096"`
	Status              AgentStatus `json:"status" gorm:"column:status;type:varchar(16);default:active;index"`
	// UseAllSkills 是否使用系统所有技能
	// true: 使用所有技能（默认）
	// false: 仅使用关联表中的技能
	UseAllSkills bool `json:"use_all_skills" gorm:"column:use_all_skills;type:boolean;default:true"`
	// UseAllTools 是否使用所有工具
	// true: 使用所有工具（默认）
	// false: 仅使用关联表中的工具
	UseAllTools bool `json:"use_all_tools" gorm:"column:use_all_tools;type:boolean;default:true"`
	// Callable 是否可作为子智能体被调用
	Callable bool `json:"callable" gorm:"column:callable;type:boolean;default:false;index"`
	// CallableDescription 作为子智能体时的工具描述（供 LLM 理解用途）
	CallableDescription string    `json:"callable_description" gorm:"column:callable_description;type:text"`
	CreatedAt           time.Time `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// AgentQuery 智能体查询参数
type AgentQuery struct {
	Name   string      // 名称模糊搜索
	Status AgentStatus // 状态筛选
}

// TableName 指定表名
func (Agent) TableName() string {
	return "agents"
}

// UsageStatsPoint 单个时间点的统计数据
type UsageStatsPoint struct {
	Time          time.Time `json:"time"`
	SessionCount  int64     `json:"session_count"`
	MessageCount  int64     `json:"message_count"`
	InputTokens   int64     `json:"input_tokens"`
	OutputTokens  int64     `json:"output_tokens"`
	TotalTokens   int64     `json:"total_tokens"`
}

// UsageStatsResponse 用量统计响应
type UsageStatsResponse struct {
	Points []UsageStatsPoint `json:"points"`
}
