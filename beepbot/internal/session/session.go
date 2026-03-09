package session

import "github.com/StellarisJAY/beepbot/internal/types"

// 默认压缩比例阈值，当 tokenUsed 达到 maxTokens * compressionRatio 时触发压缩
const defaultCompressionRatio = 0.8
const defaultMaxTokens = 1000000

// 为了提升智能体的短期记忆力，将窗口大小设置为50
const defaultWindowSize = 50

type Session interface {
	// AppendMessage 添加历史消息到会话
	AppendMessage(message types.Message) bool

	// GetHistory 获取窗口的历史消息
	GetHistory() []types.Message

	// GetSummary 获取会话摘要
	GetSummary() string

	// NeedCompress 会话是否需要压缩上下文。在agentLoop中判断需要就会触发上下文压缩
	NeedCompress() bool

	// SetSummary 设置会话摘要，并清除压缩标记
	SetSummary(summary string)

	// ClearHistory 清空历史消息，用于压缩后重置
	ClearHistory()

	Compress(keepCount int) []types.Message

	// GetTokenUsage 返回当前 token 用量
	GetTokenUsage() int64

	// GetMaxTokens 返回 token 上限
	GetMaxTokens() int64

	GetSessionKey(channelID string, groupID string, userID string) string
}
