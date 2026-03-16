package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	"github.com/StellarisJAY/beepbot/internal/agent/dify"
	"github.com/StellarisJAY/beepbot/internal/agent/react"
	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/mcp"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

type AgentManager struct {
	agentRepo    repository.AgentRepository
	botRepo      repository.BotRepository
	providerRepo repository.ProviderRepository
	sessionRepo  repository.SessionRepository
	skillRepo    repository.SkillRepository // 技能仓储
	bus          *channel.MessageBus
	encryptor    *crypto.Encryptor
	cronDeps     *tool.CronToolDeps // 定时任务工具依赖
	mcpManager   *mcp.Manager       // MCP 管理器

	config config.APIConfig

	cancelFunc context.CancelFunc
	wg         *sync.WaitGroup
}

func NewAgentManager(
	config config.APIConfig,
	db *gorm.DB,
	agentRepo repository.AgentRepository,
	botRepo repository.BotRepository,
	providerRepo repository.ProviderRepository,
	messageBus *channel.MessageBus,
	encryptor *crypto.Encryptor,
	cronDeps *tool.CronToolDeps,
	mcpManager *mcp.Manager) *AgentManager {
	return &AgentManager{
		config:       config,
		agentRepo:    agentRepo,
		botRepo:      botRepo,
		providerRepo: providerRepo,
		sessionRepo:  repository.NewSessionRepository(db),
		skillRepo:    repository.NewSkillRepository(db),
		bus:          messageBus,
		encryptor:    encryptor,
		cronDeps:     cronDeps,
		mcpManager:   mcpManager,
		wg:           &sync.WaitGroup{},
	}
}

// AgentManager的消息循环，接收所有外部消息，然后根据通道的id，找到绑定的智能体执行
func (a *AgentManager) MessageLoop(ctx context.Context) {
	ctx, cancelFunc := context.WithCancel(ctx)
	a.cancelFunc = cancelFunc
	for {
		select {
		case <-ctx.Done():
			a.cancelFunc()
			slog.Info("agent manager waiting agent runners to stop...")
			a.wg.Wait()
			slog.Info("agent manager stopped")
			return
		default:
			msg, ok := a.bus.ConsumeInbound(ctx)
			if !ok {
				return
			}

			var agentDef *types.Agent

			// 根据 AgentID 是否存在区分消息来源
			if msg.AgentID != "" {
				// 定时任务消息：直接通过 AgentID 查询智能体
				var err error
				agentDef, err = a.agentRepo.GetByID(msg.AgentID)
				if err != nil {
					slog.Error("agent manager query agent by AgentID error", "err", err, "agent_id", msg.AgentID)
					continue
				}
				// 检查智能体状态
				if agentDef.Status == types.AgentStatusInactive {
					slog.Info("agent manager unable to start inactive agent for cron job", "agent_id", msg.AgentID)
					continue
				}
			} else {
				// 机器人消息：通过 Channel 查询机器人绑定的智能体
				channelID := msg.Channel
				bot, err := a.botRepo.GetWithRelations(channelID)
				if err != nil {
					slog.Error("agent manager query channel's binding agent error", "err", err)
					continue
				}
				// 机器人禁用
				if bot.Status == types.BotStatusInactive {
					slog.Info("agent manager unable to start inactive bot's agent")
					continue
				}
				// 机器人没有绑定智能体
				if bot.Agent == nil {
					slog.Error("agent manager start agent error", "err", errors.New("bot has no binding agent"))
					a.bus.PublishOutbound(ctx, channel.OutboundMessage{
						Channel:          msg.Channel,
						Content:          "机器人没有绑定智能体，无法回复消息",
						GroupID:          msg.GroupID,
						UserID:           msg.UserID,
						ChatID:           msg.ChatID,
						InboundMessageID: msg.MessageID,
					})
					continue
				}

				// 智能体禁用，向用户发送禁用消息
				if bot.Agent.Status == types.AgentStatusInactive {
					slog.Info("agent manager unable to start inactive agent")
					a.bus.PublishOutbound(ctx, channel.OutboundMessage{
						Channel:          msg.Channel,
						Content:          "智能体已被禁用",
						GroupID:          msg.GroupID,
						UserID:           msg.UserID,
						ChatID:           msg.ChatID,
						InboundMessageID: msg.MessageID,
					})
					continue
				}
				agentDef = bot.Agent
			}

			if agentDef.External {
				go a.runExternalAgent(ctx, agentDef, msg)
			} else {
				go a.startReActAgent(ctx, agentDef, msg)
			}
		}
	}
}

