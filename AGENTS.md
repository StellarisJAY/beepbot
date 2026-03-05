# BeepBot - AI 智能代理机器人

## 项目概述

BeepBot 是一个 AI 代理机器人项目，包含后端服务和前端仪表板。后端使用 Go 语言编写，通过多种消息渠道提供智能对话能力；前端使用 Vue 3 构建，提供管理界面。

后端支持 OpenAI 和 DashScope（阿里云）等 LLM 提供商，可通过 QQ 机器人 API 或控制台界面与用户交互。具有 ReAct 风格的代理循环和工具调用能力，可以执行 shell 命令、读写文件、管理任务列表、收集系统信息，帮助用户完成各种任务。

## 项目结构

```
beepbot/                          # 项目根目录
├── beepbot/                      # 后端 Go 项目
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go           # API 服务入口点
│   │   └── standalone/
│   │       └── main.go           # 独立模式入口点
│   ├── internal/
│   │   ├── agent/
│   │   │   ├── agent.go          # 核心代理循环和消息处理
│   │   │   ├── context.go        # 代理上下文管理
│   │   │   └── standalone.go     # 独立模式代理
│   │   ├── api/
│   │   │   ├── agent_handler.go  # Agent API 处理器
│   │   │   ├── provider_handler.go # Provider API 处理器
│   │   │   ├── response.go       # API 响应工具
│   │   │   └── router.go         # 路由定义
│   │   ├── channel/
│   │   │   ├── channel.go        # 渠道接口和管理器
│   │   │   ├── message_bus.go    # 入站/出站消息总线
│   │   │   ├── qq_channel.go     # QQ 机器人渠道实现
│   │   │   └── system_channel.go # 系统/内部渠道
│   │   ├── config/
│   │   │   ├── config.go         # 配置结构和加载
│   │   │   ├── api.go            # API 模式配置
│   │   │   └── standalone.go     # 独立模式配置
│   │   ├── crypto/
│   │   │   └── encryption.go     # 加密工具
│   │   ├── database/
│   │   │   └── database.go       # 数据库连接和管理
│   │   ├── heartbeat/
│   │   │   └── heartbeat.go      # 定期心跳任务调度器
│   │   ├── logger/
│   │   │   ├── logger.go         # 日志初始化
│   │   │   └── qq.go             # QQ 渠道专用日志
│   │   ├── mcp/
│   │   │   └── mcp.go            # MCP（模型上下文协议）- 占位符
│   │   ├── memory/
│   │   │   ├── base.go           # 内存管理器工厂
│   │   │   ├── flash.go          # 内存（闪存）存储
│   │   │   └── milvus.go         # Milvus 向量数据库（占位符）
│   │   ├── provider/
│   │   │   ├── base.go           # LLM 提供商工厂
│   │   │   └── openai.go         # OpenAI 兼容提供商实现
│   │   ├── repository/
│   │   │   ├── repository.go     # 仓储基类
│   │   │   ├── agent_repository.go # Agent 仓储
│   │   │   └── provider_repository.go # Provider 仓储
│   │   ├── service/
│   │   │   ├── agent_service.go  # Agent 服务层
│   │   │   └── provider_service.go # Provider 服务层
│   │   ├── session/
│   │   │   └── session.go        # 会话管理
│   │   ├── skill/
│   │   │   └── skill.go          # 技能管理系统
│   │   ├── tool/
│   │   │   ├── base.go           # 工具接口定义
│   │   │   ├── file_system.go    # 文件读/写/列表/编辑工具
│   │   │   ├── message.go        # 消息工具
│   │   │   ├── shell.go          # Shell 执行工具
│   │   │   ├── system.go         # 系统信息工具
│   │   │   ├── todo.go           # TODO 任务管理工具
│   │   │   └── tool_registry.go  # 工具注册表
│   │   └── types/
│   │       ├── agent.go          # Agent 相关类型
│   │       ├── base.go           # 基础类型
│   │       ├── cron.go           # Cron 相关类型
│   │       ├── memory.go         # 内存接口和类型
│   │       ├── provider.go       # LLM 提供商接口和消息类型
│   │       └── session.go        # 会话数据库模型
│   ├── config.example.json       # 配置示例
│   ├── go.mod                    # Go 模块定义
│   └── .gitignore                # Git 忽略文件
├── dashboard/                    # 前端 Vue 项目
│   ├── src/
│   │   ├── App.vue               # 根组件
│   │   ├── main.ts               # 应用入口
│   │   ├── assets/
│   │   │   └── styles/           # 样式文件
│   │   │       ├── variables.css # CSS 变量（主题色）
│   │   │       └── global.css    # 全局样式
│   │   ├── components/
│   │   │   └── layout/           # 布局组件
│   │   │       ├── AppLayout.vue # 整体布局容器
│   │   │       ├── Header.vue    # 头部栏组件
│   │   │       └── Sidebar.vue   # 侧边导航栏组件
│   │   ├── router/
│   │   │   └── index.ts          # Vue Router 配置
│   │   ├── stores/               # Pinia 状态管理
│   │   │   ├── theme.ts          # 主题状态
│   │   │   └── sidebar.ts        # 侧边栏折叠状态
│   │   └── views/                # 页面组件
│   │       ├── agents/
│   │       │   └── AgentList.vue # 智能体列表页
│   │       ├── providers/
│   │       │   └── ProviderList.vue # 模型供应商列表页
│   │       ├── bots/
│   │       │   └── BotList.vue   # IM机器人列表页
│   │       └── settings/
│   │           └── Settings.vue  # 全局设置页
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
- **LLM 提供商**: OpenAI API 兼容 (OpenAI, DashScope)
- **消息渠道**: QQ 机器人 (腾讯), 控制台
- **主要依赖**:
  - `github.com/tencent-connect/botgo` - QQ 机器人 SDK
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
- SQLite 支持

#### 构建
```bash
cd beepbot

