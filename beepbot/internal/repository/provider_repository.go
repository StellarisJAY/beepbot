package repository

import (
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// ProviderRepository 供应商仓储接口
type ProviderRepository interface {
	Repository[types.Provider]

	// GetByName 根据名称获取供应商
	GetByName(name string) (*types.Provider, error)

	// GetByType 根据类型获取供应商列表
	GetByType(providerType types.ProviderType) ([]types.Provider, error)

	// GetDefaultByType 获取指定类型的默认供应商
	GetDefaultByType(providerType types.ProviderType) (*types.Provider, error)

	// SetDefault 设置默认供应商
	SetDefault(id string) error
}

type providerRepository struct {
	*BaseRepository[types.Provider]
	db *gorm.DB
}

func NewProviderRepository(db *gorm.DB) ProviderRepository {
	return &providerRepository{
		BaseRepository: NewBaseRepository[types.Provider](db),
		db:             db,
	}
}

func (r *providerRepository) GetByName(name string) (*types.Provider, error) {
	var provider types.Provider
	err := r.db.Where("name = ?", name).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *providerRepository) GetByType(providerType types.ProviderType) ([]types.Provider, error) {
	var providers []types.Provider
	err := r.db.Where("provider_type = ?", providerType).Find(&providers).Error
	return providers, err
}

func (r *providerRepository) GetDefaultByType(providerType types.ProviderType) (*types.Provider, error) {
	var provider types.Provider
	err := r.db.Where("provider_type = ? AND is_default = ?", providerType, true).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *providerRepository) SetDefault(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先获取要设置的供应商
		var provider types.Provider
		if err := tx.Where("id = ?", id).First(&provider).Error; err != nil {
			return err
		}

		// 清除同类型的其他默认标记
		if err := tx.Model(&types.Provider{}).
			Where("provider_type = ?", provider.ProviderType).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// 设置新的默认
		return tx.Model(&types.Provider{}).
			Where("id = ?", id).
			Update("is_default", true).Error
	})
}
