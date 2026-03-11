package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/StellarisJAY/beepbot/internal/cron"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// contextKey 上下文键类型
type contextKey string

const (
	// AgentIDKey 上下文中智能体ID的键
	AgentIDKey contextKey = "agentID"
	// SessionInfoKey 上下文中会话推送信息的键
	SessionInfoKey contextKey = "sessionInfo"
)

// SessionPushInfo 会话推送信息，用于定时任务主动推送
type SessionPushInfo struct {
	Channel string // 渠道类型（Bot.Platform）
	BotID   string // 机器人ID（Channel ID）
	UserID  string // 发送者ID
	GroupID string // 群聊ID（仅群聊时有值）
	ChatID  string // 会话ID（飞书 chat_id）
	AgentID string // 当前智能体ID
}

// CronServiceInterface 定时任务服务接口，用于解耦
type CronServiceInterface interface {
	CreateCronJob(agentID, name, cronExpr, message string, enabled bool) (*types.CronJob, error)
	DeleteCronJob(id string) error
	GetCronJobByID(id string) (*types.CronJob, error)
	GetCronJobsByAgentID(agentID string) ([]types.CronJob, error)
	UpdateCronJobStatus(id string, enabled bool) error
}

// CronToolDeps CronTool的依赖项
type CronToolDeps struct {
	CronRepo   repository.CronRepository
	Scheduler  *cron.Scheduler
	Validator  *CronValidator
	AgentCheck func(agentID string) bool // 检查智能体是否存在的函数
}

// CronTool 定时任务工具，允许智能体管理自己的定时任务
type CronTool struct {
	deps *CronToolDeps
}

// NewCronTool 创建新的定时任务工具
func NewCronTool(deps *CronToolDeps) *CronTool {
	if deps.Validator == nil {
		deps.Validator = DefaultCronValidator()
	}
	return &CronTool{deps: deps}
}

// Name 返回工具名称
func (t *CronTool) Name() string {
	return "cron"
}

// Description 返回工具描述
func (t *CronTool) Description() string {
	return `定时任务管理工具，用于创建、删除、查看和管理定时任务。

操作类型:
- create: 创建新的定时任务
- delete: 删除定时任务（通过名称）
- list: 列出当前智能体的所有定时任务
- get: 获取指定定时任务的详细信息
- toggle: 启用或禁用定时任务

Cron表达式格式（6字段）:
秒 分 时 日 月 周

示例:
- "0 30 * * * *" - 每小时30分执行
- "0 0 9 * * 1-5" - 周一到周五早上9点执行
- "0 */5 * * * *" - 每5分钟执行一次

注意: 执行频率最快为每分钟一次，更频繁的表达式将被拒绝。`
}

// Parameters 返回工具参数定义
func (t *CronTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"enum":        []string{"create", "delete", "list", "get", "toggle"},
				"description": "操作类型: create(创建), delete(删除), list(列表), get(详情), toggle(启用/禁用)",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "定时任务名称（创建、删除、获取详情、切换状态时必填）",
			},
			"cron_expression": map[string]any{
				"type":        "string",
				"description": "Cron表达式，6字段格式：秒 分 时 日 月 周（创建时必填）",
			},
			"message": map[string]any{
				"type":        "string",
				"description": "任务内容，比如总结文档,查询天气",
			},
			"enabled": map[string]any{
				"type":        "boolean",
				"description": "是否启用（创建和切换状态时使用）",
			},
		},
		"required": []string{"action"},
	}
}

// Execute 执行工具
func (t *CronTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	// 从上下文获取智能体ID
	agentID, ok := ctx.Value(AgentIDKey).(string)
	if !ok || agentID == "" {
		return "", errors.New("无法获取智能体ID，定时任务工具只能在智能体上下文中使用")
	}

	action, _ := params["action"].(string)
	if action == "" {
		return "", errors.New("缺少必需参数: action")
	}

	switch action {
	case "create":
		return t.handleCreate(ctx, agentID, params)
	case "delete":
		return t.handleDelete(agentID, params)
	case "list":
		return t.handleList(agentID)
	case "get":
		return t.handleGet(agentID, params)
	case "toggle":
		return t.handleToggle(agentID, params)
	default:
		return "", fmt.Errorf("未知的操作类型: %s", action)
	}
}