# 构建 API 服务
go build -o beepbot-api.exe ./cmd/api

# 构建独立模式
go build -o beepbot-standalone.exe ./cmd/standalone
```

#### 运行
```bash
cd beepbot

# 运行 API 服务
./beepbot-api.exe

# 运行独立模式
./beepbot-standalone.exe -config /path/to/config.json
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
  "providers": {
    "dashscope": {
      "api_key": "your_api_key",
      "base_url": "https://dashscope.aliyuncs.com/api/v2/apps/protocols/compatible-mode/v1"
    },
    "openai": {
      "api_key": "your_api_key",
      "base_url": "https://api.openai.com/v1"
    }
  },
  "agent": {
    "provider": "dashscope",
    "model": "qwen3.5-plus",
    "temperature": 0.2,
    "max_iterations": 50,
    "max_tokens": 4096,
    "system_prompt": "你是一个有用的助手。",
    "working_dir": "D:/data/beepbot"
  },
  "memory": {
    "window_size": 20,
    "flash": {}
  },
  "channel": {
    "console": {},
    "qq": {
      "app_id": "your_app_id",
      "app_secret": "your_app_secret"
    }
  },
  "logging": {
    "level": "info",
    "file": "",
    "format": "json",
    "qq": {
      "level": "info",
      "file": "./logs/qq_channel.log"
    }
  },
  "builtin_tools": {
    "shell": {
      "enable": true,
      "forbidden_commands": ["sudo", "su", "rm", "chmod", "chown"],
      "timeout": "30s"
    }
  },
  "heart_beat": {
    "enable": true,
    "interval": "60s",
    "heart_beat_file": "./beepbot/HEARTBEAT.md"
  },
  "beepbot_data_dir": "./beepbot"
}
```

### 配置字段

- **providers**: LLM 提供商配置
  - `dashscope`: 阿里云 DashScope API 设置
  - `openai`: OpenAI API 设置

- **agent**: 代理行为设置
  - `provider`: 要使用的 LLM 提供商 ("openai" 或 "dashscope")
  - `model`: 模型标识符
  - `temperature`: 采样温度 (0.0 - 2.0)
  - `max_iterations`: 每次请求的最大工具调用迭代次数
  - `max_tokens`: 每次响应的最大 token 数
  - `system_prompt`: 代理的系统提示词
  - `working_dir`: 文件操作的工作目录（隔离）

- **memory**: 内存管理设置
  - `window_size`: 消息历史窗口大小
  - `flash`: 使用内存存储（易失性）
  - `milvus`: Milvus 向量数据库配置（可选）

- **channel**: 消息渠道设置
  - `console`: 控制台渠道（如果存在则始终启用）
  - `qq`: QQ 机器人渠道设置（需要 app_id 和 app_secret）

- **logging**: 日志配置
  - `level`: 日志级别 (debug, info, warn, error)
  - `format`: 日志格式 (json, text)
  - `qq`: QQ 渠道专用日志设置

- **builtin_tools**: 内置工具设置
  - `shell`: Shell 工具配置
    - `enable`: 启用 shell 工具
    - `forbidden_commands`: 禁止命令列表
    - `timeout`: 命令执行超时

- **heart_beat**: 定期心跳任务
  - `enable`: 启用心跳
  - `interval`: 心跳间隔（例如 "60s", "5m", "1h"）
  - `heart_beat_file`: 包含心跳任务的文件

- **beepbot_data_dir**: BeepBot 公共数据目录，用于存储全局技能、共享数据等

## 架构

### 运行模式

后端支持两种运行模式：

1. **API 模式** (`cmd/api`): 提供 REST API 服务，供前端仪表板调用
2. **独立模式** (`cmd/standalone`): 独立运行，支持控制台和 QQ 渠道

### 消息流

1. **入站消息**: 渠道（QQ/控制台）接收消息并发布到 MessageBus
2. **代理处理**: AgentRunner 消费消息、管理会话、协调 LLM 调用
3. **工具执行**: 代理可在对话期间调用工具（shell、文件系统、TODO 管理）
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

会话通过 `{channel}:{group_id}:{user_id}` 标识。每个会话维护：
- 对话历史（滑动窗口）
- 创建/更新时间戳
- 可选摘要（供将来使用）

会话存储在内存中，重启后历史记录将清空。

会话历史采用 FIFO 策略，当达到窗口大小时：
- 普通消息：删除最早的消息
- 包含 tool_calls 的 assistant 消息：同时删除对应的 tool 结果消息

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

### 技能系统

技能系统允许为智能体定义可复用的技能指令。技能存储在两个位置：
- **全局技能目录**: `{beepbot_data_dir}/skills/` - 所有工作空间共享
- **工作空间技能目录**: `{working_dir}/skills/` - 当前工作空间专用

每个技能是一个包含 `SKILL.md` 文件的目录，格式如下：
```markdown
# 技能名称
技能简短描述（第二行到空行之前）

