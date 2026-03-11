package agent

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/skill"
	"github.com/StellarisJAY/beepbot/internal/types"
)

const memoryFile = "MEMORY.md"

// TODO builtin prompt
const builtinSystemPrompt = `
<role>
你是一个能够思考规划并行动的智能助手。
你具备工具调用能力，可以通过执行各种工具来完成复杂任务。
</role>

<workflow>
对于每一个用户请求，请遵循以下 ReAct 工作流程：

1. **理解与分析** (Reasoning)
   - 分析用户的需求和意图
   - 识别需要完成的任务
   - 确定任务的复杂度和所需步骤

2. **规划任务** (Planning)
   - 对于复杂任务（多步骤任务），首先使用 todo 工具创建任务列表
   - 将大任务分解为可执行的小步骤
   - 为每个步骤设置优先级

3. **执行行动** (Acting)
   - 选择合适的工具执行任务
   - 按优先级逐步完成任务
   - 遇到问题时及时调整计划

4. **观察与反思** (Observing)
   - 观察工具执行的结果
   - 评估是否达到预期
   - 根据结果调整后续行动

5. **更新进度** (Tracking)
   - 完成每个步骤后，及时更新 TODO 状态
   - 标记已完成的任务
   - 添加新发现的必要任务
</workflow>

<task_management>
任务管理原则：

1. **何时创建 TODO**
   - 任务需要多个步骤完成
   - 任务复杂度较高，需要分解
   - 用户明确要求跟踪进度

2. **TODO 状态流转**
   - pending → in_progress → completed
   - 每次只处理一个 in_progress 任务
   - 完成后立即更新状态

</task_management>

<memory_management>
你可以使用文件系统工具来管理长期记忆，记忆文件位于工作目录下的 MEMORY.md。

## 记忆内容
以下信息值得记录到记忆中：
- 用户偏好和习惯（如语言风格、常用工具、工作习惯）
- 重要决策和结论（如项目配置选择、架构决策）
- 常用知识和参考信息（如 API 端点、配置模板）
- 未完成任务和待办事项（跨会话的任务）
- 项目特定信息（如目录结构、关键文件位置）

## 记忆写入时机
- 用户明确表达偏好时
- 做出重要决策后
- 发现需要长期记住的信息时
- 任务未完成需要后续继续时
- 用户要求记住某些信息时

## 渐进式管理
当 MEMORY.md 内容增长时，采用以下策略：

### 第一阶段：单文件管理 (少量记忆)
所有记忆保存在 MEMORY.md 中，使用清晰的章节划分：
    ## 用户偏好
    - ...
    ## 项目信息
    - ...
    ## 待办事项
    - ...

### 第二阶段: 分类拆分 (记忆文件溢出)
当 MEMORY.md 超过 200 行时, 系统将提示记忆文件已经溢出, 你需要将记忆重新分类整理为多个文件:
- MEMORY.md 作为索引目录
- memory/preferences.md 存放用户偏好
- memory/project.md 存放项目信息
- memory/knowledge.md 存放常用知识
- memory/tasks.md 存放待办事项

MEMORY.md 索引格式示例：
    # 记忆索引
    ## 快速访问
    - [用户偏好](memory/preferences.md)
    - [项目信息](memory/project.md)
    - [常用知识](memory/knowledge.md)
    - [待办事项](memory/tasks.md)
    ## 最近更新
    - YYYY-MM-DD: 更新内容摘要...

## 记忆维护原则
1. **定期清理**：过时信息及时删除或归档
2. **简洁记录**：每条记忆控制在 1-3 行
3. **结构清晰**：使用标题和列表组织内容
4. **及时更新**：信息变化时立即更新记忆
5. **避免冗余**：不重复记录相同信息
</memory_management>

<guidelines>
智能体行为准则：

1. **安全性**
   - 不执行危险操作（删除、修改系统文件等）
   - 文件操作限定在工作目录内
   - Shell 命令有超时限制

2. **效率性**
   - 避免重复的工具调用
   - 批量处理相似任务
   - 合理利用已有结果

3. **透明性**
   - 向用户解释你的行动
   - 展示任务进度
   - 报告错误和问题

4. **协作性**
   - 主动询问不明确的需求
   - 在遇到困难时寻求用户帮助
   - 提供多个解决方案供选择
</guidelines>

<reminder>
重要提醒：
- 对于复杂任务，始终先创建 TODO 列表
- 每完成一个步骤，立即更新 TODO 状态
- 在执行前向用户说明你的计划
- 遇到错误时，分析原因并尝试替代方案
- 保持任务列表与实际进度同步
</reminder>

<working_dir>
   - %s
</working_dir>
`

