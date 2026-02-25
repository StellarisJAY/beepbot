package db

import (
	"context"
	"log/slog"

	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDatabase(config config.Config) error {
	slog.Info("Init sqlite database", "file", "beepbot.db")
	gormDB, err := gorm.Open(sqlite.Open("beepbot.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	slog.Info("Migrate database...")
	err = gormDB.Migrator().AutoMigrate(
		&types.Session{},
		&types.SessionMessage{},
	)
	if err != nil {
		return err
	}
	slog.Info("Init and migrate database done!")

	db = gormDB
	return nil
}

// CreateSession 创建一个新的会话
func CreateSession(ctx context.Context, session *types.Session) error {
	return db.WithContext(ctx).Create(session).Error
}

// GetSession 根据ID获取会话
func GetSession(ctx context.Context, id string) (*types.Session, error) {
	var session types.Session
	err := db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionByKey 根据 Key 获取会话
func GetSessionByKey(ctx context.Context, key string) (*types.Session, error) {
	var session types.Session
	err := db.WithContext(ctx).Where("key = ?", key).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListSessions 获取所有会话列表
func ListSessions(ctx context.Context, offset, limit int) ([]types.Session, error) {
	var sessions []types.Session
	err := db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// UpdateSession 更新会话信息
func UpdateSession(ctx context.Context, session *types.Session) error {
	return db.WithContext(ctx).Save(session).Error
}

// UpdateSessionSummary 更新会话摘要
func UpdateSessionSummary(ctx context.Context, id string, summary string) error {
	return db.WithContext(ctx).Model(&types.Session{}).Where("id = ?", id).Update("summary", summary).Error
}

// DeleteSession 删除会话（会同时删除相关的消息，因为需要外键关联或手动删除）
func DeleteSession(ctx context.Context, id string) error {
	// 使用事务确保原子性
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除关联的消息
		if err := tx.Where("session_id = ?", id).Delete(&types.SessionMessage{}).Error; err != nil {
			return err
		}
		// 再删除会话
		if err := tx.Where("id = ?", id).Delete(&types.Session{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteSessionByKey 根据 Key 删除会话
func DeleteSessionByKey(ctx context.Context, key string) error {
	// 先查找会话
	session, err := GetSessionByKey(ctx, key)
	if err != nil {
		return err
	}
	// 删除会话
	return DeleteSession(ctx, session.ID)
}

// CreateSessionMessage 创建一条会话消息
func CreateSessionMessage(ctx context.Context, message *types.SessionMessage) error {
	return db.WithContext(ctx).Create(message).Error
}

// GetSessionMessage 根据ID获取消息
func GetSessionMessage(ctx context.Context, id string) (*types.SessionMessage, error) {
	var message types.SessionMessage
	err := db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// ListSessionMessages 根据会话ID获取所有消息
func ListSessionMessages(ctx context.Context, sessionID string) ([]types.SessionMessage, error) {
	var messages []types.SessionMessage
	err := db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// ListSessionMessagesWithPagination 根据会话ID分页获取消息
func ListSessionMessagesWithPagination(ctx context.Context, sessionID string, offset, limit int) ([]types.SessionMessage, error) {
	var messages []types.SessionMessage
	err := db.WithContext(ctx).Where("session_id = ?", sessionID).
		Offset(offset).Limit(limit).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// UpdateSessionMessage 更新消息
func UpdateSessionMessage(ctx context.Context, message *types.SessionMessage) error {
	return db.WithContext(ctx).Save(message).Error
}

// DeleteSessionMessage 删除单条消息
func DeleteSessionMessage(ctx context.Context, id string) error {
	return db.WithContext(ctx).Where("id = ?", id).Delete(&types.SessionMessage{}).Error
}

// DeleteSessionMessages 删除指定会话的所有消息
func DeleteSessionMessages(ctx context.Context, sessionID string) error {
	return db.WithContext(ctx).Where("session_id = ?", sessionID).Delete(&types.SessionMessage{}).Error
}

// CountSessionMessages 统计指定会话的消息数量
func CountSessionMessages(ctx context.Context, sessionID string) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Model(&types.SessionMessage{}).Where("session_id = ?", sessionID).Count(&count).Error
	return count, err
}
