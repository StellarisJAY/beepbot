package service

import (
	"errors"
	"time"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度器接口
// 用于解耦 CronService 和具体的 scheduler 实现
type Scheduler interface {
	AddJob(job types.CronJob) error
	RemoveJob(jobID string)
	UpdateJob(job types.CronJob) error
}

// CronService 定时任务服务
type CronService struct {
	repo         repository.CronRepository
	agentService *AgentService
	scheduler    Scheduler
}

// NewCronService 创建定时任务服务实例
func NewCronService(repo repository.CronRepository, agentService *AgentService) *CronService {
	return &CronService{
		repo:         repo,
		agentService: agentService,
	}
}

// SetScheduler 设置调度器
func (s *CronService) SetScheduler(scheduler Scheduler) {
	s.scheduler = scheduler
}

// CreateCronJobRequest 创建定时任务请求
type CreateCronJobRequest struct {
	Name           string `json:"name" binding:"required"`
	CronExpression string `json:"cron_expression" binding:"required"`
	AgentID        string `json:"agent_id" binding:"required"`
	Message        string `json:"message" binding:"required"`
	Enabled        bool   `json:"enabled"`
}

// UpdateCronJobRequest 更新定时任务请求
type UpdateCronJobRequest struct {
	Name           string `json:"name" binding:"required"`
	CronExpression string `json:"cron_expression" binding:"required"`
	AgentID        string `json:"agent_id" binding:"required"`
	Message        string `json:"message" binding:"required"`
	Enabled        bool   `json:"enabled"`
}

// CronJobResponse 定时任务响应
type CronJobResponse struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	CronExpression string              `json:"cron_expression"`
	AgentID        string              `json:"agent_id"`
	Agent          *AgentBriefResponse `json:"agent,omitempty"`
	Message        string              `json:"message"`
	Enabled        bool                `json:"enabled"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// AgentBriefResponse 智能体简要信息响应
type AgentBriefResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ValidateCronExpression 验证 Cron 表达式是否有效
// 支持 6 字段格式（秒 分 时 日 月 周），与 scheduler 的 cron.WithSeconds() 一致
func ValidateCronExpression(expression string) bool {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expression)
	return err == nil
}

// CreateCronJob 创建定时任务
func (s *CronService) CreateCronJob(req *CreateCronJobRequest) (*types.CronJob, error) {
	// 检查名称是否已存在
	if _, err := s.repo.GetByName(req.Name); err == nil {
		return nil, errors.New("cron job name already exists")
	}

	// 验证 Cron 表达式
	if !ValidateCronExpression(req.CronExpression) {
		return nil, errors.New("invalid cron expression")
	}

	// 检查智能体是否存在
	if _, err := s.agentService.GetAgent(req.AgentID); err != nil {
		return nil, errors.New("agent not found")
	}

	job := &types.CronJob{
		ID:             types.GenerateUUIDv7(),
		Name:           req.Name,
		CronExpression: req.CronExpression,
		AgentID:        req.AgentID,
		Message:        req.Message,
		Enabled:        types.CronJobStatus(req.Enabled),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(job); err != nil {
		return nil, err
	}

	// 如果任务启用且调度器存在，添加到调度器
	if job.Enabled && s.scheduler != nil {
		if err := s.scheduler.AddJob(*job); err != nil {
			// 记录错误但不回滚数据库操作，因为数据库操作已成功
			// 可以通过日志或监控来处理这种情况
		}
	}

	return job, nil
}

// UpdateCronJob 更新定时任务
func (s *CronService) UpdateCronJob(id string, req *UpdateCronJobRequest) (*types.CronJob, error) {
	job, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 检查名称是否被其他任务使用
	if existing, err := s.repo.GetByName(req.Name); err == nil && existing.ID != id {
		return nil, errors.New("cron job name already exists")
	}

	// 验证 Cron 表达式
	if !ValidateCronExpression(req.CronExpression) {
		return nil, errors.New("invalid cron expression")
	}

	// 检查智能体是否存在
	if _, err := s.agentService.GetAgent(req.AgentID); err != nil {
		return nil, errors.New("agent not found")
	}

	job.Name = req.Name
	job.CronExpression = req.CronExpression
	job.AgentID = req.AgentID
	job.Message = req.Message
	job.Enabled = types.CronJobStatus(req.Enabled)
	job.UpdatedAt = time.Now()

	if err := s.repo.Update(job); err != nil {
		return nil, err
	}

	// 更新调度器中的任务
	if s.scheduler != nil {
		if err := s.scheduler.UpdateJob(*job); err != nil {
			// 记录错误但不回滚数据库操作
		}
	}

	return job, nil
}

// DeleteCronJob 删除定时任务
func (s *CronService) DeleteCronJob(id string) error {
	// 先从调度器移除
	if s.scheduler != nil {
		s.scheduler.RemoveJob(id)
	}

	return s.repo.Delete(id)
}

// GetCronJob 获取单个定时任务
func (s *CronService) GetCronJob(id string) (*types.CronJob, error) {
	return s.repo.GetByID(id)
}

// GetCronJobWithAgent 获取定时任务及其关联的智能体信息
func (s *CronService) GetCronJobWithAgent(id string) (*CronJobResponse, error) {
	job, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := &CronJobResponse{
		ID:             job.ID,
		Name:           job.Name,
		CronExpression: job.CronExpression,
		AgentID:        job.AgentID,
		Message:        job.Message,
		Enabled:        bool(job.Enabled),
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
	}

	// 获取关联的智能体信息
	if agent, err := s.agentService.GetAgent(job.AgentID); err == nil {
		response.Agent = &AgentBriefResponse{
			ID:   agent.ID,
			Name: agent.Name,
		}
	}

	return response, nil
}

// ListCronJobs 获取定时任务列表（支持筛选）
func (s *CronService) ListCronJobs(page, pageSize int, query *types.CronQuery) ([]CronJobResponse, int64, error) {
	jobs, total, err := s.repo.ListWithQuery(page, pageSize, query)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]CronJobResponse, 0, len(jobs))
	for _, job := range jobs {
		response := CronJobResponse{
			ID:             job.ID,
			Name:           job.Name,
			CronExpression: job.CronExpression,
			AgentID:        job.AgentID,
			Message:        job.Message,
			Enabled:        bool(job.Enabled),
			CreatedAt:      job.CreatedAt,
			UpdatedAt:      job.UpdatedAt,
		}

		// 获取关联的智能体信息
		if agent, err := s.agentService.GetAgent(job.AgentID); err == nil {
			response.Agent = &AgentBriefResponse{
				ID:   agent.ID,
				Name: agent.Name,
			}
		}

		responses = append(responses, response)
	}

	return responses, total, nil
}

// UpdateCronJobStatus 更新定时任务状态
func (s *CronService) UpdateCronJobStatus(id string, enabled bool) error {
	job, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	job.Enabled = types.CronJobStatus(enabled)
	job.UpdatedAt = time.Now()

	if err := s.repo.Update(job); err != nil {
		return err
	}

	// 更新调度器中的任务状态
	if s.scheduler != nil {
		if err := s.scheduler.UpdateJob(*job); err != nil {
			// 记录错误但不回滚数据库操作
		}
	}

	return nil
}

// GetCronJobsByAgent 获取指定智能体的定时任务
func (s *CronService) GetCronJobsByAgent(agentID string) ([]CronJobResponse, error) {
	jobs, err := s.repo.GetByAgentID(agentID)
	if err != nil {
		return nil, err
	}

	responses := make([]CronJobResponse, 0, len(jobs))
	for _, job := range jobs {
		response := CronJobResponse{
			ID:             job.ID,
			Name:           job.Name,
			CronExpression: job.CronExpression,
			AgentID:        job.AgentID,
			Message:        job.Message,
			Enabled:        bool(job.Enabled),
			CreatedAt:      job.CreatedAt,
			UpdatedAt:      job.UpdatedAt,
		}

		// 获取关联的智能体信息
		if agent, err := s.agentService.GetAgent(job.AgentID); err == nil {
			response.Agent = &AgentBriefResponse{
				ID:   agent.ID,
				Name: agent.Name,
			}
		}

		responses = append(responses, response)
	}

	return responses, nil
}
