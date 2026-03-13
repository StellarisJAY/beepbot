package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"sync"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/StellarisJAY/beepbot/internal/types"
)

// headerTransport 是一个 http.RoundTripper 包装器，用于在请求中添加自定义 headers
type headerTransport struct {
	transport http.RoundTripper
	headers   map[string]string
}

// RoundTrip 实现 http.RoundTripper 接口
func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}
	return t.transport.RoundTrip(req)
}

// newHTTPClientWithHeaders 创建带有自定义 headers 的 HTTP client
func newHTTPClientWithHeaders(headers map[string]string) *http.Client {
	transport := &headerTransport{
		transport: http.DefaultTransport,
		headers:   headers,
	}
	return &http.Client{
		Transport: transport,
	}
}

// Client 封装官方 SDK 的 MCP 客户端
type Client struct {
	client  *mcp.Client
	session *mcp.ClientSession
	server  *types.MCPServer
	tools   []*mcp.Tool
	mu      sync.RWMutex
}

// NewClient 创建 MCP 客户端
func NewClient(server *types.MCPServer) *Client {
	impl := &mcp.Implementation{
		Name:    "beepbot",
		Version: "1.0.0",
	}
	return &Client{
		client: mcp.NewClient(impl, nil),
		server: server,
		tools:  make([]*mcp.Tool, 0),
	}
}

// Initialize 初始化 MCP 连接
func (c *Client) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session != nil {
		return nil
	}

	var transport mcp.Transport

	// 解析 headers
	var headers map[string]string
	if c.server.Headers != nil {
		_ = json.Unmarshal(c.server.Headers, &headers)
	}

	switch c.server.TransportType {
	case types.MCPTransportHTTP:
		// 使用 StreamableClientTransport (2025-03-26 规范)
		httpTransport := &mcp.StreamableClientTransport{
			Endpoint: c.server.URL,
		}
		// 如果有自定义 headers，创建自定义 HTTP client
		if len(headers) > 0 {
			httpTransport.HTTPClient = newHTTPClientWithHeaders(headers)
		}
		transport = httpTransport
	case types.MCPTransportSSE:
		// 使用 SSEClientTransport (2024-11-05 规范)
		sseTransport := &mcp.SSEClientTransport{
			Endpoint: c.server.URL,
		}
		// 如果有自定义 headers，创建自定义 HTTP client
		if len(headers) > 0 {
			sseTransport.HTTPClient = newHTTPClientWithHeaders(headers)
		}
		transport = sseTransport
	case types.MCPTransportStdio:
		// 使用 CommandTransport
		var args []string
		if c.server.Args != nil {
			_ = json.Unmarshal(c.server.Args, &args)
		}
		transport = &mcp.CommandTransport{
			Command: exec.Command(c.server.Command, args...),
		}
	default:
		// 默认使用 Streamable HTTP
		httpTransport := &mcp.StreamableClientTransport{
			Endpoint: c.server.URL,
		}
		// 如果有自定义 headers，创建自定义 HTTP client
		if len(headers) > 0 {
			httpTransport.HTTPClient = newHTTPClientWithHeaders(headers)
		}
		transport = httpTransport
	}

	session, err := c.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	c.session = session

	// 获取工具列表
	if err := c.refreshToolsLocked(ctx); err != nil {
		slog.Warn("failed to refresh tools during initialization", "server", c.server.Name, "error", err)
	}

	slog.Info("MCP client initialized",
		"server", c.server.Name,
		"sessionID", session.ID())

	return nil
}

// refreshToolsLocked 刷新工具列表（需要持有锁）
func (c *Client) refreshToolsLocked(ctx context.Context) error {
	var tools []*mcp.Tool
	for tool, err := range c.session.Tools(ctx, nil) {
		if err != nil {
			return fmt.Errorf("list tools failed: %w", err)
		}
		tools = append(tools, tool)
	}
	c.tools = tools
	slog.Info("MCP tools refreshed", "server", c.server.Name, "count", len(c.tools))
	return nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session != nil {
		err := c.session.Close()
		c.session = nil
		c.tools = nil
		return err
	}
	return nil
}

// ListTools 获取工具列表
func (c *Client) ListTools() []types.MCPToolDefinition {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]types.MCPToolDefinition, len(c.tools))
	for i, t := range c.tools {
		// InputSchema 可能是 json.RawMessage 或 map[string]any
		var inputSchema map[string]any
		switch v := t.InputSchema.(type) {
		case map[string]any:
			inputSchema = v
		case json.RawMessage:
			_ = json.Unmarshal(v, &inputSchema)
		default:
			// 尝试 JSON 序列化再反序列化
			data, _ := json.Marshal(v)
			_ = json.Unmarshal(data, &inputSchema)
		}

		result[i] = types.MCPToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: inputSchema,
		}
	}
	return result
}

// RefreshTools 刷新工具列表
func (c *Client) RefreshTools(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session == nil {
		return fmt.Errorf("client not initialized")
	}
	return c.refreshToolsLocked(ctx)
}

// CallTool 调用工具
func (c *Client) CallTool(ctx context.Context, toolName string, args map[string]any) (*types.MCPToolResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.session == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	result, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	if err != nil {
		return nil, fmt.Errorf("call tool failed: %w", err)
	}

	// 转换结果
	content := make([]types.MCPContentBlock, len(result.Content))
	for i, cont := range result.Content {
		switch v := cont.(type) {
		case *mcp.TextContent:
			content[i] = types.MCPContentBlock{
				Type: "text",
				Text: v.Text,
			}
		case *mcp.ImageContent:
			// 图片内容
			content[i] = types.MCPContentBlock{
				Type: "image",
				Text: string(v.Data), // Base64 编码的图片数据
			}
		default:
			// 其他类型转为 JSON
			data, _ := json.Marshal(cont)
			content[i] = types.MCPContentBlock{
				Type: "unknown",
				Text: string(data),
			}
		}
	}

	return &types.MCPToolResult{
		Content: content,
		IsError: result.IsError,
	}, nil
}

// IsConnected 检查是否已连接
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.session != nil
}

// GetServerName 获取服务器名称
func (c *Client) GetServerName() string {
	return c.server.Name
}

// GetServerID 获取服务器 ID
func (c *Client) GetServerID() string {
	return c.server.ID
}

// ToolToDefinition 将 MCP 工具定义转换为工具定义格式
func ToolToDefinition(serverName string, tool types.MCPToolDefinition) types.ToolDefinition {
	return types.ToolDefinition{
		Type: "function",
		Function: types.ToolFunctionDefinition{
			Name:        fmt.Sprintf("mcp::%s::%s", serverName, tool.Name),
			Description: tool.Description,
			Parameters:  tool.InputSchema,
		},
	}
}

// ParseToolName 解析 MCP 工具名称，返回服务器名称和工具名称
// 格式: mcp::{server_name}::{tool_name}
func ParseToolName(fullName string) (serverName, toolName string, ok bool) {
	// 格式: mcp::{server_name}::{tool_name}
	const prefix = "mcp::"
	if len(fullName) < len(prefix) || fullName[:len(prefix)] != prefix {
		return "", "", false
	}

	rest := fullName[len(prefix):]
	// 找到 :: 分隔符
	idx := -1
	for i := 0; i < len(rest)-1; i++ {
		if rest[i] == ':' && rest[i+1] == ':' {
			idx = i
			break
		}
	}

	if idx == -1 {
		return "", "", false
	}

	return rest[:idx], rest[idx+2:], true
}