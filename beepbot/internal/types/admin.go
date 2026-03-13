package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminUser 管理员用户
type AdminUser struct {
	ID                   string    `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Username             string    `json:"username" gorm:"column:username;type:varchar(64);uniqueIndex;not null"`
	PasswordHash         string    `json:"-" gorm:"column:password_hash;type:varchar(128);not null"`
	RequirePasswordChange bool     `json:"require_password_change" gorm:"column:require_password_change;type:boolean;default:true"`
	CreatedAt            time.Time `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (AdminUser) TableName() string {
	return "admin_users"
}

// BeforeCreate GORM 钩子，在创建记录前自动设置 ID 和时间
func (a *AdminUser) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		id, _ := uuid.NewV7()
		a.ID = id.String()
	}
	now := time.Now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
		a.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate GORM 钩子，在更新记录前自动更新时间
func (a *AdminUser) BeforeUpdate(tx *gorm.DB) error {
	a.UpdatedAt = time.Now()
	return nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token               string `json:"token"`
	RequirePasswordChange bool  `json:"require_password_change"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangeUsernameRequest 修改用户名请求
type ChangeUsernameRequest struct {
	NewUsername string `json:"new_username" binding:"required,min=3,max=64"`
}

// AdminUserInfo 管理员用户信息
type AdminUserInfo struct {
	ID                   string `json:"id"`
	Username             string `json:"username"`
	RequirePasswordChange bool   `json:"require_password_change"`
}