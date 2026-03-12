# BeepBot - AI 智能代理机器人

## 项目概述

BeepBot 是一个 AI 代理机器人项目，包含后端服务和前端仪表板。后端使用 Go 语言编写，通过多种消息渠道提供智能对话能力；前端使用 Vue 3 构建，提供管理界面。

后端支持 OpenAI、DashScope（阿里云）、Ollama 等 LLM 提供商，可通过 QQ 机器人、飞书机器人与用户交互。具有 ReAct 风格的代理循环和工具调用能力，可以执行 shell 命令、读写文件、管理任务列表、收集系统信息、管理定时任务，帮助用户完成各种任务。

## 项目结构

```
beepbot/                          # 项目根目录
├── beepbot/                      # 后端 Go 项目
│   ├── cmd/
│   │   └── api/
│   │       └── main.go           # API 服务入口点
│   ├── internal/
│   │   ├── agent/
│   │   │   ├── agent.go          # 核心代理循环和消息处理
│   │   │   ├── api.go            # API 模式 AgentRunner
│   │   │   ├── context.go        # 代理上下文管理
│   │   │   ├── hooks.go          # 输出钩子接口和实现
│   │   │   └── subagent_executor.go # 子智能体执行器
│   │   ├── api/
│   │   │   ├── agent_handler.go  # Agent API 处理器
│   │   │   ├── bot_handler.go    # Bot API 处理器
│   │   │   ├── cron_handler.go   # Cron API 处理器
│   │   │   ├── provider_handler.go # Provider API 处理器
│   │   │   ├── session_handler.go # Session API 处理器
│   │   │   ├── skill_handler.go  # Skill API 处理器
│   │   │   ├── response.go       # API 响应工具
│   │   │   └── router.go         # 路由定义
│   │   ├── channel/
│   │   │   ├── channel.go        # 渠道接口
│   │   │   ├── channel_manager.go # Channel 管理器
│   │   │   ├── config.go         # Channel 配置
│   │   │   ├── constants.go      # 常量定义
│   │   │   ├── factory.go        # Channel 工厂
│   │   │   ├── feishu_channel.go # 飞书机器人渠道实现
│   │   │   ├── message_bus.go    # 入站/出站消息总线
│   │   │   ├── qq_channel.go     # QQ 机器人渠道实现
│   │   │   ├── registry.go       # Channel 注册表
│   │   │   └── system_channel.go # 系统/内部渠道
│   │   ├── config/
│   │   │   ├── config.go         # 配置结构和加载
│   │   │   └── api.go            # API 模式配置
│   │   ├── cron/
│   │   │   └── scheduler.go      # 定时任务调度器
│   │   ├── crypto/
│   │   │   └── encryption.go     # 加密工具
│   │   ├── database/
│   │   │   └── database.go       # 数据库连接和管理
│   │   ├── logger/
│   │   │   ├── logger.go         # 日志初始化
│   │   │   └── qq.go             # QQ 渠道专用日志
│   │   ├── mcp/
│   │   │   └── mcp.go            # MCP（模型上下文协议）- 占位符
│   │   ├── provider/
│   │   │   ├── base.go           # LLM 提供商工厂
│   │   │   └── openai.go         # OpenAI 兼容提供商实现
│   │   ├── repository/
│   │   │   ├── repository.go     # 仓储基类
│   │   │   ├── agent_repository.go # Agent 仓储
│   │   │   ├── bot_repository.go # Bot 仓储
│   │   │   ├── cron_repository.go # Cron 仓储
│   │   │   ├── provider_repository.go # Provider 仓储
│   │   │   ├── session_repository.go # Session 仓储
│   │   │   └── skill_repository.go # Skill 仓储
│   │   ├── service/
│   │   │   ├── agent_manager.go  # Agent 管理器
│   │   │   ├── agent_service.go  # Agent 服务层
│   │   │   ├── bot_service.go    # Bot 服务层
│   │   │   ├── cron_service.go   # Cron 服务层
│   │   │   ├── provider_service.go # Provider 服务层
│   │   │   ├── session_service.go # Session 服务层
│   │   │   └── skill_service.go  # Skill 服务层
│   │   ├── session/
│   │   │   ├── session.go        # 会话接口
│   │   │   ├── api.go            # API 模式会话实现
│   │   │   └── memory.go         # 内存会话（用于子智能体）
│   │   ├── skill/
│   │   │   └── skill.go          # 技能管理系统
│   │   ├── tool/
│   │   │   ├── base.go           # 工具接口定义
│   │   │   ├── cron.go           # 定时任务工具
│   │   │   ├── cron_validator.go # Cron 表达式验证
│   │   │   ├── file_system.go    # 文件读/写/列表/编辑工具
│   │   │   ├── message.go        # 消息工具
│   │   │   ├── shell.go          # Shell 执行工具
│   │   │   ├── subagent.go       # 子智能体调用工具
│   │   │   ├── system.go         # 系统信息工具
│   │   │   ├── todo.go           # TODO 任务管理工具
│   │   │   └── tool_registry.go  # 工具注册表
│   │   └── types/
│   │       ├── agent.go          # Agent 相关类型
│   │       ├── agent_skill.go    # Agent-Skill 关联
│   │       ├── agent_tool.go     # Agent-Tool 关联
│   │       ├── base.go           # 基础类型
│   │       ├── bot.go            # Bot 相关类型
│   │       ├── cron.go           # Cron 相关类型
│   │       ├── memory.go         # 内存接口和类型
│   │       ├── provider.go       # LLM 提供商接口和消息类型
│   │       ├── session.go        # 会话数据库模型
│   │       └── skill.go          # Skill 相关类型
│   ├── config.example.json       # 配置示例
│   ├── go.mod                    # Go 模块定义
│   └── .gitignore                # Git 忽略文件
├── dashboard/                    # 前端 Vue 项目
│   ├── src/
│   │   ├── App.vue               # 根组件
│   │   ├── main.ts               # 应用入口
│   │   ├── api/                  # API 请求模块
│   │   │   ├── agent.ts          # Agent API
│   │   │   ├── bot.ts            # Bot API
│   │   │   ├── cron.ts           # Cron API
│   │   │   ├── provider.ts       # Provider API
│   │   │   ├── session.ts        # Session API
│   │   │   └── skill.ts          # Skill API
│   │   ├── assets/
│   │   │   └── styles/           # 样式文件
│   │   │       ├── variables.css # CSS 变量（主题色）
│   │   │       └── global.css    # 全局样式
│   │   ├── components/
│   │   │   ├── AgentCreateModal.vue # 智能体创建模态框
│   │   │   ├── BotFormModal.vue  # 机器人表单模态框
│   │   │   ├── MarkdownRenderer.vue # Markdown 渲染组件
│   │   │   └── layout/           # 布局组件
│   │   │       ├── AppLayout.vue # 整体布局容器
│   │   │       ├── Header.vue    # 头部栏组件
│   │   │       └── Sidebar.vue   # 侧边导航栏组件
│   │   ├── router/
│   │   │   └── index.ts          # Vue Router 配置
│   │   ├── stores/               # Pinia 状态管理
│   │   │   ├── counter.ts        # 计数器状态
│   │   │   ├── theme.ts          # 主题状态
│   │   │   └── sidebar.ts        # 侧边栏折叠状态
│   │   ├── types/                # TypeScript 类型定义
│   │   │   ├── agent.ts
│   │   │   ├── bot.ts
│   │   │   ├── cron.ts
│   │   │   ├── provider.ts
│   │   │   ├── session.ts
│   │   │   └── skill.ts
│   │   ├── utils/
│   │   │   └── http.ts           # HTTP 请求工具
│   │   └── views/                # 页面组件
│   │       ├── agents/
│   │       │   ├── AgentList.vue # 智能体列表页
│   │       │   ├── AgentConfig.vue # 智能体配置页
│   │       │   ├── AgentDetailLayout.vue # 智能体详情布局
│   │       │   ├── AgentLogs.vue # 智能体日志页
│   │       │   ├── AgentMonitor.vue # 智能体监控页
│   │       │   └── SessionMessages.vue # 会话消息页
│   │       ├── bots/
│   │       │   └── BotList.vue   # IM机器人列表页
│   │       ├── crons/
│   │       │   └── CronList.vue  # 定时任务列表页
│   │       ├── providers/
│   │       │   └── ProviderList.vue # 模型供应商列表页
│   │       ├── settings/
│   │       │   └── Settings.vue  # 全局设置页
│   │       └── skills/
│   │           ├── SkillList.vue # 技能列表页
│   │           └── SkillDetail.vue # 技能详情页
│   ├── public/                   # 静态资源
│   ├── index.html                # HTML 入口
│   ├── vite.config.ts            # Vite 配置
│   ├── tsconfig.json             # TypeScript 配置
│   ├── tsconfig.app.json         # 应用 TypeScript 配置
│   ├── tsconfig.node.json        # Node TypeScript 配置
│   ├── eslint.config.ts          # ESLint 配置
│   ├── .prettierrc.json          # Prettier 配置
│   ├── .oxlintrc.json            # Oxlint 配置
│   ├── .editorconfig             # EditorConfig
│   ├── .gitattributes            # Git 属性
│   ├── .gitignore                # Git 忽略文件
│   ├── env.d.ts                  # 环境类型声明
│   ├── package.json              # NPM 包配置
│   ├── pnpm-lock.yaml            # pnpm 锁文件
│   └── README.md                 # 项目说明
└── AGENTS.md                     # 本文件
```

