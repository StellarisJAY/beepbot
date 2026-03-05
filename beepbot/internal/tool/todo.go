package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const TodoFileName = "TODO.md"

// TodoItem 表示一个待办事项
type TodoItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`   // "pending", "in_progress", "completed"
	Priority    string    `json:"priority"` // "low", "medium", "high"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TodoList 表示待办事项列表
type TodoList struct {
	Items []TodoItem `json:"items"`
}

// WriteTodoTool 管理待办事项的工具
type WriteTodoTool struct {
	workingDir string
	mu         sync.Mutex
}

// NewWriteTodoTool 创建一个新的 TODO 工具
func NewWriteTodoTool(workingDir string) Tool {
	absDir, _ := filepath.Abs(workingDir)
	return &WriteTodoTool{workingDir: absDir}
}

// Name 实现 Tool 接口
func (w *WriteTodoTool) Name() string {
	return "todo"
}

// Description 实现 Tool 接口
func (w *WriteTodoTool) Description() string {
	return `Manage a todo list to track task progress. 
Supports operations: add, list, update, delete, clear.
Use this tool to keep track of tasks and their completion status.`
}

// Parameters 实现 Tool 接口
func (w *WriteTodoTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"operation": map[string]any{
				"type":        "string",
				"enum":        []string{"add", "list", "update", "delete", "clear"},
				"description": "The operation to perform",
			},
			"id": map[string]any{
				"type":        "string",
				"description": "Task ID (required for update and delete operations)",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Task description (required for add operation)",
			},
			"status": map[string]any{
				"type":        "string",
				"enum":        []string{"pending", "in_progress", "completed"},
				"description": "Task status (for update operation)",
			},
			"priority": map[string]any{
				"type":        "string",
				"enum":        []string{"low", "medium", "high"},
				"description": "Task priority (optional for add operation, default: medium)",
			},
		},
		"required": []string{"operation"},
	}
}

// Execute 实现 Tool 接口
func (w *WriteTodoTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	operation, ok := params["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// 加载现有的 TODO 列表
	todoList, err := w.loadTodoList()
	if err != nil {
		return "", fmt.Errorf("failed to load todo list: %w", err)
	}

	switch operation {
	case "add":
		return w.handleAdd(todoList, params)
	case "list":
		return w.handleList(todoList)
	case "update":
		return w.handleUpdate(todoList, params)
	case "delete":
		return w.handleDelete(todoList, params)
	case "clear":
		return w.handleClear(todoList)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// handleAdd 添加新任务
func (w *WriteTodoTool) handleAdd(todoList *TodoList, params map[string]any) (string, error) {
	description, ok := params["description"].(string)
	if !ok || description == "" {
		return "", fmt.Errorf("description is required for add operation")
	}

	priority := "medium"
	if p, ok := params["priority"].(string); ok && p != "" {
		priority = p
	}

	now := time.Now()
	item := TodoItem{
		ID:          generateID(),
		Description: description,
		Status:      "pending",
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	todoList.Items = append(todoList.Items, item)

	if err := w.saveTodoList(todoList); err != nil {
		return "", fmt.Errorf("failed to save todo list: %w", err)
	}

	return fmt.Sprintf("✅ Task added successfully!\nID: %s\nDescription: %s\nPriority: %s\nStatus: pending",
		item.ID, item.Description, item.Priority), nil
}

// handleList 列出所有任务
func (w *WriteTodoTool) handleList(todoList *TodoList) (string, error) {
	if len(todoList.Items) == 0 {
		return "📋 No tasks in the todo list.", nil
	}

	result := "📋 **Todo List**\n\n"

	// 按状态和优先级排序
	sortedItems := make([]TodoItem, len(todoList.Items))
	copy(sortedItems, todoList.Items)

	// 分类显示
	pendingItems := []TodoItem{}
	inProgressItems := []TodoItem{}
	completedItems := []TodoItem{}

	for _, item := range sortedItems {
		switch item.Status {
		case "pending":
			pendingItems = append(pendingItems, item)
		case "in_progress":
			inProgressItems = append(inProgressItems, item)
		case "completed":
			completedItems = append(completedItems, item)
		}
	}

	if len(inProgressItems) > 0 {
		result += "🔄 **In Progress:**\n"
		for _, item := range inProgressItems {
			result += fmt.Sprintf("  - [%s] %s (priority: %s)\n", item.ID, item.Description, item.Priority)
		}
		result += "\n"
	}

	if len(pendingItems) > 0 {
		result += "⏳ **Pending:**\n"
		for _, item := range pendingItems {
			result += fmt.Sprintf("  - [%s] %s (priority: %s)\n", item.ID, item.Description, item.Priority)
		}
		result += "\n"
	}

	if len(completedItems) > 0 {
		result += "✅ **Completed:**\n"
		for _, item := range completedItems {
			result += fmt.Sprintf("  - [%s] %s\n", item.ID, item.Description)
		}
		result += "\n"
	}

	result += fmt.Sprintf("Total: %d tasks (%d pending, %d in progress, %d completed)",
		len(todoList.Items), len(pendingItems), len(inProgressItems), len(completedItems))

	return result, nil
}

// handleUpdate 更新任务状态
func (w *WriteTodoTool) handleUpdate(todoList *TodoList, params map[string]any) (string, error) {
	id, ok := params["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required for update operation")
	}

	status, ok := params["status"].(string)
	if !ok || status == "" {
		return "", fmt.Errorf("status is required for update operation")
	}

	// 查找并更新任务
	found := false
	var updatedItem TodoItem
	for i, item := range todoList.Items {
		if item.ID == id {
			todoList.Items[i].Status = status
			todoList.Items[i].UpdatedAt = time.Now()
			updatedItem = todoList.Items[i]
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("task with ID %s not found", id)
	}

	if err := w.saveTodoList(todoList); err != nil {
		return "", fmt.Errorf("failed to save todo list: %w", err)
	}

	statusEmoji := map[string]string{
		"pending":     "⏳",
		"in_progress": "🔄",
		"completed":   "✅",
	}

	return fmt.Sprintf("%s Task updated successfully!\nID: %s\nDescription: %s\nNew Status: %s",
		statusEmoji[status], updatedItem.ID, updatedItem.Description, status), nil
}

// handleDelete 删除任务
func (w *WriteTodoTool) handleDelete(todoList *TodoList, params map[string]any) (string, error) {
	id, ok := params["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required for delete operation")
	}

	// 查找并删除任务
	found := false
	var deletedItem TodoItem
	newItems := make([]TodoItem, 0, len(todoList.Items)-1)
	for _, item := range todoList.Items {
		if item.ID == id {
			deletedItem = item
			found = true
			continue
		}
		newItems = append(newItems, item)
	}

	if !found {
		return "", fmt.Errorf("task with ID %s not found", id)
	}

	todoList.Items = newItems

	if err := w.saveTodoList(todoList); err != nil {
		return "", fmt.Errorf("failed to save todo list: %w", err)
	}

	return fmt.Sprintf("🗑️ Task deleted successfully!\nID: %s\nDescription: %s",
		deletedItem.ID, deletedItem.Description), nil
}

// handleClear 清空所有任务
func (w *WriteTodoTool) handleClear(todoList *TodoList) (string, error) {
	count := len(todoList.Items)
	todoList.Items = []TodoItem{}

	if err := w.saveTodoList(todoList); err != nil {
		return "", fmt.Errorf("failed to save todo list: %w", err)
	}

	return fmt.Sprintf("🧹 Cleared all %d tasks from the todo list.", count), nil
}

// loadTodoList 从文件加载 TODO 列表
func (w *WriteTodoTool) loadTodoList() (*TodoList, error) {
	todoPath := filepath.Join(w.workingDir, TodoFileName)

	data, err := os.ReadFile(todoPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，返回空列表
			return &TodoList{Items: []TodoItem{}}, nil
		}
		return nil, err
	}

	var todoList TodoList
	if err := json.Unmarshal(data, &todoList); err != nil {
		return nil, fmt.Errorf("failed to parse todo list: %w", err)
	}

	return &todoList, nil
}

// saveTodoList 保存 TODO 列表到文件
func (w *WriteTodoTool) saveTodoList(todoList *TodoList) error {
	todoPath := filepath.Join(w.workingDir, TodoFileName)

	data, err := json.MarshalIndent(todoList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal todo list: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(todoPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(todoPath, data, 0755)
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%10000)
}
