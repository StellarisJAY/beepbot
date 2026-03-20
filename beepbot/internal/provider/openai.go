package provider

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type OpenAIProvider struct {
	BaseProvider
	client openai.Client
}

func NewOpenAIProvider(apiKey, baseURL, defaultModel string) LLMProvider {
	client := openai.NewClient(option.WithBaseURL(baseURL), option.WithAPIKey(apiKey))
	return &OpenAIProvider{
		BaseProvider: BaseProvider{
			apiKey:       apiKey,
			baseURL:      baseURL,
			defaultModel: defaultModel,
		},
		client: client,
	}
}

func (d *OpenAIProvider) Chat(ctx context.Context, messages []types.Message, model string, options types.ChatOptions) (*types.LLMResponse, error) {
	params := buildOpenAICompletionsParams(messages, model, options)
	result, err := d.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("invoke openai responses api error: %w", err)
	}
	response := new(types.LLMResponse)

	// 记录token用量
	response.Usage = &types.TokenUsage{
		InputTokens:     result.Usage.PromptTokens,
		OutputTokens:    result.Usage.CompletionTokens,
		TotalTokens:     result.Usage.TotalTokens,
		CacheTokens:     result.Usage.PromptTokensDetails.CachedTokens,
		ReasoningTokens: result.Usage.CompletionTokensDetails.ReasoningTokens,
	}

	response.ToolCalls = make([]types.ToolCall, 0)
	for _, tc := range result.Choices[0].Message.ToolCalls {
		response.ToolCalls = append(response.ToolCalls, types.ToolCall{
			Type: "function",
			Name: tc.Function.Name,
			Function: &types.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
			ID: tc.ID,
		})
	}
	response.FinishReason = convertOpenAIFinishReason(result.Choices[0].FinishReason)
	response.Content = result.Choices[0].Message.Content
	return response, nil
}

func (d *OpenAIProvider) Stream(ctx context.Context, messages []types.Message, model string, options types.ChatOptions) (*Stream, error) {
	params := buildOpenAICompletionsParams(messages, model, options)
	// stream模式需要打开用量统计
	params.StreamOptions.IncludeUsage = openai.Opt(true)
	stream := d.client.Chat.Completions.NewStreaming(ctx, params)
	result := NewStream()
	go func() {
		defer stream.Close()
		if stream.Err() != nil {
			result.errChan <- stream.Err()
			return
		}

		// 用于累积工具调用，key 是 index
		toolCallBuilders := make(map[int64]*struct {
			id        string
			name      string
			arguments strings.Builder
		})

		for stream.Next() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			current := stream.Current()
			response := types.LLMResponse{}
			// 记录token用量
			response.Usage = &types.TokenUsage{
				InputTokens:     current.Usage.PromptTokens,
				OutputTokens:    current.Usage.CompletionTokens,
				TotalTokens:     current.Usage.TotalTokens,
				CacheTokens:     current.Usage.PromptTokensDetails.CachedTokens,
				ReasoningTokens: current.Usage.CompletionTokensDetails.ReasoningTokens,
			}
			if len(current.Choices) == 0 {
				result.outChan <- response
				continue
			}
			// 处理工具调用增量
			for _, tc := range current.Choices[0].Delta.ToolCalls {
				index := tc.Index
				builder, exists := toolCallBuilders[index]
				if !exists {
					builder = &struct {
						id        string
						name      string
						arguments strings.Builder
					}{}
					toolCallBuilders[index] = builder
				}

				// 累积 ID（通常只在第一个 chunk 出现）
				if tc.ID != "" {
					builder.id = tc.ID
				}
				// 累积函数名（通常只在第一个 chunk 出现）
				if tc.Function.Name != "" {
					builder.name = tc.Function.Name
				}
				// 累积参数片段
				if tc.Function.Arguments != "" {
					builder.arguments.WriteString(tc.Function.Arguments)
				}
			}

			response.FinishReason = convertOpenAIFinishReason(current.Choices[0].FinishReason)
			response.Content = current.Choices[0].Delta.Content
			result.outChan <- response
		}

		// 流结束，构建完整的工具调用列表
		finalResponse := types.LLMResponse{}
		// 按 index 顺序排序工具调用
		indices := make([]int64, 0, len(toolCallBuilders))
		for idx := range toolCallBuilders {
			indices = append(indices, idx)
		}
		slices.Sort(indices)
		for _, idx := range indices {
			builder := toolCallBuilders[idx]
			finalResponse.ToolCalls = append(finalResponse.ToolCalls, types.ToolCall{
				ID:   builder.id,
				Type: "function",
				Function: &types.FunctionCall{
					Name:      builder.name,
					Arguments: builder.arguments.String(),
				},
			})
		}
		// 发送包含完整工具调用的最终响应
		if len(finalResponse.ToolCalls) > 0 {
			result.outChan <- finalResponse
		}

		// 结束，发送一个EOF信号
		result.errChan <- io.EOF
	}()

	return result, nil
}

