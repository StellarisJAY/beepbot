package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/provider"
	"github.com/StellarisJAY/beepbot/internal/session"
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
}

// TODO builtin prompt
const builtinSystemPrompt = `
	You are a helpful assistant.
`

func NewAgentRun(config config.Config, bus *channel.MessageBus) (*AgentRunner, error) {
	// 创建聊天模型接口
	llmProvider, err := provider.CreateLLMProvider(config)
	if err != nil {
		return nil, err
	}
	// 注册工具
	toolRegistry := tool.NewToolRegistry()
	// 文件系统工具需要传入工作目录，实现访问隔离
	workingDir := config.AgentConfig.WorkingDir
	toolRegistry.Register(tool.NewListDirTool(workingDir))
	toolRegistry.Register(tool.NewReadFileTool(workingDir))
	toolRegistry.Register(tool.NewWriteFileTool(workingDir))
	// 操作系统信息工具
	toolRegistry.Register(tool.NewReadSystemInfoTool())
	// shell 工具
	if config.BuiltinTools.Shell.Enable {
		toolRegistry.Register(tool.NewShellTool(config))
	}

	// 会话管理器，管理不同会话的消息缓存和长期记忆
	sessionManager := session.NewSessionManager(config)

	agentRun := &AgentRunner{
		model:          llmProvider,
		sessionManager: sessionManager,
		config:         config,
		bus:            bus,
		tools:          toolRegistry,
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
	model := a.config.AgentConfig.Model
	channelName, userID, groupID, inboundMsgID := message.Channel, message.UserID, message.GroupID, message.MessageID
	maxIterations := a.config.AgentConfig.MaxIterations

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

	// 智能体循环
	for iterations := range maxIterations {
		messages := []types.Message{
			{
				Role:    types.RoleSystem,
				Content: builtinSystemPrompt,
			},
			{
				Role:    types.RoleSystem,
				Content: a.config.AgentConfig.SystemPrompt,
			},
		}
		// TODO 上下文构建
		messages = append(messages, session.GetHistory()...)

		// 调用大模型api
		response, err := a.model.Chat(ctx, messages, model, options)
		if err != nil {
			slog.Error("model api error", "iteration", iterations, "error", err)
			continue
		}
		slog.Debug("response from model", "content", response.Content, "usage", response.Usage, "tool_calls", response.ToolCalls)

		// 没有工具调用，直接返回消息
		if len(response.ToolCalls) == 0 {
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
			return
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

		// 执行工具函数
		for _, tc := range functionCalls {
			params := make(map[string]any)
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &params); err != nil {
				slog.Error("parse function call arguments error", "tool", tc.Name, "args", tc.Function.Arguments)
				continue
			}
			result, err := a.tools.ExecuteWithContext(ctx, tc.Function.Name, params, channelName, userID)
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
			slog.Debug("Tool call success", "channel", channelName, "userID", userID, "tool", tc.Function.Name, "args", tc.Function.Arguments, "result", result)
			// 添加工具调用结果到会话
			session.AppendMessage(toolMessage)
		}
	}
}
