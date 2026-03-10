package repository

import (
	"time"

	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/google/uuid"
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

	// GetAgentSkills 获取智能体关联的技能ID列表
	GetAgentSkills(agentID string) ([]string, error)

	// SetAgentSkills 设置智能体关联的技能（替换现有）
	SetAgentSkills(agentID string, skillIDs []string) error

	// DeleteAgentSkills 删除智能体的所有技能关联
	DeleteAgentSkills(agentID string) error

	// DeleteSkillFromAllAgents 从所有智能体中删除指定技能的关联
	DeleteSkillFromAllAgents(skillID string) error
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

// GetAgentSkills 获取智能体关联的技能ID列表
func (r *agentRepository) GetAgentSkills(agentID string) ([]string, error) {
	var agentSkills []types.AgentSkill
	err := r.db.Where("agent_id = ?", agentID).Find(&agentSkills).Error
	if err != nil {
		return nil, err
	}
	skillIDs := make([]string, len(agentSkills))
	for i, as := range agentSkills {
		skillIDs[i] = as.SkillID
	}
	return skillIDs, nil
}

// SetAgentSkills 设置智能体关联的技能（替换现有）
func (r *agentRepository) SetAgentSkills(agentID string, skillIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有关联
		if err := tx.Where("agent_id = ?", agentID).Delete(&types.AgentSkill{}).Error; err != nil {
			return err
		}

		// 创建新关联
		if len(skillIDs) > 0 {
			now := time.Now()
			agentSkills := make([]types.AgentSkill, len(skillIDs))
			for i, skillID := range skillIDs {
				agentSkills[i] = types.AgentSkill{
					ID:        uuid.NewString(),
					AgentID:   agentID,
					SkillID:   skillID,
					CreatedAt: now,
				}
			}
			if err := tx.Create(&agentSkills).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteAgentSkills 删除智能体的所有技能关联
func (r *agentRepository) DeleteAgentSkills(agentID string) error {
	return r.db.Where("agent_id = ?", agentID).Delete(&types.AgentSkill{}).Error
}

// DeleteSkillFromAllAgents 从所有智能体中删除指定技能的关联
func (r *agentRepository) DeleteSkillFromAllAgents(skillID string) error {
	return r.db.Where("skill_id = ?", skillID).Delete(&types.AgentSkill{}).Error
}
