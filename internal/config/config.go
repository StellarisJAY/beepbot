package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	ProvidersConfig ProvidersConfig `json:"providers"`
	AgentConfig     AgentConfig     `json:"agent"`
	MemoryConfig    MemoryConfig    `json:"memory"`
	ChannelConfig   ChannelConfig   `json:"channel"`
	Logging         Logging         `json:"logging"`
	BuiltinTools    BuiltinTools    `json:"builtin_tools"`
	HeartBeat       HeartBeatConfig `json:"heart_beat"`
	Skills          SkillsConfig    `json:"skills"`
}

type ProvidersConfig struct {
	OpenAI    ProviderConfig `json:"openai"`
	DashScope ProviderConfig `json:"dashscope"`
}

type ProviderConfig struct {
	APIKey     string `json:"api_key"`
	BaseURL    string `json:"base_url"`
	AuthMethod string `json:"auth_method,omitempty"`
}

type AgentConfig struct {
	Temperature   float32 `json:"temperature"`    // 温度
	MaxIterations int     `json:"max_iterations"` // 单个agent循环的最大迭代次数
	MaxTokens     int64   `json:"max_tokens"`     // 最大输出tokens
	Provider      string  `json:"provider"`       // 模型供应商
	Model         string  `json:"model"`          // 模型ID
	SystemPrompt  string  `json:"system_prompt"`  // 系统提示词
	WorkingDir    string  `json:"working_dir"`    // 智能体工作的目录，与主机文件系统隔离，防止智能体随便访问文件
}

type MemoryConfig struct {
	WindowSize int                 `json:"window_size"`
	Flash      *FlashMemoryConfig  `json:"flash,omitempty"`
	Milvus     *MilvusMemoryConfig `json:"milvus,omitempty"`
}

type FlashMemoryConfig struct {
}

type MilvusMemoryConfig struct {
	Endpoint string `json:"endpoint"`
}

type ChannelConfig struct {
	Console *ConsoleChannelConfig `json:"console,omitempty"`
	QQ      *QQChannelConfig      `json:"qq,omitempty"`
}

type ConsoleChannelConfig struct {
}

type QQChannelConfig struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type Logging struct {
	Level  string         `json:"level"`
	File   string         `json:"file"`
	Format string         `json:"format"`
	QQ     ChannelLogging `json:"qq"`
}

type ChannelLogging struct {
	Level string `json:"level"`
	File  string `json:"file"`
}

type BuiltinTools struct {
	Shell ShellTool `json:"shell"`
}

type ShellTool struct {
	Enable            bool     `json:"enable"`
	ForbiddenCommands []string `json:"forbidden_commands"`
	Timeout           string   `json:"timeout"`
}

type SkillsConfig struct {
	ConfigFile string `json:"config_file"`
}

type HeartBeatConfig struct {
	Interval      string `json:"interval"`        // 心跳间隔
	Enable        bool   `json:"enable"`          // 是否开启心跳
	HeartBeatFile string `json:"heart_beat_file"` // 心跳任务文件目录
}

func FromConfigFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file error: %w", err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read config file error: %w", err)
	}

	config := Config{}
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("parse config file error: %w", err)
	}
	return &config, nil
}
