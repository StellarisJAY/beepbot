package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/provider"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/skill"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// 工具调用连续出错次数，达到这个次数将提示智能体提前结束或尝试其他方案
const toolErrorThreshold = 3

type AgentRunner struct {
	model          types.LLMProvider
	bus            *channel.MessageBus
	sessionManager *session.SessionManager
	config         config.Config
	tools          *tool.ToolRegistry
	skillManager   *skill.Manager
	modelID        string
}

func NewAgentRun(config config.Config, bus *channel.MessageBus) (*AgentRunner, error) {
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
		config:         config,
		bus:            bus,
		tools:          toolRegistry,
		skillManager:   skillManager,
		modelID:        modelID,
	}
	return agentRun, nil
}

// MessageLoop 核心消息消费循环
func (a *AgentRunner) MessageLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 接收消息总线消息
			message, ok := a.bus.ConsumeInbound(ctx)
			if !ok {
				continue
			}
			var sessionKey string
			if message.GroupID != "" {
				slog.Info("receive group message", "channel", message.Channel, "group", message.GroupID, "userID", message.UserID, "content", message.Content)
			} else {
				slog.Info("receive private message", "channel", message.Channel, "user", message.UserID, "content", message.Content)
			}
			sessionKey = session.GetSessionKey(message.Channel, message.GroupID, message.UserID)
			session := a.sessionManager.GetOrCreateSession(sessionKey)
			// 创建智能体循环gorountine处理消息
			go a.agentLoop(ctx, session, message)
		}
	}
}

func (a *AgentRunner) agentLoop(ctx context.Context, session *session.Session, message channel.InboundMessage) {
	channelName, userID, groupID, inboundMsgID := message.Channel, message.UserID, message.GroupID, message.MessageID
	maxIterations := a.config.AgentConfig.MaxIterations

	totalTokenUsage := types.TokenUsage{}

	// 限制一次agent循环的默认最大迭代次数为50
	if maxIterations == 0 {
		maxIterations = 50
	}

	options := types.ChatOptions{}

	// 将工具转换成模型需要的格式
	tools := a.tools.GetDefinitions()
	options.Tools = make([]types.ToolDefinition, 0, len(tools))

	for _, tool := range tools {
		options.Tools = append(options.Tools, types.ToolDefinition{
			Type: tool["type"].(string),
			Function: types.ToolFunctionDefinition{
				Name:        tool["function"].(map[string]any)["name"].(string),
				Parameters:  tool["function"].(map[string]any)["parameters"].(map[string]any),
				Description: tool["function"].(map[string]any)["description"].(string),
			},
		})
	}

	// 会话中增加用户新发送的消息
	session.AppendMessage(types.Message{
		Role:    types.RoleUser,
		Content: message.Content,
	})

	// 连续错误工具计数
	toolErrorCounter := make(map[string]int)

	// 上下文构建器
	contextBuilder := contextBuilder{
		systemPrompt: a.config.AgentConfig.SystemPrompt,
		skillManager: a.skillManager,
		session:      session,
	}
	contextBuilder.prebuild()

	// 智能体循环, 加5步来整理总结
	// TODO 更好的结束策略
	for iterations := range maxIterations + 5 {
		// 如果达到了最大迭代次数，添加一条提示结束任务和记录任务状态的消息
		if iterations == maxIterations {
			session.AppendMessage(types.Message{
				Role:    types.RoleSystem,
				Content: completionMessage,
			})
		}

		// 上下文
		messages := contextBuilder.buildContext()

		// 调用大模型api
		response, err := a.model.Chat(ctx, messages, a.modelID, options)
		if err != nil {
			slog.Error("model api error", "iteration", iterations, "error", err)
			a.bus.PublishOutbound(ctx, channel.OutboundMessage{
				Channel:          channelName,
				UserID:           userID,
				GroupID:          groupID,
				Content:          "调用模型失败, 请重试...",
				MessageType:      channel.TextMessage,
				InboundMessageID: inboundMsgID,
			})
			break
		}

		// 记录token用量
		totalTokenUsage.CacheTokens += response.Usage.CacheTokens
		totalTokenUsage.InputTokens += response.Usage.InputTokens
		totalTokenUsage.OutputTokens += response.Usage.OutputTokens
		totalTokenUsage.ReasoningTokens += response.Usage.ReasoningTokens
		totalTokenUsage.TotalTokens += totalTokenUsage.InputTokens + totalTokenUsage.OutputTokens + totalTokenUsage.ReasoningTokens

		// 没有工具调用，直接返回消息
		if response.FinishReason != "tool_calls" {
			session.AppendMessage(types.Message{
				Role:    types.RoleAssistant,
				Content: response.Content,
			})
			a.bus.PublishOutbound(ctx, channel.OutboundMessage{
				Channel:          channelName,
				UserID:           userID,
				GroupID:          groupID,
				Content:          response.Content,
				MessageType:      channel.MarkdownMsg,
				InboundMessageID: inboundMsgID,
			})
			break
		}

		// 获得不同类型的工具调用
		functionCalls := make([]types.ToolCall, 0, len(response.ToolCalls))
		for _, toolCall := range response.ToolCalls {
			if toolCall.Type == "function" {
				functionCalls = append(functionCalls, toolCall)
			}
			// TODO MCP
		}

		// 如果有连续错误计数的工具不在调用列表，则清空计数
		for key := range toolErrorCounter {
			if !slices.ContainsFunc(functionCalls, func(item types.ToolCall) bool {
				return item.Function.Name == key
			}) {
				toolErrorCounter[key] = 0
			}
		}

		// 将工具调用消息添加到会话
		assistantMsg := types.Message{
			Role:      types.RoleAssistant,
			Content:   response.Content,
			ToolCalls: functionCalls,
		}
		session.AppendMessage(assistantMsg)

		slog.Info("executing tools", "count", len(functionCalls))
		// 执行工具函数
		for _, tc := range functionCalls {
			params := make(map[string]any)
			var err error
			var result string
			if err = json.Unmarshal([]byte(tc.Function.Arguments), &params); err != nil {
				slog.Error("parse function call arguments error", "tool", tc.Name, "args", tc.Function.Arguments)
				err = errors.New("invalid arguments")
			} else {
				result, err = a.tools.ExecuteWithContext(ctx, tc.Function.Name, params, channelName, userID)
			}

			if err != nil {
				// 返回错误结果
				toolMessage := types.Message{
					Role:       types.RoleTool,
					Content:    err.Error(),
					ToolCallID: tc.ID,
				}
				// 检查是否超出连续工具调用错误限制
				count, ok := toolErrorCounter[tc.Function.Name]
				if !ok {
					count = 1
					toolErrorCounter[tc.Function.Name] = count
				}
				if count > toolErrorThreshold {
					toolMessage.Content = fmt.Sprintf("工具调用已连续错误:%d次, 请停止继续调用该工具, 请尝试其他方案或直接回复用户。", count)
				}
				slog.Debug("Tool call error", "channel", channelName, "userID", userID, "tool", tc.Function.Name, "args", tc.Function.Arguments, "error", err)
				session.AppendMessage(toolMessage)
				continue
			}
			// 记录工具调用结果
			toolMessage := types.Message{
				Role:       types.RoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			}
			slog.Info("Tool call success", "channel", channelName, "userID", userID, "tool", tc.Function.Name, "tool_call_id", tc.ID, "args", tc.Function.Arguments, "result", result)
			// 添加工具调用结果到会话
			session.AppendMessage(toolMessage)
		}
	}

	slog.Info("agent loop finished", "token_usage", totalTokenUsage)
}
