package repository

import (
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// SkillRepository 技能仓储接口
type SkillRepository interface {
	// Create 创建技能
	Create(skill *types.Skill) error
	// Update 更新技能
	Update(skill *types.Skill) error
	// Delete 删除技能
	Delete(id string) error
	// GetByID 根据ID获取技能
	GetByID(id string) (*types.Skill, error)
	// GetByName 根据名称获取技能
	GetByName(name string) (*types.Skill, error)
	// List 分页获取技能列表
	List(page, size int, status types.SkillStatus) ([]types.Skill, int64, error)
	// ExistsByName 检查技能名称是否存在
	ExistsByName(name string) (bool, error)

	// 文件操作
	CreateFile(file *types.SkillFile) error
	CreateFiles(files []types.SkillFile) error
	DeleteFilesBySkillID(skillID string) error
	ListFilesBySkillID(skillID string) ([]types.SkillFile, error)
	GetFileByID(id string) (*types.SkillFile, error)
}

// skillRepositoryImpl 技能仓储实现
type skillRepositoryImpl struct {
	db *gorm.DB
}

// NewSkillRepository 创建技能仓储
func NewSkillRepository(db *gorm.DB) SkillRepository {
	return &skillRepositoryImpl{db: db}
}

// Create 创建技能
func (r *skillRepositoryImpl) Create(skill *types.Skill) error {
	return r.db.Create(skill).Error
}

// Update 更新技能
func (r *skillRepositoryImpl) Update(skill *types.Skill) error {
	return r.db.Save(skill).Error
}

// Delete 删除技能
func (r *skillRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&types.Skill{}, "id = ?", id).Error
}

// GetByID 根据ID获取技能
func (r *skillRepositoryImpl) GetByID(id string) (*types.Skill, error) {
	var skill types.Skill
	err := r.db.First(&skill, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

// GetByName 根据名称获取技能
func (r *skillRepositoryImpl) GetByName(name string) (*types.Skill, error) {
	var skill types.Skill
	err := r.db.First(&skill, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

// List 分页获取技能列表
func (r *skillRepositoryImpl) List(page, size int, status types.SkillStatus) ([]types.Skill, int64, error) {
	var skills []types.Skill
	var total int64

	query := r.db.Model(&types.Skill{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := query.Order("installed_at DESC").Offset(offset).Limit(size).Find(&skills).Error; err != nil {
		return nil, 0, err
	}

	return skills, total, nil
}

// ExistsByName 检查技能名称是否存在
func (r *skillRepositoryImpl) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&types.Skill{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateFile 创建文件记录
func (r *skillRepositoryImpl) CreateFile(file *types.SkillFile) error {
	return r.db.Create(file).Error
}

// CreateFiles 批量创建文件记录
func (r *skillRepositoryImpl) CreateFiles(files []types.SkillFile) error {
	if len(files) == 0 {
		return nil
	}
	return r.db.Create(&files).Error
}

// DeleteFilesBySkillID 删除技能的所有文件记录
func (r *skillRepositoryImpl) DeleteFilesBySkillID(skillID string) error {
	return r.db.Delete(&types.SkillFile{}, "skill_id = ?", skillID).Error
}

// ListFilesBySkillID 获取技能的所有文件
func (r *skillRepositoryImpl) ListFilesBySkillID(skillID string) ([]types.SkillFile, error) {
	var files []types.SkillFile
	err := r.db.Where("skill_id = ?", skillID).Order("file_path ASC").Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetFileByID 根据ID获取文件
func (r *skillRepositoryImpl) GetFileByID(id string) (*types.SkillFile, error) {
	var file types.SkillFile
	err := r.db.First(&file, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}
