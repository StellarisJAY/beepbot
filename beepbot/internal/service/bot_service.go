package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/datatypes"
)

// ChannelManager 接口定义，用于解耦 BotService 和具体的 ChannelManager 实现
type ChannelManager interface {
	StartChannel(ctx context.Context, bot *types.Bot) error
	StopChannel(botID string) error
	RestartChannel(ctx context.Context, bot *types.Bot) error
	IsChannelRunning(botID string) bool
}

// BotService 机器人服务
type BotService struct {
	repo           repository.BotRepository
	agentService   *AgentService
	channelManager ChannelManager
}

func NewBotService(repo repository.BotRepository, agentService *AgentService) *BotService {
	return &BotService{
		repo:         repo,
		agentService: agentService,
	}
}

// SetChannelManager 设置 ChannelManager
// 用于在 API 模式下注入 ChannelManager
func (s *BotService) SetChannelManager(cm ChannelManager) {
	s.channelManager = cm
}

// CreateBotRequest 创建机器人请求
type CreateBotRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Platform    types.BotPlatform `json:"platform" binding:"required"`
	Identifier  string            `json:"identifier"`
	Config      map[string]any    `json:"config"`
	AgentID     *string           `json:"agent_id"`
}

// UpdateBotRequest 更新机器人请求
type UpdateBotRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Identifier  string          `json:"identifier"`
	Config      map[string]any  `json:"config"`
	Status      types.BotStatus `json:"status"`
}

// BindAgentRequest 绑定智能体请求
type BindAgentRequest struct {
	AgentID *string `json:"agent_id"`
}

