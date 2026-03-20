package api

import (
	"strconv"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/gin-gonic/gin"
)

// TeamHandler 团队处理器
type TeamHandler struct {
	service *service.TeamService
}

func NewTeamHandler(service *service.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

// ListTeams 获取团队列表
// GET /api/v1/teams
func (h *TeamHandler) ListTeams(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询参数
	query := &types.TeamQuery{
		Name:   c.Query("name"),
		Status: types.TeamStatus(c.Query("status")),
	}

	teams, total, err := h.service.ListTeams(page, pageSize, query)
	if err != nil {
		InternalError(c, "failed to list teams: "+err.Error())
		return
	}

	SuccessWithPage(c, teams, total, page, pageSize)
}

// GetTeam 获取单个团队
// GET /api/v1/teams/:id
func (h *TeamHandler) GetTeam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "team id is required")
		return
	}

	team, err := h.service.GetTeam(id)
	if err != nil {
		NotFound(c, "team not found")
		return
	}

	Success(c, team)
}

// CreateTeam 创建团队
// POST /api/v1/teams
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req service.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	team, err := h.service.CreateTeam(&req)
	if err != nil {
		Error(c, 500, "failed to create team: "+err.Error())
		return
	}

	Success(c, team)
}

// UpdateTeam 更新团队
// PUT /api/v1/teams/:id
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "team id is required")
		return
	}

	var req service.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	team, err := h.service.UpdateTeam(id, &req)
	if err != nil {
		Error(c, 500, "failed to update team: "+err.Error())
		return
	}

	Success(c, team)
}

// DeleteTeam 删除团队
// DELETE /api/v1/teams/:id
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "team id is required")
		return
	}

	if err := h.service.DeleteTeam(id); err != nil {
		Error(c, 500, "failed to delete team: "+err.Error())
		return
	}

	Success(c, nil)
}

// UpdateTeamStatus 更新团队状态
// PUT /api/v1/teams/:id/status
func (h *TeamHandler) UpdateTeamStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "team id is required")
		return
	}

	var req struct {
		Status types.TeamStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 验证状态
	if req.Status != types.TeamStatusActive && req.Status != types.TeamStatusInactive {
		BadRequest(c, "invalid status, must be 'active' or 'inactive'")
		return
	}

	if err := h.service.UpdateTeamStatus(id, req.Status); err != nil {
		Error(c, 500, "failed to update team status: "+err.Error())
		return
	}

	Success(c, nil)
}

// GetAgentTeams 获取智能体所属的团队列表
// GET /api/v1/teams/agent/:agent_id
func (h *TeamHandler) GetAgentTeams(c *gin.Context) {
	agentID := c.Param("agent_id")
	if agentID == "" {
		BadRequest(c, "agent id is required")
		return
	}

	teams, err := h.service.GetAgentTeams(agentID)
	if err != nil {
		Error(c, 500, "failed to get agent teams: "+err.Error())
		return
	}

	Success(c, teams)
}
