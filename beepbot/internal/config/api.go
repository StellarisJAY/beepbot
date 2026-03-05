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
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Key string `json:"key"` // Base64 编码的加密密钥，可选
}
