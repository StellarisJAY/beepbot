package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseEntity 基础实体，包含公共字段
type BaseEntity struct {
	ID        string    `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// BeforeCreate GORM 钩子，在创建记录前自动设置 ID 和时间
func (b *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		id, _ := uuid.NewV7()
		b.ID = id.String()
	}
	now := time.Now()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
		b.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate GORM 钩子，在更新记录前自动更新时间
func (b *BaseEntity) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

// GenerateUUIDv7 生成 UUIDv7
func GenerateUUIDv7() string {
	id, _ := uuid.NewV7()
	return id.String()
}
