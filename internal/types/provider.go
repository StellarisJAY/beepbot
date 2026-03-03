package types

import "context"

const (
	RoleSystem    string = "system"
	RoleAssistant string = "assistant"
	RoleTool      string = "tool"
	RoleUser      string = "user"
)

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type LLMResponse struct {
	Content      string      `json:"content"`
	ToolCalls    []ToolCall  `json:"tool_calls,omitempty"`
	FinishReason string      `json:"finish_reason"`
	Usage        *TokenUsage `json:"usage,omitempty"`
}

type TokenUsage struct {
	InputTokens     int64 `json:"input_tokens"`
	OutputTokens    int64 `json:"output_tokens"`
	CacheTokens     int64 `json:"cache_tokens"`
	ReasoningTokens int64 `json:"reasoning_tokens"`
	TotalTokens     int64 `json:"total_tokens"`
}

type ToolCall struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Function  *FunctionCall  `json:"function,omitempty"`
	Name      string         `json:"name,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type Message struct {
	Role         string     `json:"role"`
	Content      string     `json:"content"`
	ToolCallID   string     `json:"tool_call_id,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"`

	Usage *TokenUsage `json:"token_usage,omitempty"`
}

type ToolDefinition struct {
	Type     string                 `json:"type"`
	Function ToolFunctionDefinition `json:"function"`
}

type ToolFunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ChatOptions struct {
	Temperature *float32
	Reasoning   *bool
	Tools       []ToolDefinition
	MaxTokens   *int64
}

type LLMProvider interface {
	Chat(ctx context.Context, messages []Message, model string, options ChatOptions) (*LLMResponse, error)
}
