package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/StellarisJAY/beepbot/internal/agent"
	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/repository"
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
	cronDeps *tool.CronToolDeps) *AgentManager {
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

			// 启动智能体
			go a.startAgentLoop(ctx, agentDef, msg)
		}
	}
}

func (a *AgentManager) startAgentLoop(ctx context.Context, agentDef *types.Agent, msg channel.InboundMessage) {
	a.wg.Add(1)
	defer a.wg.Done()
	providerDef, err := a.providerRepo.GetByID(agentDef.ProviderID)
	if err != nil {
		slog.Error("query agent's provider error", "err", err)
		return
	}
	providerDef.APIKey, _ = a.encryptor.Decrypt(providerDef.APIKey)
	runner, err := agent.NewApiAgentRunner(agentDef, providerDef, a.bus, a.config, a.sessionRepo, a.cronDeps, a.skillRepo, a.agentRepo)
	if err != nil {
		slog.Error("create agent runner error", "err", err)
		return
	}

	if err := runner.RunWithMessage(ctx, msg); err != nil {
		slog.Error("agent loop error", "err", err)
	}
}
