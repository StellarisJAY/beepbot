# BeepBot Project Guide

## Project Overview

BeepBot is an AI-powered chatbot framework built in Go, designed to integrate with multiple messaging platforms and leverage Large Language Models (LLMs) for intelligent conversations. The project features a modular architecture supporting extensible channels, tools, and memory systems.

### Key Technologies
- **Language**: Go 1.24.5
- **Database**: SQLite with GORM ORM
- **LLM Providers**: OpenAI API compatible services (OpenAI, Alibaba Cloud DashScope)
- **Messaging Platforms**: QQ Bot (via Tencent BotGo SDK), Console
- **Vector Database**: Milvus (optional, for long-term memory)
- **Architecture Pattern**: Event-driven with message bus

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      BeepBot Server                          │
│                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Channels   │───▶│ Message Bus  │───▶│    Agent     │  │
│  │  (QQ,Console)│    │  (Inbound/   │    │   Runner     │  │
│  └──────────────┘    │   Outbound)  │    └──────────────┘  │
│                      └──────────────┘            │           │
│                                                  ▼           │
│                                        ┌──────────────┐     │
│                                        │ LLM Provider │     │
│                                        │ (OpenAI/API) │     │
│                                        └──────────────┘     │
│                                                  │           │
│                    ┌─────────────────────────────┼───────┐  │
│                    ▼                             ▼       ▼  │
│            ┌──────────────┐            ┌──────────┐ ┌─────┐│
│            │    Tools     │            │ Session  │ │Memory││
│            │  (Registry)  │            │ Manager  │ │ Mgr ││
│            └──────────────┘            └──────────┘ └─────┘│
│                                                  │           │
│                                        ┌─────────▼───────┐  │
│                                        │  SQLite Database│  │
│                                        └─────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Getting Started

### Prerequisites
- Go 1.24.5 or later
- SQLite3
- (Optional) Milvus vector database for long-term memory
- API keys for LLM providers (OpenAI or DashScope)
- QQ Bot credentials (if using QQ channel)

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd beepbot
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create a configuration file (`config.json`):
   ```json
   {
     "providers": {
       "openai": {
         "api_key": "your-openai-api-key",
         "base_url": "https://api.openai.com/v1"
       },
       "dashscope": {
         "api_key": "your-dashscope-api-key",
         "base_url": "https://dashscope.aliyuncs.com/api/v2/apps/protocols/compatible-mode/v1"
       }
     },
     "agent": {
       "temperature": 0.7,
       "max_tool_iterations": 50,
       "max_tokens": 4096,
       "model": "openai/gpt-4",
       "system_prompt": "You are a helpful assistant."
     },
     "memory": {
       "window_size": 20,
       "flash": {}
     },
     "channel": {
       "console": {},
       "qq": {
         "app_id": "your-qq-app-id",
         "app_secret": "your-qq-app-secret"
       }
     }
   }
   ```

4. Run the server:
   ```bash
   go run cmd/server/main.go -config config.json
   ```

### Basic Usage

The bot supports multiple messaging channels:

- **Console Channel**: Direct terminal interaction for testing
- **QQ Channel**: Integration with QQ messaging platform