## 技术栈

### 后端 (beepbot/)
- **语言**: Go 1.24.5
- **LLM 提供商**: OpenAI API 兼容 (OpenAI, DashScope, Ollama)
- **消息渠道**: QQ 机器人 (腾讯), 飞书机器人
- **数据库**: PostgreSQL (使用 GORM)
- **主要依赖**:
  - `github.com/tencent-connect/botgo` - QQ 机器人 SDK
  - `github.com/larksuite/oapi-sdk-go/v3` - 飞书 SDK
  - `github.com/openai/openai-go/v3` - OpenAI 客户端
  - `gorm.io/gorm` - 数据库 ORM
  - `github.com/google/uuid` - UUID 生成
  - `github.com/shirou/gopsutil/v4` - 系统信息

### 前端 (dashboard/)
- **框架**: Vue 3.5+
- **构建工具**: Vite 7
- **语言**: TypeScript 5.9
- **UI 组件库**: Ant Design Vue 4.x
- **图标库**: @ant-design/icons-vue
- **状态管理**: Pinia 3
- **路由**: Vue Router 5
- **代码规范**: ESLint + Oxlint + Prettier
- **包管理**: pnpm
- **Node.js**: ^20.19.0 || >=22.12.0

## 构建和运行

### 后端

#### 前置条件
- Go 1.24.5 或更高版本
- PostgreSQL 数据库

