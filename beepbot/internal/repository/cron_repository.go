package repository

import (
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// CronRepository 定时任务仓储接口
type CronRepository interface {
	Repository[types.CronJob]

	// ListWithQuery 带筛选条件的分页查询
	ListWithQuery(page, pageSize int, query *types.CronQuery) ([]types.CronJob, int64, error)

	// GetByName 根据名称获取定时任务
	GetByName(name string) (*types.CronJob, error)

	// GetByAgentID 根据智能体ID获取定时任务列表
	GetByAgentID(agentID string) ([]types.CronJob, error)

	// GetEnabledJobs 获取所有启用的定时任务
	GetEnabledJobs() ([]types.CronJob, error)
}

type cronRepository struct {
	*BaseRepository[types.CronJob]
	db *gorm.DB
}

// NewCronRepository 创建定时任务仓储实例
func NewCronRepository(db *gorm.DB) CronRepository {
	return &cronRepository{
		BaseRepository: NewBaseRepository[types.CronJob](db),
		db:             db,
	}
}

// ListWithQuery 带筛选条件的分页查询
func (r *cronRepository) ListWithQuery(page, pageSize int, query *types.CronQuery) ([]types.CronJob, int64, error) {
	var jobs []types.CronJob
	var total int64

	db := r.db.Model(&types.CronJob{})

	// 动态拼接筛选条件
	if query != nil {
		if query.Name != "" {
			db = db.Where("name LIKE ?", "%"+query.Name+"%")
		}
		if query.Enabled != nil {
			db = db.Where("enabled = ?", *query.Enabled)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

func (r *cronRepository) GetByName(name string) (*types.CronJob, error) {
	var job types.CronJob
	err := r.db.Where("name = ?", name).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *cronRepository) GetByAgentID(agentID string) ([]types.CronJob, error) {
	var jobs []types.CronJob
	err := r.db.Where("agent_id = ?", agentID).Find(&jobs).Error
	return jobs, err
}

func (r *cronRepository) GetEnabledJobs() ([]types.CronJob, error) {
	var jobs []types.CronJob
	err := r.db.Where("enabled = ?", true).Find(&jobs).Error
	return jobs, err
}
