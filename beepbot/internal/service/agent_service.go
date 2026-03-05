package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/datatypes"
)

// AgentService 智能体服务
type AgentService struct {
	repo            repository.AgentRepository
	providerService *ProviderService
}

func NewAgentService(repo repository.AgentRepository, providerService *ProviderService) *AgentService {
	return &AgentService{
		repo:            repo,
		providerService: providerService,
	}
}

// CreateAgentRequest 创建智能体请求
type CreateAgentRequest struct {
	Name              string  `json:"name" binding:"required"`
	Description       string  `json:"description"`
	ProviderID        string  `json:"provider_id" binding:"required"`
	Model             string  `json:"model" binding:"required"`
	SystemPrompt      string  `json:"system_prompt"`
	Temperature       float32 `json:"temperature"`
	MaxIterations     int     `json:"max_iterations"`
	MaxOutputTokens   int64   `json:"max_output_tokens"`
	WorkingDir        string  `json:"working_dir" binding:"required"`
	ContextWindowSize int     `json:"context_window_size"`
	WindowSize        int     `json:"window_size"`
	CompressionRatio  float64 `json:"compression_ratio"`
	ContextMaxTokens  int64   `json:"context_max_tokens"`
}

// UpdateAgentRequest 更新智能体请求
type UpdateAgentRequest struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ProviderID        string            `json:"provider_id"`
	Model             string            `json:"model"`
	SystemPrompt      string            `json:"system_prompt"`
	Temperature       float32           `json:"temperature"`
	MaxIterations     int               `json:"max_iterations"`
	MaxOutputTokens   int64             `json:"max_output_tokens"`
	WorkingDir        string            `json:"working_dir"`
	ContextWindowSize int               `json:"context_window_size"`
	WindowSize        int               `json:"window_size"`
	CompressionRatio  float64           `json:"compression_ratio"`
	ContextMaxTokens  int64             `json:"context_max_tokens"`
	Status            types.AgentStatus `json:"status"`
}

// CreateChannelRequest 创建渠道绑定请求
type CreateChannelRequest struct {
	ChannelType       string         `json:"channel_type" binding:"required"`
	ChannelIdentifier string         `json:"channel_identifier"`
	Config            map[string]any `json:"config"`
}

// UpdateChannelRequest 更新渠道绑定请求
type UpdateChannelRequest struct {
	ChannelIdentifier string         `json:"channel_identifier"`
	Config            map[string]any `json:"config"`
}

