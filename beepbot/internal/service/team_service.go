package service

import (
	"errors"
	"time"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// TeamService 团队服务
type TeamService struct {
	teamRepo  repository.AgentTeamRepository
	agentRepo repository.AgentRepository
	db        *gorm.DB
}

func NewTeamService(teamRepo repository.AgentTeamRepository, agentRepo repository.AgentRepository, db *gorm.DB) *TeamService {
	return &TeamService{
		teamRepo:  teamRepo,
		agentRepo: agentRepo,
		db:        db,
	}
}

// CreateTeamRequest 创建团队请求
type CreateTeamRequest struct {
	Name        string           `json:"name" binding:"required"`
	Description string           `json:"description"`
	LeaderID    string           `json:"leader_id" binding:"required"` // Leader 智能体 ID
	Members     []MemberRequest  `json:"members"`                      // 队员列表
}

// MemberRequest 成员请求
type MemberRequest struct {
	AgentID     string `json:"agent_id" binding:"required"`
	MemberName  string `json:"member_name" binding:"required"`
	Description string `json:"description"`
}

// UpdateTeamRequest 更新团队请求
type UpdateTeamRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	LeaderID    string          `json:"leader_id"`
	Members     []MemberRequest `json:"members"`
	Status      types.TeamStatus `json:"status"`
}

// TeamResponse 团队响应
type TeamResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	LeaderID    string              `json:"leader_id"`
	Leader      *TeamMemberBrief   `json:"leader,omitempty"`
	Members     []TeamMemberBrief  `json:"members,omitempty"`
	Status      types.TeamStatus   `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// TeamMemberBrief 团队成员简要信息
type TeamMemberBrief struct {
	AgentID     string `json:"agent_id"`
	AgentName   string `json:"agent_name"`
	MemberName  string `json:"member_name"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

// CreateTeam 创建团队
func (s *TeamService) CreateTeam(req *CreateTeamRequest) (*TeamResponse, error) {
	// 检查名称是否已存在
	if _, err := s.teamRepo.GetByName(req.Name); err == nil {
		return nil, errors.New("team name already exists")
	}

	// 检查 Leader 是否存在
	leader, err := s.agentRepo.GetByID(req.LeaderID)
	if err != nil {
		return nil, errors.New("leader agent not found")
	}

	// 检查成员名称是否重复（包括 Leader）
	memberNames := make(map[string]bool)
	memberNames[leader.Name] = true // Leader 默认使用 Agent 名称作为成员名称

	for _, m := range req.Members {
		if memberNames[m.MemberName] {
			return nil, errors.New("成员名称重复: " + m.MemberName)
		}
		memberNames[m.MemberName] = true
	}

	// 生成团队 ID
	teamID := repository.GenerateTeamID()
	now := time.Now()

	// 创建团队实体
	team := &types.AgentTeam{
		ID:          teamID,
		Name:        req.Name,
		Description: req.Description,
		LeaderID:    req.LeaderID,
		Status:      types.TeamStatusInactive, // 默认禁用，配置完成后启用
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 创建成员列表
	var members []types.AgentTeamMember

	// 添加 Leader 作为成员
	members = append(members, *repository.NewAgentTeamMember(
		teamID,
		req.LeaderID,
		leader.Name, // Leader 默认使用 Agent 名称作为成员名称
		types.MemberRoleLeader,
		"团队 Leader",
	))

	// 添加队员（允许与 Leader 使用相同的智能体）
	if len(req.Members) > 0 {
		for _, m := range req.Members {
			// 检查 Agent 是否存在
			agent, err := s.agentRepo.GetByID(m.AgentID)
			if err != nil {
				return nil, errors.New("member agent not found: " + m.AgentID)
			}

			members = append(members, *repository.NewAgentTeamMember(
				teamID,
				m.AgentID,
				m.MemberName,
				types.MemberRoleMember,
				m.Description,
			))

			// 记录 Agent 名称（用于响应）
			_ = agent
		}
	}

	// 创建团队及成员（使用事务）
	if err := repository.CreateTeamWithMembers(s.db, team, members); err != nil {
		return nil, err
	}

	return s.GetTeam(teamID)
}

// UpdateTeam 更新团队
func (s *TeamService) UpdateTeam(id string, req *UpdateTeamRequest) (*TeamResponse, error) {
	team, err := s.teamRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 更新名称
	if req.Name != "" {
		// 检查名称是否被其他团队使用
		if existing, err := s.teamRepo.GetByName(req.Name); err == nil && existing.ID != id {
			return nil, errors.New("team name already exists")
		}
		team.Name = req.Name
	}

	// 更新描述
	if req.Description != "" {
		team.Description = req.Description
	}

	// 更新 Leader
	if req.LeaderID != "" && req.LeaderID != team.LeaderID {
		// 检查新的 Leader Agent 是否存在
		if _, err := s.agentRepo.GetByID(req.LeaderID); err != nil {
			return nil, errors.New("leader agent not found")
		}
		team.LeaderID = req.LeaderID
	}

	// 更新状态
	if req.Status != "" {
		team.Status = req.Status
	}

	team.UpdatedAt = time.Now()

	if err := s.teamRepo.Update(team); err != nil {
		return nil, err
	}

	// 更新成员列表
	if req.Members != nil {
		// 找到 Leader 的成员名称
		var leaderName string
		if agent, err := s.agentRepo.GetByID(team.LeaderID); err == nil {
			leaderName = agent.Name
		}

		// 检查成员名称是否重复（包括 Leader）
		memberNames := make(map[string]bool)
		memberNames[leaderName] = true

		for _, m := range req.Members {
			if memberNames[m.MemberName] {
				return nil, errors.New("成员名称重复: " + m.MemberName)
			}
			memberNames[m.MemberName] = true
		}

		// 构建新的成员列表
		var newMembers []types.AgentTeamMember

		// 添加 Leader
		newMembers = append(newMembers, *repository.NewAgentTeamMember(
			id,
			team.LeaderID,
			leaderName,
			types.MemberRoleLeader,
			"团队 Leader",
		))

		// 添加队员（允许与 Leader 使用相同的智能体）
		for _, m := range req.Members {
			// 检查 Agent 是否存在
			if _, err := s.agentRepo.GetByID(m.AgentID); err != nil {
				return nil, errors.New("member agent not found: " + m.AgentID)
			}

			newMembers = append(newMembers, *repository.NewAgentTeamMember(
				id,
				m.AgentID,
				m.MemberName,
				types.MemberRoleMember,
				m.Description,
			))
		}

		if err := s.teamRepo.SetMembers(id, newMembers); err != nil {
			return nil, err
		}
	}

	return s.GetTeam(id)
}

// DeleteTeam 删除团队
func (s *TeamService) DeleteTeam(id string) error {
	// 先删除成员
	if err := s.teamRepo.DeleteMembers(id); err != nil {
		return err
	}
	// 再删除团队
	return s.teamRepo.Delete(id)
}

// GetTeam 获取团队
func (s *TeamService) GetTeam(id string) (*TeamResponse, error) {
	team, members, err := s.teamRepo.GetWithMembers(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(team, members), nil
}

// ListTeams 列出团队
func (s *TeamService) ListTeams(page, pageSize int, query *types.TeamQuery) ([]TeamResponse, int64, error) {
	teams, total, err := s.teamRepo.ListWithQuery(page, pageSize, query)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]TeamResponse, len(teams))
	for i, t := range teams {
		members, err := s.teamRepo.GetMembers(t.ID)
		if err != nil {
			return nil, 0, err
		}
		responses[i] = *s.toResponse(&teams[i], members)
	}

	return responses, total, nil
}

