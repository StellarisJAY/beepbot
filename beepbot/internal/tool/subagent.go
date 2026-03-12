package tool

import (
	"context"
	"fmt"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// SubAgentDeps 子智能体工具依赖
type SubAgentDeps struct {
	Config       config.APIConfig
	SessionRepo  repository.SessionRepository
	SkillRepo    repository.SkillRepository
	AgentRepo    repository.AgentRepository
	ProviderRepo repository.ProviderRepository
	Encryptor    *crypto.Encryptor
	CronDeps     *CronToolDeps
	Bus          *channel.MessageBus
}

// SubAgentExecutor 子智能体执行器接口
// 用于解耦 tool 包和 agent 包，避免循环依赖
type SubAgentExecutor interface {
	Execute(ctx context.Context, agentDef *types.Agent, providerDef *types.Provider, message string, parentWorkingDir string) (string, error)
}

// SubAgentTool 子智能体调用工具
type SubAgentTool struct {
	agentDef         *types.Agent
	providerDef      *types.Provider
	description      string
	executor         SubAgentExecutor
	parentWorkingDir string // 父智能体的工作目录，子智能体将继承此目录
}

// NewSubAgentTool 创建子智能体调用工具
func NewSubAgentTool(
	agentDef *types.Agent,
	providerDef *types.Provider,
	description string,
	deps *SubAgentDeps,
) *SubAgentTool {
	return &SubAgentTool{
		agentDef:    agentDef,
		providerDef: providerDef,
		description: description,
		// deps 暂时保留用于向后兼容，实际执行由 executor 完成
	}
}

// NewSubAgentToolWithExecutor 创建子智能体调用工具（使用执行器）
func NewSubAgentToolWithExecutor(
	agentDef *types.Agent,
	providerDef *types.Provider,
	description string,
	executor SubAgentExecutor,
	parentWorkingDir string,
) *SubAgentTool {
	return &SubAgentTool{
		agentDef:         agentDef,
		providerDef:      providerDef,
		description:      description,
		executor:         executor,
		parentWorkingDir: parentWorkingDir,
	}
}

func (t *SubAgentTool) Name() string {
	return fmt.Sprintf("subagent_%s", t.agentDef.Name)
}

func (t *SubAgentTool) Description() string {
	desc := t.description
	if desc == "" {
		desc = t.agentDef.Description
	}
	return fmt.Sprintf(`调用子智能体 "%s" 执行特定任务。

描述: %s

该子智能体具有独立的系统提示词、模型配置和受限的工具访问权限。
传递清晰的任务描述给它，它会返回执行结果。`, t.agentDef.Name, desc)
}

func (t *SubAgentTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task": map[string]any{
				"type":        "string",
				"description": "要交给子智能体执行的任务描述",
			},
			"context": map[string]any{
				"type":        "string",
				"description": "可选的上下文信息，帮助子智能体更好地理解任务",
			},
		},
		"required": []string{"task"},
	}
}

func (t *SubAgentTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	task, ok := params["task"].(string)
	if !ok || task == "" {
		return "", fmt.Errorf("task parameter is required")
	}

	contextInfo, _ := params["context"].(string)

	// 构建子智能体的输入消息
	message := task
	if contextInfo != "" {
		message = fmt.Sprintf("上下文:\n%s\n\n任务:\n%s", contextInfo, task)
	}

	// 使用执行器运行子智能体
	if t.executor != nil {
		return t.executor.Execute(ctx, t.agentDef, t.providerDef, message, t.parentWorkingDir)
	}

	return "", fmt.Errorf("sub-agent executor not configured")
}