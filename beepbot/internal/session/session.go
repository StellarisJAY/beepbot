package session

import "github.com/StellarisJAY/beepbot/internal/types"

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

	// Compress 压缩历史消息，使用配置的 compressionKeepSize 保留数量
	// 返回被淘汰的消息，用于生成摘要
	Compress() []types.Message

	// GetTokenUsage 返回当前 token 用量
	GetTokenUsage() int64

	// GetMaxTokens 返回 token 上限
	GetMaxTokens() int64

	GetSessionKey(sessionType types.SessionType, channelID string, chatID string, userID string) string

	// GetSessionID 返回会话 ID
	GetSessionID() string

	// GetCronJobID 返回定时任务 ID（仅定时任务会话有值）
	GetCronJobID() *string

	// GetIMContext 返回 IM 会话上下文
	GetIMContext() *types.IMSessionContext
}