func (a *AgentManager) startReActAgent(ctx context.Context, agentDef *types.Agent, message channel.InboundMessage) {
	a.wg.Add(1)
	defer a.wg.Done()

	var cronJobID *string
	if message.Channel == channel.ChannelCron {
		cronJobID = &message.UserID
	}

	providerDef, err := a.providerRepo.GetByID(agentDef.ProviderID)
	if err != nil {
		slog.Error("query agent's provider error", "err", err)
		return
	}
	providerDef.APIKey, _ = a.encryptor.Decrypt(providerDef.APIKey)

	sessionType := message.SessionType
	if sessionType == "" {
		sessionType = types.SessionTypeChat
	}

	// 生成会话 Key（使用 ChatID）
	sessionKey := session.GetApiSessionKey(sessionType, agentDef.ID, message.Channel, message.ChatID, message.UserID)

	// 构建 IM 上下文
	imContext := &types.IMSessionContext{
		UserID:  message.UserID,
		GroupID: message.GroupID,
		ChatID:  message.ChatID,
	}

	// 创建或加载会话
	sess, err := session.NewApiSession(
		a.sessionRepo,
		sessionKey,
		agentDef.ID,
		message.Channel,
		message.SessionType,
		cronJobID,
		imContext,
		agentDef.ContextMaxTokens,
		agentDef.CompressionRatio,
		agentDef.CompressionKeepSize,
	)
	if err != nil {
		slog.Error("create sesssion failed", "error", err)
		return
	}

	// 计算会话级 workDir
	sessionWorkDir, err := session.EnsureSessionWorkDir(agentDef.WorkingDir, sessionKey)
	if err != nil {
		slog.Error("create session working dir failed", "error", err)
		return
	}

	runner, err := react.NewApiAgentRunner(react.NewReactAgentParam{
		AgentDef:         agentDef,
		ProviderDef:      providerDef,
		Bus:              a.bus,
		Config:           a.config,
		SessionRepo:      a.sessionRepo,
		SkillRepo:        a.skillRepo,
		CronDeps:         a.cronDeps,
		AgentRepo:        a.agentRepo,
		ProviderRepo:     a.providerRepo,
		Encryptor:        a.encryptor,
		McpManager:       a.mcpManager,
		AllowSubAgents:   true,
		ParentWorkingDir: sessionWorkDir,
	})
	if err != nil {
		slog.Error("create agent runner error", "err", err)
		return
	}

	if err := runner.RunWithMessage(ctx, sess, message); err != nil {
		slog.Error("agent loop error", "err", err)
	}
}

func (a *AgentManager) runExternalAgent(ctx context.Context, agentDef *types.Agent, message channel.InboundMessage) {
	a.wg.Add(1)
	defer a.wg.Done()

	var cronJobID *string
	if message.Channel == channel.ChannelCron {
		cronJobID = &message.UserID
	}

	sessionType := message.SessionType
	if sessionType == "" {
		sessionType = types.SessionTypeChat
	}

	// 生成会话 Key（使用 ChatID）
	sessionKey := session.GetApiSessionKey(sessionType, agentDef.ID, message.Channel, message.ChatID, message.UserID)

	// 构建 IM 上下文
	imContext := &types.IMSessionContext{
		UserID:  message.UserID,
		GroupID: message.GroupID,
		ChatID:  message.ChatID,
	}

	sess, err := session.NewExternalSession(
		a.sessionRepo,
		sessionKey,
		agentDef.ID,
		message.Channel,
		message.SessionType,
		cronJobID,
		imContext,
	)
	if err != nil {
		slog.Error("create external session failed", "error", err)
		return
	}

	// 解析外部配置
	if agentDef.ExternalType != types.ExternalTypeDify {
		slog.Error("unsupported external agent type", "type", agentDef.ExternalType)
		return
	}

	// 解析 Dify 配置
	var difyConfig types.DifyConfig
	if agentDef.ExternalConfig != nil {
		// 解密配置
		var configMap map[string]any
		if err := json.Unmarshal(agentDef.ExternalConfig, &configMap); err != nil {
			slog.Error("parse external config failed", "error", err)
			return
		}

		// 解密敏感字段
		if apiKey, ok := configMap["api_key"].(string); ok && apiKey != "" {
			decrypted, err := a.encryptor.Decrypt(apiKey)
			if err != nil {
				slog.Error("decrypt api_key failed", "error", err)
				return
			}
			difyConfig.APIKey = decrypted
		}

		// 提取配置
		if baseURL, ok := configMap["base_url"].(string); ok {
			difyConfig.BaseURL = baseURL
		}
	}

	if difyConfig.BaseURL == "" || difyConfig.APIKey == "" {
		slog.Error("dify config is incomplete", "base_url", difyConfig.BaseURL)
		return
	}

	runner := dify.NewDifyRunner(difyConfig.BaseURL, difyConfig.APIKey, "blocking", a.bus)

	if err := runner.RunWithMessage(ctx, sess, message); err != nil {
		slog.Error("agent loop error", "err", err)
	}
}
