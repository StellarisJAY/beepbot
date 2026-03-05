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

	// GetByChannel 根据渠道类型和标识获取智能体
	GetByChannel(channelType, channelIdentifier string) (*types.Agent, error)

	// CreateChannel 创建渠道绑定
	CreateChannel(channel *types.AgentChannel) error

	// UpdateChannel 更新渠道绑定
	UpdateChannel(channel *types.AgentChannel) error

	// DeleteChannel 删除渠道绑定
	DeleteChannel(id string) error

	// GetChannelByID 根据ID获取渠道绑定
	GetChannelByID(id string) (*types.AgentChannel, error)

	// GetChannelsByAgent 获取智能体的所有渠道绑定
	GetChannelsByAgent(agentID string) ([]types.AgentChannel, error)
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
	err := r.db.Preload("Provider").Preload("Channels").
		Where("id = ?", id).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *agentRepository) GetByChannel(channelType, channelIdentifier string) (*types.Agent, error) {
	var agent types.Agent
	err := r.db.Joins("JOIN agent_channels ON agents.id = agent_channels.agent_id").
		Where("agent_channels.channel_type = ? AND agent_channels.channel_identifier = ?",
			channelType, channelIdentifier).
		First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *agentRepository) CreateChannel(channel *types.AgentChannel) error {
	return r.db.Create(channel).Error
}

func (r *agentRepository) UpdateChannel(channel *types.AgentChannel) error {
	return r.db.Save(channel).Error
}

func (r *agentRepository) DeleteChannel(id string) error {
	return r.db.Where("id = ?", id).Delete(&types.AgentChannel{}).Error
}

func (r *agentRepository) GetChannelByID(id string) (*types.AgentChannel, error) {
	var channel types.AgentChannel
	err := r.db.Where("id = ?", id).First(&channel).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *agentRepository) GetChannelsByAgent(agentID string) ([]types.AgentChannel, error) {
	var channels []types.AgentChannel
	err := r.db.Where("agent_id = ?", agentID).Find(&channels).Error
	return channels, err
}
