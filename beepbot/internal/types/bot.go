package types

import (
	"time"

	"gorm.io/datatypes"
)

// BotStatus 机器人状态
type BotStatus string

const (
	BotStatusActive   BotStatus = "active"
	BotStatusInactive BotStatus = "inactive"
)

// BotPlatform 机器人平台类型
type BotPlatform string

const (
	BotPlatformQQ     BotPlatform = "qq"
	BotPlatformFeishu BotPlatform = "feishu"
	// 未来可扩展: discord, telegram, wechat 等
)

// Bot IM机器人配置
type Bot struct {
	ID          string         `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name        string         `json:"name" gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	Description string         `json:"description" gorm:"column:description;type:text"`
	Platform    BotPlatform    `json:"platform" gorm:"column:platform;type:varchar(32);not null;index"`
	Identifier  string         `json:"identifier" gorm:"column:identifier;type:varchar(128)"`
	Config      datatypes.JSON `json:"config" gorm:"column:config;type:jsonb"`
	AgentID     *string        `json:"agent_id" gorm:"column:agent_id;type:varchar(64);index"`
	Status      BotStatus      `json:"status" gorm:"column:status;type:varchar(16);default:active;index"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`

	// 关联
	Agent *Agent `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

// BotQuery 机器人查询参数
type BotQuery struct {
	Name     string      // 名称模糊搜索
	Status   BotStatus   // 状态筛选
	Platform BotPlatform // 平台筛选
}

// TableName 指定表名
func (Bot) TableName() string {
	return "bots"
}