// AgentResponse 智能体响应
type AgentResponse struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ProviderID        string            `json:"provider_id"`
	Provider          *ProviderResponse `json:"provider,omitempty"`
	Model             string            `json:"model"`
	SystemPrompt      string            `json:"system_prompt"`
	Temperature       float32           `json:"temperature"`
	MaxIterations     int               `json:"max_iterations"`
	MaxOutputTokens   int64             `json:"max_output_tokens"`
	WorkingDir        string            `json:"working_dir"`
	ContextWindowSize int               `json:"context_window_size"`
	WindowSize        int               `json:"window_size"`
	CompressionRatio  float64           `json:"compression_ratio"`
	ContextMaxTokens  int64             `json:"context_max_tokens"`
	Status            types.AgentStatus `json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Channels          []ChannelResponse `json:"channels,omitempty"`
}

// ChannelResponse 渠道绑定响应
type ChannelResponse struct {
	ID                string         `json:"id"`
	AgentID           string         `json:"agent_id"`
	ChannelType       string         `json:"channel_type"`
	ChannelIdentifier string         `json:"channel_identifier"`
	Config            map[string]any `json:"config"`
	CreatedAt         time.Time      `json:"created_at"`
}

// CreateAgent 创建智能体
func (s *AgentService) CreateAgent(req *CreateAgentRequest) (*types.Agent, error) {
	// 检查名称是否已存在
	if _, err := s.repo.GetByName(req.Name); err == nil {
		return nil, errors.New("agent name already exists")
	}

	// 检查供应商是否存在
	if _, err := s.providerService.GetProvider(req.ProviderID); err != nil {
		return nil, errors.New("provider not found")
	}

	// 设置默认值
	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}
	maxIterations := req.MaxIterations
	if maxIterations == 0 {
		maxIterations = 50
	}
	maxOutputTokens := req.MaxOutputTokens
	if maxOutputTokens == 0 {
		maxOutputTokens = 4096
	}
	contextWindowSize := req.ContextWindowSize
	if contextWindowSize == 0 {
		contextWindowSize = 20
	}
	windowSize := req.WindowSize
	if windowSize == 0 {
		windowSize = 20
	}
	compressionRatio := req.CompressionRatio
	if compressionRatio == 0 {
		compressionRatio = 0.7
	}
	contextMaxTokens := req.ContextMaxTokens
	if contextMaxTokens == 0 {
		contextMaxTokens = 4096
	}

	agent := &types.Agent{
		ID:                types.GenerateUUIDv7(),
		Name:              req.Name,
		Description:       req.Description,
		ProviderID:        req.ProviderID,
		Model:             req.Model,
		SystemPrompt:      req.SystemPrompt,
		Temperature:       temperature,
		MaxIterations:     maxIterations,
		MaxOutputTokens:   maxOutputTokens,
		WorkingDir:        req.WorkingDir,
		ContextWindowSize: contextWindowSize,
		WindowSize:        windowSize,
		CompressionRatio:  compressionRatio,
		ContextMaxTokens:  contextMaxTokens,
		Status:            types.AgentStatusActive,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.Create(agent); err != nil {
		return nil, err
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

	if req.ContextWindowSize != 0 {
		agent.ContextWindowSize = req.ContextWindowSize
	}

	if req.WindowSize != 0 {
		agent.WindowSize = req.WindowSize
	}

	if req.CompressionRatio != 0 {
		agent.CompressionRatio = req.CompressionRatio
	}

	if req.ContextMaxTokens != 0 {
		agent.ContextMaxTokens = req.ContextMaxTokens
	}

	if req.Status != "" {
		agent.Status = req.Status
	}

	agent.UpdatedAt = time.Now()

	if err := s.repo.Update(agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// DeleteAgent 删除智能体
func (s *AgentService) DeleteAgent(id string) error {
	// 先删除关联的渠道绑定
	channels, err := s.repo.GetChannelsByAgent(id)
	if err != nil {
		return err
	}
	for _, ch := range channels {
		_ = s.repo.DeleteChannel(ch.ID)
	}

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

// ListAgents 列出智能体
func (s *AgentService) ListAgents(page, pageSize int) ([]AgentResponse, int64, error) {
	agents, total, err := s.repo.List(page, pageSize)
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

// GetAgentByChannel 根据渠道获取智能体
func (s *AgentService) GetAgentByChannel(channelType, channelIdentifier string) (*AgentResponse, error) {
	agent, err := s.repo.GetByChannel(channelType, channelIdentifier)
	if err != nil {
		return nil, err
	}

	return s.toResponse(agent), nil
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

// CreateChannel 创建渠道绑定
func (s *AgentService) CreateChannel(agentID string, req *CreateChannelRequest) (*types.AgentChannel, error) {
	// 检查智能体是否存在
	if _, err := s.repo.GetByID(agentID); err != nil {
		return nil, errors.New("agent not found")
	}

	// 处理 Config
	var config datatypes.JSON
	if req.Config != nil {
		data, _ := json.Marshal(req.Config)
		config = data
	}

	channel := &types.AgentChannel{
		ID:                types.GenerateUUIDv7(),
		AgentID:           agentID,
		ChannelType:       req.ChannelType,
		ChannelIdentifier: req.ChannelIdentifier,
		Config:            config,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.CreateChannel(channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// UpdateChannel 更新渠道绑定
func (s *AgentService) UpdateChannel(id string, req *UpdateChannelRequest) (*types.AgentChannel, error) {
	channel, err := s.repo.GetChannelByID(id)
	if err != nil {
		return nil, err
	}

	if req.ChannelIdentifier != "" {
		channel.ChannelIdentifier = req.ChannelIdentifier
	}

	if req.Config != nil {
		data, _ := json.Marshal(req.Config)
		channel.Config = data
	}

	if err := s.repo.UpdateChannel(channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// DeleteChannel 删除渠道绑定
func (s *AgentService) DeleteChannel(id string) error {
	return s.repo.DeleteChannel(id)
}

// GetChannelsByAgent 获取智能体的所有渠道绑定
func (s *AgentService) GetChannelsByAgent(agentID string) ([]ChannelResponse, error) {
	channels, err := s.repo.GetChannelsByAgent(agentID)
	if err != nil {
		return nil, err
	}

	responses := make([]ChannelResponse, len(channels))
	for i, ch := range channels {
		responses[i] = *s.channelToResponse(&ch)
	}

	return responses, nil
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
		ID:                agent.ID,
		Name:              agent.Name,
		Description:       agent.Description,
		ProviderID:        agent.ProviderID,
		Model:             agent.Model,
		SystemPrompt:      agent.SystemPrompt,
		Temperature:       agent.Temperature,
		MaxIterations:     agent.MaxIterations,
		MaxOutputTokens:   agent.MaxOutputTokens,
		WorkingDir:        agent.WorkingDir,
		ContextWindowSize: agent.ContextWindowSize,
		WindowSize:        agent.WindowSize,
		CompressionRatio:  agent.CompressionRatio,
		ContextMaxTokens:  agent.ContextMaxTokens,
		Status:            agent.Status,
		CreatedAt:         agent.CreatedAt,
		UpdatedAt:         agent.UpdatedAt,
	}
}

// toResponseWithRelations 转换为响应格式（包含关联数据）
func (s *AgentService) toResponseWithRelations(agent *types.Agent) *AgentResponse {
	response := s.toResponse(agent)

	// 添加供应商信息
	if agent.Provider != nil {
		response.Provider = s.providerService.toResponse(agent.Provider)
	}

	return response
}

// channelToResponse 转换渠道绑定为响应格式
func (s *AgentService) channelToResponse(channel *types.AgentChannel) *ChannelResponse {
	var config map[string]any
	if channel.Config != nil {
		_ = json.Unmarshal(channel.Config, &config)
	}

	return &ChannelResponse{
		ID:                channel.ID,
		AgentID:           channel.AgentID,
		ChannelType:       channel.ChannelType,
		ChannelIdentifier: channel.ChannelIdentifier,
		Config:            config,
		CreatedAt:         channel.CreatedAt,
	}
}
