package agent

import (
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/skill"
	"github.com/StellarisJAY/beepbot/internal/types"
)

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
`

type contextBuilder struct {
	systemPrompt     string
	skillManager     *skill.Manager
	session          *session.Session
	prebuiltMessages []types.Message
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
			Content: builtinSystemPrompt, // 内置系统提示词
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

func (b *contextBuilder) buildContext() []types.Message {
	// 历史消息
	return append(b.prebuiltMessages, b.session.GetHistory()...)
}
