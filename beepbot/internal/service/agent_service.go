package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// AgentService 智能体服务
type AgentService struct {
	repo            repository.AgentRepository
	skillRepo       repository.SkillRepository
	providerService *ProviderService
}

func NewAgentService(repo repository.AgentRepository, skillRepo repository.SkillRepository, providerService *ProviderService) *AgentService {
	return &AgentService{
		repo:            repo,
		skillRepo:       skillRepo,
		providerService: providerService,
	}
}

// AgentDefaults 智能体默认配置
type AgentDefaults struct {
	SystemPrompt        string  `json:"system_prompt"`
	Temperature         float32 `json:"temperature"`
	MaxIterations       int     `json:"max_iterations"`
	MaxOutputTokens     int64   `json:"max_output_tokens"`
	CompressionRatio    float64 `json:"compression_ratio"`
	CompressionKeepSize int     `json:"compression_keep_size"`
	ContextMaxTokens    int64   `json:"context_max_tokens"`
}

// ValidationResult 校验结果
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// 默认配置常量
const (
	DefaultSystemPrompt        = "你是一名智能助手"
	DefaultTemperature         = 0.7
	DefaultMaxIterations       = 50
	DefaultMaxOutputTokens     = 4096
	DefaultCompressionRatio    = 0.7
	DefaultCompressionKeepSize = 5
	DefaultContextMaxTokens    = 40960
)

// CreateAgentRequest 创建智能体请求
type CreateAgentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	// 以下字段可选，创建时可留空
	ProviderID          string  `json:"provider_id"`
	Model               string  `json:"model"`
	SystemPrompt        string  `json:"system_prompt"`
	WorkingDir          string  `json:"working_dir"`
	Temperature         float32 `json:"temperature"`
	MaxIterations       int     `json:"max_iterations"`
	MaxOutputTokens     int64   `json:"max_output_tokens"`
	CompressionRatio    float64 `json:"compression_ratio"`
	CompressionKeepSize int     `json:"compression_keep_size"`
	ContextMaxTokens    int64   `json:"context_max_tokens"`
	// UseAllSkills 是否使用所有技能（默认 true）
	UseAllSkills bool `json:"use_all_skills"`
	// SkillIDs 关联的技能ID列表
	SkillIDs []string `json:"skill_ids"`
}

// GetAgentDefaults 获取智能体默认配置
func (s *AgentService) GetAgentDefaults() *AgentDefaults {
	return &AgentDefaults{
		SystemPrompt:        DefaultSystemPrompt,
		Temperature:         DefaultTemperature,
		MaxIterations:       DefaultMaxIterations,
		MaxOutputTokens:     DefaultMaxOutputTokens,
		CompressionRatio:    DefaultCompressionRatio,
		CompressionKeepSize: DefaultCompressionKeepSize,
		ContextMaxTokens:    DefaultContextMaxTokens,
	}
}

