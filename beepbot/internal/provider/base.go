package provider

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/types"
)

const DashScopeBaseURL = "https://dashscope.aliyuncs.com/api/v2/apps/protocols/compatible-mode/v1"
const OpenAIBaseURl = "https://openai.org/"

type BaseProvider struct {
	apiKey       string
	baseURL      string
	defaultModel string
}

func CreateLLMProviderFromApi(provider *types.Provider, model string) (types.LLMProvider, error) {
	providerType := provider.ProviderType
	model = strings.ToLower(model)
	apiKey := provider.APIKey
	baseURL := provider.BaseURL
	switch providerType {
	case "openai":
		slog.Info("Using OpenAI provider", "model", model)
		if strings.TrimSpace(baseURL) == "" {
			baseURL = OpenAIBaseURl
		}
		return NewOpenAIProvider(apiKey, baseURL, model), nil
	case "dashscope":
		slog.Info("Using DashScope provider", "model", model)
		if strings.TrimSpace(baseURL) == "" {
			baseURL = DashScopeBaseURL
		}
		return NewOpenAIProvider(apiKey, baseURL, model), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s, available providers: openai, dashscope", providerType)
	}
}
