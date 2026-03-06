package api

import (
	"strconv"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/gin-gonic/gin"
)

// AgentHandler 智能体处理器
type AgentHandler struct {
	service *service.AgentService
}

func NewAgentHandler(service *service.AgentService) *AgentHandler {
	return &AgentHandler{service: service}
}

// ListAgents 获取智能体列表
// GET /api/v1/agents
func (h *AgentHandler) ListAgents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	agents, total, err := h.service.ListAgents(page, pageSize)
	if err != nil {
		InternalError(c, "failed to list agents: "+err.Error())
		return
	}

	SuccessWithPage(c, agents, total, page, pageSize)
}

// GetAgent 获取单个智能体
// GET /api/v1/agents/:id
func (h *AgentHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "agent id is required")
		return
	}

	agent, err := h.service.GetAgentWithRelations(id)
	if err != nil {
		NotFound(c, "agent not found")
		return
	}

	Success(c, agent)
}

// CreateAgent 创建智能体
// POST /api/v1/agents
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req service.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	agent, err := h.service.CreateAgent(&req)
	if err != nil {
		Error(c, 500, "failed to create agent: "+err.Error())
		return
	}

	Success(c, agent)
}

// UpdateAgent 更新智能体
// PUT /api/v1/agents/:id
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "agent id is required")
		return
	}

	var req service.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	agent, err := h.service.UpdateAgent(id, &req)
	if err != nil {
		Error(c, 500, "failed to update agent: "+err.Error())
		return
	}

	Success(c, agent)
}

// DeleteAgent 删除智能体
// DELETE /api/v1/agents/:id
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "agent id is required")
		return
	}

	if err := h.service.DeleteAgent(id); err != nil {
		Error(c, 500, "failed to delete agent: "+err.Error())
		return
	}

	Success(c, nil)
}

// UpdateAgentStatus 更新智能体状态
// PUT /api/v1/agents/:id/status
func (h *AgentHandler) UpdateAgentStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "agent id is required")
		return
	}

	var req struct {
		Status types.AgentStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 验证状态
	if req.Status != types.AgentStatusActive && req.Status != types.AgentStatusInactive {
		BadRequest(c, "invalid status, must be 'active' or 'inactive'")
		return
	}

	if err := h.service.UpdateAgentStatus(id, req.Status); err != nil {
		Error(c, 500, "failed to update agent status: "+err.Error())
		return
	}

	Success(c, nil)
}

// GetActiveAgents 获取所有活跃的智能体
// GET /api/v1/agents/active
func (h *AgentHandler) GetActiveAgents(c *gin.Context) {
	agents, err := h.service.GetActiveAgents()
	if err != nil {
		Error(c, 500, "failed to get active agents: "+err.Error())
		return
	}

	Success(c, agents)
}

// GetAgentDefaults 获取智能体默认配置
// GET /api/v1/agents/defaults
func (h *AgentHandler) GetAgentDefaults(c *gin.Context) {
	defaults := h.service.GetAgentDefaults()
	Success(c, defaults)
}

// ValidateAgent 校验智能体配置是否完整
// POST /api/v1/agents/:id/validate
func (h *AgentHandler) ValidateAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "agent id is required")
		return
	}

	agent, err := h.service.GetAgent(id)
	if err != nil {
		NotFound(c, "agent not found")
		return
	}

	result := h.service.ValidateAgentConfig(agent)
	Success(c, result)
}
