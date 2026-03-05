package database

import (
	"fmt"
	"log/slog"

	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDatabase 初始化数据库连接
func InitDatabase(cfg config.DatabaseConfig, loggingConfig config.Logging) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
		cfg.SSLMode,
	)

	// 配置 GORM 日志
	gormLogger := logger.Default
	if loggingConfig.Level == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	slog.Info("Connected to PostgreSQL database", "host", cfg.Host, "dbname", cfg.DBName)

	// 自动迁移
	err = db.AutoMigrate(
		&types.Provider{},
		&types.Agent{},
		&types.AgentChannel{},
		&types.Session{},
		&types.SessionMessage{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	slog.Info("Database migration completed")

	return db, nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
