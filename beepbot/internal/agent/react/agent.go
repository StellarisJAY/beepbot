package react

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/skill"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// 工具调用连续出错次数，达到这个次数将提示智能体提前结束或尝试其他方案
const toolErrorThreshold = 3

const (
	builtinCommandCompact = "/compact"
	builtinCommandContext = "/context"
	builtinCommandClear   = "/clear"
)

type ReActAgentRunner struct {
	model        types.LLMProvider
	bus          *channel.MessageBus
	tools        *tool.ToolRegistry
	skillManager *skill.Manager
	modelID      string
	agentID      string // 智能体ID，用于工具执行时的身份识别

	maxIterations int
	systemPrompt  string
	workingDir    string

	maxTokens           int64
	compressionRatio    float64
	compressionKeepSize int

	onMessageRecv        MessageRecvHook
	onChatCompletion     ChatCompletionHook
	onContextCompression ContextCompressionHook
	onToolUsage          ToolUsageHook
	onLoopFinish         LoopFinishHook
}

type MessageRecvHook func(message channel.InboundMessage)
type ChatCompletionHook func(response types.LLMResponse)
type ContextCompressionHook func()
type ToolUsageHook func(tool types.ToolCall, result string, err error, duration time.Duration)
type LoopFinishHook func(totalIterations int, tokenUsage types.TokenUsage)

