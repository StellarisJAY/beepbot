package react

import (
	"context"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/mcp"
	"github.com/StellarisJAY/beepbot/internal/provider"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/skill"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// ApiAgentRunner API 模式的 AgentRunner，包含会话仓储
type ApiAgentRunner struct {
	*ReActAgentRunner
	sessionRepo repository.SessionRepository
	agentDef    *types.Agent
	providerDef *types.Provider
	// 工具注册依赖
	config       config.APIConfig
	cronDeps     *tool.CronToolDeps
	skillRepo    repository.SkillRepository
	agentRepo    repository.AgentRepository
	providerRepo repository.ProviderRepository
	encryptor    *crypto.Encryptor
	mcpManager   *mcp.Manager
}

type NewReactAgentParam struct {
	AgentDef     *types.Agent
	ProviderDef  *types.Provider
	Bus          *channel.MessageBus
	Config       config.APIConfig
	SessionRepo  repository.SessionRepository
	CronDeps     *tool.CronToolDeps
	SkillRepo    repository.SkillRepository
	AgentRepo    repository.AgentRepository
	ProviderRepo repository.ProviderRepository
	Encryptor    *crypto.Encryptor
	McpManager   *mcp.Manager

	AllowSubAgents   bool
	ParentWorkingDir string
}

// NewApiAgentRunner 创建 API 模式的 AgentRunner
func NewApiAgentRunner(param NewReactAgentParam) (*ApiAgentRunner, error) {
	// 创建聊天模型接口
	llmProvider, err := provider.CreateLLMProviderFromApi(param.ProviderDef, param.AgentDef.Model)
	if err != nil {
		return nil, err
	}
	modelID := strings.ToLower(param.AgentDef.Model)
	// 注册工具
	toolRegistry := tool.NewToolRegistry()
	// 文件系统工具需要传入工作目录，实现访问隔离
	// 子智能体继承父智能体的工作目录
	workingDir := param.AgentDef.WorkingDir
	if param.ParentWorkingDir != "" {
		workingDir = param.ParentWorkingDir
	}

	// 注册工具
	registerTools(toolRegistry, toolRegistrationParams{
		agentDef:       param.AgentDef,
		workingDir:     workingDir,
		config:         param.Config,
		cronDeps:       param.CronDeps,
		skillRepo:      param.SkillRepo,
		agentRepo:      param.AgentRepo,
		providerRepo:   param.ProviderRepo,
		encryptor:      param.Encryptor,
		bus:            param.Bus,
		allowSubAgents: param.AllowSubAgents,
		mcpManager:     param.McpManager,
	})

	// 创建技能管理器
	skillManager := skill.NewManager(param.SkillRepo, param.AgentRepo, param.AgentDef.ID, param.Config.DataDir)

	// 获取会话配置参数
	maxTokens := param.AgentDef.ContextMaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	compressionRatio := param.AgentDef.CompressionRatio
	if compressionRatio <= 0 || compressionRatio >= 1 {
		compressionRatio = 0.7
	}
	compressionKeepSize := param.AgentDef.CompressionKeepSize
	if compressionKeepSize <= 0 {
		compressionKeepSize = 5
	}

	runner := &ReActAgentRunner{
		model:         llmProvider,
		bus:           param.Bus,
		tools:         toolRegistry,
		skillManager:  skillManager,
		modelID:       modelID,
		agentID:       param.AgentDef.ID,
		systemPrompt:  param.AgentDef.SystemPrompt,
		maxIterations: param.AgentDef.MaxIterations,
		workingDir:    param.AgentDef.WorkingDir,

		maxTokens:           maxTokens,
		compressionRatio:    compressionRatio,
		compressionKeepSize: compressionKeepSize,
	}

	return &ApiAgentRunner{
		ReActAgentRunner: runner,
		sessionRepo:      param.SessionRepo,
		agentDef:         param.AgentDef,
		providerDef:      param.ProviderDef,
		config:           param.Config,
		cronDeps:         param.CronDeps,
		skillRepo:        param.SkillRepo,
		agentRepo:        param.AgentRepo,
		providerRepo:     param.ProviderRepo,
		encryptor:        param.Encryptor,
		mcpManager:       param.McpManager,
	}, nil
}

// RunWithMessage 使用消息创建会话并运行 AgentLoop
func (r *ApiAgentRunner) RunWithMessage(ctx context.Context, sess session.Session, message channel.InboundMessage) error {
	// 运行 AgentLoop，传递 nil 使用默认的 BusOutputHook
	r.AgentLoop(ctx, sess, message, nil)
	return nil
}
