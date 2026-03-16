package react

import (
	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/mcp"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// toolRegistrationParams 工具注册参数
type toolRegistrationParams struct {
	agentDef       *types.Agent
	workingDir     string
	config         config.APIConfig
	cronDeps       *tool.CronToolDeps
	skillRepo      repository.SkillRepository
	agentRepo      repository.AgentRepository
	providerRepo   repository.ProviderRepository
	encryptor      *crypto.Encryptor
	bus            *channel.MessageBus
	allowSubAgents bool
	mcpManager     *mcp.Manager
}

// registerTools 注册工具到工具注册表
func registerTools(toolRegistry *tool.ToolRegistry, params toolRegistrationParams) {
	var toolNames []string
	// 根据配置决定注册哪些工具
	if params.agentDef.UseAllTools {
		toolNames = []string{"list_dir", "read_file", "write_file", "edit_file", "shell", "cron"}
	} else {
		// 只注册配置的工具
		toolNames, _ = params.agentRepo.GetAgentTools(params.agentDef.ID)
	}

	// 系统内置工具
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

	// 注册 Sub-Agent 工具（仅当 allowSubAgents 为 true 时）
	if params.allowSubAgents && params.agentRepo != nil && params.providerRepo != nil && params.encryptor != nil {
		// 创建子智能体执行器
		executor := NewSubAgentExecutor(
			params.config,
			params.skillRepo,
			params.agentRepo,
			params.providerRepo,
			params.encryptor,
			params.cronDeps,
			params.bus,
			params.mcpManager,
		)

		// 查询所有能够被调用的智能体，每个智能体创建单独的sub-agent工具
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
