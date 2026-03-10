package agent

import (
	"context"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/provider"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/skill"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// NewApiRunner 创建API模式的智能体运行器
func NewApiRunner(agentDef *types.Agent, providerDef *types.Provider, bus *channel.MessageBus, config config.APIConfig, cronDeps *tool.CronToolDeps) (*AgentRunner, error) {
	// 创建聊天模型接口
	llmProvider, err := provider.CreateLLMProviderFromApi(providerDef, agentDef.Model)
	if err != nil {
		return nil, err
	}
	modelID := strings.ToLower(agentDef.Model)
	// 注册工具
	toolRegistry := tool.NewToolRegistry()
	// 文件系统工具需要传入工作目录，实现访问隔离
	workingDir := agentDef.WorkingDir
	toolRegistry.Register(tool.NewListDirTool(workingDir, config.DataDir))
	toolRegistry.Register(tool.NewReadFileTool(workingDir, config.DataDir))
	toolRegistry.Register(tool.NewWriteFileTool(workingDir, config.DataDir))
	toolRegistry.Register(tool.NewEditFileTool(workingDir, config.DataDir))
	// 待办管理工具
	toolRegistry.Register(tool.NewWriteTodoTool(workingDir))
	// 操作系统信息工具
	toolRegistry.Register(tool.NewReadSystemInfoTool())
	toolRegistry.Register(tool.NewShellToolFromApi(workingDir))
	// 定时任务工具（仅在API模式下注册）
	if cronDeps != nil {
		toolRegistry.Register(tool.NewCronTool(cronDeps))
	}

	// 创建技能管理器
	skillManager := skill.NewManager(config.DataDir, workingDir)

	agentRun := &AgentRunner{
		model:         llmProvider,
		bus:           bus,
		tools:         toolRegistry,
		skillManager:  skillManager,
		modelID:       modelID,
		agentID:       agentDef.ID,
		systemPrompt:  agentDef.SystemPrompt,
		maxIterations: agentDef.MaxIterations,
		workingDir:    agentDef.WorkingDir,
	}
	return agentRun, nil
}

// ApiAgentRunner API 模式的 AgentRunner，包含会话仓储
type ApiAgentRunner struct {
	*AgentRunner
	sessionRepo      repository.SessionRepository
	agentDef         *types.Agent
	windowSize       int
	maxTokens        int64
	compressionRatio float64
}

// NewApiAgentRunner 创建 API 模式的 AgentRunner
func NewApiAgentRunner(
	agentDef *types.Agent,
	providerDef *types.Provider,
	bus *channel.MessageBus,
	config config.APIConfig,
	sessionRepo repository.SessionRepository,
	cronDeps *tool.CronToolDeps,
) (*ApiAgentRunner, error) {
	runner, err := NewApiRunner(agentDef, providerDef, bus, config, cronDeps)
	if err != nil {
		return nil, err
	}

	// 获取会话配置参数
	windowSize := agentDef.WindowSize
	if windowSize <= 0 {
		windowSize = 20
	}
	maxTokens := agentDef.ContextMaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	compressionRatio := agentDef.CompressionRatio
	if compressionRatio <= 0 || compressionRatio >= 1 {
		compressionRatio = 0.7
	}

	return &ApiAgentRunner{
		AgentRunner:      runner,
		sessionRepo:      sessionRepo,
		agentDef:         agentDef,
		windowSize:       windowSize,
		maxTokens:        maxTokens,
		compressionRatio: compressionRatio,
	}, nil
}

// RunWithMessage 使用消息创建会话并运行 AgentLoop
func (r *ApiAgentRunner) RunWithMessage(ctx context.Context, message channel.InboundMessage) error {
	// 生成会话 Key
	sessionKey := r.GetSessionKey(message.Channel, message.GroupID, message.UserID)

	// 创建或加载会话
	sess, err := r.createSession(sessionKey, message.Channel)
	if err != nil {
		return err
	}

	// 运行 AgentLoop
	r.AgentLoop(ctx, sess, message)
	return nil
}

// createSession 创建或加载 API 会话
func (r *ApiAgentRunner) createSession(sessionKey string, botID string) (session.Session, error) {
	return session.NewApiSession(
		r.sessionRepo,
		sessionKey,
		r.agentDef.ID,
		botID,
		r.windowSize,
		r.maxTokens,
		r.compressionRatio,
	)
}

// GetSessionKey 生成会话 Key
func (r *ApiAgentRunner) GetSessionKey(channelID, groupID, userID string) string {
	return session.GetApiSessionKey(r.agentDef.ID, channelID, groupID, userID)
}
