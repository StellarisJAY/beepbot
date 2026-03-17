package react

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/mcp"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// SubAgentExecutor 子智能体执行器实现
type SubAgentExecutor struct {
	config       config.APIConfig
	skillRepo    repository.SkillRepository    // 技能repo
	agentRepo    repository.AgentRepository    // 智能体表repo
	providerRepo repository.ProviderRepository // 模型供应商repo
	encryptor    *crypto.Encryptor             // 解密工具，用于解密供应商key
	cronDeps     *tool.CronToolDeps            // cron工具依赖
	bus          *channel.MessageBus           // 消息总线
	mcpManager   *mcp.Manager                  // MCP管理器
}

// Execute 执行子智能体
// parentWorkingDir: 父智能体的工作目录，子智能体将继承此目录
func (e *SubAgentExecutor) Execute(ctx context.Context, agentDef *types.Agent, providerDef *types.Provider, message string, parentWorkingDir string) (string, error) {
	slog.Info("SubAgentExecutor executing", "agent", agentDef.Name, "message_len", len(message), "parentWorkingDir", parentWorkingDir)

	// 创建子智能体运行器（不注册 Sub-Agent 工具，防止递归）
	// 使用父智能体的工作目录
	runner, err := NewApiAgentRunner(NewReactAgentParam{
		AgentDef:         agentDef,
		ProviderDef:      providerDef,
		Bus:              e.bus,
		Config:           e.config,
		SkillRepo:        e.skillRepo,
		CronDeps:         e.cronDeps,
		AgentRepo:        e.agentRepo,
		ProviderRepo:     e.providerRepo,
		Encryptor:        e.encryptor,
		McpManager:       e.mcpManager,
		AllowSubAgents:   false,            // 不允许递归调用子智能体
		ParentWorkingDir: parentWorkingDir, // 继承父智能体的工作空间
	})
	if err != nil {
		return "", fmt.Errorf("failed to create sub-agent runner: %w", err)
	}

	// 创建内存会话
	sess := session.NewInMemorySession()

	// 创建收集器钩子
	collector := NewCollectorHook()

	// 构建输入消息
	inboundMsg := channel.InboundMessage{
		Content: message,
	}

	// 运行智能体循环
	runner.AgentLoop(ctx, sess, inboundMsg, collector)

	// 检查是否有错误
	if err := collector.GetError(); err != nil {
		return "", fmt.Errorf("sub-agent execution failed: %w", err)
	}

	// 获取结果
	result := collector.GetResult()
	if result == "" {
		result = "子智能体执行完成，但没有返回结果。"
	}

	slog.Info("SubAgentExecutor completed", "agent", agentDef.Name, "result_len", len(result))
	return result, nil
}
