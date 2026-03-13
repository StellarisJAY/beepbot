package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// Manager MCP 管理器
type Manager struct {
	clients map[string]*Client // serverID -> client
	repo    repository.MCPServerRepository
	mu      sync.RWMutex
}

// NewManager 创建 MCP 管理器
func NewManager(repo repository.MCPServerRepository) *Manager {
	return &Manager{
		clients: make(map[string]*Client),
		repo:    repo,
	}
}

// StartServer 启动 MCP 服务器连接
func (m *Manager) StartServer(ctx context.Context, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已经启动
	if client, exists := m.clients[serverID]; exists && client.IsConnected() {
		return nil
	}

	// 获取服务器配置
	server, err := m.repo.GetByID(serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// 创建客户端
	client := NewClient(server)

	// 初始化连接
	if err := client.Initialize(ctx); err != nil {
		return fmt.Errorf("initialize client failed: %w", err)
	}

	m.clients[serverID] = client

	// 更新服务器状态
	if err := m.repo.UpdateStatus(serverID, types.MCPServerStatusActive); err != nil {
		slog.Warn("failed to update server status", "serverID", serverID, "error", err)
	}

	slog.Info("MCP server started", "server", server.Name)
	return nil
}

// StopServer 停止 MCP 服务器连接
func (m *Manager) StopServer(serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[serverID]
	if !exists {
		return nil
	}

	// 关闭客户端
	if err := client.Close(); err != nil {
		slog.Warn("failed to close client", "serverID", serverID, "error", err)
	}

	delete(m.clients, serverID)

	// 更新服务器状态
	if err := m.repo.UpdateStatus(serverID, types.MCPServerStatusInactive); err != nil {
		slog.Warn("failed to update server status", "serverID", serverID, "error", err)
	}

	slog.Info("MCP server stopped", "serverID", serverID)
	return nil
}

// StartAllServers 启动所有活跃的 MCP 服务器
func (m *Manager) StartAllServers(ctx context.Context) error {
	servers, err := m.repo.GetActiveServers()
	if err != nil {
		return fmt.Errorf("get active servers failed: %w", err)
	}

	var startErrors []error
	for _, server := range servers {
		if err := m.StartServer(ctx, server.ID); err != nil {
			startErrors = append(startErrors, fmt.Errorf("start server %s failed: %w", server.Name, err))
		}
	}

	if len(startErrors) > 0 {
		return fmt.Errorf("some servers failed to start: %v", startErrors)
	}

	return nil
}

// StopAllServers 停止所有 MCP 服务器
func (m *Manager) StopAllServers() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for serverID, client := range m.clients {
		if err := client.Close(); err != nil {
			slog.Warn("failed to close client", "serverID", serverID, "error", err)
		}
		delete(m.clients, serverID)
	}

	slog.Info("All MCP servers stopped")
}

// GetAllTools 获取所有 MCP 工具定义
func (m *Manager) GetAllTools() []types.ToolDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tools []types.ToolDefinition
	for _, client := range m.clients {
		if !client.IsConnected() {
			continue
		}
		serverName := client.GetServerName()
		for _, tool := range client.ListTools() {
			tools = append(tools, ToolToDefinition(serverName, tool))
		}
	}
	return tools
}

// GetServerTools 获取指定服务器的工具列表
func (m *Manager) GetServerTools(serverID string) ([]types.MCPToolDefinition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[serverID]
	if !exists {
		return nil, fmt.Errorf("server not connected")
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("server not connected")
	}

	return client.ListTools(), nil
}

// CallTool 调用 MCP 工具
func (m *Manager) CallTool(ctx context.Context, serverID, toolName string, args map[string]any) (*types.MCPToolResult, error) {
	m.mu.RLock()
	client, exists := m.clients[serverID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server not connected: %s", serverID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("server not connected: %s", serverID)
	}

	return client.CallTool(ctx, toolName, args)
}

// CallToolByFullName 通过完整工具名称调用工具
func (m *Manager) CallToolByFullName(ctx context.Context, fullName string, args map[string]any) (*types.MCPToolResult, error) {
	serverName, toolName, ok := ParseToolName(fullName)
	if !ok {
		return nil, fmt.Errorf("invalid MCP tool name: %s", fullName)
	}

	// 查找服务器
	m.mu.RLock()
	var targetClient *Client
	for _, client := range m.clients {
		if client.GetServerName() == serverName {
			targetClient = client
			break
		}
	}
	m.mu.RUnlock()

	if targetClient == nil {
		return nil, fmt.Errorf("MCP server not found: %s", serverName)
	}

	if !targetClient.IsConnected() {
		return nil, fmt.Errorf("MCP server not connected: %s", serverName)
	}

	return targetClient.CallTool(ctx, toolName, args)
}

// RefreshTools 刷新所有服务器的工具列表
func (m *Manager) RefreshTools(ctx context.Context) error {
	m.mu.RLock()
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.mu.RUnlock()

	var errors []error
	for _, client := range clients {
		if err := client.RefreshTools(ctx); err != nil {
			errors = append(errors, fmt.Errorf("refresh tools for server %s failed: %w", client.GetServerName(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some servers failed to refresh: %v", errors)
	}
	return nil
}

// GetClient 获取指定服务器的客户端
func (m *Manager) GetClient(serverID string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, exists := m.clients[serverID]
	return client, exists
}

// GetClientByServerName 通过服务器名称获取客户端
func (m *Manager) GetClientByServerName(serverName string) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, client := range m.clients {
		if client.GetServerName() == serverName {
			return client
		}
	}
	return nil
}

// IsServerConnected 检查服务器是否已连接
func (m *Manager) IsServerConnected(serverID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, exists := m.clients[serverID]
	if !exists {
		return false
	}
	return client.IsConnected()
}

// ListConnectedServers 列出已连接的服务器
func (m *Manager) ListConnectedServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var servers []string
	for serverID, client := range m.clients {
		if client.IsConnected() {
			servers = append(servers, serverID)
		}
	}
	return servers
}

// Reconnect 重连服务器
func (m *Manager) Reconnect(ctx context.Context, serverID string) error {
	// 先停止
	_ = m.StopServer(serverID)
	// 再启动
	return m.StartServer(ctx, serverID)
}
