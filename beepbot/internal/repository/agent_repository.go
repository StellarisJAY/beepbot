package repository

import (
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// AgentRepository 智能体仓储接口
type AgentRepository interface {
	Repository[types.Agent]

	// GetByName 根据名称获取智能体
	GetByName(name string) (*types.Agent, error)

	// GetByStatus 根据状态获取智能体列表
	GetByStatus(status types.AgentStatus) ([]types.Agent, error)

	// GetActiveAgents 获取所有活跃的智能体
	GetActiveAgents() ([]types.Agent, error)

	// GetWithRelations 获取智能体及其关联数据
	GetWithRelations(id string) (*types.Agent, error)
}

type agentRepository struct {
	*BaseRepository[types.Agent]
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) AgentRepository {
	return &agentRepository{
		BaseRepository: NewBaseRepository[types.Agent](db),
		db:             db,
	}
}

func (r *agentRepository) GetByName(name string) (*types.Agent, error) {
	var agent types.Agent
	err := r.db.Where("name = ?", name).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *agentRepository) GetByStatus(status types.AgentStatus) ([]types.Agent, error) {
	var agents []types.Agent
	err := r.db.Where("status = ?", status).Find(&agents).Error
	return agents, err
}

func (r *agentRepository) GetActiveAgents() ([]types.Agent, error) {
	return r.GetByStatus(types.AgentStatusActive)
}

func (r *agentRepository) GetWithRelations(id string) (*types.Agent, error) {
	var agent types.Agent
	err := r.db.Where("id = ?", id).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}
