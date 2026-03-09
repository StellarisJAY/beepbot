package service

import (
	"encoding/json"
	"time"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// SessionListItem 会话列表项
type SessionListItem struct {
	ID           string            `json:"id"`
	Key          string            `json:"key"`
	BotID        string            `json:"bot_id"`
	BotName      string            `json:"bot_name"`
	BotPlatform  types.BotPlatform `json:"bot_platform"`
	Summary      string            `json:"summary"`
	MessageCount int64             `json:"message_count"`
	TotalTokens  int64             `json:"total_tokens"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// SessionService 会话服务
type SessionService struct {
	sessionRepo repository.SessionRepository
	botRepo     repository.BotRepository
}

// NewSessionService 创建会话服务
func NewSessionService(sessionRepo repository.SessionRepository, botRepo repository.BotRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		botRepo:     botRepo,
	}
}

// GetSessionsByAgent 获取智能体的会话列表
func (s *SessionService) GetSessionsByAgent(agentID string, page, pageSize int) ([]SessionListItem, int64, error) {
	// 1. 查询会话列表
	sessions, total, err := s.sessionRepo.GetSessionsByAgentID(agentID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 2. 收集所有 BotID 和 SessionID
	botIDs := make([]string, 0, len(sessions))
	sessionIDs := make([]string, 0, len(sessions))
	for _, session := range sessions {
		if session.BotID != "" {
			botIDs = append(botIDs, session.BotID)
		}
		sessionIDs = append(sessionIDs, session.ID)
	}

	// 3. 批量查询 Bot 信息
	botMap := make(map[string]*types.Bot)
	if len(botIDs) > 0 {
		bots, err := s.botRepo.FindByIDs(botIDs)
		if err == nil {
			for i := range bots {
				botMap[bots[i].ID] = &bots[i]
			}
		}
	}

	// 4. 批量获取会话统计信息
	statsMap, err := s.sessionRepo.GetSessionStats(sessionIDs)
	if err != nil {
		// 统计信息获取失败不影响主流程，记录日志即可
		statsMap = make(map[string]repository.SessionStats)
	}

	// 5. 组装响应数据
	items := make([]SessionListItem, 0, len(sessions))
	for _, session := range sessions {
		item := SessionListItem{
			ID:        session.ID,
			Key:       session.Key,
			BotID:     session.BotID,
			Summary:   session.Summary,
			CreatedAt: session.CreatedAt,
			UpdatedAt: session.UpdatedAt,
		}
		if bot, ok := botMap[session.BotID]; ok {
			item.BotName = bot.Name
			item.BotPlatform = bot.Platform
		}
		// 填充统计信息
		if stats, ok := statsMap[session.ID]; ok {
			item.MessageCount = stats.MessageCount
			item.TotalTokens = stats.TotalTokens
		}
		items = append(items, item)
	}

	return items, total, nil
}

// MessageListItem 消息列表项
type MessageListItem struct {
	ID           string           `json:"id"`
	Role         string           `json:"role"`
	Content      string           `json:"content"`
	ToolCalls    []types.ToolCall `json:"tool_calls,omitempty"`
	ToolCallID   string           `json:"tool_call_id,omitempty"`
	Name         string           `json:"name,omitempty"`
	InputTokens  int64            `json:"input_tokens,omitempty"`
	OutputTokens int64            `json:"output_tokens,omitempty"`
	TotalTokens  int64            `json:"total_tokens,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}

// MessagesResponse 消息列表响应
type MessagesResponse struct {
	Messages []MessageListItem `json:"messages"`
	Total    int64             `json:"total"`
	HasMore  bool              `json:"has_more"`
}

// GetSessionMessages 获取会话消息列表
func (s *SessionService) GetSessionMessages(sessionID string, beforeID string, limit int) (*MessagesResponse, error) {
	// 默认限制
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 查询消息
	messages, total, err := s.sessionRepo.GetMessagesPaginated(sessionID, beforeID, limit)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	items := make([]MessageListItem, 0, len(messages))
	for _, msg := range messages {
		item := MessageListItem{
			ID:           msg.ID,
			Role:         msg.Role,
			Content:      msg.Content,
			ToolCallID:   msg.ToolCallID,
			InputTokens:  msg.InputTokens,
			OutputTokens: msg.OutputTokens,
			TotalTokens:  msg.TotalTokens,
			CreatedAt:    msg.CreatedAt,
		}
		// 解析 ToolCalls JSON 字符串
		if msg.ToolCalls != "" {
			var toolCalls []types.ToolCall
			if err := json.Unmarshal([]byte(msg.ToolCalls), &toolCalls); err == nil {
				item.ToolCalls = toolCalls
			}
		}
		items = append(items, item)
	}

	// 判断是否还有更多消息
	hasMore := int64(len(messages)) < total

	return &MessagesResponse{
		Messages: items,
		Total:    total,
		HasMore:  hasMore,
	}, nil
}
