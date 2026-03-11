package provider

import (
	"context"
	"fmt"

	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type OpenAIProvider struct {
	BaseProvider
	client openai.Client
}

func NewOpenAIProvider(apiKey, baseURL, defaultModel string) types.LLMProvider {
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
	response.FinishReason = result.Choices[0].FinishReason
	response.Content = result.Choices[0].Message.Content
	return response, nil
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
