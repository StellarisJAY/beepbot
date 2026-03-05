package provider

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/types"
)

const DashScopeBaseURL = "https://dashscope.aliyuncs.com/api/v2/apps/protocols/compatible-mode/v1"
const OpenAIBaseURl = "https://openai.org/"

type BaseProvider struct {
	apiKey       string
	baseURL      string
	defaultModel string
}

func CreateLLMProvider(config config.StandaloneConfig) (types.LLMProvider, error) {
	provider := strings.ToLower(config.AgentConfig.Provider)
	model := config.AgentConfig.Model

	switch provider {
	case "openai":
		slog.Info("Using OpenAI provider", "model", model)
		apiKey := config.ProvidersConfig.OpenAI.APIKey
		baseURL := config.ProvidersConfig.OpenAI.BaseURL
		if strings.TrimSpace(baseURL) == "" {
			baseURL = OpenAIBaseURl
		}
		return NewOpenAIProvider(apiKey, baseURL, model), nil
	case "dashscope":
		slog.Info("Using DashScope provider", "model", model)
		apiKey := config.ProvidersConfig.DashScope.APIKey
		baseURL := config.ProvidersConfig.DashScope.BaseURL
		if strings.TrimSpace(baseURL) == "" {
			baseURL = DashScopeBaseURL
		}
		return NewOpenAIProvider(apiKey, baseURL, model), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s, available providers: openai, dashscope", provider)
	}
}