Messages flow through the system:
1. Channel receives message → Message Bus (Inbound)
2. Agent consumes message → Processes with LLM
3. LLM response (with optional tool calls) → Agent executes tools
4. Final response → Message Bus (Outbound) → Channel sends to user

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/provider/...
```

## Project Structure

```
beepbot/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── agent/
│   │   ├── agent.go            # Agent runner and message loop
│   │   └── context.go          # Agent context management
│   ├── channel/
│   │   ├── channel.go          # Channel interface and manager
│   │   ├── message_bus.go      # Inbound/outbound message routing
│   │   └── qq_channel.go       # QQ bot channel implementation
│   ├── config/
│   │   └── config.go           # Configuration structures and loader
│   ├── db/
│   │   └── sqlite.go           # Database operations with GORM
│   ├── memory/
│   │   ├── base.go             # Memory manager factory
│   │   ├── flash.go            # In-memory storage (ephemeral)
│   │   └── milvus.go           # Milvus vector DB integration
│   ├── provider/
│   │   ├── base.go             # LLM provider factory
│   │   └── openai.go           # OpenAI API implementation
│   ├── session/
│   │   └── session.go          # Session management
│   ├── tool/
│   │   ├── base.go             # Tool interface
│   │   ├── tool_registry.go    # Tool registration and execution
│   │   ├── time.go             # Time utility tool
│   │   ├── file_system.go      # File operations tool
│   │   └── message.go          # Messaging tool
│   └── types/
│       ├── provider.go         # LLM types (Message, ToolCall, etc.)
│       ├── session.go          # Session types
│       ├── memory.go           # Memory types
│       └── cron.go             # Scheduled task types
├── design/
│   └── flow.drawio              # Architecture diagram
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── config.json                  # Configuration file (create from template)
└── beepbot.db                   # SQLite database (auto-created)
```

### Key Files and Their Roles

| File | Purpose |
|------|---------|
| `cmd/server/main.go` | Main entry point; initializes config, database, channels, and agent |
| `internal/agent/agent.go` | Core agent logic with message loop and tool execution cycle |
| `internal/channel/message_bus.go` | Message routing between channels and agent |
| `internal/provider/openai.go` | OpenAI API client implementation |
| `internal/session/session.go` | Conversation session state management |
| `internal/tool/tool_registry.go` | Tool registration and execution engine |
| `internal/db/sqlite.go` | Database layer for sessions and messages |

### Important Configuration

Configuration is loaded from `config.json` with the following main sections:

- **providers**: API keys and endpoints for OpenAI/DashScope
- **agent**: Model settings, temperature, max tokens, system prompt
- **memory**: Memory type (flash/milvus) and window size
- **channel**: Channel-specific configurations (QQ app credentials)

## Development Workflow

### Coding Standards

- Follow standard Go conventions and idiomatic Go code
- Use meaningful variable and function names
- Add comments for exported functions and types
- Use `slog` package for structured logging
- Handle errors explicitly; don't ignore them
- Use context for cancellation and timeouts

### Project Conventions

1. **Package Organization**: Internal packages are not exported; use interfaces for abstraction
2. **Error Handling**: Return errors to callers; use structured logging for errors
3. **Configuration**: Use JSON config file; avoid hardcoding values
4. **Testing**: Write unit tests for critical components, especially providers and tools
5. **Logging**: Use `slog` with context keys for structured, leveled logging

### Build and Deployment

```bash
# Build for current platform
go build -o beepbot cmd/server/main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o beepbot-linux cmd/server/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o beepbot.exe cmd/server/main.go
```

### Adding New Tools

1. Create a new file in `internal/tool/` (e.g., `my_tool.go`)
2. Implement the `Tool` interface:
   ```go
   type Tool interface {
       Name() string
       Description() string
       Parameters() map[string]any
       Execute(ctx context.Context, params map[string]any) (string, error)
   }
   ```
3. Register the tool in `internal/agent/agent.go`:
   ```go
   toolRegistry.Register(&MyTool{})
   ```

### Adding New Channels

1. Create a new file in `internal/channel/` (e.g., `discord_channel.go`)
2. Implement the `Channel` interface:
   ```go
   type Channel interface {
       ID() string
       IsAvailable() bool
       Send(ctx context.Context, message OutboundMessage) error
       IsAllowed(senderID string) bool
       Start(ctx context.Context) error
       Stop()
   }
   ```
3. Add channel config in `internal/config/config.go`
4. Initialize in `ChannelManager.InitChannels()`

### Adding New LLM Providers

1. Create a new file in `internal/provider/`
2. Implement the `LLMProvider` interface:
   ```go
   type LLMProvider interface {
       Chat(ctx context.Context, messages []Message, model string, options ChatOptions) (*LLMResponse, error)
   }
   ```
3. Add provider detection in `CreateLLMProvider()` function
4. Add provider config in `internal/config/config.go`

## Key Concepts

### Agent Loop

The core processing cycle in `internal/agent/agent.go`:

1. **Consume Message**: Receive inbound message from message bus
2. **Session Lookup**: Find or create conversation session
3. **Context Building**: Build message history with system prompts
4. **LLM Call**: Send messages to LLM provider
5. **Tool Execution**: If LLM returns tool calls, execute them
6. **Response**: Send final response back through message bus
7. **Iteration**: Repeat until no tool calls or max iterations reached

### Message Bus

Event-driven architecture for decoupling channels from agent:

- **Inbound Queue**: Messages from users (channel → agent)
- **Outbound Queue**: Responses to users (agent → channel)
- Channels publish inbound messages and consume outbound messages
- Agent consumes inbound messages and publishes outbound messages

### Session Management

Sessions maintain conversation state:

- **Session Key**: `{channelID}:{userID}` format
- **History**: Windowed message history (configurable size)
- **Summary**: Optional conversation summary for context compression
- **Persistence**: Stored in SQLite database via GORM

### Memory Systems

Two memory implementations:

1. **Flash Memory**: In-memory, ephemeral storage (lost on restart)
2. **Milvus Memory**: Vector database for long-term semantic search

Memory is used for:
- Conversation context
- Long-term knowledge retrieval
- User preferences

### Tool System

Extensible function calling:

- **Tool Registry**: Central registry for available tools
- **Tool Interface**: Standard interface for all tools
- **Execution Context**: Tools receive channel and user context
- **Examples**: Time, FileSystem (read/write), Message sending

### LLM Provider Abstraction

Unified interface for multiple LLM providers:

- OpenAI API compatible (OpenAI, DashScope)
- Supports function calling (tools)
- Token usage tracking
- Temperature and max tokens configuration

## Common Tasks

### Testing a New Tool Locally

1. Start the bot with console channel enabled:
   ```json
   {
     "channel": {
       "console": {}
     }
   }
   ```

2. Run the server: `go run cmd/server/main.go`

3. Type messages in the console to test

4. Check logs for tool execution results

### Debugging Message Flow

1. Enable debug logging (slog level)
2. Check message bus consumption in `agent.go`
3. Verify channel is publishing to inbound queue
4. Check outbound queue dispatch in `channel.go`

### Updating Database Schema

1. Modify types in `internal/types/`
2. GORM auto-migration runs on startup
3. For complex migrations, create manual migration scripts

### Adding a New API Endpoint

The current architecture doesn't have an HTTP API layer. To add one:

1. Create `internal/api/` package
2. Use standard `net/http` or a framework like `gin`
3. Integrate with existing services (agent, session manager)
4. Start HTTP server in `main.go`

## Troubleshooting

### Common Issues

#### Bot doesn't respond to messages
- **Check**: Channel is started and available (`IsAvailable()`)
- **Check**: Message bus is consuming messages
- **Check**: LLM provider API key is valid
- **Logs**: Look for errors in agent loop

#### Tool execution fails
- **Check**: Tool is registered in registry
- **Check**: Tool parameters match definition
- **Check**: Tool execution context (channel, userID)
- **Logs**: Check tool execution duration and errors

#### Database errors
- **Check**: SQLite file permissions (`beepbot.db`)
- **Check**: GORM migration logs on startup
- **Check**: Database file is not locked by another process

#### QQ channel connection issues
- **Check**: QQ app ID and secret are correct
- **Check**: Network connectivity to QQ servers
- **Check**: Bot has required permissions
- **Logs**: Look for WebSocket connection errors

### Debugging Tips

1. **Enable Verbose Logging**: 
   - Set log level to debug in code
   - Use `slog.SetLogLoggerLevel(slog.LevelDebug)`

2. **Inspect Message Flow**:
   - Add logging in message bus publish/consume
   - Check session history after each message

3. **Test LLM Provider**:
   ```bash
   go test -v ./internal/provider/ -run TestOpenAI
   ```

4. **Database Inspection**:
   ```bash
   sqlite3 beepbot.db
   .tables
   SELECT * FROM sessions;
   ```

5. **Tool Testing**:
   - Create unit tests for tools
   - Test with mock context

### Performance Considerations

- **Session Window Size**: Larger windows = more tokens per request
- **Tool Iterations**: Max 50 iterations by default (configurable)
- **Database**: SQLite may lock under high concurrency; consider PostgreSQL for production
- **Memory**: Flash memory is lost on restart; use Milvus for persistence

## References

### Documentation
- [Go Documentation](https://golang.org/doc/)
- [GORM Guide](https://gorm.io/docs/)
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)
- [Tencent BotGo SDK](https://github.com/tencent-connect/botgo)
- [Milvus Documentation](https://milvus.io/docs)

### Key Dependencies
- `gorm.io/gorm` - ORM library
- `gorm.io/driver/sqlite` - SQLite driver
- `github.com/openai/openai-go` - OpenAI Go SDK
- `github.com/tencent-connect/botgo` - QQ Bot SDK
- `github.com/google/uuid` - UUID generation

### Architecture Decisions
- **Why SQLite?**: Simple, embedded database for single-instance deployments
- **Why Message Bus?**: Decouples channels from agent, enables future scaling
- **Why Tool Registry?**: Extensible, type-safe function calling
- **Why Session-per-User?**: Isolated conversation contexts

---

**Note**: This guide is auto-generated based on codebase analysis. Some sections may need verification or updates as the project evolves. Please review and update this file as needed.