// convertOpenAIFinishReason 将 OpenAI 的 finish_reason 转换为统一的 FinishReason
// OpenAI finish_reason:
// - stop: 正常结束
// - tool_calls: 工具调用
// - length: 达到最大 token
// - content_filter: 内容被过滤
func convertOpenAIFinishReason(reason string) types.FinishReason {
	switch reason {
	case "stop":
		return types.FinishReasonStop
	case "tool_calls":
		return types.FinishReasonToolCall
	case "length":
		return types.FinishReasonMaxTokens
	case "content_filter":
		return types.FinishReasonContentFilter
	default:
		return types.FinishReasonUnknown
	}
}

func buildOpenAICompletionsParams(messages []types.Message, model string, options types.ChatOptions) openai.ChatCompletionNewParams {
	// 1. 转换消息
	chatMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case types.RoleSystem:
			// 系统消息
			chatMessages = append(chatMessages, openai.SystemMessage(msg.Content))
		case types.RoleUser:
			// 用户消息
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))

		case types.RoleAssistant:

			assistantMessage := openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{},
			}
			// 助手消息 - 需要处理可能的 ToolCalls
			if len(msg.ToolCalls) > 0 {
				// 有工具调用的助手消息
				toolCalls := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(msg.ToolCalls))

				for _, tc := range msg.ToolCalls {
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID: tc.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						},
					})
				}
				assistantMessage.OfAssistant.ToolCalls = toolCalls
			} else {
				assistantMessage.OfAssistant.ToolCalls = nil
			}
			assistantMessage.OfAssistant.Content.OfString = openai.Opt(msg.Content)
			chatMessages = append(chatMessages, assistantMessage)
		case types.RoleTool:
			// 工具响应消息
			chatMessages = append(chatMessages, openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					ToolCallID: msg.ToolCallID,
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.Opt(msg.Content),
					},
				},
			})
		}
	}

	// 2. 构建参数
	params := openai.ChatCompletionNewParams{
		Model:    model,
		Messages: chatMessages,
	}
	params.SetExtraFields(map[string]any{})

	// 3. 可选参数
	// 温度
	if options.Temperature != nil {
		params.Temperature = openai.Opt(float64(*options.Temperature))
	}
	// 最大token数
	if options.MaxTokens != nil {
		params.MaxTokens = openai.Opt(*options.MaxTokens)
	}
	// thinking
	if options.Reasoning != nil && *options.Reasoning == true {
		params.ReasoningEffort = "medium"
		params.ExtraFields()["enable_thinking"] = true
	} else {
		params.ReasoningEffort = "none"
		params.ExtraFields()["enable_thinking"] = false
	}

	// 系统内置工具
	if len(options.Tools) > 0 {
		params.Tools = make([]openai.ChatCompletionToolUnionParam, 0, len(options.Tools))
		for _, tool := range options.Tools {
			params.Tools = append(params.Tools, openai.ChatCompletionToolUnionParam{
				OfFunction: &openai.ChatCompletionFunctionToolParam{
					Function: shared.FunctionDefinitionParam{
						Name:        tool.Function.Name,
						Description: openai.Opt(tool.Function.Description),
						Parameters:  shared.FunctionParameters(tool.Function.Parameters),
						Strict:      openai.Opt(true),
					},
				},
			})
		}
	}

	return params
}
