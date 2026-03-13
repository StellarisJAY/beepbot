package agent

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

// toolRegistrationParams 工具注册参数
type toolRegistrationParams struct {
	agentDef      *types.Agent
	workingDir    string
	config        config.APIConfig
	cronDeps      *tool.CronToolDeps
	skillRepo     repository.SkillRepository
	agentRepo     repository.AgentRepository
	providerRepo  repository.ProviderRepository
	encryptor     *crypto.Encryptor
	bus           *channel.MessageBus
	allowSubAgents bool
	mcpManager    *mcp.Manager
}

// registerTools 注册工具到工具注册表
func registerTools(toolRegistry *tool.ToolRegistry, params toolRegistrationParams) {
	// 根据配置决定注册哪些工具
	if params.agentDef.UseAllTools {
		// 注册所有工具（默认行为）
		toolRegistry.Register(tool.NewListDirTool(params.workingDir, params.config.DataDir))
		toolRegistry.Register(tool.NewReadFileTool(params.workingDir, params.config.DataDir))
		toolRegistry.Register(tool.NewWriteFileTool(params.workingDir, params.config.DataDir))
		toolRegistry.Register(tool.NewEditFileTool(params.workingDir, params.config.DataDir))
		// 待办管理工具
		toolRegistry.Register(tool.NewWriteTodoTool(params.workingDir))
		// Shell工具
		// TODO 沙箱环境
		toolRegistry.Register(tool.NewShellToolFromApi(params.workingDir))
		// 定时任务工具（仅在API模式下注册）
		if params.cronDeps != nil {
			toolRegistry.Register(tool.NewCronTool(params.cronDeps))
		}
	} else {
		// 只注册配置的工具
		toolNames, err := params.agentRepo.GetAgentTools(params.agentDef.ID)
		if err == nil {
			for _, name := range toolNames {
				switch name {
				case "list_dir":
					toolRegistry.Register(tool.NewListDirTool(params.workingDir, params.config.DataDir))
				case "read_file":
					toolRegistry.Register(tool.NewReadFileTool(params.workingDir, params.config.DataDir))
				case "write_file":
					toolRegistry.Register(tool.NewWriteFileTool(params.workingDir, params.config.DataDir))
				case "edit_file":
					toolRegistry.Register(tool.NewEditFileTool(params.workingDir, params.config.DataDir))
				case "todo_write":
					toolRegistry.Register(tool.NewWriteTodoTool(params.workingDir))
				case "shell":
					toolRegistry.Register(tool.NewShellToolFromApi(params.workingDir))
				case "cron":
					if params.cronDeps != nil {
						toolRegistry.Register(tool.NewCronTool(params.cronDeps))
					}
				}
			}
		}
	}

	// 注册 Sub-Agent 工具（仅当 allowSubAgents 为 true 时）
	if params.allowSubAgents && params.agentRepo != nil && params.providerRepo != nil && params.encryptor != nil {
		// 创建子智能体执行器
		executor := NewSubAgentExecutor(params.config, params.skillRepo, params.agentRepo, params.providerRepo, params.encryptor, params.cronDeps, params.bus)

		callableAgents, err := params.agentRepo.GetCallableAgents()
		if err == nil {
			for _, agent := range callableAgents {
				// 排除自身
				if agent.ID == params.agentDef.ID {
					continue
				}
				// 跳过禁用的智能体
				if agent.Status != types.AgentStatusActive {
					continue
				}
				// 获取智能体的 Provider 配置
				prov, err := params.providerRepo.GetByID(agent.ProviderID)
				if err != nil {
					continue
				}
				// 解密 API Key
				prov.APIKey, _ = params.encryptor.Decrypt(prov.APIKey)

				// 创建 Sub-Agent 工具（使用执行器）
				// 传入当前工作目录，子智能体将继承此目录
				subAgentTool := tool.NewSubAgentToolWithExecutor(
					&agent,
					prov,
					agent.CallableDescription,
					executor,
					params.workingDir,
				)
				toolRegistry.Register(subAgentTool)
			}
		}
	}

	// 注册 MCP 工具（仅当智能体启用了 MCP 且 mcpManager 不为空时）
	if params.mcpManager != nil && params.agentDef.EnableMCP {
		mcpTools := params.mcpManager.GetAllTools()
		for _, toolDef := range mcpTools {
			// 解析工具名称获取服务器信息
			serverName, toolName, ok := mcp.ParseToolName(toolDef.Function.Name)
			if !ok {
				continue
			}
			// 通过 serverName 查找对应的客户端
			client := params.mcpManager.GetClientByServerName(serverName)
			if client == nil {
				continue
			}
			// 创建 MCP 工具定义
			mcpToolDef := types.MCPToolDefinition{
				Name:        toolName,
				Description: toolDef.Function.Description,
				InputSchema: toolDef.Function.Parameters,
			}
			// 创建工具代理
			proxy := mcp.NewToolProxy(params.mcpManager, client.GetServerID(), serverName, mcpToolDef)
			toolRegistry.Register(proxy)
		}
	}
}

