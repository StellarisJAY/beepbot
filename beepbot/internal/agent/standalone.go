package agent

import (
	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/provider"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/skill"
	"github.com/StellarisJAY/beepbot/internal/tool"
)

func NewStandaloneRun(config config.StandaloneConfig, bus *channel.MessageBus) (*AgentRunner, error) {
	// 创建聊天模型接口
	llmProvider, err := provider.CreateLLMProvider(config)
	if err != nil {
		return nil, err
	}
	modelID := config.AgentConfig.Model
	// 注册工具
	toolRegistry := tool.NewToolRegistry()
	// 文件系统工具需要传入工作目录，实现访问隔离
	workingDir := config.AgentConfig.WorkingDir
	toolRegistry.Register(tool.NewListDirTool(workingDir, config.DataDir))
	toolRegistry.Register(tool.NewReadFileTool(workingDir, config.DataDir))
	toolRegistry.Register(tool.NewWriteFileTool(workingDir, config.DataDir))
	toolRegistry.Register(tool.NewEditFileTool(workingDir, config.DataDir))
	// TODO 任务管理工具
	toolRegistry.Register(tool.NewWriteTodoTool(workingDir))
	// 操作系统信息工具
	toolRegistry.Register(tool.NewReadSystemInfoTool())
	// shell 工具
	if config.BuiltinTools.Shell.Enable {
		toolRegistry.Register(tool.NewShellTool(config))
	}

	// 会话管理器，管理不同会话的消息缓存和长期记忆
	sessionManager := session.NewSessionManager(config)

	// 创建技能管理器
	skillManager := skill.NewManager(config.DataDir, workingDir)

	agentRun := &AgentRunner{
		model:          llmProvider,
		sessionManager: sessionManager,
		bus:            bus,
		tools:          toolRegistry,
		skillManager:   skillManager,
		modelID:        modelID,
		systemPrompt:   config.AgentConfig.SystemPrompt,
		maxIterations:  config.AgentConfig.MaxIterations,
		workingDir:     config.AgentConfig.WorkingDir,
	}
	return agentRun, nil
}
