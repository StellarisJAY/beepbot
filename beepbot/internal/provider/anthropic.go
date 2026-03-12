package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const AnthropicDefaultBaseURL = "https://api.anthropic.com"

type AnthropicProvider struct {
	BaseProvider
	client anthropic.Client
}

func NewAnthropicProvider(apiKey, baseURL, defaultModel string) types.LLMProvider {
	if baseURL == "" {
		baseURL = AnthropicDefaultBaseURL
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != AnthropicDefaultBaseURL {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := anthropic.NewClient(opts...)
	return &AnthropicProvider{
		BaseProvider: BaseProvider{
			apiKey:       apiKey,
			baseURL:      baseURL,
			defaultModel: defaultModel,
		},
		client: client,
	}
}

func (p *AnthropicProvider) Chat(ctx context.Context, messages []types.Message, model string, options types.ChatOptions) (*types.LLMResponse, error) {
	params, err := buildAnthropicParams(messages, model, options)
	if err != nil {
		return nil, fmt.Errorf("build anthropic params error: %w", err)
	}

	result, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic api error: %w", err)
	}

	return convertAnthropicResponse(result), nil
}

func buildAnthropicParams(messages []types.Message, model string, options types.ChatOptions) (anthropic.MessageNewParams, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 4096, // Anthropic 要求必填
	}

	// 转换消息
	anthropicMessages := make([]anthropic.MessageParam, 0)
	var systemPrompt string

	for _, msg := range messages {
		switch msg.Role {
		case types.RoleSystem:
			// Anthropic 的 system prompt 作为独立参数
			systemPrompt += msg.Content + "\n"
		case types.RoleUser:
			anthropicMessages = append(anthropicMessages,
				anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case types.RoleAssistant:
			// 处理助手消息（可能包含工具调用）
			content := make([]anthropic.ContentBlockParamUnion, 0)
			if msg.Content != "" {
				content = append(content, anthropic.NewTextBlock(msg.Content))
			}
			for _, tc := range msg.ToolCalls {
				// 解析 Arguments JSON 为 map
				var argsMap map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &argsMap); err != nil {
					argsMap = make(map[string]interface{})
				}
				// NewToolUseBlock 签名: (id string, input any, name string)
				content = append(content, anthropic.NewToolUseBlock(
					tc.ID,
					argsMap,
					tc.Function.Name,
				))
			}
			anthropicMessages = append(anthropicMessages,
				anthropic.NewAssistantMessage(content...))
		case types.RoleTool:
			// Anthropic 中工具结果放在 user 消息的 ToolResultBlock 中
			content := []anthropic.ContentBlockParamUnion{
				anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, false),
			}
			anthropicMessages = append(anthropicMessages,
				anthropic.NewUserMessage(content...))
		}
	}

	// 设置 system prompt
	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
	}

	params.Messages = anthropicMessages

	// 可选参数
	if options.Temperature != nil {
		params.Temperature = anthropic.Float(float64(*options.Temperature))
	}
	if options.MaxTokens != nil {
		params.MaxTokens = *options.MaxTokens
	}

	// 工具定义
	if len(options.Tools) > 0 {
		tools := make([]anthropic.ToolUnionParam, len(options.Tools))
		for i, tool := range options.Tools {
			tools[i] = anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        tool.Function.Name,
					Description: anthropic.String(tool.Function.Description),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: tool.Function.Parameters,
					},
				},
			}
		}
		params.Tools = tools
	}

	return params, nil
}

func convertAnthropicResponse(result *anthropic.Message) *types.LLMResponse {
	response := &types.LLMResponse{
		FinishReason: convertAnthropicStopReason(result.StopReason),
		Usage: &types.TokenUsage{
			InputTokens:  result.Usage.InputTokens,
			OutputTokens: result.Usage.OutputTokens,
		},
	}

	toolCalls := make([]types.ToolCall, 0)
	var content string

	for _, block := range result.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			content += b.Text
		case anthropic.ToolUseBlock:
			argsJSON, _ := json.Marshal(b.Input)
			toolCalls = append(toolCalls, types.ToolCall{
				ID:   b.ID,
				Type: "function",
				Name: b.Name,
				Function: &types.FunctionCall{
					Name:      b.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	response.Content = content
	response.ToolCalls = toolCalls

	return response
}

// convertAnthropicStopReason 将 Anthropic 的 stop_reason 转换为统一的 FinishReason
// Anthropic stop_reason:
// - end_turn: 正常结束
// - tool_use: 工具调用
// - max_tokens: 达到最大 token
// - stop_sequence: 遇到停止序列
func convertAnthropicStopReason(reason anthropic.StopReason) types.FinishReason {
	switch reason {
	case anthropic.StopReasonEndTurn:
		return types.FinishReasonStop
	case anthropic.StopReasonToolUse:
		return types.FinishReasonToolCall
	case anthropic.StopReasonMaxTokens:
		return types.FinishReasonMaxTokens
	case anthropic.StopReasonStopSequence:
		return types.FinishReasonStopSequence
	default:
		return types.FinishReasonUnknown
	}
}