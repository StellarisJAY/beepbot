package react

import (
	"fmt"
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
2. 你完成的重要操作与结果
3. 需要记忆的重要信息, 比如用户的偏好等

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
	cronJob          bool // 智能体是否是定时任务触发
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

	// 让智能体知道自己是定时任务触发的
	if b.cronJob {
		messages = append(messages, types.Message{
			Role:    types.RoleSystem,
			Content: "<trigger_by>cron\n\n<rule>禁止嵌套创建定时任务</rule>\n\n</trigger_by>\n\n",
		})
	}

	// 注入系统当前时间
	messages = append(messages, types.Message{
		Role:    types.RoleSystem,
		Content: fmt.Sprintf("<system_time>%s</system_time>\n\n", time.Now().Format(time.DateTime)),
	})
	return messages
}
