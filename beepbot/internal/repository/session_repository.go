package repository

import (
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// SessionStats 会话统计信息
type SessionStats struct {
	SessionID    string
	MessageCount int64
	TotalTokens  int64
}

// SessionRepository 会话仓储接口
type SessionRepository interface {
	// 会话操作
	CreateSession(session *types.Session) error
	GetSessionByKey(key string) (*types.Session, error)
	GetSessionByID(id string) (*types.Session, error)
	UpdateSession(session *types.Session) error
	UpdateSessionSummary(sessionID string, summary string) error
	UpdateSessionContextTokens(sessionID string, tokens int64) error
	DeleteSession(id string) error

	// GetSessionsByAgentID 根据智能体ID分页查询会话列表（支持筛选）
	GetSessionsByAgentID(agentID string, page, pageSize int, query *types.SessionQuery) ([]types.Session, int64, error)

	// GetWebChatSessions 获取前端聊天会话列表
	GetWebChatSessions(agentID, userID string, page, pageSize int) ([]types.Session, int64, error)

	// GetSessionStats 批量获取会话统计信息（消息数量和token用量）
	GetSessionStats(sessionIDs []string) (map[string]SessionStats, error)

	// 消息操作
	AppendMessage(sessionID string, message *types.SessionMessage) error
	GetMessages(sessionID string, limit int) ([]types.SessionMessage, error)
	GetMessagesPaginated(sessionID string, beforeID string, limit int) ([]types.SessionMessage, int64, error)
	GetOldestMessages(sessionID string, limit int) ([]types.SessionMessage, error)
	DeleteMessages(sessionID string, messageIDs []string) error
	DeleteOldestMessages(sessionID string, count int) error
	ClearMessages(sessionID string) error
	CountMessages(sessionID string) (int64, error)

	// Token 用量
	GetTokenUsage(sessionID string) (int64, error)

	// 窗口内消息操作
	CountMessagesInWindow(sessionID string) (int64, error)
	GetOldestMessagesInWindow(sessionID string, limit int) ([]types.SessionMessage, error)
	GetMessagesInWindow(sessionID string, limit int) ([]types.SessionMessage, error)
	EvictMessages(sessionID string, messageIDs []string) error
	EvictMessagesInWindow(sessionID string) error
}

// SessionRepositoryImpl 会话仓储实现
type SessionRepositoryImpl struct {
	db *gorm.DB
}

// NewSessionRepository 创建会话仓储
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &SessionRepositoryImpl{db: db}
}

// CreateSession 创建会话
func (r *SessionRepositoryImpl) CreateSession(session *types.Session) error {
	return r.db.Create(session).Error
}

