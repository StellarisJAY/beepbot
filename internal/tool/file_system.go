package tool

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// fileSystemTool 文件系统工具的公共基础结构
type fileSystemTool struct {
	workingDir     string // 智能体工作目录
	beepbotDataDir string // beepbot公共数据目录
}

// resolvePath 将相对路径转换成工作目录内的路径，判断绝对路径是否是可用访问的路径
func (f *fileSystemTool) resolvePath(path string) (string, error) {
	absPath := path
	var err error
	// 如果不是绝对路径，则是工作目录内部的相对路径，做工作目录安全检查
	if !filepath.IsAbs(path) {
		// 清理路径，移除前缀的 / 或 ./
		cleanPath := strings.TrimPrefix(path, "./")
		// 与工作目录拼接
		fullPath := filepath.Join(f.workingDir, cleanPath)
		// 转换为绝对路径并清理（解析 ..）
		absPath, err = filepath.Abs(fullPath)
		if err != nil {
			return "", fmt.Errorf("invalid path: %w", err)
		}
	}
	// 安全检查：确保解析后的路径仍在工作目录内或公共数据目录
	if !strings.HasPrefix(absPath, f.workingDir) && !strings.HasPrefix(absPath, f.beepbotDataDir) {
		return "", fmt.Errorf("access denied: path %q attempts to escape working directory and beepbot data directory", path)
	}
	return absPath, nil
}

// ReadFileTool 读取文件内容工具
type ReadFileTool struct {
	fileSystemTool
}

func NewReadFileTool(workingDir string, beepbotDataDir string) *ReadFileTool {
	workingDirAbs, _ := filepath.Abs(workingDir)
	beepbotDataDirAbs, _ := filepath.Abs(beepbotDataDir)
	return &ReadFileTool{
		fileSystemTool: fileSystemTool{workingDir: workingDirAbs, beepbotDataDir: beepbotDataDirAbs},
	}
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

// ListDirTool 列出目录内容工具
type ListDirTool struct {
	fileSystemTool
}

func NewListDirTool(workingDir string, beepbotDataDir string) *ListDirTool {
	workingDirAbs, _ := filepath.Abs(workingDir)
	beepbotDataDirAbs, _ := filepath.Abs(beepbotDataDir)
	return &ListDirTool{
		fileSystemTool: fileSystemTool{workingDir: workingDirAbs, beepbotDataDir: beepbotDataDirAbs},
	}
}

func (r *ListDirTool) Name() string {
	return "list_dir"
}

func (r *ListDirTool) Description() string {
	return "List the files in a directory (relative to working directory or )"
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

// WriteFileTool 写入文件内容工具
type WriteFileTool struct {
	fileSystemTool
}

func NewWriteFileTool(workingDir string, beepbotDataDir string) *WriteFileTool {
	workingDirAbs, _ := filepath.Abs(workingDir)
	beepbotDataDirAbs, _ := filepath.Abs(beepbotDataDir)
	return &WriteFileTool{
		fileSystemTool: fileSystemTool{workingDir: workingDirAbs, beepbotDataDir: beepbotDataDirAbs},
	}
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
