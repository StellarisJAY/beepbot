package config

// APIConfig API 服务配置
type APIConfig struct {
	Port       int              `json:"port"`             // API 服务端口
	Logging    Logging          `json:"logging"`          // 日志配置
	DataDir    string           `json:"beepbot_data_dir"` // 公共数据目录，包含公共技能
	Database   DatabaseConfig   `json:"database"`         // 数据库配置
	Encryption EncryptionConfig `json:"encryption"`       // 加密配置
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `json:"type"` // postgres
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
	Migrate  bool   `json:"migrate"` // 是否自动迁移数据库
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Key string `json:"key"` // Base64 编码的加密密钥，可选
}

// Logging 日志配置
type Logging struct {
	Level  string         `json:"level"`
	File   string         `json:"file"`
	Format string         `json:"format"`
	QQ     ChannelLogging `json:"qq"`
}

// ChannelLogging 频道日志配置
type ChannelLogging struct {
	Level string `json:"level"`
	File  string `json:"file"`
}
