package cron

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron           *cron.Cron
	repo           repository.CronRepository
	bus            *channel.MessageBus
	channelManager *channel.ChannelManager // 用于获取 Channel 实例
	entryMap       map[string]cron.EntryID // job ID -> entry ID
	mu             sync.RWMutex
}

// NewScheduler 创建调度器
func NewScheduler(repo repository.CronRepository, bus *channel.MessageBus, channelManager *channel.ChannelManager) *Scheduler {
	return &Scheduler{
		cron:           cron.New(cron.WithSeconds()), // 支持秒级 cron
		repo:           repo,
		bus:            bus,
		channelManager: channelManager,
		entryMap:       make(map[string]cron.EntryID),
	}
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) error {
	// 加载所有启用的定时任务
	jobs, err := s.repo.GetEnabledJobs()
	if err != nil {
		return fmt.Errorf("failed to load enabled cron jobs: %w", err)
	}

	for _, job := range jobs {
		if err := s.addJob(job); err != nil {
			slog.Error("failed to add cron job", "id", job.ID, "name", job.Name, "error", err)
		}
	}

	s.cron.Start()
	slog.Info("Cron scheduler started", "jobs_count", len(s.entryMap))
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("Cron scheduler stopped")
}

// AddJob 添加定时任务到调度器
func (s *Scheduler) AddJob(job types.CronJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addJob(job)
}

// addJob 内部添加任务方法（不加锁）
func (s *Scheduler) addJob(job types.CronJob) error {
	// 如果已存在，先移除
	if entryID, exists := s.entryMap[job.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.entryMap, job.ID)
	}

	// 添加新任务
	entryID, err := s.cron.AddFunc(job.CronExpression, func() {
		s.triggerJob(job)
	})
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	s.entryMap[job.ID] = entryID
	slog.Info("Cron job added to scheduler", "id", job.ID, "name", job.Name, "cron", job.CronExpression)
	return nil
}

// RemoveJob 从调度器移除定时任务
func (s *Scheduler) RemoveJob(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entryMap[jobID]; exists {
		s.cron.Remove(entryID)
		delete(s.entryMap, jobID)
		slog.Info("Cron job removed from scheduler", "id", jobID)
	}
}

// UpdateJob 更新调度器中的定时任务
func (s *Scheduler) UpdateJob(job types.CronJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 先移除旧任务
	if entryID, exists := s.entryMap[job.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.entryMap, job.ID)
	}

	// 如果任务启用，重新添加
	if job.Enabled {
		return s.addJob(job)
	}

	slog.Info("Cron job updated in scheduler", "id", job.ID, "name", job.Name, "enabled", job.Enabled)
	return nil
}

// triggerJob 触发定时任务执行
func (s *Scheduler) triggerJob(job types.CronJob) {
	slog.Info("Cron job triggered", "id", job.ID, "name", job.Name, "agent_id", job.AgentID)

	// 检查是否有推送信息（通过智能体对话创建的定时任务）
	if job.PushBotID != nil && job.PushChatID != nil && s.channelManager != nil {
		// 尝试获取 Channel
		ch, exists := s.channelManager.GetChannel(*job.PushBotID)
		if !exists {
			slog.Warn("Channel not found for push, falling back to cron channel",
				"job_id", job.ID, "bot_id", *job.PushBotID)
			s.triggerCronChannel(job)
			return
		}

		// 检查是否支持主动推送
		if !ch.CanPushProactively() {
			slog.Warn("Channel does not support proactive push, falling back to cron channel",
				"job_id", job.ID, "channel", *job.PushChannel)
			s.triggerCronChannel(job)
			return
		}

		// 构造带会话信息的 InboundMessage，智能体处理完成后会推送到原会话
		msg := channel.InboundMessage{
			Channel:   *job.PushBotID, // 使用 BotID 作为 Channel（用于 OutboundMessage 分发）
			UserID:    ptrToString(job.PushUserID),
			GroupID:   ptrToString(job.PushGroupID),
			ChatID:    *job.PushChatID,
			MessageID: fmt.Sprintf("cron-push-%s-%d", job.ID, time.Now().Unix()),
			Content:   job.Message,
			AgentID:   job.AgentID,
		}

		slog.Info("Cron job with push info, routing to original session",
			"job_id", job.ID, "bot_id", *job.PushBotID, "chat_id", *job.PushChatID)

		// 发送到 MessageBus，智能体处理完成后会推送到原会话
		s.bus.PublishInbound(context.Background(), msg)
		return
	}

	// 无推送信息，走 cron channel（平台手动创建的定时任务）
	s.triggerCronChannel(job)
}

// triggerCronChannel 通过 cron channel 触发定时任务
func (s *Scheduler) triggerCronChannel(job types.CronJob) {
	msg := channel.InboundMessage{
		Channel:   channel.ChannelCron,
		UserID:    job.ID,
		GroupID:   "",
		MessageID: fmt.Sprintf("cron-%s-%d", job.ID, time.Now().Unix()),
		Content:   job.Message,
		AgentID:   job.AgentID,
	}
	s.bus.PublishInbound(context.Background(), msg)
}

// ptrToString 安全地将 *string 转换为 string
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Reload 重新加载所有定时任务
func (s *Scheduler) Reload(ctx context.Context) error {
	s.mu.Lock()
	// 停止所有任务
	for _, entryID := range s.entryMap {
		s.cron.Remove(entryID)
	}
	s.entryMap = make(map[string]cron.EntryID)
	s.mu.Unlock()

	// 重新加载
	return s.Start(ctx)
}

// GetEntryCount 获取当前调度器中的任务数量
func (s *Scheduler) GetEntryCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entryMap)
}
