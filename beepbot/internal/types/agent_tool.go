package types

import "time"

// AgentTool 智能体与工具的关联表
// 注意：不使用数据库外键约束，通过应用层逻辑维护数据一致性
type AgentTool struct {
	ID        string    `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	AgentID   string    `json:"agent_id" gorm:"column:agent_id;type:varchar(64);not null;uniqueIndex:idx_agent_tool"`
	ToolName  string    `json:"tool_name" gorm:"column:tool_name;type:varchar(64);not null;uniqueIndex:idx_agent_tool"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (AgentTool) TableName() string {
	return "agent_tools"
}