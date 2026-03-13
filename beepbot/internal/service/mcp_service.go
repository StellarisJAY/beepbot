package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/StellarisJAY/beepbot/internal/mcp"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/datatypes"
)

// MCPService MCP 服务
type MCPService struct {
	repo    repository.MCPServerRepository
	manager *mcp.Manager
}

// NewMCPService 创建 MCP 服务
func NewMCPService(repo repository.MCPServerRepository, manager *mcp.Manager) *MCPService {
	return &MCPService{
		repo:    repo,
		manager: manager,
	}
}

// CreateMCPServerRequest 创建 MCP 服务器请求
type CreateMCPServerRequest struct {
	Name          string            `json:"name" binding:"required"`
	Description   string            `json:"description"`
	TransportType string            `json:"transport_type"` // "sse" or "stdio"
	URL           string            `json:"url"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	Env           map[string]string `json:"env"`
	Headers       map[string]string `json:"headers"` // HTTP 请求头（用于认证）
}

// UpdateMCPServerRequest 更新 MCP 服务器请求
type UpdateMCPServerRequest struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	TransportType string            `json:"transport_type"`
	URL           string            `json:"url"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	Env           map[string]string `json:"env"`
	Headers       map[string]string `json:"headers"`
}

// MCPServerResponse MCP 服务器响应
type MCPServerResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	TransportType types.MCPTransportType `json:"transport_type"`
	URL           string                 `json:"url"`
	Command       string                 `json:"command"`
	Args          []string               `json:"args,omitempty"`
	Env           map[string]string      `json:"env,omitempty"`
	Headers       map[string]string      `json:"headers,omitempty"`
	Status        types.MCPServerStatus  `json:"status"`
	Tools         []MCPToolInfo          `json:"tools,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// MCPToolInfo MCP 工具信息
type MCPToolInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// CreateServer 创建 MCP 服务器
func (s *MCPService) CreateServer(req *CreateMCPServerRequest) (*types.MCPServer, error) {
	// 检查名称是否已存在
	if _, err := s.repo.GetByName(req.Name); err == nil {
		return nil, errors.New("server name already exists")
	}

	// 设置默认传输类型
	transportType := types.MCPTransportSSE
	if req.TransportType == "stdio" {
		transportType = types.MCPTransportStdio
	}

	// 序列化 Args
	var argsJSON datatypes.JSON
	if req.Args != nil {
		data, _ := json.Marshal(req.Args)
		argsJSON = data
	}

	// 序列化 Env
	var envJSON datatypes.JSON
	if req.Env != nil {
		data, _ := json.Marshal(req.Env)
		envJSON = data
	}

	// 序列化 Headers
	var headersJSON datatypes.JSON
	if req.Headers != nil {
		data, _ := json.Marshal(req.Headers)
		headersJSON = data
	}

	server := &types.MCPServer{
		Name:          req.Name,
		Description:   req.Description,
		TransportType: transportType,
		URL:           req.URL,
		Command:       req.Command,
		Args:          argsJSON,
		Env:           envJSON,
		Headers:       headersJSON,
		Status:        types.MCPServerStatusInactive,
	}

	// 设置 ID 和时间
	server.ID = types.GenerateUUIDv7()
	now := time.Now()
	server.CreatedAt = now
	server.UpdatedAt = now

	if err := s.repo.Create(server); err != nil {
		return nil, err
	}

	return server, nil
}

// UpdateServer 更新 MCP 服务器
func (s *MCPService) UpdateServer(id string, req *UpdateMCPServerRequest) (*types.MCPServer, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 检查是否需要重启
	needRestart := false
	if s.manager.IsServerConnected(id) {
		// 如果修改了连接相关的配置，需要重启
		if req.URL != "" && req.URL != server.URL {
			needRestart = true
		}
		if req.TransportType != "" && types.MCPTransportType(req.TransportType) != server.TransportType {
			needRestart = true
		}
		if req.Headers != nil {
			needRestart = true
		}
	}

	// 更新字段
	if req.Name != "" {
		// 检查名称是否被其他服务器使用
		if existing, err := s.repo.GetByName(req.Name); err == nil && existing.ID != id {
			return nil, errors.New("server name already exists")
		}
		server.Name = req.Name
	}

	if req.Description != "" {
		server.Description = req.Description
	}

	if req.TransportType != "" {
		switch req.TransportType {
		case "stdio":
			server.TransportType = types.MCPTransportStdio
		case "sse":
			server.TransportType = types.MCPTransportSSE
		default:
			server.TransportType = types.MCPTransportHTTP
		}
	}

	if req.URL != "" {
		server.URL = req.URL
	}

	if req.Command != "" {
		server.Command = req.Command
	}

	if req.Args != nil {
		data, _ := json.Marshal(req.Args)
		server.Args = data
	}

	if req.Env != nil {
		data, _ := json.Marshal(req.Env)
		server.Env = data
	}

	if req.Headers != nil {
		data, _ := json.Marshal(req.Headers)
		server.Headers = data
	}

	server.UpdatedAt = time.Now()

	if err := s.repo.Update(server); err != nil {
		return nil, err
	}

	// 如果需要重启，先停止再启动
	if needRestart {
		_ = s.manager.StopServer(id)
	}

	return server, nil
}

// DeleteServer 删除 MCP 服务器
func (s *MCPService) DeleteServer(id string) error {
	// 先停止服务器
	_ = s.manager.StopServer(id)
	return s.repo.Delete(id)
}

// GetServer 获取 MCP 服务器
func (s *MCPService) GetServer(id string) (*MCPServerResponse, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(server), nil
}

// ListServers 列出 MCP 服务器
func (s *MCPService) ListServers(page, pageSize int, query *types.MCPServerQuery) ([]MCPServerResponse, int64, error) {
	servers, total, err := s.repo.ListWithQuery(page, pageSize, query)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]MCPServerResponse, len(servers))
	for i, server := range servers {
		responses[i] = *s.toResponse(&server)
	}

	return responses, total, nil
}

// StartServer 启动 MCP 服务器
func (s *MCPService) StartServer(id string) error {
	// 检查服务器是否存在
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 检查是否已连接
	if s.manager.IsServerConnected(id) {
		return nil
	}

	// 启动服务器
	return s.manager.StartServer(context.Background(), id)
}

// StopServer 停止 MCP 服务器
func (s *MCPService) StopServer(id string) error {
	return s.manager.StopServer(id)
}

// TestConnection 测试 MCP 服务器连接
func (s *MCPService) TestConnection(id string) error {
	// 获取服务器配置
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 创建临时客户端进行测试
	client := mcp.NewClient(server)
	defer client.Close()

	// 尝试初始化连接
	return client.Initialize(context.Background())
}

// GetServerTools 获取服务器提供的工具列表
func (s *MCPService) GetServerTools(id string) ([]MCPToolInfo, error) {
	tools, err := s.manager.GetServerTools(id)
	if err != nil {
		return nil, err
	}

	result := make([]MCPToolInfo, len(tools))
	for i, tool := range tools {
		result[i] = MCPToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		}
	}

	return result, nil
}

// GetManager 获取 MCP 管理器
func (s *MCPService) GetManager() *mcp.Manager {
	return s.manager
}

// toResponse 转换为响应格式
func (s *MCPService) toResponse(server *types.MCPServer) *MCPServerResponse {
	var args []string
	if server.Args != nil {
		_ = json.Unmarshal(server.Args, &args)
	}

	var env map[string]string
	if server.Env != nil {
		_ = json.Unmarshal(server.Env, &env)
	}

	var headers map[string]string
	if server.Headers != nil {
		_ = json.Unmarshal(server.Headers, &headers)
	}

	response := &MCPServerResponse{
		ID:            server.ID,
		Name:          server.Name,
		Description:   server.Description,
		TransportType: server.TransportType,
		URL:           server.URL,
		Command:       server.Command,
		Args:          args,
		Env:           env,
		Headers:       headers,
		Status:        server.Status,
		CreatedAt:     server.CreatedAt,
		UpdatedAt:     server.UpdatedAt,
	}

	// 如果服务器已连接，获取工具列表
	if s.manager.IsServerConnected(server.ID) {
		tools, _ := s.manager.GetServerTools(server.ID)
		response.Tools = make([]MCPToolInfo, len(tools))
		for i, tool := range tools {
			response.Tools[i] = MCPToolInfo{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.InputSchema,
			}
		}
	}

	return response
}

// StartAllActiveServers 启动所有活跃的服务器
func (s *MCPService) StartAllActiveServers() error {
	return s.manager.StartAllServers(context.Background())
}

// StopAllServers 停止所有服务器
func (s *MCPService) StopAllServers() {
	s.manager.StopAllServers()
}

// GetAllMCPTools 获取所有 MCP 工具定义
func (s *MCPService) GetAllMCPTools() []types.ToolDefinition {
	return s.manager.GetAllTools()
}

// ReconnectServer 重连服务器
func (s *MCPService) ReconnectServer(id string) error {
	return s.manager.Reconnect(context.Background(), id)
}
