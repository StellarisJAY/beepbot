package tool

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

type ToolRegistry struct {
	tools map[string]Tool
	mutex sync.RWMutex
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
		mutex: sync.RWMutex{},
	}
}

func (r *ToolRegistry) Register(tool Tool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.tools[tool.Name()] = tool
}

func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, params map[string]any) (string, error) {
	return r.ExecuteWithContext(ctx, name, params, "", "")
}

// ExecuteWithContext 带有上下文信息的工具执行，上下文中包含发送消息的用户信息
func (r *ToolRegistry) ExecuteWithContext(ctx context.Context, name string, params map[string]any, channel string, userID string) (string, error) {
	r.mutex.RLock()
	tool, ok := r.tools[name]
	r.mutex.RUnlock()
	if !ok {
		slog.Error("execute on not exist tools", "tool", name)
		return "", errors.New("tool not found")
	}
	start := time.Now()
	// 将channel和用户信息加入到工具上下文
	ctx = context.WithValue(ctx, "channel", channel)
	ctx = context.WithValue(ctx, "userID", userID)
	result, err := tool.Execute(ctx, params)
	duration := time.Since(start)
	if err != nil {
		slog.Info("execute tool error", "duration", duration, "error", err)
	} else {
		slog.Info("execute tool success", "tool", name, "duration", duration, "result", result)
	}
	return result, err
}

func (r *ToolRegistry) GetDefinitions() []map[string]any {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	definitions := make([]map[string]any, 0, len(r.tools))
	for _, tool := range r.tools {
		definitions = append(definitions, ToolToDefinition(tool))
	}
	return definitions
}

func (r *ToolRegistry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.tools)
}
