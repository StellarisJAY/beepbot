package channel

import (
	"encoding/json"
	"fmt"

	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/datatypes"
)

// QQChannelConfig QQ Channel 配置
type QQChannelConfig struct {
	AppID         string   `json:"app_id"`
	AppSecret     string   `json:"app_secret"`
	AllowedUsers  []string `json:"allowed_users,omitempty"`
	AllowedGroups []string `json:"allowed_groups,omitempty"`
}

// FeishuChannelConfig 飞书 Channel 配置
type FeishuChannelConfig struct {
	AppID         string   `json:"app_id"`
	AppSecret     string   `json:"app_secret"`
	EncryptKey    string   `json:"encrypt_key,omitempty"`
	AllowedUsers  []string `json:"allowed_users,omitempty"`
	AllowedGroups []string `json:"allowed_groups,omitempty"`
}

// ParseQQChannelConfig 从 Bot.Config 解析 QQ 配置
func ParseQQChannelConfig(config datatypes.JSON) (*QQChannelConfig, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	var cfg QQChannelConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse QQ channel config: %w", err)
	}

	if cfg.AppID == "" {
		return nil, fmt.Errorf("app_id is required")
	}
	if cfg.AppSecret == "" {
		return nil, fmt.Errorf("app_secret is required")
	}

	return &cfg, nil
}

// ChannelConfig 通用 Channel 配置接口
type ChannelConfig interface {
	Validate() error
}

// Validate 验证 QQ Channel 配置
func (c *QQChannelConfig) Validate() error {
	if c.AppID == "" {
		return fmt.Errorf("app_id is required")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("app_secret is required")
	}
	return nil
}

// ParseFeishuChannelConfig 从 Bot.Config 解析飞书配置
func ParseFeishuChannelConfig(config datatypes.JSON) (*FeishuChannelConfig, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	var cfg FeishuChannelConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse Feishu channel config: %w", err)
	}

	if cfg.AppID == "" {
		return nil, fmt.Errorf("app_id is required")
	}
	if cfg.AppSecret == "" {
		return nil, fmt.Errorf("app_secret is required")
	}

	return &cfg, nil
}

// Validate 验证飞书 Channel 配置
func (c *FeishuChannelConfig) Validate() error {
	if c.AppID == "" {
		return fmt.Errorf("app_id is required")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("app_secret is required")
	}
	return nil
}

// ParseChannelConfig 根据平台类型解析配置
func ParseChannelConfig(platform types.BotPlatform, config datatypes.JSON) (ChannelConfig, error) {
	switch platform {
	case types.BotPlatformQQ:
		return ParseQQChannelConfig(config)
	case types.BotPlatformFeishu:
		return ParseFeishuChannelConfig(config)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}
