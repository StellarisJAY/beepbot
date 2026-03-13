package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/types"
)

// ToolProxy MCP 工具代理，实现 tool.Tool 接口
type ToolProxy struct {
	manager    *Manager
	serverID   string
	serverName string
	toolName   string
	toolDef    types.MCPToolDefinition
}

// NewToolProxy 创建 MCP 工具代理
func NewToolProxy(manager *Manager, serverID, serverName string, toolDef types.MCPToolDefinition) *ToolProxy {
	return &ToolProxy{
		manager:    manager,
		serverID:   serverID,
		serverName: serverName,
		toolName:   toolDef.Name,
		toolDef:    toolDef,
	}
}

// Name 返回工具名称（带前缀）
func (p *ToolProxy) Name() string {
	return fmt.Sprintf("mcp_%s_%s", p.serverName, p.toolName)
}

// Description 返回工具描述
func (p *ToolProxy) Description() string {
	return p.toolDef.Description
}

// Parameters 返回工具参数定义
func (p *ToolProxy) Parameters() map[string]any {
	return p.toolDef.InputSchema
}

// Execute 执行工具调用
func (p *ToolProxy) Execute(ctx context.Context, params map[string]any) (string, error) {
	result, err := p.manager.CallTool(ctx, p.serverID, p.toolName, params)
	if err != nil {
		return "", fmt.Errorf("MCP tool call failed: %w", err)
	}

	// 构建结果字符串
	var resultBuilder strings.Builder
	for _, content := range result.Content {
		switch content.Type {
		case "text":
			resultBuilder.WriteString(content.Text)
		default:
			// 其他类型转为 JSON
			data, _ := json.Marshal(content)
			resultBuilder.WriteString(string(data))
		}
	}

	output := resultBuilder.String()
	if result.IsError {
		return "", fmt.Errorf("MCP tool error: %s", output)
	}

	return output, nil
}

// ToolProxiesToDefinitions 将工具代理列表转换为工具定义列表
func ToolProxiesToDefinitions(proxies []*ToolProxy) []types.ToolDefinition {
	definitions := make([]types.ToolDefinition, len(proxies))
	for i, proxy := range proxies {
		definitions[i] = types.ToolDefinition{
			Type: "function",
			Function: types.ToolFunctionDefinition{
				Name:        proxy.Name(),
				Description: proxy.Description(),
				Parameters:  proxy.Parameters(),
			},
		}
	}
	return definitions
}