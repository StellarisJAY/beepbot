package react

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSEEvent SSE 事件结构
type SSEEvent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// SSEOutputHook SSE 输出钩子，用于流式输出到 HTTP 响应
type SSEOutputHook struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	mutex   sync.Mutex
	done    bool
}

// NewSSEOutputHook 创建 SSE 输出钩子
func NewSSEOutputHook(w http.ResponseWriter) (*SSEOutputHook, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming unsupported")
	}

	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	return &SSEOutputHook{
		writer:  w,
		flusher: flusher,
	}, nil
}

// sendEvent 发送 SSE 事件
func (h *SSEOutputHook) sendEvent(eventType string, content string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.done {
		return
	}

	event := SSEEvent{
		Type:    eventType,
		Content: content,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	fmt.Fprintf(h.writer, "data: %s\n\n", data)
	h.flusher.Flush()
}

// OnError 调用模型失败时调用
func (h *SSEOutputHook) OnError(ctx context.Context, err error) {
	h.sendEvent("error", err.Error())
	h.Close()
}

// OnResponse 有最终响应内容时调用
func (h *SSEOutputHook) OnResponse(ctx context.Context, content string) {
	h.sendEvent("message", content)
	h.sendEvent("done", "")
	h.Close()
}

// OnIntermediateContent 有中间内容时调用（流式输出的内容）
func (h *SSEOutputHook) OnIntermediateContent(ctx context.Context, content string) {
	// 流式输出的内容发送为 message 事件，前端会累积显示
	h.sendEvent("message", content)
}

// OnToolCall 工具调用时调用
func (h *SSEOutputHook) OnToolCall(ctx context.Context, toolName string, args string) {
	h.sendEvent("tool_call", fmt.Sprintf("%s(%s)", toolName, truncateToolArgs(args, 100)))
}

// OnToolResult 工具执行结果
func (h *SSEOutputHook) OnToolResult(ctx context.Context, toolName string, result string, err error) {
	if err != nil {
		h.sendEvent("tool_error", fmt.Sprintf("%s: %s", toolName, err.Error()))
	} else {
		h.sendEvent("tool_result", fmt.Sprintf("%s: %s", toolName, truncateToolArgs(result, 200)))
	}
}

// Close 关闭 SSE 流
func (h *SSEOutputHook) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.done = true
}

// IsDone 返回是否已完成
func (h *SSEOutputHook) IsDone() bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.done
}

// SendEvent 公开方法，允许外部发送 SSE 事件
func (h *SSEOutputHook) SendEvent(eventType string, content string) {
	h.sendEvent(eventType, content)
}

// truncateToolArgs 截断工具参数/结果，用于展示
func truncateToolArgs(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}