package api

import (
	"strconv"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/gin-gonic/gin"
)

// MCPHandler MCP 处理器
type MCPHandler struct {
	service *service.MCPService
}

// NewMCPHandler 创建 MCP 处理器
func NewMCPHandler(service *service.MCPService) *MCPHandler {
	return &MCPHandler{service: service}
}

// ListMCPServers 获取 MCP 服务器列表
// GET /api/v1/mcp
func (h *MCPHandler) ListMCPServers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询参数
	query := &types.MCPServerQuery{
		Name:   c.Query("name"),
		Status: types.MCPServerStatus(c.Query("status")),
	}

	servers, total, err := h.service.ListServers(page, pageSize, query)
	if err != nil {
		InternalError(c, "failed to list MCP servers: "+err.Error())
		return
	}

	SuccessWithPage(c, servers, total, page, pageSize)
}

// GetMCPServer 获取单个 MCP 服务器
// GET /api/v1/mcp/:id
func (h *MCPHandler) GetMCPServer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "server id is required")
		return
	}

	server, err := h.service.GetServer(id)
	if err != nil {
		NotFound(c, "MCP server not found")
		return
	}

	Success(c, server)
}

// CreateMCPServer 创建 MCP 服务器
// POST /api/v1/mcp
func (h *MCPHandler) CreateMCPServer(c *gin.Context) {
	var req service.CreateMCPServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 验证传输类型
	if req.TransportType != "" && req.TransportType != "sse" && req.TransportType != "stdio" {
		BadRequest(c, "invalid transport type, must be 'sse' or 'stdio'")
		return
	}

	// SSE 传输需要 URL
	if req.TransportType == "" || req.TransportType == "sse" {
		if req.URL == "" {
			BadRequest(c, "URL is required for SSE transport")
			return
		}
	}

	// Stdio 传输需要 Command
	if req.TransportType == "stdio" {
		if req.Command == "" {
			BadRequest(c, "Command is required for Stdio transport")
			return
		}
	}

	server, err := h.service.CreateServer(&req)
	if err != nil {
		Error(c, 500, "failed to create MCP server: "+err.Error())
		return
	}

	Success(c, server)
}

// UpdateMCPServer 更新 MCP 服务器
// PUT /api/v1/mcp/:id
func (h *MCPHandler) UpdateMCPServer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "server id is required")
		return
	}

	var req service.UpdateMCPServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	server, err := h.service.UpdateServer(id, &req)
	if err != nil {
		Error(c, 500, "failed to update MCP server: "+err.Error())
		return
	}

	Success(c, server)
}

// DeleteMCPServer 删除 MCP 服务器
// DELETE /api/v1/mcp/:id
func (h *MCPHandler) DeleteMCPServer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "server id is required")
		return
	}

	if err := h.service.DeleteServer(id); err != nil {
		Error(c, 500, "failed to delete MCP server: "+err.Error())
		return
	}

	Success(c, nil)
}

// StartMCPServer 启动 MCP 服务器连接
// PUT /api/v1/mcp/:id/start
func (h *MCPHandler) StartMCPServer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "server id is required")
		return
	}

	if err := h.service.StartServer(id); err != nil {
		Error(c, 500, "failed to start MCP server: "+err.Error())
		return
	}

	Success(c, nil)
}

// StopMCPServer 停止 MCP 服务器连接
// PUT /api/v1/mcp/:id/stop
func (h *MCPHandler) StopMCPServer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "server id is required")
		return
	}

	if err := h.service.StopServer(id); err != nil {
		Error(c, 500, "failed to stop MCP server: "+err.Error())
		return
	}

	Success(c, nil)
}

// TestMCPConnection 测试 MCP 服务器连接
// POST /api/v1/mcp/:id/test
func (h *MCPHandler) TestMCPConnection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "server id is required")
		return
	}

	if err := h.service.TestConnection(id); err != nil {
		Error(c, 500, "connection test failed: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "connection test successful"})
}

// GetMCPServerTools 获取 MCP 服务器提供的工具列表
// GET /api/v1/mcp/:id/tools
func (h *MCPHandler) GetMCPServerTools(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "server id is required")
		return
	}

	tools, err := h.service.GetServerTools(id)
	if err != nil {
		Error(c, 500, "failed to get server tools: "+err.Error())
		return
	}

	Success(c, tools)
}

// ReconnectMCPServer 重连 MCP 服务器
// POST /api/v1/mcp/:id/reconnect
func (h *MCPHandler) ReconnectMCPServer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "server id is required")
		return
	}

	if err := h.service.ReconnectServer(id); err != nil {
		Error(c, 500, "failed to reconnect MCP server: "+err.Error())
		return
	}

	Success(c, nil)
}