// GetSessionByKey 根据 Key 获取会话
func (r *SessionRepositoryImpl) GetSessionByKey(key string) (*types.Session, error) {
	var session types.Session
	err := r.db.Where("key = ?", key).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionByID 根据 ID 获取会话
func (r *SessionRepositoryImpl) GetSessionByID(id string) (*types.Session, error) {
	var session types.Session
	err := r.db.Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession 更新会话
func (r *SessionRepositoryImpl) UpdateSession(session *types.Session) error {
	return r.db.Save(session).Error
}

// UpdateSessionSummary 只更新会话的 summary 字段
func (r *SessionRepositoryImpl) UpdateSessionSummary(sessionID string, summary string) error {
	return r.db.Model(&types.Session{}).Where("id = ?", sessionID).Update("summary", summary).Error
}

// UpdateSessionContextTokens 更新会话的上下文 token 大小
func (r *SessionRepositoryImpl) UpdateSessionContextTokens(sessionID string, tokens int64) error {
	return r.db.Model(&types.Session{}).Where("id = ?", sessionID).Update("last_context_tokens", tokens).Error
}

// GetSessionsByAgentID 根据智能体ID分页查询会话列表（支持筛选）
func (r *SessionRepositoryImpl) GetSessionsByAgentID(agentID string, page, pageSize int, query *types.SessionQuery) ([]types.Session, int64, error) {
	var sessions []types.Session
	var total int64

	offset := (page - 1) * pageSize

	// 基础查询
	db := r.db.Model(&types.Session{}).Where("sessions.agent_id = ?", agentID)

	// 动态拼接筛选条件
	if query != nil {
		// 会话类型筛选
		if query.SessionType != "" {
			db = db.Where("sessions.session_type = ?", query.SessionType)
		}
		// 平台筛选（需要 JOIN bots 表）
		if query.Platform != "" {
			db = db.Joins("JOIN bots ON bots.id = sessions.bot_id").Where("bots.platform = ?", query.Platform)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按更新时间倒序
	err := db.Order("sessions.updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&sessions).Error

	return sessions, total, err
}

// DeleteSession 删除会话
func (r *SessionRepositoryImpl) DeleteSession(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先删除关联的消息
		if err := tx.Where("session_id = ?", id).Delete(&types.SessionMessage{}).Error; err != nil {
			return err
		}
		// 再删除会话
		return tx.Where("id = ?", id).Delete(&types.Session{}).Error
	})
}

// AppendMessage 添加消息
func (r *SessionRepositoryImpl) AppendMessage(sessionID string, message *types.SessionMessage) error {
	message.SessionID = sessionID
	return r.db.Create(message).Error
}

// GetMessages 获取会话消息（按创建时间升序）
func (r *SessionRepositoryImpl) GetMessages(sessionID string, limit int) ([]types.SessionMessage, error) {
	var messages []types.SessionMessage
	query := r.db.Where("session_id = ?", sessionID).Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&messages).Error
	return messages, err
}

// GetMessagesPaginated 分页获取消息（支持向上翻页）
// beforeID: 加载此ID之前的消息，为空则加载最新消息
// limit: 每次加载的消息数量
// 返回消息按创建时间升序排列，total 为会话消息总数
func (r *SessionRepositoryImpl) GetMessagesPaginated(sessionID string, beforeID string, limit int) ([]types.SessionMessage, int64, error) {
	var messages []types.SessionMessage
	var total int64

	// 统计总数
	if err := r.db.Model(&types.SessionMessage{}).Where("session_id = ?", sessionID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.Where("session_id = ?", sessionID)

	// 如果指定了 beforeID，获取该消息之前的消息
	if beforeID != "" {
		// 先获取 beforeID 消息的创建时间
		var beforeMessage types.SessionMessage
		if err := r.db.Select("created_at").Where("id = ?", beforeID).First(&beforeMessage).Error; err != nil {
			return nil, 0, err
		}
		query = query.Where("created_at < ?", beforeMessage.CreatedAt)
	}

	// 按创建时间倒序获取 limit 条，然后反转顺序
	err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	// 反转消息顺序，使其按时间升序
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, total, nil
}

// GetOldestMessages 获取最早的消息
func (r *SessionRepositoryImpl) GetOldestMessages(sessionID string, limit int) ([]types.SessionMessage, error) {
	var messages []types.SessionMessage
	err := r.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// DeleteMessages 删除指定消息
func (r *SessionRepositoryImpl) DeleteMessages(sessionID string, messageIDs []string) error {
	return r.db.Where("session_id = ? AND id IN ?", sessionID, messageIDs).
		Delete(&types.SessionMessage{}).Error
}

// DeleteOldestMessages 删除最早的 N 条消息
func (r *SessionRepositoryImpl) DeleteOldestMessages(sessionID string, count int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var ids []string
		err := tx.Model(&types.SessionMessage{}).
			Where("session_id = ?", sessionID).
			Order("created_at ASC").
			Limit(count).
			Pluck("id", &ids).Error
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		return tx.Where("id IN ?", ids).Delete(&types.SessionMessage{}).Error
	})
}

// ClearMessages 清空会话消息
func (r *SessionRepositoryImpl) ClearMessages(sessionID string) error {
	return r.db.Where("session_id = ?", sessionID).Delete(&types.SessionMessage{}).Error
}

// CountMessages 统计消息数量
func (r *SessionRepositoryImpl) CountMessages(sessionID string) (int64, error) {
	var count int64
	err := r.db.Model(&types.SessionMessage{}).
		Where("session_id = ?", sessionID).
		Count(&count).Error
	return count, err
}

// GetTokenUsage 获取会话的 token 用量（聚合计算）
func (r *SessionRepositoryImpl) GetTokenUsage(sessionID string) (int64, error) {
	var total int64
	err := r.db.Model(&types.SessionMessage{}).
		Where("session_id = ?", sessionID).
		Select("COALESCE(SUM(total_tokens), 0)").
		Scan(&total).Error
	return total, err
}

// GetSessionStats 批量获取会话统计信息（消息数量和token用量）
func (r *SessionRepositoryImpl) GetSessionStats(sessionIDs []string) (map[string]SessionStats, error) {
	result := make(map[string]SessionStats)
	if len(sessionIDs) == 0 {
		return result, nil
	}

	// 使用单个查询获取所有会话的统计信息
	type statsRow struct {
		SessionID    string
		MessageCount int64
		TotalTokens  int64
	}
	var rows []statsRow

	err := r.db.Model(&types.SessionMessage{}).
		Select("session_id, COUNT(*) as message_count, COALESCE(SUM(total_tokens), 0) as total_tokens").
		Where("session_id IN ?", sessionIDs).
		Group("session_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.SessionID] = SessionStats{
			SessionID:    row.SessionID,
			MessageCount: row.MessageCount,
			TotalTokens:  row.TotalTokens,
		}
	}

	// 对于没有消息的会话，初始化为0
	for _, id := range sessionIDs {
		if _, exists := result[id]; !exists {
			result[id] = SessionStats{
				SessionID:    id,
				MessageCount: 0,
				TotalTokens:  0,
			}
		}
	}

	return result, nil
}

// CountMessagesInWindow 统计窗口内消息数量
func (r *SessionRepositoryImpl) CountMessagesInWindow(sessionID string) (int64, error) {
	var count int64
	err := r.db.Model(&types.SessionMessage{}).
		Where("session_id = ? AND in_window = ?", sessionID, true).
		Count(&count).Error
	return count, err
}

// GetOldestMessagesInWindow 获取窗口内最早的消息
func (r *SessionRepositoryImpl) GetOldestMessagesInWindow(sessionID string, limit int) ([]types.SessionMessage, error) {
	var messages []types.SessionMessage
	err := r.db.Where("session_id = ? AND in_window = ?", sessionID, true).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetMessagesInWindow 获取窗口内消息（按创建时间升序）
func (r *SessionRepositoryImpl) GetMessagesInWindow(sessionID string, limit int) ([]types.SessionMessage, error) {
	var messages []types.SessionMessage
	query := r.db.Where("session_id = ? AND in_window = ?", sessionID, true).Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&messages).Error
	return messages, err
}

// EvictMessages 标记消息为窗口外
func (r *SessionRepositoryImpl) EvictMessages(sessionID string, messageIDs []string) error {
	return r.db.Model(&types.SessionMessage{}).
		Where("session_id = ? AND id IN ?", sessionID, messageIDs).
		Update("in_window", false).Error
}

// EvictMessagesInWindow 将所有窗口中的消息淘汰掉
func (r *SessionRepositoryImpl) EvictMessagesInWindow(sessionID string) error {
	return r.db.Model(&types.SessionMessage{}).
		Where("session_id = ?", sessionID).
		Update("in_window", false).Error
}

// GetWebChatSessions 获取前端聊天会话列表
// 查询条件：agentID + sessionType=chat + botID=web + userID（通过 im_context）
func (r *SessionRepositoryImpl) GetWebChatSessions(agentID, userID string, page, pageSize int) ([]types.Session, int64, error) {
	var sessions []types.Session
	var total int64

	offset := (page - 1) * pageSize

	// 查询条件：
	// - agent_id = ?
	// - session_type = 'chat'
	// - bot_id = 'web'
	// - im_context->>'user_id' = ?
	query := r.db.Model(&types.Session{}).
		Where("agent_id = ? AND session_type = ? AND bot_id = ?", agentID, types.SessionTypeChat, "web").
		Where("im_context->>'user_id' = ?", userID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按更新时间倒序
	err := query.Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&sessions).Error

	return sessions, total, err
}