#### 构建
```bash
cd beepbot

# 构建 API 服务
go build -o beepbot-api.exe ./cmd/api
```

#### 运行
```bash
cd beepbot

# 运行 API 服务
./beepbot-api.exe -config /path/to/config.json
```

### 前端

#### 前置条件
- Node.js 20.19+ 或 22.12+
- pnpm

#### 安装依赖
```bash
cd dashboard
pnpm install
```

#### 开发
```bash
cd dashboard
pnpm dev
```

#### 构建
```bash
cd dashboard
pnpm build
```

#### 预览
```bash
cd dashboard
pnpm preview
```

#### 代码检查
```bash
cd dashboard
pnpm lint      # 运行 ESLint 和 Oxlint
pnpm format    # 使用 Prettier 格式化
```

## 配置

### 后端配置

应用使用 JSON 配置。复制 `config.example.json` 为 `config.json` 并自定义：

```json
{
  "port": 8080,
  "beepbot_data_dir": "/data/beepbot",
  "database": {
    "type": "postgres",
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "password",
    "dbname": "beepbot",
    "sslmode": "disable"
  },
  "encryption": {
    "key": ""
  },
  "logging": {
    "level": "info",
    "file": "",
    "format": "json",
    "qq": {
      "level": "info",
      "file": "./logs/qq_channel.log"
    }
  }
}
```

### 配置字段

- **port**: API 服务端口
- **beepbot_data_dir**: BeepBot 公共数据目录，用于存储全局技能、共享数据等
- **database**: 数据库配置
  - `type`: 数据库类型 (postgres)
  - `host`: 数据库主机
  - `port`: 数据库端口
  - `user`: 数据库用户
  - `password`: 数据库密码
  - `dbname`: 数据库名称
  - `sslmode`: SSL 模式
- **encryption**: 加密配置
  - `key`: Base64 编码的加密密钥（可选，不提供则自动生成）
- **logging**: 日志配置
  - `level`: 日志级别 (debug, info, warn, error)
  - `file`: 日志文件路径
  - `format`: 日志格式 (json, text)
  - `qq`: QQ 渠道专用日志设置

## 架构

### 运行模式

后端仅支持 API 模式运行，提供 REST API 服务供前端仪表板调用。

