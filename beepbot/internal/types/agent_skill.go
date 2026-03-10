package types

import "time"

// AgentSkill 智能体与技能的关联表
// 注意：不使用数据库外键约束，通过应用层逻辑维护数据一致性
type AgentSkill struct {
	ID        string    `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	AgentID   string    `json:"agent_id" gorm:"column:agent_id;type:varchar(64);not null;uniqueIndex:idx_agent_skill"`
	SkillID   string    `json:"skill_id" gorm:"column:skill_id;type:varchar(64);not null;uniqueIndex:idx_agent_skill"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (AgentSkill) TableName() string {
	return "agent_skills"
}
