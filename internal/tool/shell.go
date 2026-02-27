package tool

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/StellarisJAY/beepbot/internal/config"
)

type ShellTool struct {
	workingDir        string
	allowedCommands   []string
	forbiddenCommands []string
	timeout           time.Duration
	description       string
	system            string // 当前的操作系统环境
}

// Description implements [Tool].
func (s *ShellTool) Description() string {
	return s.description
}
func (s *ShellTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	// 1. 参数验证
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("invalid or empty command parameter")
	}

	command = strings.TrimSpace(command)
	slog.Info("Shell tool executing", "command", command, "working_dir", s.workingDir)

	// 2. 安全检查
	if s.isForbidden(command) {
		err := fmt.Errorf("command is forbidden: %s", command)
		slog.Warn("Shell tool blocked forbidden command", "command", command)
		return "", err
	}

	// 3. 超时控制
	execCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 4. 执行命令
	var cmd *exec.Cmd
	switch s.system {
	case "linux":
		cmd = exec.CommandContext(execCtx, "sh", "-c", command)
	case "windows":
		cmd = exec.CommandContext(execCtx, "cmd", "/c", "chcp 65001 && "+command)
	default:
		cmd = exec.CommandContext(execCtx, "sh", "-c", command)
	}

	cmd.Dir = s.workingDir

	output, err := cmd.CombinedOutput()

	// 5. 处理结果
	result := string(output)
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timeout after %s", s.timeout)
		}
		slog.Error("Shell tool execution failed", "command", command, "error", err)
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	slog.Info("Shell tool execution completed", "command", command, "output_len", len(result))
	return result, nil
}

func (s *ShellTool) isForbidden(command string) bool {
	command = strings.TrimSpace(command)
	lowerCmd := strings.ToLower(command)

	for _, forbidden := range s.forbiddenCommands {
		forbidden = strings.TrimSpace(forbidden)
		if forbidden == "" {
			continue
		}

		// 检查命令是否以 forbidden 开头
		if strings.HasPrefix(lowerCmd, strings.ToLower(forbidden)) {
			return true
		}
	}

	return false
}

// Name implements [Tool].
func (s *ShellTool) Name() string {
	return "shell"
}

// Parameters implements [Tool].
func (s *ShellTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"description": "the shell command to execute. eg: \"cat main.py\",\"cat data.csv | grep beep\"",
			},
		},
		"required": []string{"command"},
	}
}

func NewShellTool(config config.Config) Tool {
	shellToolConfig := config.BuiltinTools.Shell
	timeout, err := time.ParseDuration(shellToolConfig.Timeout)
	if err != nil {
		slog.Error("Invalid shell tool timeout, using default", "invalid", config.BuiltinTools.Shell.Timeout, "default", "30s")
		timeout = 30 * time.Second
	}

	tool := &ShellTool{
		forbiddenCommands: shellToolConfig.ForbiddenCommands,
		timeout:           timeout,
		workingDir:        config.AgentConfig.WorkingDir,
		system:            runtime.GOOS,
	}
	tool.description = tool.buildDescription()
	return tool
}

func (s *ShellTool) buildDescription() string {
	forbiddenList := "无"
	if len(s.forbiddenCommands) > 0 {
		forbiddenList = strings.Join(s.forbiddenCommands, "\n")
	}
	slog.Info("forbidden shell commands", "cmds", s.forbiddenCommands)
	return fmt.Sprintf(`执行 shell 命令并在指定工作目录中运行。

<capabilities>
- 执行文件操作: ls, cat, head, tail, grep, find
- 文本处理: awk, sed, sort, uniq
- 目录操作: mkdir, pwd, tree
- 支持管道: cat file.txt | grep pattern
- 支持重定向: echo "content" > file.txt
- 脚本语言: python, sh
- 编译工具: gcc, go, rustc, javac
- 运行用户程序: example.exe, ./main
</capabilities>

当前操作系统信息:
<os>
%s
</os>

<working-directory>
%s
</working-directory>

以下命令被禁止执行，调用将返回错误:
<forbidden-commands>
%s
</forbidden-commands>

以下行为被认为是危险操作, 禁止执行:
<dangerous-operations>
- 删除文件或目录
- 下载
- 修改环境变量
- 关闭重启系统
- 访问不明网页
- 其他可能产生危险的操作
</dangerous-operations>

<limits>
- 脚本运行限制: 运行前请仔细阅读代码, 判断是否存在危险操作, 对于存在危险操作的脚本禁止使用shell工具执行。
- 编译工具限制: 编译前请仔细阅读代码, 判断是否存在危险操作, 对于存在危险操作的程序禁止使用shell工具编译。
- 运行用户程序限制: 用户应该描述程序的作用, 你需要判断运行是否存在危险操作, 对于存在危险操作的程序禁止使用shell工具运行。
</limits>

<examples>
正确用法:
- "ls -la" 列出当前目录文件
- "cat README.md" 查看文件内容
- "find . -name '*.go'" 查找 Go 文件
- "grep -r 'pattern' src/" 搜索代码

错误用法（会被拒绝）:
- "sudo apt install xxx" 使用 sudo
- "rm -rf /" 危险删除命令
- "python remove_everything.py" 危险操作脚本禁止运行
- "go build download_unknown_file.go" 危险操作程序禁止编译
- "test.exe" 潜在危险操作程序, 用户必须描述程序作用否则禁止运行
</examples>

<important-notes>
1. 命令在工作目录内执行，无法访问外部路径
2. 命令执行超时限制为 %s，超时将终止
3. 输出过长可能被截断
</important-notes>`,
		s.system,
		s.workingDir,
		forbiddenList,
		s.timeout.String(),
	)
}