// handleCreate 处理创建定时任务
func (t *CronTool) handleCreate(ctx context.Context, agentID string, params map[string]any) (string, error) {
	name, _ := params["name"].(string)
	if name == "" {
		return "", errors.New("创建定时任务需要提供 name 参数")
	}

	cronExpr, _ := params["cron_expression"].(string)
	if cronExpr == "" {
		return "", errors.New("创建定时任务需要提供 cron_expression 参数")
	}

	message, _ := params["message"].(string)
	if message == "" {
		return "", errors.New("创建定时任务需要提供 message 参数")
	}

	enabled := true
	if e, ok := params["enabled"]; ok {
		switch v := e.(type) {
		case bool:
			enabled = v
		case string:
			enabled = v == "true"
		}
	}

	// 验证cron表达式格式和频率
	if _, err := t.deps.Validator.Validate(cronExpr); err != nil {
		return "", fmt.Errorf("cron表达式验证失败: %w", err)
	}

	// 检查名称是否已存在
	if existing, err := t.deps.CronRepo.GetByName(name); err == nil && existing != nil {
		return "", fmt.Errorf("定时任务名称 '%s' 已存在", name)
	}

	// 创建定时任务
	job := &types.CronJob{
		ID:             types.GenerateUUIDv7(),
		Name:           name,
		CronExpression: cronExpr,
		AgentID:        agentID,
		Message:        message,
		Enabled:        types.CronJobStatus(enabled),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 从上下文获取会话推送信息
	if sessionInfo, ok := ctx.Value(SessionInfoKey).(*SessionPushInfo); ok && sessionInfo != nil {
		job.PushChannel = &sessionInfo.Channel
		job.PushBotID = &sessionInfo.BotID
		job.PushUserID = &sessionInfo.UserID
		if sessionInfo.GroupID != "" {
			job.PushGroupID = &sessionInfo.GroupID
		}
		if sessionInfo.ChatID != "" {
			job.PushChatID = &sessionInfo.ChatID
		}
	}

	if err := t.deps.CronRepo.Create(job); err != nil {
		return "", fmt.Errorf("创建定时任务失败: %w", err)
	}

	// 如果任务启用且调度器存在，添加到调度器
	if job.Enabled && t.deps.Scheduler != nil {
		if err := t.deps.Scheduler.AddJob(*job); err != nil {
			// 记录错误但不回滚数据库操作
		}
	}

	result := map[string]any{
		"success": true,
		"message": "定时任务创建成功",
		"job":     t.jobToMap(job),
	}

	return toJSON(result)
}

// handleDelete 处理删除定时任务
func (t *CronTool) handleDelete(agentID string, params map[string]any) (string, error) {
	name, _ := params["name"].(string)
	if name == "" {
		return "", errors.New("删除定时任务需要提供 name 参数")
	}

	// 通过名称查找任务
	job, err := t.deps.CronRepo.GetByName(name)
	if err != nil {
		return "", fmt.Errorf("未找到名为 '%s' 的定时任务", name)
	}

	// 验证任务属于当前智能体
	if job.AgentID != agentID {
		return "", errors.New("无权删除此定时任务：任务不属于当前智能体")
	}

	// 从调度器移除
	if t.deps.Scheduler != nil {
		t.deps.Scheduler.RemoveJob(job.ID)
	}

	// 删除任务
	if err := t.deps.CronRepo.Delete(job.ID); err != nil {
		return "", fmt.Errorf("删除定时任务失败: %w", err)
	}

	result := map[string]any{
		"success": true,
		"message": fmt.Sprintf("定时任务 '%s' 已删除", name),
	}

	return toJSON(result)
}

// handleList 处理列出定时任务
func (t *CronTool) handleList(agentID string) (string, error) {
	jobs, err := t.deps.CronRepo.GetByAgentID(agentID)
	if err != nil {
		return "", fmt.Errorf("获取定时任务列表失败: %w", err)
	}

	jobList := make([]map[string]any, 0, len(jobs))
	for _, job := range jobs {
		jobList = append(jobList, map[string]any{
			"id":              job.ID,
			"name":            job.Name,
			"cron_expression": job.CronExpression,
			"message":         job.Message,
			"enabled":         bool(job.Enabled),
			"created_at":      job.CreatedAt.Format(time.RFC3339),
			"updated_at":      job.UpdatedAt.Format(time.RFC3339),
		})
	}

	result := map[string]any{
		"success": true,
		"count":   len(jobList),
		"jobs":    jobList,
	}

	return toJSON(result)
}

// handleGet 处理获取定时任务详情
func (t *CronTool) handleGet(agentID string, params map[string]any) (string, error) {
	name, _ := params["name"].(string)
	if name == "" {
		return "", errors.New("获取定时任务详情需要提供 name 参数")
	}

	// 通过名称查找任务
	job, err := t.deps.CronRepo.GetByName(name)
	if err != nil {
		return "", fmt.Errorf("未找到名为 '%s' 的定时任务", name)
	}

	// 验证任务属于当前智能体
	if job.AgentID != agentID {
		return "", errors.New("无权查看此定时任务：任务不属于当前智能体")
	}

	// 获取下次执行时间
	nextRuns := t.getNextRunTimes(job.CronExpression, 5)

	result := map[string]any{
		"success":   true,
		"job":       t.jobToMap(job),
		"next_runs": nextRuns,
	}

	return toJSON(result)
}

// handleToggle 处理启用/禁用定时任务
func (t *CronTool) handleToggle(agentID string, params map[string]any) (string, error) {
	name, _ := params["name"].(string)
	if name == "" {
		return "", errors.New("切换定时任务状态需要提供 name 参数")
	}

	enabled, ok := params["enabled"].(bool)
	if !ok {
		// 尝试从字符串转换
		if e, ok := params["enabled"].(string); ok {
			enabled = e == "true"
		} else {
			return "", errors.New("切换定时任务状态需要提供 enabled 参数")
		}
	}

	// 通过名称查找任务
	job, err := t.deps.CronRepo.GetByName(name)
	if err != nil {
		return "", fmt.Errorf("未找到名为 '%s' 的定时任务", name)
	}

	// 验证任务属于当前智能体
	if job.AgentID != agentID {
		return "", errors.New("无权操作此定时任务：任务不属于当前智能体")
	}

	// 更新状态
	job.Enabled = types.CronJobStatus(enabled)
	job.UpdatedAt = time.Now()

	if err := t.deps.CronRepo.Update(job); err != nil {
		return "", fmt.Errorf("更新定时任务状态失败: %w", err)
	}

	// 更新调度器中的任务
	if t.deps.Scheduler != nil {
		if enabled {
			_ = t.deps.Scheduler.AddJob(*job)
		} else {
			t.deps.Scheduler.RemoveJob(job.ID)
		}
	}

	status := "已禁用"
	if enabled {
		status = "已启用"
	}

	result := map[string]any{
		"success": true,
		"message": fmt.Sprintf("定时任务 '%s' %s", name, status),
		"job": map[string]any{
			"id":              job.ID,
			"name":            name,
			"cron_expression": job.CronExpression,
			"enabled":         enabled,
		},
	}

	return toJSON(result)
}

// jobToMap 将CronJob转换为map
func (t *CronTool) jobToMap(job *types.CronJob) map[string]any {
	return map[string]any{
		"id":              job.ID,
		"name":            job.Name,
		"cron_expression": job.CronExpression,
		"agent_id":        job.AgentID,
		"message":         job.Message,
		"enabled":         bool(job.Enabled),
		"created_at":      job.CreatedAt.Format(time.RFC3339),
		"updated_at":      job.UpdatedAt.Format(time.RFC3339),
	}
}

// getNextRunTimes 获取接下来几次执行时间
func (t *CronTool) getNextRunTimes(cronExpr string, count int) []string {
	schedule, err := t.deps.Validator.Validate(cronExpr)
	if err != nil {
		return nil
	}

	times := make([]string, 0, count)
	now := time.Now()
	next := schedule.Next(now)
	for i := 0; i < count; i++ {
		times = append(times, next.Format(time.RFC3339))
		next = schedule.Next(next)
	}
	return times
}

// toJSON 将结果转换为JSON字符串
func toJSON(data map[string]any) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化结果失败: %w", err)
	}
	return string(bytes), nil
}
