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

	// 会话推送信息（用于智能体处理完成后推送响应）
	// 只有用户通过智能体对话创建的定时任务才有推送信息
	PushChannel *string `json:"push_channel,omitempty" gorm:"column:push_channel;type:varchar(64)"`   // 推送渠道类型：qq/feishu
	PushBotID   *string `json:"push_bot_id,omitempty" gorm:"column:push_bot_id;type:varchar(64)"`     // 推送机器人ID
	PushUserID  *string `json:"push_user_id,omitempty" gorm:"column:push_user_id;type:varchar(64)"`   // 推送目标用户ID
	PushGroupID *string `json:"push_group_id,omitempty" gorm:"column:push_group_id;type:varchar(64)"` // 推送目标群ID（群聊时）
	PushChatID  *string `json:"push_chat_id,omitempty" gorm:"column:push_chat_id;type:varchar(64)"`   // 会话ID（飞书 chat_id）

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// CronQuery 定时任务查询参数
type CronQuery struct {
	Name    string // 名称模糊搜索
	Enabled *bool  // 是否启用（使用指针区分空值和 false）
}

// TableName 指定表名
func (CronJob) TableName() string {
	return "cron_jobs"
}
