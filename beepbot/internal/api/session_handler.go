package api

import (
	"strconv"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/gin-gonic/gin"
)

// SessionHandler 会话处理器
type SessionHandler struct {
	service *service.SessionService
}

// NewSessionHandler 创建会话处理器
func NewSessionHandler(service *service.SessionService) *SessionHandler {
	return &SessionHandler{service: service}
}

// GetAgentSessions 获取智能体的会话列表
// GET /api/v1/agents/:id/sessions
func (h *SessionHandler) GetAgentSessions(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		BadRequest(c, "agent id is required")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询参数
	query := &types.SessionQuery{
		SessionType: types.SessionType(c.Query("session_type")),
		Platform:    types.BotPlatform(c.Query("platform")),
	}

	// 查询会话列表
	items, total, err := h.service.GetSessionsByAgent(agentID, page, pageSize, query)
	if err != nil {
		InternalError(c, "failed to get sessions: "+err.Error())
		return
	}

	SuccessWithPage(c, items, total, page, pageSize)
}

// GetSessionMessages 获取会话消息列表
// GET /api/v1/sessions/:id/messages
func (h *SessionHandler) GetSessionMessages(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		BadRequest(c, "session id is required")
		return
	}

	// 解析分页参数
	beforeID := c.Query("before_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 查询消息列表
	response, err := h.service.GetSessionMessages(sessionID, beforeID, limit)
	if err != nil {
		InternalError(c, "failed to get messages: "+err.Error())
		return
	}

	Success(c, response)
}
