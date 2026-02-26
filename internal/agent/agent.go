package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/provider"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
)

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
	toolRegistry.Register(&tool.TimeTool{})
	// 文件系统工具需要传入工作目录，实现访问隔离
	workingDir := config.AgentConfig.WorkingDir
	toolRegistry.Register(tool.NewListDirTool(workingDir))
	toolRegistry.Register(tool.NewReadFileTool(workingDir))
	toolRegistry.Register(tool.NewWriteFileTool(workingDir))

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
			slog.Info("receive message", "channel", message.Channel, "user", message.UserID, "content", message.Content)
			// 找到消息所属会话，没有则新建会话
			sessionKey := session.GetSessionKey(message.Channel, message.UserID)
			session := a.sessionManager.GetOrCreateSession(sessionKey)
			// 创建智能体循环gorountine处理消息
			go a.agentLoop(ctx, session, message)
		}
	}
}

func (a *AgentRunner) agentLoop(ctx context.Context, session *session.Session, message channel.InboundMessage) {
	model := a.config.AgentConfig.Model
	channelName, userID := message.Channel, message.UserID
	maxIterations := a.config.AgentConfig.MaxToolIterations

	if maxIterations == 0 {
		maxIterations = 50
	}

	// 先发一条消息让用户知道机器人的状态
	a.bus.PublishOutbound(ctx, channel.OutboundMessage{
		Channel:     channelName,
		UserID:      userID,
		Content:     "机器人正在处理，请稍候...",
		MessageType: channel.TextMessage,
	})

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
				Channel:     channelName,
				UserID:      userID,
				Content:     response.Content,
				MessageType: channel.MarkdownMsg,
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
			// 发布一条工具调用成功消息给用户，让用户知道智能体的执行过程
			a.publishToolCallMessage(ctx, message.Channel, message.UserID, tc, result)
			// 添加工具调用结果到会话
			session.AppendMessage(toolMessage)
		}
	}
}

func (a *AgentRunner) publishToolCallMessage(ctx context.Context, channelName, userID string, tc types.ToolCall, result string) {
	// 压缩结果
	truncatedResult := result[:min(len(result), 20)] + "..."
	msg := channel.OutboundMessage{
		Channel:     channelName,
		UserID:      userID,
		Content:     fmt.Sprintf("工具=\"%s\"调用成功，结果:\"%s\"", tc.Function.Name, truncatedResult),
		MessageType: channel.TextMessage,
	}
	a.bus.PublishOutbound(ctx, msg)
}
