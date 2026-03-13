package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/StellarisJAY/beepbot/internal/types"
)

// AdminRepository 管理员仓储接口
type AdminRepository interface {
	Create(user *types.AdminUser) error
	Update(user *types.AdminUser) error
	GetByID(id string) (*types.AdminUser, error)
	GetByUsername(username string) (*types.AdminUser, error)
	Exists() (bool, error)
}

// adminRepository 管理员仓储实现
type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository 创建管理员仓储
func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

// Create 创建管理员用户
func (r *adminRepository) Create(user *types.AdminUser) error {
	return r.db.Create(user).Error
}

// Update 更新管理员用户
func (r *adminRepository) Update(user *types.AdminUser) error {
	return r.db.Save(user).Error
}

// GetByID 根据 ID 获取管理员用户
func (r *adminRepository) GetByID(id string) (*types.AdminUser, error) {
	var user types.AdminUser
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取管理员用户
func (r *adminRepository) GetByUsername(username string) (*types.AdminUser, error) {
	var user types.AdminUser
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Exists 检查是否存在管理员用户
func (r *adminRepository) Exists() (bool, error) {
	var count int64
	err := r.db.Model(&types.AdminUser{}).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}