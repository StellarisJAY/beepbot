package types

import (
	"time"

	"gorm.io/datatypes"
)

// MCPTransportType MCP 传输类型
type MCPTransportType string

const (
	MCPTransportSSE   MCPTransportType = "sse"   // SSE 传输 (2024-11-05 规范)
	MCPTransportHTTP  MCPTransportType = "http"  // Streamable HTTP 传输 (2025-03-26 规范)
	MCPTransportStdio MCPTransportType = "stdio" // Stdio 传输
)

// MCPServerStatus MCP 服务器状态
type MCPServerStatus string

const (
	MCPServerStatusActive   MCPServerStatus = "active"   // 活跃
	MCPServerStatusInactive MCPServerStatus = "inactive" // 未激活
)

// MCPServer MCP 服务器配置
type MCPServer struct {
	ID           string           `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name         string           `json:"name" gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	Description  string           `json:"description" gorm:"column:description;type:text"`
	TransportType MCPTransportType `json:"transport_type" gorm:"column:transport_type;type:varchar(16);not null;default:'sse'"`
	URL          string           `json:"url" gorm:"column:url;type:varchar(512)"`           // SSE 传输的 URL
	Command      string           `json:"command" gorm:"column:command;type:text"`           // Stdio 传输的启动命令
	Args         datatypes.JSON   `json:"args" gorm:"column:args;type:jsonb"`                // Stdio 传输的参数列表
	Env          datatypes.JSON   `json:"env" gorm:"column:env;type:jsonb"`                  // 环境变量
	Headers      datatypes.JSON   `json:"headers" gorm:"column:headers;type:jsonb"`          // HTTP 请求头（用于认证）
	Status       MCPServerStatus  `json:"status" gorm:"column:status;type:varchar(16);not null;default:'inactive'"`
	CreatedAt    time.Time        `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt    time.Time        `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (MCPServer) TableName() string {
	return "mcp_servers"
}

// MCPServerQuery MCP 服务器查询参数
type MCPServerQuery struct {
	Name   string           // 名称模糊搜索
	Status MCPServerStatus  // 状态筛选
}

// MCPToolDefinition MCP 工具定义（从 MCP 服务器获取）
type MCPToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// MCPToolResult MCP 工具调用结果
type MCPToolResult struct {
	Content []MCPContentBlock `json:"content"`
	IsError bool              `json:"isError"`
}

// MCPContentBlock MCP 内容块
type MCPContentBlock struct {
	Type string `json:"type"` // "text", "image", "resource"
	Text string `json:"text,omitempty"`
}