package repository

import (
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// BotRepository 机器人仓储接口
type BotRepository interface {
	Repository[types.Bot]

	// ListWithQuery 带筛选条件的分页查询
	ListWithQuery(page, pageSize int, query *types.BotQuery) ([]types.Bot, int64, error)

	// GetByName 根据名称获取 Bot
	GetByName(name string) (*types.Bot, error)

	// GetByPlatform 根据平台获取 Bot 列表
	GetByPlatform(platform types.BotPlatform) ([]types.Bot, error)

	// GetByStatus 根据状态获取 Bot 列表
	GetByStatus(status types.BotStatus) ([]types.Bot, error)

	// GetUnbound 获取未绑定智能体的 Bot
	GetUnbound() ([]types.Bot, error)

	// GetByAgent 根据智能体ID获取 Bot 列表
	GetByAgent(agentID string) ([]types.Bot, error)

	// GetWithRelations 获取 Bot 及其关联的智能体
	GetWithRelations(id string) (*types.Bot, error)

	// BindAgent 绑定智能体
	BindAgent(botID string, agentID *string) error

	// FindByIDs 根据ID列表批量查询 Bot
	FindByIDs(ids []string) ([]types.Bot, error)
}

type botRepository struct {
	*BaseRepository[types.Bot]
	db *gorm.DB
}

func NewBotRepository(db *gorm.DB) BotRepository {
	return &botRepository{
		BaseRepository: NewBaseRepository[types.Bot](db),
		db:             db,
	}
}

// ListWithQuery 带筛选条件的分页查询
func (r *botRepository) ListWithQuery(page, pageSize int, query *types.BotQuery) ([]types.Bot, int64, error) {
	var bots []types.Bot
	var total int64

	db := r.db.Model(&types.Bot{})

	// 动态拼接筛选条件
	if query != nil {
		if query.Name != "" {
			db = db.Where("name LIKE ?", "%"+query.Name+"%")
		}
		if query.Status != "" {
			db = db.Where("status = ?", query.Status)
		}
		if query.Platform != "" {
			db = db.Where("platform = ?", query.Platform)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&bots).Error; err != nil {
		return nil, 0, err
	}

	return bots, total, nil
}

func (r *botRepository) GetByName(name string) (*types.Bot, error) {
	var bot types.Bot
	err := r.db.Where("name = ?", name).First(&bot).Error
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *botRepository) GetByPlatform(platform types.BotPlatform) ([]types.Bot, error) {
	var bots []types.Bot
	err := r.db.Where("platform = ?", platform).Find(&bots).Error
	return bots, err
}

func (r *botRepository) GetByStatus(status types.BotStatus) ([]types.Bot, error) {
	var bots []types.Bot
	err := r.db.Where("status = ?", status).Find(&bots).Error
	return bots, err
}

func (r *botRepository) GetUnbound() ([]types.Bot, error) {
	var bots []types.Bot
	err := r.db.Where("agent_id IS NULL").Find(&bots).Error
	return bots, err
}

func (r *botRepository) GetByAgent(agentID string) ([]types.Bot, error) {
	var bots []types.Bot
	err := r.db.Where("agent_id = ?", agentID).Find(&bots).Error
	return bots, err
}

func (r *botRepository) GetWithRelations(id string) (*types.Bot, error) {
	var bot types.Bot
	err := r.db.Preload("Agent").
		Where("id = ?", id).First(&bot).Error
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *botRepository) BindAgent(botID string, agentID *string) error {
	return r.db.Model(&types.Bot{}).Where("id = ?", botID).Update("agent_id", agentID).Error
}

func (r *botRepository) FindByIDs(ids []string) ([]types.Bot, error) {
	if len(ids) == 0 {
		return []types.Bot{}, nil
	}
	var bots []types.Bot
	err := r.db.Where("id IN ?", ids).Find(&bots).Error
	return bots, err
}