### 核心模块

#### Agent（智能体）
智能体是系统的核心，每个智能体可以：
- 绑定一个 LLM 提供商和模型
- 配置系统提示词、温度、最大迭代次数等参数
- 关联多个技能
- 绑定多个机器人渠道
- 配置为可调用（Callable）作为子智能体被其他智能体调用
- 配置工具访问权限（使用所有工具或指定工具）

#### Provider（模型供应商）
支持多种 LLM 提供商：
- OpenAI
- DashScope（阿里云）
- Ollama（本地部署）

#### Bot（机器人）
支持多种消息平台：
- QQ 机器人（腾讯官方 SDK）
- 飞书机器人

#### Cron（定时任务）
支持基于 Cron 表达式的定时任务：
- 关联智能体
- 定时触发智能体执行指定消息

#### Skill（技能）
技能系统允许为智能体定义可复用的技能指令：
- 存储在数据库中
- 支持版本管理
- 可关联到多个智能体

### 消息流

1. **入站消息**: 渠道（QQ/飞书）接收消息并发布到 MessageBus
2. **代理处理**: AgentRunner 消费消息、管理会话、协调 LLM 调用
3. **工具执行**: 代理可在对话期间调用工具（shell、文件系统、TODO 管理、定时任务）
4. **出站消息**: 响应发布到 MessageBus 并分发到相应渠道

### 代理循环 (ReAct 模式)

代理遵循 ReAct（推理 + 行动）模式：

1. 接收用户消息并添加到会话历史
2. 使用系统提示词和对话历史构建上下文
3. 调用带有可用工具的 LLM
4. 如果请求工具调用：
   - 执行带有安全检查的工具
   - 将结果添加到对话
   - 从步骤 3 重复（最多 max_iterations 次）
5. 向用户返回最终响应

### 会话管理

会话通过 `{agent_id}:{bot_id}:{channel}:{group_id}:{user_id}` 标识。每个会话维护：
- 对话历史（滑动窗口）
- 创建/更新时间戳
- 可选摘要（供上下文压缩使用）

会话持久化到数据库，支持：
- 上下文窗口管理
- Token 用量统计
- 上下文压缩（当达到阈值时）

### 工具系统

工具实现 `Tool` 接口：
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]any
    Execute(ctx context.Context, params map[string]any) (string, error)
}
```

内置工具：
- `read_file`: 从工作目录读取文件内容
- `write_file`: 向工作目录写入文件
- `edit_file`: 编辑文件（精确搜索替换）
- `list_dir`: 列出目录内容
- `shell`: 执行 shell 命令（有限制）
- `system_info`: 获取系统信息
- `todo`: 任务管理工具（添加、列出、更新、删除、清除任务）
- `cron`: 定时任务管理工具
- `subagent_xxx`: 子智能体调用工具（动态生成，xxx 为智能体名称）

### 子智能体系统 (Sub-Agent)

子智能体系统允许智能体调用其他智能体作为工具，实现任务分解和专业分工。

#### 核心概念

**可调用智能体 (Callable Agent)**：
- 智能体可配置 `callable=true` 标记为可被其他智能体调用
- `callable_description` 字段提供工具描述，帮助 LLM 理解用途
- 只有活跃状态的可调用智能体会被注册为工具

**工具命名**：
- 子智能体工具名称格式：`subagent_{智能体名称}`
- 例如：智能体 "code-reviewer" 的工具名为 `subagent_code-reviewer`

#### 工作流程

```
父智能体                    子智能体执行器                 子智能体
   │                              │                          │
   │──调用 subagent_xxx 工具────→│                          │
   │                              │──创建 InMemorySession──→│
   │                              │──创建 Runner (禁用递归)─→│
   │                              │──运行 AgentLoop────────→│
   │                              │                          │
   │                              │←──返回执行结果──────────│
   │←──返回工具结果──────────────│                          │
