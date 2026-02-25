package tool

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileTool 读取文件内容工具
type ReadFileTool struct {
	workingDir string
}

func NewReadFileTool(workingDir string) *ReadFileTool {
	return &ReadFileTool{workingDir: workingDir}
}

func (r *ReadFileTool) Name() string {
	return "read_file"
}

func (r *ReadFileTool) Description() string {
	return "Read the content of a file (relative to working directory)"
}

func (r *ReadFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"description": "relative path to the file (e.g., 'data.txt' or 'docs/readme.md')",
			},
		},
		"required": []string{"path"},
	}
}

func (r *ReadFileTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := params["path"].(string)

	// 验证并解析路径
	validPath, err := r.resolvePath(path)
	if err != nil {
		return "", err
	}

	file, err := os.Open(validPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// resolvePath 将相对路径解析为工作目录内的绝对路径，防止路径遍历
func (r *ReadFileTool) resolvePath(path string) (string, error) {
	// 清理路径，移除前缀的 / 或 ./
	cleanPath := strings.TrimPrefix(path, "/")
	cleanPath = strings.TrimPrefix(cleanPath, "./")

	// 与工作目录拼接
	fullPath := filepath.Join(r.workingDir, cleanPath)

	// 转换为绝对路径并清理（解析 ..）
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// 安全检查：确保解析后的路径仍在工作目录内
	// 这可以防止通过 ../ 进行路径遍历攻击
	if !strings.HasPrefix(absPath, r.workingDir) {
		return "", fmt.Errorf("access denied: path %q attempts to escape working directory", path)
	}

	return absPath, nil
}

// ListDirTool 列出目录内容工具
type ListDirTool struct {
	workingDir string
}

func NewListDirTool(workingDir string) *ListDirTool {
	return &ListDirTool{workingDir: workingDir}
}

func (r *ListDirTool) Name() string {
	return "list_dir"
}

func (r *ListDirTool) Description() string {
	return "List the files in a directory (relative to working directory)"
}

func (r *ListDirTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"description": "relative path to the directory (e.g., '.' or 'docs')",
			},
		},
		"required": []string{"path"},
	}
}

func (r *ListDirTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := params["path"].(string)

	// 验证并解析路径
	validPath, err := r.resolvePath(path)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(validPath)
	if err != nil {
		return "", err
	}
	result := strings.Builder{}
	for _, entry := range entries {
		result.WriteString(fmt.Sprintf("name:%q,dir:%v\n\n", entry.Name(), entry.IsDir()))
	}
	return result.String(), nil
}

// resolvePath 将相对路径解析为工作目录内的绝对路径
func (r *ListDirTool) resolvePath(path string) (string, error) {
	cleanPath := strings.TrimPrefix(path, "/")
	cleanPath = strings.TrimPrefix(cleanPath, "./")
	fullPath := filepath.Join(r.workingDir, cleanPath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if !strings.HasPrefix(absPath, r.workingDir) {
		return "", fmt.Errorf("access denied: path %q attempts to escape working directory", path)
	}
	return absPath, nil
}

// WriteFileTool 写入文件内容工具
type WriteFileTool struct {
	workingDir string
}

func NewWriteFileTool(workingDir string) *WriteFileTool {
	return &WriteFileTool{workingDir: workingDir}
}

func (r *WriteFileTool) Name() string {
	return "write_file"
}

func (r *WriteFileTool) Description() string {
	return "Write text content to a file (relative to working directory)"
}

func (r *WriteFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"description": "relative path to the file (e.g., 'data.txt' or 'docs/readme.md')",
			},
			"content": map[string]any{
				"description": "text content to write",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (r *WriteFileTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := params["path"].(string)
	content := params["content"].(string)

	// 验证并解析路径
	validPath, err := r.resolvePath(path)
	if err != nil {
		return "", err
	}

	file, err := os.OpenFile(validPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.Write([]byte(content))
	if err != nil {
		return "", err
	}
	return "write success", nil
}

// resolvePath 将相对路径解析为工作目录内的绝对路径
func (r *WriteFileTool) resolvePath(path string) (string, error) {
	cleanPath := strings.TrimPrefix(path, "/")
	cleanPath = strings.TrimPrefix(cleanPath, "./")
	fullPath := filepath.Join(r.workingDir, cleanPath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if !strings.HasPrefix(absPath, r.workingDir) {
		return "", fmt.Errorf("access denied: path %q attempts to escape working directory", path)
	}
	return absPath, nil
}