// BotResponse 机器人响应
type BotResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Platform    types.BotPlatform `json:"platform"`
	Identifier  string            `json:"identifier"`
	Config      map[string]any    `json:"config"`
	AgentID     *string           `json:"agent_id"`
	Agent       *AgentResponse    `json:"agent,omitempty"`
	Status      types.BotStatus   `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CreateBot 创建机器人
func (s *BotService) CreateBot(ctx context.Context, req *CreateBotRequest) (*types.Bot, error) {
	// 检查名称是否已存在
	if _, err := s.repo.GetByName(req.Name); err == nil {
		return nil, errors.New("bot name already exists")
	}

	// 如果指定了智能体，检查智能体是否存在
	if req.AgentID != nil && *req.AgentID != "" {
		if _, err := s.agentService.GetAgent(*req.AgentID); err != nil {
			return nil, errors.New("agent not found")
		}
	}

	// 处理 Config
	var config datatypes.JSON
	if req.Config != nil {
		data, _ := json.Marshal(req.Config)
		config = data
	}

	bot := &types.Bot{
		ID:          types.GenerateUUIDv7(),
		Name:        req.Name,
		Description: req.Description,
		Platform:    req.Platform,
		Identifier:  req.Identifier,
		Config:      config,
		AgentID:     req.AgentID,
		Status:      types.BotStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(bot); err != nil {
		return nil, err
	}

	// 如果 Bot 状态为 active，启动 Channel
	if bot.Status == types.BotStatusActive && s.channelManager != nil {
		if err := s.channelManager.StartChannel(ctx, bot); err != nil {
			slog.Error("failed to start channel for new bot", "bot_id", bot.ID, "error", err)
			// 不返回错误，Bot 已创建成功，只是 Channel 启动失败
		}
	}

	return bot, nil
}

// UpdateBot 更新机器人
func (s *BotService) UpdateBot(ctx context.Context, id string, req *UpdateBotRequest) (*types.Bot, error) {
	bot, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	oldStatus := bot.Status
	configChanged := false

	if req.Name != "" {
		// 检查名称是否被其他机器人使用
		if existing, err := s.repo.GetByName(req.Name); err == nil && existing.ID != id {
			return nil, errors.New("bot name already exists")
		}
		bot.Name = req.Name
	}

	if req.Description != "" {
		bot.Description = req.Description
	}

	if req.Identifier != "" {
		bot.Identifier = req.Identifier
	}

	if req.Config != nil {
		data, _ := json.Marshal(req.Config)
		bot.Config = data
		configChanged = true
	}

	if req.Status != "" {
		bot.Status = req.Status
	}

	bot.UpdatedAt = time.Now()

	if err := s.repo.Update(bot); err != nil {
		return nil, err
	}

	// 处理 Channel 生命周期
	if s.channelManager != nil {
		// 状态变更
		if oldStatus != bot.Status {
			if bot.Status == types.BotStatusInactive {
				// active -> inactive: 停止 Channel
				if err := s.channelManager.StopChannel(bot.ID); err != nil {
					slog.Error("failed to stop channel", "bot_id", bot.ID, "error", err)
				}
			} else if bot.Status == types.BotStatusActive {
				// inactive -> active: 启动 Channel
				if err := s.channelManager.StartChannel(ctx, bot); err != nil {
					slog.Error("failed to start channel", "bot_id", bot.ID, "error", err)
				}
			}
		} else if configChanged && bot.Status == types.BotStatusActive {
			// 配置变更且 Bot 处于 active 状态，重启 Channel
			if err := s.channelManager.RestartChannel(ctx, bot); err != nil {
				slog.Error("failed to restart channel", "bot_id", bot.ID, "error", err)
			}
		}
	}

	return bot, nil
}

// DeleteBot 删除机器人
func (s *BotService) DeleteBot(id string) error {
	// 先停止 Channel
	if s.channelManager != nil {
		if err := s.channelManager.StopChannel(id); err != nil {
			slog.Warn("failed to stop channel before delete", "bot_id", id, "error", err)
		}
	}
	return s.repo.Delete(id)
}

// GetBot 获取机器人
func (s *BotService) GetBot(id string) (*BotResponse, error) {
	bot, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(bot), nil
}

// GetBotWithRelations 获取机器人及其关联数据
func (s *BotService) GetBotWithRelations(id string) (*BotResponse, error) {
	bot, err := s.repo.GetWithRelations(id)
	if err != nil {
		return nil, err
	}

	return s.toResponseWithRelations(bot), nil
}

// ListBots 列出机器人（支持筛选）
func (s *BotService) ListBots(page, pageSize int, query *types.BotQuery) ([]BotResponse, int64, error) {
	bots, total, err := s.repo.ListWithQuery(page, pageSize, query)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]BotResponse, len(bots))
	for i, b := range bots {
		responses[i] = *s.toResponse(&b)
	}

	return responses, total, nil
}

// BindAgent 绑定智能体
func (s *BotService) BindAgent(botID string, agentID *string) error {
	// 检查机器人是否存在
	if _, err := s.repo.GetByID(botID); err != nil {
		return errors.New("bot not found")
	}

	// 如果指定了智能体，检查智能体是否存在
	if agentID != nil && *agentID != "" {
		if _, err := s.agentService.GetAgent(*agentID); err != nil {
			return errors.New("agent not found")
		}
	}

	return s.repo.BindAgent(botID, agentID)
}

// GetUnboundBots 获取未绑定智能体的机器人
func (s *BotService) GetUnboundBots() ([]BotResponse, error) {
	bots, err := s.repo.GetUnbound()
	if err != nil {
		return nil, err
	}

	responses := make([]BotResponse, len(bots))
	for i, b := range bots {
		responses[i] = *s.toResponse(&b)
	}

	return responses, nil
}

// GetBotsByAgent 根据智能体ID获取机器人列表
func (s *BotService) GetBotsByAgent(agentID string) ([]BotResponse, error) {
	bots, err := s.repo.GetByAgent(agentID)
	if err != nil {
		return nil, err
	}

	responses := make([]BotResponse, len(bots))
	for i, b := range bots {
		responses[i] = *s.toResponse(&b)
	}

	return responses, nil
}

// GetBotsByPlatform 根据平台获取机器人列表
func (s *BotService) GetBotsByPlatform(platform types.BotPlatform) ([]BotResponse, error) {
	bots, err := s.repo.GetByPlatform(platform)
	if err != nil {
		return nil, err
	}

	responses := make([]BotResponse, len(bots))
	for i, b := range bots {
		responses[i] = *s.toResponse(&b)
	}

	return responses, nil
}

// UpdateBotStatus 更新机器人状态
func (s *BotService) UpdateBotStatus(ctx context.Context, id string, status types.BotStatus) error {
	bot, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	oldStatus := bot.Status
	bot.Status = status
	bot.UpdatedAt = time.Now()

	if err := s.repo.Update(bot); err != nil {
		return err
	}

	// 处理 Channel 生命周期
	if s.channelManager != nil && oldStatus != status {
		if status == types.BotStatusInactive {
			// active -> inactive: 停止 Channel
			if err := s.channelManager.StopChannel(id); err != nil {
				slog.Error("failed to stop channel", "bot_id", id, "error", err)
			}
		} else if status == types.BotStatusActive {
			// inactive -> active: 启动 Channel
			if err := s.channelManager.StartChannel(ctx, bot); err != nil {
				slog.Error("failed to start channel", "bot_id", id, "error", err)
			}
		}
	}

	return nil
}

// toResponse 转换为响应格式
func (s *BotService) toResponse(bot *types.Bot) *BotResponse {
	var config map[string]any
	if bot.Config != nil {
		_ = json.Unmarshal(bot.Config, &config)
	}

	return &BotResponse{
		ID:          bot.ID,
		Name:        bot.Name,
		Description: bot.Description,
		Platform:    bot.Platform,
		Identifier:  bot.Identifier,
		Config:      config,
		AgentID:     bot.AgentID,
		Status:      bot.Status,
		CreatedAt:   bot.CreatedAt,
		UpdatedAt:   bot.UpdatedAt,
	}
}

// toResponseWithRelations 转换为响应格式（包含关联数据）
func (s *BotService) toResponseWithRelations(bot *types.Bot) *BotResponse {
	response := s.toResponse(bot)

	if bot.Agent != nil {
		response.Agent = s.agentService.toResponse(bot.Agent)
	}

	return response
}
