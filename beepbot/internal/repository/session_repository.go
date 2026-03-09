package repository

import (
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// SessionRepository 会话仓储接口
type SessionRepository interface {
	// 会话操作
	CreateSession(session *types.Session) error
	GetSessionByKey(key string) (*types.Session, error)
	GetSessionByID(id string) (*types.Session, error)
	UpdateSession(session *types.Session) error
	DeleteSession(id string) error

	// 消息操作
	AppendMessage(sessionID string, message *types.SessionMessage) error
	GetMessages(sessionID string, limit int) ([]types.SessionMessage, error)
	GetOldestMessages(sessionID string, limit int) ([]types.SessionMessage, error)
	DeleteMessages(sessionID string, messageIDs []string) error
	DeleteOldestMessages(sessionID string, count int) error
	ClearMessages(sessionID string) error
	CountMessages(sessionID string) (int64, error)

	// Token 用量
	GetTokenUsage(sessionID string) (int64, error)
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