```

#### 核心组件

**SubAgentExecutor** (`agent/subagent_executor.go`)：
- 负责创建和执行子智能体
- 创建内存会话（不持久化到数据库）
- 使用 CollectorHook 收集执行结果

**SubAgentTool** (`tool/subagent.go`)：
- 实现工具接口，动态生成工具定义
- 参数：`task` (必需) - 任务描述，`context` (可选) - 上下文信息
- 调用执行器运行子智能体

**OutputHook 接口** (`agent/hooks.go`)：
```go
type OutputHook interface {
    OnError(ctx context.Context, err error)
    OnResponse(ctx context.Context, content string)
    OnIntermediateContent(ctx context.Context, content string)
}
```
- `BusOutputHook`: 对话场景，通过 MessageBus 发送消息给用户
- `CollectorHook`: 子智能体场景，仅收集结果不发送消息

**InMemorySession** (`session/memory.go`)：
- 子智能体使用的内存会话实现
- 不持久化到数据库
- 不支持压缩，不追踪 Token 用量

#### 安全机制

**递归防护**：
- 子智能体执行时 `allowSubAgents=false`，禁止继续调用其他子智能体
- 防止无限递归和资源耗尽

**工作目录继承**：
- 子智能体继承父智能体的 `working_dir`
- 确保文件操作在相同的安全边界内

**会话隔离**：
- 子智能体使用独立的内存会话
- 不影响父智能体的会话历史

#### 配置示例

智能体配置（设为可调用）：
```json
{
  "name": "code-reviewer",
  "callable": true,
  "callable_description": "代码审查专家，可以审查代码质量、发现潜在问题",
  "system_prompt": "你是一个专业的代码审查专家...",
  ...
}
```

父智能体调用示例：
```
父智能体决定需要代码审查 → 调用 subagent_code-reviewer 工具 →
传入任务描述和代码上下文 → 子智能体执行并返回结果
```

### 技能系统

技能系统允许为智能体定义可复用的技能指令。技能存储在数据库中，通过文件系统管理：

每个技能是一个包含 `SKILL.md` 文件的目录，格式如下：
```markdown
# 技能名称
技能简短描述（第二行到空行之前）

