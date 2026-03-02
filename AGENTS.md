# BeepBot - AI 智能代理机器人

## 项目概述

BeepBot 是一个用 Go 语言编写的 AI 代理机器人，通过多种消息渠道提供智能对话能力。它支持 OpenAI 和 DashScope（阿里云）等 LLM 提供商，可通过 QQ 机器人 API 或控制台界面与用户交互。

该机器人具有 ReAct 风格的代理循环和工具调用能力，可以执行 shell 命令、读写文件、管理任务列表、收集系统信息，帮助用户完成各种任务。

## 技术栈

- **语言**: Go 1.24.5
- **LLM 提供商**: OpenAI API 兼容 (OpenAI, DashScope)
- **消息渠道**: QQ 机器人 (腾讯), 控制台
- **主要依赖**:
  - `github.com/tencent-connect/botgo` - QQ 机器人 SDK
  - `github.com/openai/openai-go/v3` - OpenAI 客户端
  - `gorm.io/gorm` - 数据库 ORM
  - `github.com/google/uuid` - UUID 生成
  - `github.com/shirou/gopsutil/v4` - 系统信息

## 项目结构

```
beepbot/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口点
├── internal/
│   ├── agent/
│   │   ├── agent.go             # 核心代理循环和消息处理
│   │   └── context.go           # 代理上下文管理
│   ├── channel/
│   │   ├── channel.go           # 渠道接口和管理器
│   │   ├── message_bus.go       # 入站/出站消息总线
│   │   ├── qq_channel.go        # QQ 机器人渠道实现
│   │   └── system_channel.go    # 系统/内部渠道
│   ├── config/
│   │   └── config.go            # 配置结构和加载
│   ├── heartbeat/
│   │   └── heartbeat.go         # 定期心跳任务调度器
│   ├── logger/
│   │   ├── logger.go            # 日志初始化
│   │   └── qq.go                # QQ 渠道专用日志
│   ├── memory/
│   │   ├── base.go              # 内存管理器工厂
│   │   ├── flash.go             # 内存（闪存）存储
│   │   └── milvus.go            # Milvus 向量数据库（占位符）
│   ├── mcp/
│   │   └── mcp.go               # MCP（模型上下文协议）- 占位符
│   ├── provider/
│   │   ├── base.go              # LLM 提供商工厂
│   │   └── openai.go            # OpenAI 兼容提供商实现
│   ├── session/
│   │   └── session.go           # 会话管理
│   ├── skill/
│   │   └── skill.go             # 技能管理系统
│   ├── tool/
│   │   ├── base.go              # 工具接口定义
│   │   ├── file_system.go       # 文件读/写/列表/编辑工具
│   │   ├── message.go           # 消息工具
│   │   ├── shell.go             # Shell 执行工具
│   │   ├── system.go            # 系统信息工具
│   │   ├── todo.go              # TODO 任务管理工具
│   │   └── tool_registry.go     # 工具注册表
│   └── types/
│       ├── cron.go              # Cron 相关类型
│       ├── memory.go            # 内存接口和类型
│       ├── provider.go          # LLM 提供商接口和消息类型
│       └── session.go           # 会话数据库模型
├── beepbot/
│   ├── HEARTBEAT.md             # 心跳任务定义文件
│   └── skills/                  # 全局技能目录
├── beepbot.db                   # SQLite 数据库文件（生成）
├── logs/                        # 日志文件目录（生成）
├── config.json                  # 配置文件
├── config.example.json          # 配置示例
├── go.mod                       # Go 模块定义
└── go.sum                       # Go 模块校验和
```

## 构建和运行

### 前置条件
- Go 1.24.5 或更高版本
- SQLite 支持

### 构建
```bash
cd cmd/server
go build -o beepbot.exe
```

或在项目根目录：
```bash
go build -o beepbot.exe ./cmd/server
```

### 运行
```bash
# 使用默认配置文件 (config.json)
./beepbot.exe

# 使用自定义配置文件
./beepbot.exe -config /path/to/config.json
```

## 配置

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

### 代码风格
- 遵循标准 Go 约定
- 使用有意义的变量名（英文）
- 为导出函数和类型添加注释
- 使用 `log/slog` 进行结构化日志

### 添加新工具
1. 在 `internal/tool/` 中创建新文件
2. 实现 `Tool` 接口
3. 在 `agent.NewAgentRun()` 中注册
4. 如需要添加配置

### 添加新渠道
1. 在 `internal/channel/` 中创建新文件
2. 实现 `Channel` 接口
3. 在 `ChannelManager.InitChannels()` 中注册
4. 在 `internal/config/` 中添加配置结构

### 添加新 LLM 提供商
1. 在 `internal/provider/` 中创建提供商实现
2. 实现 `types.LLMProvider` 接口
3. 在 `provider.CreateLLMProvider()` 中注册

## 测试

目前，项目不包含自动化测试。添加测试时：
- 将测试文件放在源文件旁边，后缀为 `_test.go`
- 使用表驱动测试处理多个测试用例
- 模拟外部依赖（LLM API、数据库）

## 部署

### Windows
```bash
go build -ldflags "-s -w" -o beepbot.exe ./cmd/server
```

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o beepbot ./cmd/server
```

### Docker (未来)
可以添加 Dockerfile 用于容器化部署。

## 重要说明

1. **工作目录**: 配置中的 `working_dir` 隔离文件操作。确保它存在并具有适当的权限。

2. **数据目录**: `beepbot_data_dir` 用于存储全局技能、共享数据等公共资源。文件工具可以访问此目录。

3. **QQ 机器人**: 需要腾讯 QQ 机器人平台的有效 AppID 和 AppSecret。消息将 `.` 替换为 `·` 以避免 URL 检测问题。

4. **心跳**: 心跳系统从 markdown 文件读取任务，并按配置间隔触发代理处理。适用于自动化和定时任务。

5. **API 密钥**: 切勿将 API 密钥提交到版本控制。在生产环境中使用环境变量或安全密钥管理。

6. **TODO 管理**: TODO 工具将任务列表存储在工作目录的 `TODO.md` 文件中，支持添加、列出、更新、删除和清除操作。

## 许可证

请参阅仓库获取许可证信息。
