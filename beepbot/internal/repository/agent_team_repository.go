package repository

import (
	"time"

	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentTeamRepository 团队仓储接口
type AgentTeamRepository interface {
	Repository[types.AgentTeam]

	// ListWithQuery 带筛选条件的分页查询
	ListWithQuery(page, pageSize int, query *types.TeamQuery) ([]types.AgentTeam, int64, error)

	// GetByName 根据名称获取团队
	GetByName(name string) (*types.AgentTeam, error)

	// GetWithMembers 获取团队及其成员信息
	GetWithMembers(id string) (*types.AgentTeam, []types.AgentTeamMember, error)

	// GetMembers 获取团队成员列表
	GetMembers(teamID string) ([]types.AgentTeamMember, error)

	// SetMembers 设置团队成员（替换现有）
	SetMembers(teamID string, members []types.AgentTeamMember) error

	// DeleteMembers 删除团队的所有成员
	DeleteMembers(teamID string) error

	// GetLeaderTeams 获取某智能体作为 Leader 的团队列表
	GetLeaderTeams(agentID string) ([]types.AgentTeam, error)

	// GetMemberTeams 获取某智能体作为成员的团队列表
	GetMemberTeams(agentID string) ([]types.AgentTeam, error)
}

type agentTeamRepository struct {
	*BaseRepository[types.AgentTeam]
	db *gorm.DB
}

func NewAgentTeamRepository(db *gorm.DB) AgentTeamRepository {
	return &agentTeamRepository{
		BaseRepository: NewBaseRepository[types.AgentTeam](db),
		db:             db,
	}
}

// ListWithQuery 带筛选条件的分页查询
func (r *agentTeamRepository) ListWithQuery(page, pageSize int, query *types.TeamQuery) ([]types.AgentTeam, int64, error) {
	var teams []types.AgentTeam
	var total int64

	db := r.db.Model(&types.AgentTeam{})

	// 动态拼接筛选条件
	if query != nil {
		if query.Name != "" {
			db = db.Where("name LIKE ?", "%"+query.Name+"%")
		}
		if query.Status != "" {
			db = db.Where("status = ?", query.Status)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&teams).Error; err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// GetByName 根据名称获取团队
func (r *agentTeamRepository) GetByName(name string) (*types.AgentTeam, error) {
	var team types.AgentTeam
	err := r.db.Where("name = ?", name).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetWithMembers 获取团队及其成员信息
func (r *agentTeamRepository) GetWithMembers(id string) (*types.AgentTeam, []types.AgentTeamMember, error) {
	team, err := r.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	members, err := r.GetMembers(id)
	if err != nil {
		return nil, nil, err
	}

	return team, members, nil
}

// GetMembers 获取团队成员列表
func (r *agentTeamRepository) GetMembers(teamID string) ([]types.AgentTeamMember, error) {
	var members []types.AgentTeamMember
	err := r.db.Where("team_id = ?", teamID).Order("created_at ASC").Find(&members).Error
	return members, err
}

// SetMembers 设置团队成员（替换现有）
func (r *agentTeamRepository) SetMembers(teamID string, members []types.AgentTeamMember) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有成员
		if err := tx.Where("team_id = ?", teamID).Delete(&types.AgentTeamMember{}).Error; err != nil {
			return err
		}

		// 创建新成员
		if len(members) > 0 {
			now := time.Now()
			for i := range members {
				if members[i].CreatedAt.IsZero() {
					members[i].CreatedAt = now
				}
			}
			if err := tx.Create(&members).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteMembers 删除团队的所有成员
func (r *agentTeamRepository) DeleteMembers(teamID string) error {
	return r.db.Where("team_id = ?", teamID).Delete(&types.AgentTeamMember{}).Error
}

// GetLeaderTeams 获取某智能体作为 Leader 的团队列表
func (r *agentTeamRepository) GetLeaderTeams(agentID string) ([]types.AgentTeam, error) {
	var teams []types.AgentTeam
	err := r.db.Where("leader_id = ?", agentID).Find(&teams).Error
	return teams, err
}

// GetMemberTeams 获取某智能体作为成员的团队列表
func (r *agentTeamRepository) GetMemberTeams(agentID string) ([]types.AgentTeam, error) {
	var teams []types.AgentTeam
	err := r.db.Table("agent_teams").
		Select("agent_teams.*").
		Joins("JOIN agent_team_members ON agent_team_members.team_id = agent_teams.id").
		Where("agent_team_members.agent_id = ?", agentID).
		Find(&teams).Error
	return teams, err
}

// DeleteTeamAndMembers 删除团队及其成员（事务操作）
func DeleteTeamAndMembers(db *gorm.DB, teamID string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 删除成员
		if err := tx.Where("team_id = ?", teamID).Delete(&types.AgentTeamMember{}).Error; err != nil {
			return err
		}
		// 删除团队
		if err := tx.Where("id = ?", teamID).Delete(&types.AgentTeam{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// CreateTeamWithMembers 创建团队及其成员（事务操作）
func CreateTeamWithMembers(db *gorm.DB, team *types.AgentTeam, members []types.AgentTeamMember) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 创建团队
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		// 创建成员
		if len(members) > 0 {
			now := time.Now()
			for i := range members {
				if members[i].TeamID == "" {
					members[i].TeamID = team.ID
				}
				if members[i].CreatedAt.IsZero() {
					members[i].CreatedAt = now
				}
			}
			if err := tx.Create(&members).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// NewAgentTeamMember 创建团队成员
func NewAgentTeamMember(teamID, agentID, memberName string, role types.MemberRole, description string) *types.AgentTeamMember {
	return &types.AgentTeamMember{
		TeamID:      teamID,
		AgentID:     agentID,
		MemberName:  memberName,
		Role:        role,
		Description: description,
		CreatedAt:   time.Now(),
	}
}

// GenerateTeamID 生成团队 ID
func GenerateTeamID() string {
	return uuid.NewString()
}
