package types

import (
	"time"
)

// CronJobStatus 定时任务状态
type CronJobStatus bool

const (
	CronJobEnabled  CronJobStatus = true
	CronJobDisabled CronJobStatus = false
)

// CronJob 定时任务
type CronJob struct {
	ID             string        `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name           string        `json:"name" gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	CronExpression string        `json:"cron_expression" gorm:"column:cron_expression;type:varchar(64);not null"`
	AgentID        string        `json:"agent_id" gorm:"column:agent_id;type:varchar(64);index"`
	Message        string        `json:"message" gorm:"column:message;type:text"`
	Enabled        CronJobStatus `json:"enabled" gorm:"column:enabled;type:boolean;default:true;index"`
	CreatedAt      time.Time     `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt      time.Time     `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (CronJob) TableName() string {
	return "cron_jobs"
}
