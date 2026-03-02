package tool

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

// EditFileTool 编辑文件工具
type EditFileTool struct {
	fileSystemTool
}

func NewEditFileTool(workingDir string, beepbotDataDir string) *EditFileTool {
	workingDirAbs, _ := filepath.Abs(workingDir)
	beepbotDataDirAbs, _ := filepath.Abs(beepbotDataDir)
	return &EditFileTool{
		fileSystemTool: fileSystemTool{workingDir: workingDirAbs, beepbotDataDir: beepbotDataDirAbs},
	}
}

func (e *EditFileTool) Name() string {
	return "edit_file"
}

func (e *EditFileTool) Description() string {
	return "Edit an existing file with various operations: replace_text, replace_lines, insert_lines, delete_lines, append. The file must exist before editing."
}

func (e *EditFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"description": "relative path to the file (e.g., 'data.txt' or 'docs/readme.md')",
			},
			"operation": map[string]any{
				"description": "edit operation type: 'replace_text' (search and replace text), 'replace_lines' (replace lines in range), 'insert_lines' (insert before specified line), 'delete_lines' (delete lines in range), 'append' (append to end of file)",
			},
			"search": map[string]any{
				"description": "text to search for (required for replace_text operation)",
			},
			"replacement": map[string]any{
				"description": "replacement text (required for replace_text operation)",
			},
			"all": map[string]any{
				"description": "replace all occurrences when true (optional for replace_text, default false)",
			},
			"start_line": map[string]any{
				"description": "start line number, 1-based (required for replace_lines and delete_lines operations)",
			},
			"end_line": map[string]any{
				"description": "end line number, 1-based inclusive (required for replace_lines and delete_lines operations)",
			},
			"line": map[string]any{
				"description": "line number to insert before, 1-based (required for insert_lines operation)",
			},
			"content": map[string]any{
				"description": "content to insert or replace with (required for replace_lines, insert_lines, and append operations)",
			},
		},
		"required": []string{"path", "operation"},
	}
}

