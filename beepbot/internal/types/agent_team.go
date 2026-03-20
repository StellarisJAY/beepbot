package types

import (
	"time"
)

// TeamStatus 团队状态
type TeamStatus string

const (
	TeamStatusActive   TeamStatus = "active"
	TeamStatusInactive TeamStatus = "inactive"
)

// MemberRole 成员角色
type MemberRole string

const (
	MemberRoleLeader MemberRole = "leader"
	MemberRoleMember MemberRole = "member"
)

// AgentTeam 智能体团队
type AgentTeam struct {
	ID          string     `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name        string     `json:"name" gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	Description string     `json:"description" gorm:"column:description;type:text"`
	LeaderID    string     `json:"leader_id" gorm:"column:leader_id;type:varchar(64);index"`
	Status      TeamStatus `json:"status" gorm:"column:status;type:varchar(16);default:active;index"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (AgentTeam) TableName() string {
	return "agent_teams"
}

// AgentTeamMember 团队成员
type AgentTeamMember struct {
	TeamID      string     `json:"team_id" gorm:"column:team_id;primaryKey;type:varchar(64)"`
	AgentID     string     `json:"agent_id" gorm:"column:agent_id;primaryKey;type:varchar(64)"`
	MemberName  string     `json:"member_name" gorm:"column:member_name;type:varchar(64);not null"`
	Role        MemberRole `json:"role" gorm:"column:role;type:varchar(32);not null"`
	Description string     `json:"description" gorm:"column:description;type:text"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (AgentTeamMember) TableName() string {
	return "agent_team_members"
}

// TeamQuery 团队查询参数
type TeamQuery struct {
	Name   string     // 名称模糊搜索
	Status TeamStatus // 状态筛选
}

// TeamMemberBrief 团队成员简要信息
type TeamMemberBrief struct {
	AgentID     string `json:"agent_id"`
	AgentName   string `json:"agent_name"`
	MemberName  string `json:"member_name"`
	Role        string `json:"role"`
	Description string `json:"description"`
}