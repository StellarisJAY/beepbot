package provider

// Ollama 默认 Base URL（本地 Ollama 服务）
const OllamaDefaultBaseURL = "http://localhost:11434/v1"

// OllamaProvider Ollama 提供商实现
// Ollama 支持 OpenAI 兼容 API，因此直接复用 OpenAI 实现
type OllamaProvider struct {
	*OpenAIProvider
}

// NewOllamaProvider 创建 Ollama 提供商
// Ollama 不需要 API Key，可以传入空字符串
func NewOllamaProvider(apiKey, baseURL, defaultModel string) LLMProvider {
	if baseURL == "" {
		baseURL = OllamaDefaultBaseURL
	}
	// Ollama 不需要 API Key，但 OpenAI 客户端需要非空值
	if apiKey == "" {
		apiKey = "ollama"
	}
	return &OllamaProvider{
		OpenAIProvider: NewOpenAIProvider(apiKey, baseURL, defaultModel).(*OpenAIProvider),
	}
}
