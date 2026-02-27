package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
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
	params := buildOpenAIResponsesParams(messages, model, options)
	result, err := d.client.Responses.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("invoke openai responses api error: %w", err)
	}
	if result.Error.Code != "" {
		return nil, fmt.Errorf("openai api error, %s:%s", result.Error.Code, result.Error.Message)
	}
	response := new(types.LLMResponse)
	// 记录token用量
	response.Usage = &types.TokenUsage{
		InputTokens:     result.Usage.InputTokens,
		OutputTokens:    result.Usage.OutputTokens,
		CacheTokens:     result.Usage.InputTokensDetails.CachedTokens,
		ReasoningTokens: result.Usage.OutputTokensDetails.ReasoningTokens,
	}

	response.ToolCalls = make([]types.ToolCall, 0)
	outputText := strings.Builder{}
	for _, output := range result.Output {
		if output.Type == "function_call" {
			fc := output.AsFunctionCall()
			response.ToolCalls = append(response.ToolCalls, types.ToolCall{
				Type: "function",
				Name: fc.Name,
				Function: &types.FunctionCall{
					Name:      fc.Name,
					Arguments: fc.Arguments,
				},
				ID: fc.CallID,
			})
		}
		if output.Type == "message" {
			for _, content := range output.Content {
				if content.Type == "output_text" {
					outputText.WriteString(content.Text)
				}
			}
		}
	}
	response.Content = outputText.String()

	return response, nil
}

func buildOpenAIResponsesParams(messages []types.Message, model string, options types.ChatOptions) responses.ResponseNewParams {
	var inputItems responses.ResponseInputParam
	var instructions string

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			instructions = msg.Content
		case "user":
			if msg.ToolCallID != "" {
				inputItems = append(inputItems, responses.ResponseInputItemUnionParam{
					OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
						CallID: msg.ToolCallID,
						Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{OfString: openai.Opt(msg.Content)},
					},
				})
			} else {
				inputItems = append(inputItems, responses.ResponseInputItemUnionParam{
					OfMessage: &responses.EasyInputMessageParam{
						Role:    responses.EasyInputMessageRoleUser,
						Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt(msg.Content)},
					},
				})
			}
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				if msg.Content != "" {
					inputItems = append(inputItems, responses.ResponseInputItemUnionParam{
						OfMessage: &responses.EasyInputMessageParam{
							Role:    responses.EasyInputMessageRoleAssistant,
							Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt(msg.Content)},
						},
					})
				}
				for _, toolCall := range msg.ToolCalls {
					inputItems = append(inputItems, responses.ResponseInputItemUnionParam{
						OfFunctionCall: &responses.ResponseFunctionToolCallParam{
							CallID:    toolCall.ID,
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						},
					})
				}
			} else {
				inputItems = append(inputItems, responses.ResponseInputItemUnionParam{
					OfMessage: &responses.EasyInputMessageParam{
						Role:    responses.EasyInputMessageRoleAssistant,
						Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt(msg.Content)},
					},
				})
			}
		case "tool":
			inputItems = append(inputItems, responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: msg.ToolCallID,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{OfString: openai.Opt(msg.Content)},
					Status: "completed",
				},
			})
		}
	}

	params := responses.ResponseNewParams{
		Model: model,
		Input: responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems},
	}

	if instructions != "" {
		params.Instructions = openai.Opt(instructions)
	}

	if options.MaxTokens != nil {
		params.MaxOutputTokens = openai.Opt(*options.MaxTokens)
	}

	if len(options.Tools) > 0 {
		params.Tools = make([]responses.ToolUnionParam, 0, len(options.Tools))
		for _, tool := range options.Tools {
			if tool.Type == "function" {
				params.Tools = append(params.Tools, responses.ToolUnionParam{
					OfFunction: &responses.FunctionToolParam{
						Name:        tool.Function.Name,
						Parameters:  tool.Function.Parameters,
						Description: openai.Opt(tool.Function.Description),
					},
				})
			}
			if tool.Type == "mcp" {
				// TODO MCP
			}
		}
	}
	// 使用内置联网搜索功能
	params.Tools = append(params.Tools, responses.ToolParamOfWebSearch("web_search"))
	return params
}
