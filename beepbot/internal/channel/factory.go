package channel

import (
	"fmt"

	"github.com/StellarisJAY/beepbot/internal/types"
)

// ChannelFactory Channel 工厂接口
type ChannelFactory interface {
	// CreateChannel 根据 Bot 配置创建 Channel
	CreateChannel(bot *types.Bot, bus *MessageBus) (Channel, error)
}

// QQChannelFactory QQ Channel 工厂
type QQChannelFactory struct{}

// NewQQChannelFactory 创建 QQ Channel 工厂
func NewQQChannelFactory() *QQChannelFactory {
	return &QQChannelFactory{}
}

// CreateChannel 根据 Bot 配置创建 QQ Channel
func (f *QQChannelFactory) CreateChannel(bot *types.Bot, bus *MessageBus) (Channel, error) {
	cfg, err := ParseQQChannelConfig(bot.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse QQ channel config: %w", err)
	}
	// 机器人ID唯一，所以机器人ID就是channelID
	channelID := fmt.Sprintf("%s", bot.ID)
	return NewQQBotChannel(cfg.AppID, cfg.AppSecret, channelID, bus), nil
}

// FeishuChannelFactory 飞书 Channel 工厂
type FeishuChannelFactory struct{}

// NewFeishuChannelFactory 创建飞书 Channel 工厂
func NewFeishuChannelFactory() *FeishuChannelFactory {
	return &FeishuChannelFactory{}
}

// CreateChannel 根据 Bot 配置创建飞书 Channel
func (f *FeishuChannelFactory) CreateChannel(bot *types.Bot, bus *MessageBus) (Channel, error) {
	cfg, err := ParseFeishuChannelConfig(bot.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Feishu channel config: %w", err)
	}
	channelID := fmt.Sprintf("%s", bot.ID)
	return NewFeishuChannel(cfg.AppID, cfg.AppSecret, cfg.EncryptKey, channelID, bus, cfg.AllowedUsers, cfg.AllowedGroups), nil
}

// ChannelFactoryRegistry 工厂注册表
type ChannelFactoryRegistry struct {
	factories map[types.BotPlatform]ChannelFactory
}

// NewChannelFactoryRegistry 创建工厂注册表
func NewChannelFactoryRegistry() *ChannelFactoryRegistry {
	registry := &ChannelFactoryRegistry{
		factories: make(map[types.BotPlatform]ChannelFactory),
	}

	// 注册默认工厂
	registry.Register(types.BotPlatformQQ, NewQQChannelFactory())
	registry.Register(types.BotPlatformFeishu, NewFeishuChannelFactory())

	return registry
}

// Register 注册工厂
func (r *ChannelFactoryRegistry) Register(platform types.BotPlatform, factory ChannelFactory) {
	r.factories[platform] = factory
}

// Get 获取工厂
func (r *ChannelFactoryRegistry) Get(platform types.BotPlatform) (ChannelFactory, bool) {
	factory, exists := r.factories[platform]
	return factory, exists
}

// CreateChannel 根据平台类型创建 Channel
func (r *ChannelFactoryRegistry) CreateChannel(bot *types.Bot, bus *MessageBus) (Channel, error) {
	factory, exists := r.Get(bot.Platform)
	if !exists {
		return nil, fmt.Errorf("unsupported platform: %s", bot.Platform)
	}
	return factory.CreateChannel(bot, bus)
}
