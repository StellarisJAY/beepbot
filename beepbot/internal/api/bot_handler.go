package api

import (
	"strconv"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/gin-gonic/gin"
)

// BotHandler 机器人处理器
type BotHandler struct {
	service *service.BotService
}

func NewBotHandler(service *service.BotService) *BotHandler {
	return &BotHandler{service: service}
}

// ListBots 获取机器人列表
// GET /api/v1/bots
func (h *BotHandler) ListBots(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	bots, total, err := h.service.ListBots(page, pageSize)
	if err != nil {
		InternalError(c, "failed to list bots: "+err.Error())
		return
	}

	SuccessWithPage(c, bots, total, page, pageSize)
}

// GetBot 获取单个机器人
// GET /api/v1/bots/:id
func (h *BotHandler) GetBot(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "bot id is required")
		return
	}

	bot, err := h.service.GetBotWithRelations(id)
	if err != nil {
		NotFound(c, "bot not found")
		return
	}

	Success(c, bot)
}

// CreateBot 创建机器人
// POST /api/v1/bots
func (h *BotHandler) CreateBot(c *gin.Context) {
	var req service.CreateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 验证平台类型
	if req.Platform != types.BotPlatformQQ {
		BadRequest(c, "unsupported platform: "+string(req.Platform))
		return
	}

	bot, err := h.service.CreateBot(&req)
	if err != nil {
		Error(c, 500, "failed to create bot: "+err.Error())
		return
	}

	Success(c, bot)
}

// UpdateBot 更新机器人
// PUT /api/v1/bots/:id
func (h *BotHandler) UpdateBot(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "bot id is required")
		return
	}

	var req service.UpdateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	bot, err := h.service.UpdateBot(id, &req)
	if err != nil {
		Error(c, 500, "failed to update bot: "+err.Error())
		return
	}

	Success(c, bot)
}

// DeleteBot 删除机器人
// DELETE /api/v1/bots/:id
func (h *BotHandler) DeleteBot(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "bot id is required")
		return
	}

	if err := h.service.DeleteBot(id); err != nil {
		Error(c, 500, "failed to delete bot: "+err.Error())
		return
	}

	Success(c, nil)
}

// UpdateBotStatus 更新机器人状态
// PUT /api/v1/bots/:id/status
func (h *BotHandler) UpdateBotStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "bot id is required")
		return
	}

	var req struct {
		Status types.BotStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 验证状态
	if req.Status != types.BotStatusActive && req.Status != types.BotStatusInactive {
		BadRequest(c, "invalid status, must be 'active' or 'inactive'")
		return
	}

	if err := h.service.UpdateBotStatus(id, req.Status); err != nil {
		Error(c, 500, "failed to update bot status: "+err.Error())
		return
	}

	Success(c, nil)
}

// BindAgent 绑定智能体
// PUT /api/v1/bots/:id/agent
func (h *BotHandler) BindAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "bot id is required")
		return
	}

	var req service.BindAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if err := h.service.BindAgent(id, req.AgentID); err != nil {
		Error(c, 500, "failed to bind agent: "+err.Error())
		return
	}

	Success(c, nil)
}

// GetUnboundBots 获取未绑定智能体的机器人
// GET /api/v1/bots/unbound
func (h *BotHandler) GetUnboundBots(c *gin.Context) {
	bots, err := h.service.GetUnboundBots()
	if err != nil {
		Error(c, 500, "failed to get unbound bots: "+err.Error())
		return
	}

	Success(c, bots)
}

// GetBotsByPlatform 根据平台获取机器人
// GET /api/v1/bots/platform/:platform
func (h *BotHandler) GetBotsByPlatform(c *gin.Context) {
	platform := types.BotPlatform(c.Param("platform"))
	if platform == "" {
		BadRequest(c, "platform is required")
		return
	}

	// 验证平台类型
	if platform != types.BotPlatformQQ {
		BadRequest(c, "unsupported platform: "+string(platform))
		return
	}

	bots, err := h.service.GetBotsByPlatform(platform)
	if err != nil {
		Error(c, 500, "failed to get bots by platform: "+err.Error())
		return
	}

	Success(c, bots)
}
