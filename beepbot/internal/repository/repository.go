package repository

import (
	"errors"

	"gorm.io/gorm"
)

// ErrNotFound 记录不存在错误
var ErrNotFound = errors.New("record not found")

// Repository 基础仓储接口
type Repository[T any] interface {
	Create(entity *T) error
	Update(entity *T) error
	Delete(id string) error
	GetByID(id string) (*T, error)
	List(page, pageSize int) ([]T, int64, error)
}

// BaseRepository 基础仓储实现
type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

func (r *BaseRepository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

func (r *BaseRepository[T]) Delete(id string) error {
	var entity T
	return r.db.Where("id = ?", id).Delete(&entity).Error
}

func (r *BaseRepository[T]) GetByID(id string) (*T, error) {
	var entity T
	err := r.db.Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) List(page, pageSize int) ([]T, int64, error) {
	var entities []T
	var total int64

	offset := (page - 1) * pageSize
	if err := r.db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Offset(offset).Limit(pageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}
