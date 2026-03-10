package types

import (
	"path/filepath"
	"time"
)

// SkillStatus 技能状态
type SkillStatus string

const (
	SkillStatusActive   SkillStatus = "active"   // 正常可用
	SkillStatusInactive SkillStatus = "inactive" // 已禁用
)

// Skill 技能
type Skill struct {
	ID          string      `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name        string      `json:"name" gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	Description string      `json:"description" gorm:"column:description;type:text"`
	Version     string      `json:"version" gorm:"column:version;type:varchar(32)"`     // 技能版本
	Author      string      `json:"author" gorm:"column:author;type:varchar(128)"`      // 作者
	Path        string      `json:"path" gorm:"column:path;type:varchar(512);not null"` // 技能目录相对路径（相对于 DataDir/skills）
	Status      SkillStatus `json:"status" gorm:"column:status;type:varchar(16);default:active;index"`
	InstalledAt time.Time   `json:"installed_at" gorm:"column:installed_at;type:timestamptz;not null"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"column:updated_at;type:timestamptz;not null"`
}

// TableName 指定表名
func (Skill) TableName() string {
	return "skills"
}

// GetFullPath 获取技能的完整路径
// 完整路径 = DataDir/skills/Path
func (s *Skill) GetFullPath(dataDir string) string {
	return filepath.Join(dataDir, "skills", s.Path)
}

// SkillFile 技能文件
type SkillFile struct {
	ID        string    `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	SkillID   string    `json:"skill_id" gorm:"column:skill_id;type:varchar(64);not null;index"`
	FileName  string    `json:"file_name" gorm:"column:file_name;type:varchar(256);not null"`
	FilePath  string    `json:"file_path" gorm:"column:file_path;type:varchar(512);not null"` // 文件相对路径（相对于技能目录）
	FileType  string    `json:"file_type" gorm:"column:file_type;type:varchar(32)"`           // md, txt, json, etc.
	FileSize  int64     `json:"file_size" gorm:"column:file_size;type:integer"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamptz;not null"`

	// 关联
	Skill Skill `json:"skill" gorm:"foreignKey:SkillID"`
}

// TableName 指定表名
func (SkillFile) TableName() string {
	return "skill_files"
}

// GetFullPath 获取文件的完整路径
// 完整路径 = DataDir/skills/Skill.Path/FilePath
func (f *SkillFile) GetFullPath(dataDir string, skillPath string) string {
	return filepath.Join(dataDir, "skills", skillPath, f.FilePath)
}