// AgentLoop ReAct智能体循环
// 需要传入会话、输入消息和输出hook
// 会话负责历史消息管理
// 输入消息来源有IM机器人、定时任务和调用sub-agent
// 输出hook用来处理不同的消息输出逻辑，IM机器人、定时任务、sub-agent的输出逻辑有区别
func (a *ReActAgentRunner) AgentLoop(ctx context.Context, sess session.Session, message channel.InboundMessage, output OutputHook) {
	// 处理内置命令
	if ok := a.handleBuiltinCommand(ctx, sess, message); ok {
		slog.Debug("handle builtin command done, skip agent loop", "command", message.Content)
		return
	}
	channelName, userID, groupID, inboundMsgID := message.Channel, message.UserID, message.GroupID, message.MessageID
	maxIterations := a.maxIterations

	// 整个智能体循环的token用量统计
	totalTokenUsage := types.TokenUsage{}

	// 如果没有提供 output hook，创建一个默认的 BusOutputHook
	if output == nil {
		output = NewBusOutputHook(a.bus, channelName, userID, groupID, message.ChatID, inboundMsgID)
	}

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

	// 连续错误工具计数
	toolErrorCounter := make(map[string]int)

	// 上下文构建器
	contextBuilder := contextBuilder{
		systemPrompt:    a.systemPrompt,  // 用户配置的系统提示词
		skillManager:    a.skillManager,  // 技能信息
		session:         sess,            // 会话信息
		workingDir:      a.workingDir,    // 工作目录
		userInstruction: message.Content, // 用户请求
	}

	sess.AppendMessage(types.Message{
		Role:    types.RoleUser,
		Content: message.Content,
	})

	// 提前构建固定的上下文内容
	contextBuilder.prebuild()

	iterationCount := 1

	// 智能体循环, 加5步来整理总结
	// TODO 更好的结束策略
	for iterations := range maxIterations + 5 {
		iterationCount++
		// 检查是否需要压缩上下文
		if sess.NeedCompress() {
			if a.onContextCompression != nil {
				a.onContextCompression()
			}
			slog.Info("context compression triggered", "tokenUsed", sess.GetTokenUsage())
			a.compressContext(ctx, sess)
		}

		// 如果达到了最大迭代次数，添加一条提示结束任务和记录任务状态的消息
		if iterations == maxIterations {
			sess.AppendMessage(types.Message{
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
			output.OnError(ctx, err)
			break
		}

		if a.onChatCompletion != nil {
			a.onChatCompletion(*response)
		}

		// 记录token用量
		tokenUsage := response.Usage
		totalTokenUsage.OutputTokens += tokenUsage.OutputTokens

		// 没有工具调用，直接返回消息
		if response.FinishReason != types.FinishReasonToolCall {
			sess.AppendMessage(types.Message{
				Role:    types.RoleAssistant,
				Content: response.Content,
				Usage:   tokenUsage,
			})
			output.OnResponse(ctx, response.Content)
			break
		}

		if response.Content != "" {
			output.OnIntermediateContent(ctx, response.Content)
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
			Usage:     tokenUsage,
		}
		sess.AppendMessage(assistantMsg)

		slog.Info("executing tools", "count", len(functionCalls))

		// 构建会话推送信息，用于工具执行时传递上下文
		sessionInfo := &tool.SessionPushInfo{
			Channel: channelName,
			BotID:   channelName, // Channel 就是 BotID
			UserID:  userID,
			GroupID: groupID,
			ChatID:  message.ChatID,
			AgentID: a.agentID,
		}

		// 执行工具函数
		for _, tc := range functionCalls {
			params := make(map[string]any)
			var err error
			var result string
			if err = json.Unmarshal([]byte(tc.Function.Arguments), &params); err != nil {
				slog.Error("parse function call arguments error", "tool", tc.Name, "args", tc.Function.Arguments)
				err = errors.New("invalid arguments")
			} else {
				result, err = a.tools.ExecuteWithContext(ctx, tc.Function.Name, params, sessionInfo)
			}

			if a.onToolUsage != nil {
				a.onToolUsage(tc, result, err, time.Second)
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
				sess.AppendMessage(toolMessage)
				continue
			}
			// 记录工具调用结果
			toolMessage := types.Message{
				Role:       types.RoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			}
			slog.Info("Tool call success", "channel", channelName, "userID", userID, "tool", tc.Function.Name, "tool_call_id", tc.ID, "args", truncateText(tc.Function.Arguments, 20), "result", truncateText(result, 50))
			// 添加工具调用结果到会话
			sess.AppendMessage(toolMessage)
		}
	}

	slog.Info("agent loop finished", "output_tokens", totalTokenUsage.OutputTokens, "context_used", sess.GetTokenUsage())
	if a.onLoopFinish != nil {
		a.onLoopFinish(iterationCount, totalTokenUsage)
	}

	// 异步更新摘要
	a.generateSummary(ctx, sess.GetHistory(), sess)
}

// handleBuiltinCommand 处理内置命令，返回消息是否是内置命令，不是则需要执行智能体循环
func (a *ReActAgentRunner) handleBuiltinCommand(ctx context.Context, sess session.Session, message channel.InboundMessage) bool {
	outboundMessage := channel.OutboundMessage{
		Channel:          message.Channel,
		UserID:           message.UserID,
		GroupID:          message.GroupID,
		ChatID:           message.ChatID,
		InboundMessageID: message.MessageID,
		MessageType:      channel.TextMessage,
	}

	switch message.Content {
	case builtinCommandCompact: // 压缩上下文
		before, after := a.compressContext(ctx, sess)
		outboundMessage.Content = fmt.Sprintf("已压缩上下文, 压缩前: %d tokens, 压缩后: %d tokens, 压缩率: %.1f%%", before, after, float64(after)/float64(before)*100.0)
		a.bus.PublishOutbound(ctx, outboundMessage)
		return true
	case builtinCommandContext: // 查看上下文用量
		used, max := sess.GetTokenUsage(), sess.GetMaxTokens()
		outboundMessage.Content = fmt.Sprintf("已用上下文: %d tokens, 使用率: %.1f%%", used, float64(used)/float64(max)*100.0)
		a.bus.PublishOutbound(ctx, outboundMessage)
		return true
	case builtinCommandClear: // 清空消息窗口
		sess.ClearHistory()
		outboundMessage.Content = "已清空消息列表"
		a.bus.PublishOutbound(ctx, outboundMessage)
		return true
	default:
		return false
	}
}

// compressContext 压缩会话上下文
// 1. 生成摘要
// 2. 压缩消息窗口
// 返回压缩前的上下文大小和压缩后的上下文大小
func (a *ReActAgentRunner) compressContext(ctx context.Context, sess session.Session) (before int64, after int64) {
	before = sess.GetTokenUsage()
	// 用窗口中的所有消息生成摘要
	history := sess.GetHistory()
	// 生成摘要
	a.generateSummary(ctx, history, sess)
	// 压缩消息列表
	sess.Compress()
	after = sess.GetTokenUsage()
	return
}

func (a *ReActAgentRunner) generateSummary(ctx context.Context, history []types.Message, sess session.Session) {
	genSummaryStartAt := time.Now()
	// 构建摘要请求
	var historyText strings.Builder
	for _, msg := range history {
		switch msg.Role {
		case types.RoleUser:
			fmt.Fprintf(&historyText, "用户: %s\n", msg.Content)
		case types.RoleAssistant:
			fmt.Fprintf(&historyText, "助手: %s\n", msg.Content)
			// 处理工具调用
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					if tc.Function != nil {
						fmt.Fprintf(&historyText, "  调用工具: %s, 参数: %s\n",
							tc.Function.Name,
							truncateText(tc.Function.Arguments, 200))
					}
				}
			}
		case types.RoleTool:
			// 工具结果可能很长，需要截断
			fmt.Fprintf(&historyText, "工具结果: %s\n", truncateText(msg.Content, 200))
		}
	}

	// 调用 LLM 生成摘要
	summaryMessages := []types.Message{
		{Role: types.RoleSystem, Content: compressionPrompt},                                                         // 压缩提示词
		{Role: types.RoleSystem, Content: fmt.Sprintf("<last-summary>\n\n%s</last-summary>\n\n", sess.GetSummary())}, // 之前的摘要
		{Role: types.RoleUser, Content: historyText.String()},                                                        // 历史消息
	}

	response, err := a.model.Chat(ctx, summaryMessages, a.modelID, types.ChatOptions{})
	if err != nil {
		slog.Error("failed to generate summary for compression", "error", err)
		// 压缩失败，但历史已经被清理，继续运行
		return
	}

	// 保存摘要到会话
	sess.SetSummary(response.Content)
	slog.Info("generated session summary", "summaryLength", len(response.Content), "elapsed", time.Since(genSummaryStartAt).String())
}

// truncateText 截断文本，保留指定最大长度
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "...[已截断]"
}
