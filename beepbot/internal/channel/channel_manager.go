package channel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/StellarisJAY/beepbot/internal/types"
)

// 错误定义
var (
	ErrChannelAlreadyRunning = errors.New("channel is already running")
	ErrChannelNotRunning     = errors.New("channel is not running")
	ErrUnsupportedPlatform   = errors.New("unsupported platform")
)

// ChannelManager Channel 管理器
// 负责管理 Channel 的生命周期，提供动态启停接口
type ChannelManager struct {
	registry        *ChannelRegistry
	factoryRegistry *ChannelFactoryRegistry
	bus             *MessageBus

	ctx    context.Context
	cancel context.CancelFunc
}

// NewChannelManager 创建新的 Channel 管理器
func NewChannelManager(bus *MessageBus) *ChannelManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ChannelManager{
		registry:        NewChannelRegistry(),
		factoryRegistry: NewChannelFactoryRegistry(),
		bus:             bus,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// StartChannel 启动指定 Bot 的 Channel
// 返回 nil 表示成功，返回 error 表示失败
func (m *ChannelManager) StartChannel(ctx context.Context, bot *types.Bot) error {
	// 检查是否已存在
	if _, exists := m.registry.Get(bot.ID); exists {
		return ErrChannelAlreadyRunning
	}

	// 创建 Channel
	ch, err := m.factoryRegistry.CreateChannel(bot, m.bus)
	if err != nil {
		return fmt.Errorf("create channel failed: %w", err)
	}

	// 创建 Channel 上下文
	channelCtx, cancel := context.WithCancel(m.ctx)

	// 启动 Channel
	if err := ch.Start(channelCtx); err != nil {
		cancel()
		return fmt.Errorf("start channel failed: %w", err)
	}

	// 注册到 registry
	m.registry.Register(bot.ID, ch, cancel)

	slog.Info("channel started", "bot_id", bot.ID, "platform", bot.Platform)
	return nil
}

// StopChannel 停止指定 Bot 的 Channel
func (m *ChannelManager) StopChannel(botID string) error {
	ch, cancel, exists := m.registry.Unregister(botID)
	if !exists {
		return ErrChannelNotRunning
	}

	// 调用 Channel 的 Stop 方法
	ch.Stop()

	// 取消上下文
	if cancel != nil {
		cancel()
	}

	slog.Info("channel stopped", "bot_id", botID)
	return nil
}

// RestartChannel 重启指定 Bot 的 Channel
func (m *ChannelManager) RestartChannel(ctx context.Context, bot *types.Bot) error {
	// 先停止（忽略不存在的错误）
	_ = m.StopChannel(bot.ID)

	// 再启动
	return m.StartChannel(ctx, bot)
}

// GetChannel 获取运行中的 Channel
func (m *ChannelManager) GetChannel(botID string) (Channel, bool) {
	return m.registry.Get(botID)
}

// ListRunningChannels 列出所有运行中的 Channel ID
func (m *ChannelManager) ListRunningChannels() []string {
	return m.registry.ListIDs()
}

// IsChannelRunning 检查 Channel 是否正在运行
func (m *ChannelManager) IsChannelRunning(botID string) bool {
	_, exists := m.registry.Get(botID)
	return exists
}

// StartAllActiveChannels 启动所有 active 状态的 Bot Channel
// 返回所有错误，不中断启动过程
func (m *ChannelManager) StartAllActiveChannels(ctx context.Context, bots []types.Bot) []error {
	var errs []error

	for i := range bots {
		bot := &bots[i]
		if bot.Status != types.BotStatusActive {
			slog.Debug("skipping inactive bot", "bot_id", bot.ID, "status", bot.Status)
			continue
		}

		if err := m.StartChannel(ctx, bot); err != nil {
			errs = append(errs, fmt.Errorf("bot %s: %w", bot.ID, err))
			slog.Error("failed to start channel", "bot_id", bot.ID, "error", err)
		}
	}

	return errs
}

// StopAllChannels 停止所有 Channel
func (m *ChannelManager) StopAllChannels() {
	ids := m.registry.ListIDs()
	for _, id := range ids {
		if err := m.StopChannel(id); err != nil {
			slog.Error("failed to stop channel", "bot_id", id, "error", err)
		}
	}
	slog.Info("all channels stopped")
}

// DispatchOutbound 分发出站消息
func (m *ChannelManager) DispatchOutbound(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("channel manager outbound dispatcher stopped")
			return
		default:
			msg, ok := m.bus.ConsumeOutbound(ctx)
			if !ok {
				continue
			}
			botID := msg.Channel
			channel, exists := m.registry.Get(botID)
			if !exists {
				slog.Warn("channel not found for outbound message", "bot_id", botID, "channel", msg.Channel)
				continue
			}

			if !channel.IsAvailable() {
				slog.Warn("channel is not available", "bot_id", botID)
				continue
			}

			if !channel.IsAllowed(msg.UserID) {
				slog.Warn("user not allowed", "bot_id", botID, "user_id", msg.UserID)
				continue
			}

			if err := channel.Send(ctx, msg); err != nil {
				slog.Error("failed to send outbound message", "bot_id", botID, "error", err)
			}
		}
	}
}

// Shutdown 关闭 ChannelManager
func (m *ChannelManager) Shutdown() {
	m.StopAllChannels()
	m.cancel()
}