// NewApiRunner 创建API模式的智能体运行器
// allowSubAgents: 是否注册 Sub-Agent 工具（父智能体为 true，子智能体为 false 防止递归）
// parentWorkingDir: 父智能体的工作目录，子智能体继承此目录；为空时使用智能体自己的工作目录
// mcpManager: MCP 管理器，用于注册 MCP 工具；为 nil 时不注册 MCP 工具
func NewApiRunner(
	agentDef *types.Agent,
	providerDef *types.Provider,
	bus *channel.MessageBus,
	config config.APIConfig,
	cronDeps *tool.CronToolDeps,
	skillRepo repository.SkillRepository,
	agentRepo repository.AgentRepository,
	providerRepo repository.ProviderRepository,
	encryptor *crypto.Encryptor,
	allowSubAgents bool,
	parentWorkingDir string,
	mcpManager *mcp.Manager,
) (*AgentRunner, error) {
	// 创建聊天模型接口
	llmProvider, err := provider.CreateLLMProviderFromApi(providerDef, agentDef.Model)
	if err != nil {
		return nil, err
	}
	modelID := strings.ToLower(agentDef.Model)
	// 注册工具
	toolRegistry := tool.NewToolRegistry()
	// 文件系统工具需要传入工作目录，实现访问隔离
	// 子智能体继承父智能体的工作目录
	workingDir := agentDef.WorkingDir
	if parentWorkingDir != "" {
		workingDir = parentWorkingDir
	}

	// 注册工具
	registerTools(toolRegistry, toolRegistrationParams{
		agentDef:       agentDef,
		workingDir:     workingDir,
		config:         config,
		cronDeps:       cronDeps,
		skillRepo:      skillRepo,
		agentRepo:      agentRepo,
		providerRepo:   providerRepo,
		encryptor:      encryptor,
		bus:            bus,
		allowSubAgents: allowSubAgents,
		mcpManager:     mcpManager,
	})

	// 创建技能管理器
	skillManager := skill.NewManager(skillRepo, agentRepo, agentDef.ID, config.DataDir)

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
	sessionRepo         repository.SessionRepository
	agentDef            *types.Agent
	providerDef         *types.Provider
	maxTokens           int64
	compressionRatio    float64
	compressionKeepSize int
	// 工具注册依赖
	config       config.APIConfig
	cronDeps     *tool.CronToolDeps
	skillRepo    repository.SkillRepository
	agentRepo    repository.AgentRepository
	providerRepo repository.ProviderRepository
	encryptor    *crypto.Encryptor
	mcpManager   *mcp.Manager
}