详细指令内容...
```

### 渠道系统

渠道实现不同消息平台的 `Channel` 接口：
- `qq`: 通过 WebSocket 的 QQ 机器人（腾讯官方 SDK）
- `system`: 用于心跳和自动化的内部系统渠道

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
│  ⚙️ 设置   │  │                                             │   │
│            │  └─────────────────────────────────────────────┘   │
└────────────┴────────────────────────────────────────────────────┘
```

- **Header**: 头部栏，包含折叠按钮、Logo、深浅色主题切换开关
- **Sidebar**: 侧边导航栏，支持展开/折叠，包含四个导航项
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
| `/providers` | ProviderList.vue | 模型供应商列表，卡片风格展示 |
| `/bots` | BotList.vue | IM机器人列表，卡片风格展示 |
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

### QQ 机器人安全
- 使用 OAuth2 基于令牌的认证
- 令牌自动刷新
- WebSocket 连接安全管理的

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
3. 在 `agent.NewAgentRun()` 中注册
4. 如需要添加配置

### 添加新渠道
1. 在 `beepbot/internal/channel/` 中创建新文件
2. 实现 `Channel` 接口
3. 在 `ChannelManager.InitChannels()` 中注册
4. 在 `beepbot/internal/config/` 中添加配置结构

### 添加新 LLM 提供商
1. 在 `beepbot/internal/provider/` 中创建提供商实现
2. 实现 `types.LLMProvider` 接口
3. 在 `provider.CreateLLMProvider()` 中注册

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
go build -ldflags "-s -w" -o beepbot-standalone.exe ./cmd/standalone
```

#### Linux
```bash
cd beepbot
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o beepbot-api ./cmd/api
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o beepbot-standalone ./cmd/standalone
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

1. **工作目录**: 配置中的 `working_dir` 隔离文件操作。确保它存在并具有适当的权限。

2. **数据目录**: `beepbot_data_dir` 用于存储全局技能、共享数据等公共资源。文件工具可以访问此目录。

3. **QQ 机器人**: 需要腾讯 QQ 机器人平台的有效 AppID 和 AppSecret。消息将 `.` 替换为 `·` 以避免 URL 检测问题。

4. **心跳**: 心跳系统从 markdown 文件读取任务，并按配置间隔触发代理处理。适用于自动化和定时任务。

5. **API 密钥**: 切勿将 API 密钥提交到版本控制。在生产环境中使用环境变量或安全密钥管理。

6. **TODO 管理**: TODO 工具将任务列表存储在工作目录的 `TODO.md` 文件中，支持添加、列出、更新、删除和清除操作。

7. **前后端分离**: 前端 dashboard 是独立的 Vue 项目，需要单独安装依赖和构建。

## 许可证

请参阅仓库获取许可证信息。
