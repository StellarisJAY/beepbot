package types

import (
	"time"

	"gorm.io/datatypes"
)

// ProviderType 供应商类型
type ProviderType string

const (
	ProviderTypeOpenAI    ProviderType = "openai"
	ProviderTypeDashScope ProviderType = "dashscope"
	ProviderOllama        ProviderType = "ollama"
	ProviderTypeAnthropic ProviderType = "anthropic"
	ProviderTypeDeepSeek  ProviderType = "deepseek"
)

// Provider 模型供应商配置
type Provider struct {
	ID           string         `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name         string         `json:"name" gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	ProviderType ProviderType   `json:"provider_type" gorm:"column:provider_type;type:varchar(32);not null;index"`
	APIKey       string         `json:"-" gorm:"column:api_key;type:varchar(512);not null"` // 加密存储，不输出到 JSON
	APIKeyMasked string         `json:"api_key_masked" gorm:"-"`                            // 脱敏显示，仅用于 API 响应
	BaseURL      string         `json:"base_url" gorm:"column:base_url;type:varchar(512)"`
	ExtraConfig  datatypes.JSON `json:"extra_config" gorm:"column:extra_config;type:jsonb"`
	IsDefault    bool           `json:"is_default" gorm:"column:is_default;type:boolean;default:false"`
	CreatedAt    time.Time      `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (Provider) TableName() string {
	return "providers"
}

// ProviderQuery 供应商查询参数
type ProviderQuery struct {
	Name         string       // 名称模糊搜索
	ProviderType ProviderType // 类型筛选
}

const (
	RoleSystem    string = "system"
	RoleAssistant string = "assistant"
	RoleTool      string = "tool"
	RoleUser      string = "user"
)

// FinishReason 统一的结束原因类型
type FinishReason string

const (
	// FinishReasonStop 正常结束，对话完成
	FinishReasonStop FinishReason = "stop"
	// FinishReasonToolCall 工具调用
	FinishReasonToolCall FinishReason = "tool_call"
	// FinishReasonMaxTokens 达到最大 token 限制
	FinishReasonMaxTokens FinishReason = "max_tokens"
	// FinishReasonContentFilter 内容被过滤（安全审核）
	FinishReasonContentFilter FinishReason = "content_filter"
	// FinishReasonStopSequence 遇到停止序列
	FinishReasonStopSequence FinishReason = "stop_sequence"
	// FinishReasonError 发生错误
	FinishReasonError FinishReason = "error"
	// FinishReasonUnknown 未知原因
	FinishReasonUnknown FinishReason = "unknown"
)

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type LLMResponse struct {
	Content      string       `json:"content"`
	ToolCalls    []ToolCall   `json:"tool_calls,omitempty"`
	FinishReason FinishReason `json:"finish_reason"`
	Usage        *TokenUsage  `json:"usage,omitempty"`
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
