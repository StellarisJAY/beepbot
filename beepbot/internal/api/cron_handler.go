package api

import (
	"strconv"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/gin-gonic/gin"
)

// CronHandler 定时任务处理器
type CronHandler struct {
	service *service.CronService
}

// NewCronHandler 创建定时任务处理器实例
func NewCronHandler(service *service.CronService) *CronHandler {
	return &CronHandler{service: service}
}

// ListCronJobs 获取定时任务列表
// GET /api/v1/crons
func (h *CronHandler) ListCronJobs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	jobs, total, err := h.service.ListCronJobs(page, pageSize)
	if err != nil {
		InternalError(c, "failed to list cron jobs: "+err.Error())
		return
	}

	SuccessWithPage(c, jobs, total, page, pageSize)
}

// GetCronJob 获取单个定时任务
// GET /api/v1/crons/:id
func (h *CronHandler) GetCronJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "cron job id is required")
		return
	}

	job, err := h.service.GetCronJobWithAgent(id)
	if err != nil {
		NotFound(c, "cron job not found")
		return
	}

	Success(c, job)
}

// CreateCronJob 创建定时任务
// POST /api/v1/crons
func (h *CronHandler) CreateCronJob(c *gin.Context) {
	var req service.CreateCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	job, err := h.service.CreateCronJob(&req)
	if err != nil {
		Error(c, 500, "failed to create cron job: "+err.Error())
		return
	}

	Success(c, job)
}

// UpdateCronJob 更新定时任务
// PUT /api/v1/crons/:id
func (h *CronHandler) UpdateCronJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "cron job id is required")
		return
	}

	var req service.UpdateCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	job, err := h.service.UpdateCronJob(id, &req)
	if err != nil {
		Error(c, 500, "failed to update cron job: "+err.Error())
		return
	}

	Success(c, job)
}

// DeleteCronJob 删除定时任务
// DELETE /api/v1/crons/:id
func (h *CronHandler) DeleteCronJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "cron job id is required")
		return
	}

	if err := h.service.DeleteCronJob(id); err != nil {
		Error(c, 500, "failed to delete cron job: "+err.Error())
		return
	}

	Success(c, nil)
}

// UpdateCronJobStatus 更新定时任务状态
// PUT /api/v1/crons/:id/status
func (h *CronHandler) UpdateCronJobStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "cron job id is required")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if err := h.service.UpdateCronJobStatus(id, req.Enabled); err != nil {
		Error(c, 500, "failed to update cron job status: "+err.Error())
		return
	}

	Success(c, nil)
}

// GetCronJobsByAgent 获取指定智能体的定时任务
// GET /api/v1/crons/agent/:agent_id
func (h *CronHandler) GetCronJobsByAgent(c *gin.Context) {
	agentID := c.Param("agent_id")
	if agentID == "" {
		BadRequest(c, "agent id is required")
		return
	}

	jobs, err := h.service.GetCronJobsByAgent(agentID)
	if err != nil {
		Error(c, 500, "failed to get cron jobs by agent: "+err.Error())
		return
	}

	Success(c, jobs)
}