// UpdateTeamStatus 更新团队状态
func (s *TeamService) UpdateTeamStatus(id string, status types.TeamStatus) error {
	team, err := s.teamRepo.GetByID(id)
	if err != nil {
		return err
	}

	team.Status = status
	team.UpdatedAt = time.Now()

	return s.teamRepo.Update(team)
}

// GetAgentTeams 获取智能体所属的团队列表
func (s *TeamService) GetAgentTeams(agentID string) ([]TeamResponse, error) {
	// 获取作为 Leader 的团队
	leaderTeams, err := s.teamRepo.GetLeaderTeams(agentID)
	if err != nil {
		return nil, err
	}

	// 获取作为成员的团队
	memberTeams, err := s.teamRepo.GetMemberTeams(agentID)
	if err != nil {
		return nil, err
	}

	// 合并并去重
	teamMap := make(map[string]*types.AgentTeam)
	for i := range leaderTeams {
		teamMap[leaderTeams[i].ID] = &leaderTeams[i]
	}
	for i := range memberTeams {
		if _, exists := teamMap[memberTeams[i].ID]; !exists {
			teamMap[memberTeams[i].ID] = &memberTeams[i]
		}
	}

	// 转换为响应
	responses := make([]TeamResponse, 0, len(teamMap))
	for _, team := range teamMap {
		members, err := s.teamRepo.GetMembers(team.ID)
		if err != nil {
			continue
		}
		responses = append(responses, *s.toResponse(team, members))
	}

	return responses, nil
}

// toResponse 转换为响应格式
func (s *TeamService) toResponse(team *types.AgentTeam, members []types.AgentTeamMember) *TeamResponse {
	response := &TeamResponse{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		LeaderID:    team.LeaderID,
		Status:      team.Status,
		CreatedAt:   team.CreatedAt,
		UpdatedAt:   team.UpdatedAt,
	}

	// 处理成员信息
	var leader *TeamMemberBrief
	var teamMembers []TeamMemberBrief

	for _, m := range members {
		// 获取 Agent 名称
		agentName := ""
		if agent, err := s.agentRepo.GetByID(m.AgentID); err == nil {
			agentName = agent.Name
		}

		memberBrief := TeamMemberBrief{
			AgentID:     m.AgentID,
			AgentName:   agentName,
			MemberName:  m.MemberName,
			Role:        string(m.Role),
			Description: m.Description,
		}

		if m.Role == types.MemberRoleLeader {
			leader = &memberBrief
		} else {
			teamMembers = append(teamMembers, memberBrief)
		}
	}

	response.Leader = leader
	response.Members = teamMembers

	return response
}