func (e *EditFileTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("path parameter is required")
	}

	operation, ok := params["operation"].(string)
	if !ok || operation == "" {
		return "", fmt.Errorf("operation parameter is required")
	}

	// 验证并解析路径
	validPath, err := e.resolvePath(path)
	if err != nil {
		return "", err
	}

	// 检查文件是否存在
	info, err := os.Stat(validPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", path)
	}
	if err != nil {
		return "", fmt.Errorf("failed to access file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// 根据操作类型执行相应编辑
	switch operation {
	case "replace_text":
		return e.executeReplaceText(validPath, params)
	case "replace_lines":
		return e.executeReplaceLines(validPath, params)
	case "insert_lines":
		return e.executeInsertLines(validPath, params)
	case "delete_lines":
		return e.executeDeleteLines(validPath, params)
	case "append":
		return e.executeAppend(validPath, params)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// executeReplaceText 执行搜索替换操作
func (e *EditFileTool) executeReplaceText(path string, params map[string]any) (string, error) {
	search, ok := params["search"].(string)
	if !ok {
		return "", fmt.Errorf("search parameter is required for replace_text operation")
	}

	replacement, ok := params["replacement"].(string)
	if !ok {
		return "", fmt.Errorf("replacement parameter is required for replace_text operation")
	}

	// 读取文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	replaceAll := false
	if all, ok := params["all"].(bool); ok {
		replaceAll = all
	}

	var newContent string
	var count int
	if replaceAll {
		count = strings.Count(contentStr, search)
		newContent = strings.ReplaceAll(contentStr, search, replacement)
	} else {
		count = 0
		if strings.Contains(contentStr, search) {
			count = 1
			newContent = strings.Replace(contentStr, search, replacement, 1)
		} else {
			newContent = contentStr
		}
	}

	if count == 0 {
		return "", fmt.Errorf("search text not found in file")
	}

	// 写入文件
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if replaceAll {
		return fmt.Sprintf("replaced %d occurrence(s)", count), nil
	}
	return "replaced 1 occurrence", nil
}

// executeReplaceLines 替换指定行范围的内容
func (e *EditFileTool) executeReplaceLines(path string, params map[string]any) (string, error) {
	startLine, err := e.getIntParam(params, "start_line")
	if err != nil {
		return "", fmt.Errorf("start_line parameter is required for replace_lines operation: %w", err)
	}

	endLine, err := e.getIntParam(params, "end_line")
	if err != nil {
		return "", fmt.Errorf("end_line parameter is required for replace_lines operation: %w", err)
	}

	content, ok := params["content"].(string)
	if !ok {
		return "", fmt.Errorf("content parameter is required for replace_lines operation")
	}

	// 读取文件所有行
	lines, err := e.readFileLines(path)
	if err != nil {
		return "", err
	}

	// 验证行号
	if startLine < 1 || startLine > len(lines) {
		return "", fmt.Errorf("start_line %d is out of range (1-%d)", startLine, len(lines))
	}
	if endLine < 1 || endLine > len(lines) {
		return "", fmt.Errorf("end_line %d is out of range (1-%d)", endLine, len(lines))
	}
	if startLine > endLine {
		return "", fmt.Errorf("start_line (%d) cannot be greater than end_line (%d)", startLine, endLine)
	}

	// 构建新内容
	var newLines []string
	// 添加 startLine 之前的内容
	newLines = append(newLines, lines[:startLine-1]...)
	// 添加新内容
	newLines = append(newLines, content)
	// 添加 endLine 之后的内容
	if endLine < len(lines) {
		newLines = append(newLines, lines[endLine:]...)
	}

	// 写入文件
	err = e.writeFileLines(path, newLines)
	if err != nil {
		return "", err
	}

	replacedCount := endLine - startLine + 1
	return fmt.Sprintf("replaced %d line(s) (line %d-%d)", replacedCount, startLine, endLine), nil
}

// executeInsertLines 在指定行之前插入内容
func (e *EditFileTool) executeInsertLines(path string, params map[string]any) (string, error) {
	line, err := e.getIntParam(params, "line")
	if err != nil {
		return "", fmt.Errorf("line parameter is required for insert_lines operation: %w", err)
	}

	content, ok := params["content"].(string)
	if !ok {
		return "", fmt.Errorf("content parameter is required for insert_lines operation")
	}

	// 读取文件所有行
	lines, err := e.readFileLines(path)
	if err != nil {
		return "", err
	}

	// 验证行号 (允许在文件末尾之后插入)
	if line < 1 || line > len(lines)+1 {
		return "", fmt.Errorf("line %d is out of range (1-%d)", line, len(lines)+1)
	}

	// 构建新内容
	var newLines []string
	// 添加插入位置之前的内容
	newLines = append(newLines, lines[:line-1]...)
	// 添加新内容
	newLines = append(newLines, content)
	// 添加插入位置之后的内容
	newLines = append(newLines, lines[line-1:]...)

	// 写入文件
	err = e.writeFileLines(path, newLines)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("inserted content before line %d", line), nil
}

// executeDeleteLines 删除指定行范围
func (e *EditFileTool) executeDeleteLines(path string, params map[string]any) (string, error) {
	startLine, err := e.getIntParam(params, "start_line")
	if err != nil {
		return "", fmt.Errorf("start_line parameter is required for delete_lines operation: %w", err)
	}

	endLine, err := e.getIntParam(params, "end_line")
	if err != nil {
		return "", fmt.Errorf("end_line parameter is required for delete_lines operation: %w", err)
	}

	// 读取文件所有行
	lines, err := e.readFileLines(path)
	if err != nil {
		return "", err
	}

	// 验证行号
	if startLine < 1 || startLine > len(lines) {
		return "", fmt.Errorf("start_line %d is out of range (1-%d)", startLine, len(lines))
	}
	if endLine < 1 || endLine > len(lines) {
		return "", fmt.Errorf("end_line %d is out of range (1-%d)", endLine, len(lines))
	}
	if startLine > endLine {
		return "", fmt.Errorf("start_line (%d) cannot be greater than end_line (%d)", startLine, endLine)
	}

	// 构建新内容
	var newLines []string
	// 添加删除范围之前的内容
	newLines = append(newLines, lines[:startLine-1]...)
	// 添加删除范围之后的内容
	if endLine < len(lines) {
		newLines = append(newLines, lines[endLine:]...)
	}

	// 写入文件
	err = e.writeFileLines(path, newLines)
	if err != nil {
		return "", err
	}

	deletedCount := endLine - startLine + 1
	return fmt.Sprintf("deleted %d line(s) (line %d-%d)", deletedCount, startLine, endLine), nil
}

// executeAppend 在文件末尾追加内容
func (e *EditFileTool) executeAppend(path string, params map[string]any) (string, error) {
	content, ok := params["content"].(string)
	if !ok {
		return "", fmt.Errorf("content parameter is required for append operation")
	}

	// 打开文件以追加
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open file for append: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("failed to append to file: %w", err)
	}

	return "content appended successfully", nil
}

// readFileLines 读取文件所有行
func (e *EditFileTool) readFileLines(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 分割成行，保留空行
	lines := strings.Split(string(content), "\n")
	return lines, nil
}

// writeFileLines 将行写入文件
func (e *EditFileTool) writeFileLines(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// getIntParam 获取整数参数，支持 int 和 float64 类型（JSON 解析可能将数字解析为 float64）
func (e *EditFileTool) getIntParam(params map[string]any, key string) (int, error) {
	val, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("parameter %s not found", key)
	}

	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("parameter %s is not a valid number", key)
	}
}
