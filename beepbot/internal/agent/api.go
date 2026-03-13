package agent

import (
	"context"
	"log/slog"
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

	// 根据配置决定注册哪些工具
	if agentDef.UseAllTools {
		// 注册所有工具（默认行为）
		toolRegistry.Register(tool.NewListDirTool(workingDir, config.DataDir))
		toolRegistry.Register(tool.NewReadFileTool(workingDir, config.DataDir))
		toolRegistry.Register(tool.NewWriteFileTool(workingDir, config.DataDir))
		toolRegistry.Register(tool.NewEditFileTool(workingDir, config.DataDir))
		// 待办管理工具
		toolRegistry.Register(tool.NewWriteTodoTool(workingDir))
		// Shell工具
		// TODO 沙箱环境
		toolRegistry.Register(tool.NewShellToolFromApi(workingDir))
		// 定时任务工具（仅在API模式下注册）
		if cronDeps != nil {
			toolRegistry.Register(tool.NewCronTool(cronDeps))
		}
	} else {
		// 只注册配置的工具
		toolNames, err := agentRepo.GetAgentTools(agentDef.ID)
		if err == nil {
			for _, name := range toolNames {
				switch name {
				case "list_dir":
					toolRegistry.Register(tool.NewListDirTool(workingDir, config.DataDir))
				case "read_file":
					toolRegistry.Register(tool.NewReadFileTool(workingDir, config.DataDir))
				case "write_file":
					toolRegistry.Register(tool.NewWriteFileTool(workingDir, config.DataDir))
				case "edit_file":
					toolRegistry.Register(tool.NewEditFileTool(workingDir, config.DataDir))
				case "todo_write":
					toolRegistry.Register(tool.NewWriteTodoTool(workingDir))
				case "shell":
					toolRegistry.Register(tool.NewShellToolFromApi(workingDir))
				case "cron":
					if cronDeps != nil {
						toolRegistry.Register(tool.NewCronTool(cronDeps))
					}
				}
			}
		}
	}

	// 注册 Sub-Agent 工具（仅当 allowSubAgents 为 true 时）
	if allowSubAgents && agentRepo != nil && providerRepo != nil && encryptor != nil {
		// 创建子智能体执行器
		executor := NewSubAgentExecutor(config, skillRepo, agentRepo, providerRepo, encryptor, cronDeps, bus)

		callableAgents, err := agentRepo.GetCallableAgents()
		if err == nil {
			for _, agent := range callableAgents {
				// 排除自身
				if agent.ID == agentDef.ID {
					continue
				}
				// 跳过禁用的智能体
				if agent.Status != types.AgentStatusActive {
					continue
				}
				// 获取智能体的 Provider 配置
				prov, err := providerRepo.GetByID(agent.ProviderID)
				if err != nil {
					continue
				}
				// 解密 API Key
				prov.APIKey, _ = encryptor.Decrypt(prov.APIKey)

				// 创建 Sub-Agent 工具（使用执行器）
				// 传入当前工作目录，子智能体将继承此目录
				subAgentTool := tool.NewSubAgentToolWithExecutor(
					&agent,
					prov,
					agent.CallableDescription,
					executor,
					workingDir,
				)
				toolRegistry.Register(subAgentTool)
			}
		}
	}

	// 注册 MCP 工具（仅当智能体启用了 MCP 且 mcpManager 不为空时）
	if mcpManager != nil && agentDef.EnableMCP {
		mcpTools := mcpManager.GetAllTools()
		slog.Info("mcp tools", "count", len(mcpTools))
		for _, toolDef := range mcpTools {
			// 解析工具名称获取服务器信息
			serverName, toolName, ok := mcp.ParseToolName(toolDef.Function.Name)
			if !ok {
				slog.Info("parse mcp name failed", "toolDef", toolDef.Type)
				continue
			}
			// 通过 serverName 查找对应的客户端（GetClient 需要 serverID，但这里只有 serverName）
			client := mcpManager.GetClientByServerName(serverName)
			if client == nil {
				slog.Info("mcp client not exist", "serverName", serverName)
				continue
			}
			slog.Info("register mcp tool", "name", toolName)
			// 创建 MCP 工具定义
			mcpToolDef := types.MCPToolDefinition{
				Name:        toolName,
				Description: toolDef.Function.Description,
				InputSchema: toolDef.Function.Parameters,
			}
			// 创建工具代理
			proxy := mcp.NewToolProxy(mcpManager, client.GetServerID(), serverName, mcpToolDef)
			toolRegistry.Register(proxy)
		}
	}

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
	maxTokens           int64
	compressionRatio    float64
	compressionKeepSize int
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
		maxTokens:           maxTokens,
		compressionRatio:    compressionRatio,
		compressionKeepSize: compressionKeepSize,
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

	// 创建或加载会话
	sess, err := r.createSession(sessionKey, message.Channel, sessionType)
	if err != nil {
		return err
	}

	// 运行 AgentLoop，传递 nil 使用默认的 BusOutputHook
	r.AgentLoop(ctx, sess, message, nil)
	return nil
}

// createSession 创建或加载 API 会话
func (r *ApiAgentRunner) createSession(sessionKey string, botID string, sessionType types.SessionType) (session.Session, error) {
	return session.NewApiSession(
		r.sessionRepo,
		sessionKey,
		r.agentDef.ID,
		botID,
		sessionType,
		r.maxTokens,
		r.compressionRatio,
		r.compressionKeepSize,
	)
}

// GetSessionKey 生成会话 Key
func (r *ApiAgentRunner) GetSessionKey(sessionType types.SessionType, channelID, chatID, userID string) string {
	return session.GetApiSessionKey(sessionType, r.agentDef.ID, channelID, chatID, userID)
}
