package api

import (
	"encoding/json"
	"log/slog"

	"github.com/StellarisJAY/beepbot/internal/agent/react"
	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/mcp"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/tool"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/gin-gonic/gin"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	agentService    *service.AgentService
	sessionService  *service.SessionService
	agentRepo       repository.AgentRepository
	providerRepo    repository.ProviderRepository
	sessionRepo     repository.SessionRepository
	skillRepo       repository.SkillRepository
	encryptor       *crypto.Encryptor
	config          config.APIConfig
	cronDeps        *tool.CronToolDeps
	mcpManager      *mcp.Manager
	messageBus      *channel.MessageBus
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(
	agentService *service.AgentService,
	sessionService *service.SessionService,
	agentRepo repository.AgentRepository,
	providerRepo repository.ProviderRepository,
	sessionRepo repository.SessionRepository,
	skillRepo repository.SkillRepository,
	encryptor *crypto.Encryptor,
	config config.APIConfig,
	cronDeps *tool.CronToolDeps,
	mcpManager *mcp.Manager,
	messageBus *channel.MessageBus,
) *ChatHandler {
	return &ChatHandler{
		agentService:    agentService,
		sessionService:  sessionService,
		agentRepo:       agentRepo,
		providerRepo:    providerRepo,
		sessionRepo:     sessionRepo,
		skillRepo:       skillRepo,
		encryptor:       encryptor,
		config:          config,
		cronDeps:        cronDeps,
		mcpManager:      mcpManager,
		messageBus:      messageBus,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	AgentID   string  `json:"agent_id" binding:"required"`
	Message   string  `json:"message" binding:"required"`
	SessionID *string `json:"session_id,omitempty"`
}

// Chat 流式聊天
// POST /api/v1/chat
func (h *ChatHandler) Chat(c *gin.Context) {
	// 获取用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "未登录")
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取智能体
	agentDef, err := h.agentRepo.GetByID(req.AgentID)
	if err != nil {
		NotFound(c, "智能体不存在")
		return
	}

	// 检查智能体状态
	if agentDef.Status == types.AgentStatusInactive {
		BadRequest(c, "智能体已禁用")
		return
	}

	// 检查是否为外部智能体
	if agentDef.External {
		BadRequest(c, "外部智能体暂不支持前端聊天")
		return
	}

	// 获取供应商
	providerDef, err := h.providerRepo.GetByID(agentDef.ProviderID)
	if err != nil {
		InternalServerError(c, "获取供应商失败")
		return
	}
	// 解密 API Key
	providerDef.APIKey, _ = h.encryptor.Decrypt(providerDef.APIKey)

	// 生成或使用会话 ID
	sessionID := ""
	if req.SessionID != nil && *req.SessionID != "" {
		sessionID = *req.SessionID
	}

	// 生成会话 Key: chat:{agentID}:web:{sessionID}:{userID}
	sessionKey := session.GetApiSessionKey(types.SessionTypeChat, agentDef.ID, channel.ChannelWeb, sessionID, userID.(string))

	// 构建 IM 上下文
	imContext := &types.IMSessionContext{
		UserID: userID.(string),
	}

	// 创建或加载会话
	sess, err := session.NewApiSession(
		h.sessionRepo,
		sessionKey,
		agentDef.ID,
		channel.ChannelWeb,
		types.SessionTypeChat,
		nil,
		imContext,
		agentDef.ContextMaxTokens,
		agentDef.CompressionRatio,
		agentDef.CompressionKeepSize,
	)
	if err != nil {
		InternalServerError(c, "创建会话失败")
		return
	}

	// 返回会话 ID 给前端（用于后续请求）
	if sessionID == "" {
		sessionID = sess.GetSessionID()
	}

	// 创建 SSE 输出钩子
	sseHook, err := react.NewSSEOutputHook(c.Writer)
	if err != nil {
		InternalServerError(c, "不支持流式输出")
		return
	}

	// 计算会话级工作目录
	sessionWorkDir, err := session.EnsureSessionWorkDir(agentDef.WorkingDir, sessionKey)
	if err != nil {
		slog.Error("create session working dir failed", "error", err)
		InternalServerError(c, "创建工作目录失败")
		return
	}

	// 创建 Agent Runner
	runner, err := react.NewApiAgentRunner(react.NewReactAgentParam{
		AgentDef:         agentDef,
		ProviderDef:      providerDef,
		Bus:              h.messageBus,
		Config:           h.config,
		SessionRepo:      h.sessionRepo,
		SkillRepo:        h.skillRepo,
		CronDeps:         h.cronDeps,
		AgentRepo:        h.agentRepo,
		ProviderRepo:     h.providerRepo,
		Encryptor:        h.encryptor,
		McpManager:       h.mcpManager,
		AllowSubAgents:   true,
		ParentWorkingDir: sessionWorkDir,
	})
	if err != nil {
		InternalServerError(c, "创建智能体运行器失败")
		return
	}

	// 构建输入消息
	inboundMsg := channel.InboundMessage{
		Channel:     channel.ChannelWeb,
		UserID:      userID.(string),
		Content:     req.Message,
		SessionType: types.SessionTypeChat,
	}

	// 发送会话 ID 给前端
	sseHook.SendEvent("session_id", sessionID)

	// 在 goroutine 中运行智能体循环
	go func() {
		runner.AgentLoop(c.Request.Context(), sess, inboundMsg, sseHook)
	}()

	// 等待完成或客户端断开
	<-c.Request.Context().Done()
	sseHook.Close()
}

// ChatSessionsRequest 获取会话列表请求
type ChatSessionsRequest struct {
	AgentID string `json:"agent_id" binding:"required"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

// GetSessions 获取前端聊天会话列表
// POST /api/v1/chat/sessions
func (h *ChatHandler) GetSessions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "未登录")
		return
	}

	var req ChatSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 || size > 100 {
		size = 10
	}

	// 查询该智能体该用户的前端聊天会话
	sessions, total, err := h.sessionRepo.GetWebChatSessions(req.AgentID, userID.(string), page, size)
	if err != nil {
		InternalServerError(c, "查询会话失败")
		return
	}

	// 转换为响应格式
	items := make([]service.SessionListItem, 0, len(sessions))
	for _, sess := range sessions {
		item := service.SessionListItem{
			ID:          sess.ID,
			Key:         sess.Key,
			Summary:     sess.Summary,
			SessionType: sess.SessionType,
			CreatedAt:   sess.CreatedAt,
			UpdatedAt:   sess.UpdatedAt,
		}
		items = append(items, item)
	}

	SuccessWithPage(c, items, total, page, size)
}

// ChatMessagesRequest 获取会话消息请求
type ChatMessagesRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Limit     int    `json:"limit"`
	BeforeID  string `json:"before_id,omitempty"`
}

// GetMessages 获取会话消息
// POST /api/v1/chat/messages
func (h *ChatHandler) GetMessages(c *gin.Context) {
	var req ChatMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 查询消息
	messages, total, err := h.sessionRepo.GetMessagesPaginated(req.SessionID, req.BeforeID, limit)
	if err != nil {
		InternalServerError(c, "查询消息失败")
		return
	}

	// 转换为响应格式
	items := make([]service.MessageListItem, 0, len(messages))
	for _, msg := range messages {
		item := service.MessageListItem{
			ID:           msg.ID,
			Role:         msg.Role,
			Content:      msg.Content,
			ToolCallID:   msg.ToolCallID,
			InputTokens:  msg.InputTokens,
			OutputTokens: msg.OutputTokens,
			TotalTokens:  msg.TotalTokens,
			InWindow:     msg.InWindow,
			CreatedAt:    msg.CreatedAt,
		}
		// 解析 ToolCalls
		if msg.ToolCalls != "" {
			var toolCalls []types.ToolCall
			if err := parseJSON(msg.ToolCalls, &toolCalls); err == nil {
				item.ToolCalls = toolCalls
			}
		}
		items = append(items, item)
	}

	hasMore := int64(len(messages)) < total
	Success(c, &service.MessagesResponse{
		Messages: items,
		Total:    total,
		HasMore:  hasMore,
	})
}

// parseJSON 解析 JSON 字符串
func parseJSON(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}