详细指令内容...
```

### 渠道系统

渠道实现不同消息平台的 `Channel` 接口：
- `qq`: 通过 WebSocket 的 QQ 机器人（腾讯官方 SDK）
- `feishu`: 飞书机器人（WebSocket 长连接）
- `system`: 用于定时任务的内部系统渠道

### 前端架构

前端采用 Vue 3 + TypeScript 技术栈：
- **路由**: 使用 Vue Router 管理页面导航
- **状态管理**: 使用 Pinia 管理全局状态
- **构建**: 使用 Vite 进行开发和生产构建
- **代码规范**: ESLint + Oxlint 进行代码检查，Prettier 进行格式化

#### 页面布局

前端采用经典的左侧导航 + 顶部栏 + 主内容区布局：

```
┌─────────────────────────────────────────────────────────────────┐
│  Header (64px)                                                  │
│  [☰ 折叠按钮] [BeepBot Logo]              [🌙 深浅色开关]       │
├────────────┬────────────────────────────────────────────────────┤
│  Sidebar   │                    Main Content                    │
│  (200px)   │                                                    │
│            │  ┌─────────────────────────────────────────────┐   │
│  🤖 智能体  │  │                                             │   │
│  🔌 供应商  │  │              路由页面内容                    │   │
│  💬 机器人  │  │              (卡片列表展示)                  │   │
│  ⏰ 定时任务│  │                                             │   │
│  📚 技能   │  │                                             │   │
│  ⚙️ 设置   │  │                                             │   │
│            │  └─────────────────────────────────────────────┘   │
└────────────┴────────────────────────────────────────────────────┘
```

- **Header**: 头部栏，包含折叠按钮、Logo、深浅色主题切换开关
- **Sidebar**: 侧边导航栏，支持展开/折叠，包含六个导航项
- **Main Content**: 主内容区，展示路由页面内容

#### 主题系统

前端支持深色/浅色主题切换：
- 使用 CSS 变量定义主题色
- 主题状态通过 Pinia Store 管理
- 主题偏好保存到 localStorage
- Ant Design Vue 组件主题自动跟随

#### 页面组件

| 路由 | 页面 | 说明 |
|------|------|------|
| `/agents` | AgentList.vue | 智能体列表，卡片风格展示 |
| `/agents/:id/edit` | AgentConfig.vue | 智能体配置页 |
| `/agents/:id/logs` | AgentLogs.vue | 智能体日志页 |
| `/agents/:id/monitor` | AgentMonitor.vue | 智能体监控页 |
| `/providers` | ProviderList.vue | 模型供应商列表，卡片风格展示 |
| `/bots` | BotList.vue | IM机器人列表，卡片风格展示 |
| `/crons` | CronList.vue | 定时任务列表页 |
| `/skills` | SkillList.vue | 技能列表页 |
| `/skills/:id` | SkillDetail.vue | 技能详情页 |
| `/settings` | Settings.vue | 全局设置，表单风格 |

## 安全注意事项

### Shell 命令安全
- 禁止命令列表防止危险操作（sudo、rm、chmod 等）
- 命令按禁止前缀检查
- 超时防止长时间运行命令
- 工作目录隔离限制文件系统访问

### 文件系统隔离
- 所有文件操作限制在配置的 `working_dir` 或 `beepbot_data_dir` 内
- 路径遍历保护解析和验证路径
- 访问工作目录外路径的尝试被拒绝

### API 密钥安全
- Provider 的 API 密钥加密存储在数据库中
- API 响应中返回脱敏后的密钥
- 加密密钥可配置或自动生成

### QQ 机器人安全
- 使用 OAuth2 基于令牌的认证
- 令牌自动刷新
- WebSocket 连接安全管理

### 飞书机器人安全
- 支持加密密钥验证消息
- 支持用户和群组白名单

## 开发指南

### 后端代码风格
- 遵循标准 Go 约定
- 使用有意义的变量名（英文）
- 为导出函数和类型添加注释
- 使用 `log/slog` 进行结构化日志

### 前端代码风格
- 遵循 Vue 3 组合式 API 风格
- 使用 TypeScript 进行类型安全
- 使用 ESLint 和 Prettier 保持代码一致性

### 添加新工具
1. 在 `beepbot/internal/tool/` 中创建新文件
2. 实现 `Tool` 接口
3. 在 `agent.NewApiRunner()` 中注册
4. 如需要添加配置

### 添加新渠道
1. 在 `beepbot/internal/channel/` 中创建新文件
2. 实现 `Channel` 接口
3. 在 `ChannelFactoryRegistry` 中注册工厂函数
4. 在 `beepbot/internal/types/bot.go` 中添加平台类型

### 添加新 LLM 提供商
1. 在 `beepbot/internal/provider/` 中创建提供商实现
2. 实现 `types.LLMProvider` 接口
3. 在 `provider.CreateLLMProviderFromApi()` 中注册
4. 在 `beepbot/internal/types/provider.go` 中添加提供商类型

## 测试

目前，项目不包含自动化测试。添加测试时：
- 后端：将测试文件放在源文件旁边，后缀为 `_test.go`
- 前端：使用 Vitest 进行单元测试
- 使用表驱动测试处理多个测试用例
- 模拟外部依赖（LLM API、数据库）

## 部署

### 后端

#### Windows
```bash
cd beepbot
go build -ldflags "-s -w" -o beepbot-api.exe ./cmd/api
```

#### Linux
```bash
cd beepbot
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o beepbot-api ./cmd/api
```

### 前端

```bash
cd dashboard
pnpm build
```

构建产物将生成在 `dashboard/dist/` 目录。

### Docker (未来)
可以添加 Dockerfile 用于容器化部署。

## 重要说明

1. **工作目录**: 每个智能体有自己的 `working_dir` 隔离文件操作。确保它存在并具有适当的权限。

2. **数据目录**: `beepbot_data_dir` 用于存储全局技能、共享数据等公共资源。文件工具可以访问此目录。

3. **QQ 机器人**: 需要腾讯 QQ 机器人平台的有效 AppID 和 AppSecret。

4. **飞书机器人**: 需要飞书开放平台的有效 AppID 和 AppSecret。

5. **API 密钥**: 切勿将 API 密钥提交到版本控制。在生产环境中使用环境变量或安全密钥管理。

6. **TODO 管理**: TODO 工具将任务列表存储在工作目录的 `TODO.md` 文件中，支持添加、列出、更新、删除和清除操作。

7. **前后端分离**: 前端 dashboard 是独立的 Vue 项目，需要单独安装依赖和构建。

8. **数据库**: 使用 PostgreSQL 数据库，确保数据库已创建并可访问。

9. **定时任务**: 定时任务通过 Cron 表达式配置，由后端调度器自动触发。

10. **子智能体调用**: 子智能体使用内存会话，不持久化到数据库。子智能体继承父智能体的工作目录，且不能递归调用其他子智能体。

11. **工具权限**: 智能体可以配置 `use_all_tools` 来决定使用所有工具还是仅使用指定工具。

## 许可证

请参阅仓库获取许可证信息。