// ValidateAgentConfig 校验智能体配置是否完整
func (s *AgentService) ValidateAgentConfig(agent *AgentResponse) *ValidationResult {
	var errs []string

	if agent.ProviderID == "" {
		errs = append(errs, "供应商不能为空")
	}
	if agent.Model == "" {
		errs = append(errs, "模型不能为空")
	}
	if agent.WorkingDir == "" {
		errs = append(errs, "工作目录不能为空")
	}

	return &ValidationResult{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}

// UpdateAgentRequest 更新智能体请求
type UpdateAgentRequest struct {
	Name                string            `json:"name"`
	Description         string            `json:"description"`
	ProviderID          string            `json:"provider_id"`
	Model               string            `json:"model"`
	SystemPrompt        string            `json:"system_prompt"`
	Temperature         float32           `json:"temperature"`
	MaxIterations       int               `json:"max_iterations"`
	MaxOutputTokens     int64             `json:"max_output_tokens"`
	WorkingDir          string            `json:"working_dir"`
	CompressionRatio    float64           `json:"compression_ratio"`
	CompressionKeepSize int               `json:"compression_keep_size"`
	ContextMaxTokens    int64             `json:"context_max_tokens"`
	Status              types.AgentStatus `json:"status"`
	// UseAllSkills 是否使用所有技能
	UseAllSkills *bool `json:"use_all_skills,omitempty"`
	// SkillIDs 关联的技能ID列表（仅当 use_all_skills 为 false 时有效）
	SkillIDs []string `json:"skill_ids,omitempty"`
}

// SkillBrief 技能简要信息
type SkillBrief struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AgentResponse 智能体响应
type AgentResponse struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Description         string            `json:"description"`
	ProviderID          string            `json:"provider_id"`
	Provider            *ProviderResponse `json:"provider,omitempty"`
	Model               string            `json:"model"`
	SystemPrompt        string            `json:"system_prompt"`
	Temperature         float32           `json:"temperature"`
	MaxIterations       int               `json:"max_iterations"`
	MaxOutputTokens     int64             `json:"max_output_tokens"`
	WorkingDir          string            `json:"working_dir"`
	CompressionRatio    float64           `json:"compression_ratio"`
	CompressionKeepSize int               `json:"compression_keep_size"`
	ContextMaxTokens    int64             `json:"context_max_tokens"`
	Status              types.AgentStatus `json:"status"`
	// UseAllSkills 是否使用所有技能
	UseAllSkills bool `json:"use_all_skills"`
	// Skills 关联的技能列表（简要信息）
	Skills    []SkillBrief `json:"skills,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// CreateAgent 创建智能体
func (s *AgentService) CreateAgent(req *CreateAgentRequest) (*types.Agent, error) {
	// 检查名称是否已存在
	if _, err := s.repo.GetByName(req.Name); err == nil {
		return nil, errors.New("agent name already exists")
	}

	// 如果指定了供应商，检查是否存在
	if req.ProviderID != "" {
		if _, err := s.providerService.GetProvider(req.ProviderID); err != nil {
			return nil, errors.New("provider not found")
		}
	}

	// 生成智能体ID（先生成，用于工作目录）
	agentID := types.GenerateUUIDv7()

	// 获取默认配置
	defaults := s.GetAgentDefaults()

	// 设置默认值
	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = defaults.SystemPrompt
	}

	workingDir := req.WorkingDir
	if workingDir == "" {
		workingDir = fmt.Sprintf("/data/agents/%s", agentID)
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = defaults.Temperature
	}
	maxIterations := req.MaxIterations
	if maxIterations == 0 {
		maxIterations = defaults.MaxIterations
	}
	maxOutputTokens := req.MaxOutputTokens
	if maxOutputTokens == 0 {
		maxOutputTokens = defaults.MaxOutputTokens
	}
	compressionRatio := req.CompressionRatio
	if compressionRatio == 0 {
		compressionRatio = defaults.CompressionRatio
	}
	compressionKeepSize := req.CompressionKeepSize
	if compressionKeepSize == 0 {
		compressionKeepSize = defaults.CompressionKeepSize
	}
	contextMaxTokens := req.ContextMaxTokens
	if contextMaxTokens == 0 {
		contextMaxTokens = defaults.ContextMaxTokens
	}

	agent := &types.Agent{
		ID:                  agentID,
		Name:                req.Name,
		Description:         req.Description,
		ProviderID:          req.ProviderID, // 可为空
		Model:               req.Model,      // 可为空
		SystemPrompt:        systemPrompt,
		Temperature:         temperature,
		MaxIterations:       maxIterations,
		MaxOutputTokens:     maxOutputTokens,
		WorkingDir:          workingDir,
		CompressionRatio:    compressionRatio,
		CompressionKeepSize: compressionKeepSize,
		ContextMaxTokens:    contextMaxTokens,
		Status:              types.AgentStatusInactive, // 默认禁用，配置完成后启用
		UseAllSkills:        req.UseAllSkills,          // 默认为 false，前端应传 true
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.repo.Create(agent); err != nil {
		return nil, err
	}

	// 如果不使用所有技能且指定了技能列表，创建技能关联
	if !req.UseAllSkills && len(req.SkillIDs) > 0 {
		if err := s.repo.SetAgentSkills(agentID, req.SkillIDs); err != nil {
			// 记录错误但不回滚智能体创建
			// 可以考虑添加日志
		}
	}

	return agent, nil
}

// UpdateAgent 更新智能体
func (s *AgentService) UpdateAgent(id string, req *UpdateAgentRequest) (*types.Agent, error) {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		// 检查名称是否被其他智能体使用
		if existing, err := s.repo.GetByName(req.Name); err == nil && existing.ID != id {
			return nil, errors.New("agent name already exists")
		}
		agent.Name = req.Name
	}

	if req.Description != "" {
		agent.Description = req.Description
	}

	if req.ProviderID != "" {
		// 检查供应商是否存在
		if _, err := s.providerService.GetProvider(req.ProviderID); err != nil {
			return nil, errors.New("provider not found")
		}
		agent.ProviderID = req.ProviderID
	}

	if req.Model != "" {
		agent.Model = req.Model
	}

	if req.SystemPrompt != "" {
		agent.SystemPrompt = req.SystemPrompt
	}

	if req.Temperature != 0 {
		agent.Temperature = req.Temperature
	}

	if req.MaxIterations != 0 {
		agent.MaxIterations = req.MaxIterations
	}

	if req.MaxOutputTokens != 0 {
		agent.MaxOutputTokens = req.MaxOutputTokens
	}

	if req.WorkingDir != "" {
		agent.WorkingDir = req.WorkingDir
	}

	if req.CompressionRatio != 0 {
		agent.CompressionRatio = req.CompressionRatio
	}

	if req.CompressionKeepSize != 0 {
		agent.CompressionKeepSize = req.CompressionKeepSize
	}

	if req.ContextMaxTokens != 0 {
		agent.ContextMaxTokens = req.ContextMaxTokens
	}

	if req.Status != "" {
		agent.Status = req.Status
	}

	// 更新 UseAllSkills 字段
	if req.UseAllSkills != nil {
		agent.UseAllSkills = *req.UseAllSkills
	}

	agent.UpdatedAt = time.Now()

	if err := s.repo.Update(agent); err != nil {
		return nil, err
	}

	// 更新技能关联（仅当 UseAllSkills 为 false 且提供了 SkillIDs 时）
	if req.UseAllSkills != nil && !*req.UseAllSkills && req.SkillIDs != nil {
		if err := s.repo.SetAgentSkills(id, req.SkillIDs); err != nil {
			return nil, err
		}
	}

	return agent, nil
}

// DeleteAgent 删除智能体
func (s *AgentService) DeleteAgent(id string) error {
	// 先删除技能关联
	if err := s.repo.DeleteAgentSkills(id); err != nil {
		return err
	}
	// 再删除智能体
	return s.repo.Delete(id)
}

// GetAgent 获取智能体
func (s *AgentService) GetAgent(id string) (*AgentResponse, error) {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(agent), nil
}

// GetAgentWithRelations 获取智能体及其关联数据
func (s *AgentService) GetAgentWithRelations(id string) (*AgentResponse, error) {
	agent, err := s.repo.GetWithRelations(id)
	if err != nil {
		return nil, err
	}

	return s.toResponseWithRelations(agent), nil
}

// GetAgentByName 根据名称获取智能体
func (s *AgentService) GetAgentByName(name string) (*AgentResponse, error) {
	agent, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}

	return s.toResponse(agent), nil
}

// ListAgents 列出智能体（支持筛选）
func (s *AgentService) ListAgents(page, pageSize int, query *types.AgentQuery) ([]AgentResponse, int64, error) {
	agents, total, err := s.repo.ListWithQuery(page, pageSize, query)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]AgentResponse, len(agents))
	for i, a := range agents {
		responses[i] = *s.toResponse(&a)
	}

	return responses, total, nil
}

// GetActiveAgents 获取所有活跃的智能体
func (s *AgentService) GetActiveAgents() ([]AgentResponse, error) {
	agents, err := s.repo.GetActiveAgents()
	if err != nil {
		return nil, err
	}

	responses := make([]AgentResponse, len(agents))
	for i, a := range agents {
		responses[i] = *s.toResponse(&a)
	}

	return responses, nil
}

// UpdateAgentStatus 更新智能体状态
func (s *AgentService) UpdateAgentStatus(id string, status types.AgentStatus) error {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	agent.Status = status
	agent.UpdatedAt = time.Now()

	return s.repo.Update(agent)
}

// GetAgentEntity 获取智能体实体
func (s *AgentService) GetAgentEntity(id string) (*types.Agent, error) {
	return s.repo.GetByID(id)
}

// GetAgentEntityWithProvider 获取智能体实体及其供应商
func (s *AgentService) GetAgentEntityWithProvider(id string) (*types.Agent, error) {
	return s.repo.GetWithRelations(id)
}

// toResponse 转换为响应格式
func (s *AgentService) toResponse(agent *types.Agent) *AgentResponse {
	return &AgentResponse{
		ID:                  agent.ID,
		Name:                agent.Name,
		Description:         agent.Description,
		ProviderID:          agent.ProviderID,
		Model:               agent.Model,
		SystemPrompt:        agent.SystemPrompt,
		Temperature:         agent.Temperature,
		MaxIterations:       agent.MaxIterations,
		MaxOutputTokens:     agent.MaxOutputTokens,
		WorkingDir:          agent.WorkingDir,
		CompressionRatio:    agent.CompressionRatio,
		CompressionKeepSize: agent.CompressionKeepSize,
		ContextMaxTokens:    agent.ContextMaxTokens,
		Status:              agent.Status,
		UseAllSkills:        agent.UseAllSkills,
		CreatedAt:           agent.CreatedAt,
		UpdatedAt:           agent.UpdatedAt,
	}
}

// toResponseWithRelations 转换为响应格式（包含关联数据）
func (s *AgentService) toResponseWithRelations(agent *types.Agent) *AgentResponse {
	response := s.toResponse(agent)

	// 如果有 ProviderID，查询供应商信息
	if agent.ProviderID != "" {
		if provider, err := s.providerService.GetProviderEntity(agent.ProviderID); err == nil {
			response.Provider = s.providerService.toResponse(provider)
		}
	}

	// 如果不使用所有技能，查询关联的技能列表
	if !agent.UseAllSkills && s.skillRepo != nil {
		skills, err := s.getSkillBriefs(agent.ID)
		if err == nil {
			response.Skills = skills
		}
	}

	return response
}

// getSkillBriefs 获取智能体关联的技能简要信息
func (s *AgentService) getSkillBriefs(agentID string) ([]SkillBrief, error) {
	skillIDs, err := s.repo.GetAgentSkills(agentID)
	if err != nil {
		return nil, err
	}

	if len(skillIDs) == 0 {
		return []SkillBrief{}, nil
	}

	skills := make([]SkillBrief, 0, len(skillIDs))
	for _, id := range skillIDs {
		if skill, err := s.skillRepo.GetByID(id); err == nil {
			skills = append(skills, SkillBrief{
				ID:          skill.ID,
				Name:        skill.Name,
				Description: skill.Description,
			})
		}
	}

	return skills, nil
}

// GetAgentSkills 获取智能体关联的技能列表
func (s *AgentService) GetAgentSkills(agentID string) ([]SkillBrief, error) {
	skillIDs, err := s.repo.GetAgentSkills(agentID)
	if err != nil {
		return nil, err
	}

	// 如果没有关联技能，返回空列表
	if len(skillIDs) == 0 {
		return []SkillBrief{}, nil
	}

	// 如果有 SkillRepository，获取技能详情
	if s.skillRepo != nil {
		return s.getSkillBriefs(agentID)
	}

	// 否则只返回 ID
	skills := make([]SkillBrief, len(skillIDs))
	for i, id := range skillIDs {
		skills[i] = SkillBrief{ID: id}
	}

	return skills, nil
}

// UpdateAgentSkills 更新智能体技能配置
func (s *AgentService) UpdateAgentSkills(agentID string, useAllSkills bool, skillIDs []string) error {
	// 更新 UseAllSkills 字段
	agent, err := s.repo.GetByID(agentID)
	if err != nil {
		return err
	}

	agent.UseAllSkills = useAllSkills
	agent.UpdatedAt = time.Now()

	if err := s.repo.Update(agent); err != nil {
		return err
	}

	// 更新技能关联
	if !useAllSkills && len(skillIDs) > 0 {
		return s.repo.SetAgentSkills(agentID, skillIDs)
	} else if !useAllSkills {
		// 如果不使用所有技能且没有指定技能，清空关联
		return s.repo.DeleteAgentSkills(agentID)
	}

	return nil
}