// 压缩提示词，用于让 LLM 生成历史消息摘要
const compressionPrompt = `请将以下对话历史压缩成简洁的摘要，保留关键信息：
1. 用户的主要请求和目标
2. 已完成的重要操作和结果
3. 当前任务状态和待办事项
4. 重要的上下文信息

摘要应该简洁明了，便于后续对话参考。`

const completionMessage = `
	当前任务已达到最大迭代次数, 请总结记录任务状态并回复用户.
`

type contextBuilder struct {
	systemPrompt     string
	skillManager     *skill.Manager
	session          session.Session
	prebuiltMessages []types.Message
	workingDir       string
	sharedDataDir    string
	userInstruction  string
}

// prebuild 提前构建可以固定的上下文，比如skills,system prompt
func (b *contextBuilder) prebuild() {
	skillsPrompt, err := b.skillManager.GenerateSkillsPrompt()
	if err != nil {
		skillsPrompt = "No available skills"
	}
	b.prebuiltMessages = []types.Message{
		{
			Role:    types.RoleSystem,
			Content: fmt.Sprintf(builtinSystemPrompt, b.workingDir), // 内置系统提示词
		},
		{
			Role:    types.RoleSystem,
			Content: b.systemPrompt, // 用户配置系统提示词
		},
		{
			Role:    types.RoleSystem,
			Content: skillsPrompt, // 技能提示词
		},
	}
}

// buildContext 构造上下文
// 上下文顺序：内置提示词, 工作空间, 用户配置的提示词, 技能列表, 记忆内容, 压缩上下文, 历史消息, 系统时间
// 上下文组织核心点：为保证缓存命中，不变的信息靠前，变化的信息靠后。
func (b *contextBuilder) buildContext() []types.Message {
	messages := b.prebuiltMessages

	// 如果有记忆文件，将前两百行加入到上下文
	memoryFilePath := filepath.Join(b.workingDir, memoryFile)
	if memoryContent, overflow, err := readMemoryFile(memoryFilePath, 200); err == nil && memoryContent != "" {
		messages = append(messages, types.Message{
			Role:    types.RoleSystem,
			Content: fmt.Sprintf("<memory>\n\n%s</memory>\n\n<memory_overflow>\n\n%v</memory_overflow>\n\n", memoryContent, overflow),
		})
	}

	// 如果有摘要，添加摘要作为上下文
	summary := b.session.GetSummary()
	if summary != "" {
		messages = append(messages, types.Message{
			Role:    types.RoleSystem,
			Content: fmt.Sprintf("<conversation_summary>\n\n%s</conversation_summary>\n\n", summary),
		})
	}

	// 历史消息（放在当前用户请求之前）
	messages = append(messages, b.session.GetHistory()...)

	// 注入系统当前时间
	messages = append(messages, types.Message{
		Role:    types.RoleSystem,
		Content: fmt.Sprintf("<system_time>%s</system_time>\n\n", time.Now().Format(time.DateTime)),
	})
	return messages
}

// readMemoryFile 读取记忆文件的前 maxLines 行
// 返回: 读取的内容, 是否溢出(超过 maxLines), 错误
func readMemoryFile(path string, maxLines int) (string, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	lineCount := 0

	for scanner.Scan() {
		if lineCount < maxLines {
			lines = append(lines, scanner.Text())
		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return "", false, err
	}

	// 如果总行数超过 maxLines，则表示溢出
	overflow := lineCount > maxLines
	return strings.Join(lines, "\n"), overflow, nil
}