// NewApiAgentRunner 创建 API 模式的 AgentRunner
func NewApiAgentRunner(
	agentDef *types.Agent,
	providerDef *types.Provider,
	bus *channel.MessageBus,
	config config.APIConfig,
	sessionRepo repository.SessionRepository,
	cronDeps *tool.CronToolDeps,
	skillRepo repository.SkillRepository,
	agentRepo repository.AgentRepository,
	providerRepo repository.ProviderRepository,
	encryptor *crypto.Encryptor,
	mcpManager *mcp.Manager,
) (*ApiAgentRunner, error) {
	// allowSubAgents=true，父智能体可以调用子智能体
	// parentWorkingDir="" 表示父智能体使用自己的工作目录
	runner, err := NewApiRunner(agentDef, providerDef, bus, config, cronDeps, skillRepo, agentRepo, providerRepo, encryptor, true, "", mcpManager)
	if err != nil {
		return nil, err
	}

	// 获取会话配置参数
	maxTokens := agentDef.ContextMaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	compressionRatio := agentDef.CompressionRatio
	if compressionRatio <= 0 || compressionRatio >= 1 {
		compressionRatio = 0.7
	}
	compressionKeepSize := agentDef.CompressionKeepSize
	if compressionKeepSize <= 0 {
		compressionKeepSize = 5
	}

	return &ApiAgentRunner{
		AgentRunner:         runner,
		sessionRepo:         sessionRepo,
		agentDef:            agentDef,
		providerDef:         providerDef,
		maxTokens:           maxTokens,
		compressionRatio:    compressionRatio,
		compressionKeepSize: compressionKeepSize,
		config:              config,
		cronDeps:            cronDeps,
		skillRepo:           skillRepo,
		agentRepo:           agentRepo,
		providerRepo:        providerRepo,
		encryptor:           encryptor,
		mcpManager:          mcpManager,
	}, nil
}

// RunWithMessage 使用消息创建会话并运行 AgentLoop
func (r *ApiAgentRunner) RunWithMessage(ctx context.Context, message channel.InboundMessage) error {
	// 获取会话类型，默认为聊天类型
	sessionType := message.SessionType
	if sessionType == "" {
		sessionType = types.SessionTypeChat
	}

	// 生成会话 Key（使用 ChatID）
	sessionKey := r.GetSessionKey(sessionType, message.Channel, message.ChatID, message.UserID)

	// 提取 cronJobID（仅定时任务会话）
	var cronJobID *string
	if sessionType == types.SessionTypeCron && message.Channel == channel.ChannelCron {
		cronJobID = &message.UserID
	}

	// 构建 IM 上下文
	imContext := &types.IMSessionContext{
		UserID:  message.UserID,
		GroupID: message.GroupID,
		ChatID:  message.ChatID,
	}

	// 创建或加载会话
	sess, err := r.createSession(sessionKey, message.Channel, sessionType, cronJobID, imContext)
	if err != nil {
		return err
	}

	// 计算会话级 workDir
	sessionWorkDir, err := session.EnsureSessionWorkDir(r.agentDef.WorkingDir, sessionKey)
	if err != nil {
		return err
	}

	// 重新注册工具（使用会话级 workDir）
	toolRegistry := tool.NewToolRegistry()
	registerTools(toolRegistry, toolRegistrationParams{
		agentDef:       r.agentDef,
		workingDir:     sessionWorkDir,
		config:         r.config,
		cronDeps:       r.cronDeps,
		skillRepo:      r.skillRepo,
		agentRepo:      r.agentRepo,
		providerRepo:   r.providerRepo,
		encryptor:      r.encryptor,
		bus:            r.bus,
		allowSubAgents: true, // 允许调用子智能体
		mcpManager:     r.mcpManager,
	})

	// 更新工具注册表和工作目录
	r.AgentRunner.tools = toolRegistry
	r.AgentRunner.workingDir = sessionWorkDir

	// 运行 AgentLoop，传递 nil 使用默认的 BusOutputHook
	r.AgentLoop(ctx, sess, message, nil)
	return nil
}

// createSession 创建或加载 API 会话
func (r *ApiAgentRunner) createSession(
	sessionKey string,
	botID string,
	sessionType types.SessionType,
	cronJobID *string,
	imContext *types.IMSessionContext,
) (session.Session, error) {
	return session.NewApiSession(
		r.sessionRepo,
		sessionKey,
		r.agentDef.ID,
		botID,
		sessionType,
		cronJobID,
		imContext,
		r.maxTokens,
		r.compressionRatio,
		r.compressionKeepSize,
	)
}

// GetSessionKey 生成会话 Key
func (r *ApiAgentRunner) GetSessionKey(sessionType types.SessionType, channelID, chatID, userID string) string {
	return session.GetApiSessionKey(sessionType, r.agentDef.ID, channelID, chatID, userID)
}