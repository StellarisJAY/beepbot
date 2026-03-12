package provider

import (
	"github.com/StellarisJAY/beepbot/internal/types"
)

// DeepSeek 默认 Base URL
const DeepSeekDefaultBaseURL = "https://api.deepseek.com/v1"

// DeepSeekProvider DeepSeek 提供商实现
// DeepSeek 支持 OpenAI 兼容 API，因此直接复用 OpenAI 实现
type DeepSeekProvider struct {
	*OpenAIProvider
}

// NewDeepSeekProvider 创建 DeepSeek 提供商
func NewDeepSeekProvider(apiKey, baseURL, defaultModel string) types.LLMProvider {
	if baseURL == "" {
		baseURL = DeepSeekDefaultBaseURL
	}
	return &DeepSeekProvider{
		OpenAIProvider: NewOpenAIProvider(apiKey, baseURL, defaultModel).(*OpenAIProvider),
	}
}