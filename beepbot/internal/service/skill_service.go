package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/google/uuid"
)

// 文件大小限制 (10MB)
const maxUploadSize = 10 * 1024 * 1024

// 允许的文件扩展名
var allowedExtensions = map[string]bool{
	".md":   true,
	".txt":  true,
	".json": true,
	".yaml": true,
	".yml":  true,
	".tmpl": true,
	".html": true,
	".css":  true,
	".js":   true,
	".ts":   true,
	".py":   true,
	".go":   true,
	".java": true,
	".xml":  true,
	".csv":  true,
	".svg":  true,
}

// SkillService 技能服务
type SkillService struct {
	repo      repository.SkillRepository
	agentRepo repository.AgentRepository // 用于清理技能关联
	dataDir   string                     // 配置中的 DataDir
}

// NewSkillService 创建技能服务
func NewSkillService(repo repository.SkillRepository, agentRepo repository.AgentRepository, dataDir string) *SkillService {
	return &SkillService{
		repo:      repo,
		agentRepo: agentRepo,
		dataDir:   dataDir,
	}
}

// getSkillsDir 获取技能存储根目录
func (s *SkillService) getSkillsDir() string {
	return filepath.Join(s.dataDir, "skills")
}

// getSkillFullPath 获取技能的完整路径
func (s *SkillService) getSkillFullPath(relativePath string) string {
	return filepath.Join(s.getSkillsDir(), relativePath)
}

// UploadSkillResult 上传技能结果
type UploadSkillResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Author      string `json:"author"`
}

// SkillWithFiles 技能及其文件
type SkillWithFiles struct {
	*types.Skill
	Files []types.SkillFile `json:"files"`
}

// ListSkills 获取技能列表
func (s *SkillService) ListSkills(page, size int, status types.SkillStatus) ([]types.Skill, int64, error) {
	return s.repo.List(page, size, status)
}

// GetSkill 获取技能详情
func (s *SkillService) GetSkill(id string) (*types.Skill, error) {
	return s.repo.GetByID(id)
}

// GetSkillWithFiles 获取技能及其文件
func (s *SkillService) GetSkillWithFiles(id string) (*SkillWithFiles, error) {
	skill, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	files, err := s.repo.ListFilesBySkillID(id)
	if err != nil {
		return nil, err
	}

	return &SkillWithFiles{
		Skill: skill,
		Files: files,
	}, nil
}

// GetSkillFiles 获取技能文件列表
func (s *SkillService) GetSkillFiles(skillID string) ([]types.SkillFile, error) {
	return s.repo.ListFilesBySkillID(skillID)
}

// GetFileContent 获取文件内容
func (s *SkillService) GetFileContent(skillID, fileID string) (*FileContentResult, error) {
	// 获取技能信息
	skill, err := s.repo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	// 获取文件记录
	file, err := s.repo.GetFileByID(fileID)
	if err != nil {
		return nil, err
	}

	// 验证文件属于该技能
	if file.SkillID != skillID {
		return nil, errors.New("file does not belong to this skill")
	}

	// 构建完整路径
	fullPath := file.GetFullPath(s.dataDir, skill.Path)

	// 读取文件内容
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &FileContentResult{
		ID:       file.ID,
		FileName: file.FileName,
		FileType: file.FileType,
		Content:  string(content),
		FilePath: file.FilePath,
		FileSize: file.FileSize,
	}, nil
}

// FileContentResult 文件内容结果
type FileContentResult struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
	Content  string `json:"content"`
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
}

// UploadSkill 上传并安装技能
func (s *SkillService) UploadSkill(file multipart.File, header *multipart.FileHeader) (*UploadSkillResult, error) {
	// 1. 验证文件大小
	if header.Size > maxUploadSize {
		return nil, errors.New("file size exceeds maximum limit (10MB)")
	}

	// 2. 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".zip" {
		return nil, errors.New("only ZIP files are allowed")
	}

	// 3. 创建临时目录
	tempDir, err := os.MkdirTemp("", "skill-upload-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // 清理临时目录

	// 4. 保存上传的文件
	zipPath := filepath.Join(tempDir, "upload.zip")
	if err := saveUploadedFile(file, zipPath); err != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}

	// 5. 解压 ZIP 文件到临时目录
	extractDir := filepath.Join(tempDir, "extracted")
	if err := unzipFile(zipPath, extractDir); err != nil {
		return nil, fmt.Errorf("failed to unzip file: %w", err)
	}

	// 6. 查找 SKILL.md 文件
	skillMDPath, skillContent, err := findSkillMD(extractDir)
	if err != nil {
		return nil, err
	}

	// 7. 解析技能元数据（只解析内容，不依赖目录名）
	metadata := parseSkillMetadataFromContent(skillContent)

	// 8. 检查技能名称是否有效
	if metadata.Name == "" {
		return nil, errors.New("skill name is required in SKILL.md (first line should be '# SkillName')")
	}

	// 9. 检查是否已存在同名技能
	exists, err := s.repo.ExistsByName(metadata.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing skill: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("skill with name '%s' already exists", metadata.Name)
	}

	// 10. 生成技能 ID 和相对路径（使用技能名称作为目录名）
	skillID := uuid.NewString()
	relativePath := sanitizeName(metadata.Name)

	// 11. 确定技能源目录（SKILL.md 所在的目录）
	skillSourceDir := filepath.Dir(skillMDPath)

	// 12. 移动到最终的技能目录（使用技能名称命名）
	skillDir := s.getSkillFullPath(relativePath)
	if err := os.Rename(skillSourceDir, skillDir); err != nil {
		// 如果 Rename 失败（跨文件系统），尝试复制
		if err := copyDir(skillSourceDir, skillDir); err != nil {
			return nil, fmt.Errorf("failed to move skill directory: %w", err)
		}
	}

	// 13. 创建数据库记录
	now := time.Now()
	skill := &types.Skill{
		ID:          skillID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Version:     metadata.Version,
		Author:      metadata.Author,
		Path:        relativePath,
		Status:      types.SkillStatusActive,
		InstalledAt: now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(skill); err != nil {
		// 回滚：删除已创建的目录
		os.RemoveAll(skillDir)
		return nil, fmt.Errorf("failed to create skill record: %w", err)
	}

	// 12. 记录文件列表
	files, err := s.scanSkillFiles(skillID, skillDir)
	if err != nil {
		slog.Warn("failed to scan skill files", "error", err)
	} else if len(files) > 0 {
		if err := s.repo.CreateFiles(files); err != nil {
			slog.Warn("failed to create file records", "error", err)
		}
	}

	return &UploadSkillResult{
		ID:          skill.ID,
		Name:        skill.Name,
		Description: skill.Description,
		Version:     skill.Version,
		Author:      skill.Author,
	}, nil
}

// DeleteSkill 删除技能
func (s *SkillService) DeleteSkill(id string) error {
	// 获取技能信息
	skill, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 删除与智能体的关联
	if s.agentRepo != nil {
		if err := s.agentRepo.DeleteSkillFromAllAgents(id); err != nil {
			slog.Warn("failed to delete skill associations from agents", "error", err, "skill_id", id)
		}
	}

	// 删除文件系统目录
	skillDir := skill.GetFullPath(s.dataDir)
	if err := os.RemoveAll(skillDir); err != nil {
		slog.Warn("failed to remove skill directory", "error", err, "path", skillDir)
	}

	// 删除文件记录
	if err := s.repo.DeleteFilesBySkillID(id); err != nil {
		slog.Warn("failed to delete file records", "error", err)
	}

	// 删除数据库记录
	return s.repo.Delete(id)
}

// UpdateSkillStatus 更新技能状态
func (s *SkillService) UpdateSkillStatus(id string, status types.SkillStatus) error {
	skill, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	skill.Status = status
	skill.UpdatedAt = time.Now()

	return s.repo.Update(skill)
}

// scanSkillFiles 扫描技能目录下的所有文件
func (s *SkillService) scanSkillFiles(skillID, skillDir string) ([]types.SkillFile, error) {
	var files []types.SkillFile

	err := filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(skillDir, path)
		if err != nil {
			return err
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if !allowedExtensions[ext] {
			return nil // 跳过不允许的文件类型
		}

		// 获取文件类型
		fileType := strings.TrimPrefix(ext, ".")

		files = append(files, types.SkillFile{
			ID:        uuid.NewString(),
			SkillID:   skillID,
			FileName:  info.Name(),
			FilePath:  relPath,
			FileType:  fileType,
			FileSize:  info.Size(),
			CreatedAt: time.Now(),
		})

		return nil
	})

	return files, err
}

// saveUploadedFile 保存上传的文件
func saveUploadedFile(src multipart.File, dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, src)
	return err
}

// unzipFile 解压 ZIP 文件
func unzipFile(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// 检查是否有顶级目录
	hasTopLevelDir := false
	var topDir string

	for _, f := range r.File {
		parts := strings.Split(f.Name, "/")
		if len(parts) > 1 {
			hasTopLevelDir = true
			topDir = parts[0]
			break
		}
	}

	for _, f := range r.File {
		// 跳过 macOS 的特殊文件
		if strings.HasPrefix(f.Name, "__MACOSX/") || strings.Contains(f.Name, ".DS_Store") {
			continue
		}

		// 构建目标路径
		var targetPath string
		if hasTopLevelDir {
			// 去掉顶级目录
			relPath := strings.TrimPrefix(f.Name, topDir+"/")
			if relPath == "" {
				continue
			}
			targetPath = filepath.Join(dst, relPath)
		} else {
			targetPath = filepath.Join(dst, f.Name)
		}

		// 防止路径遍历攻击
		if !filepath.IsAbs(targetPath) {
			absDst, _ := filepath.Abs(dst)
			absTarget, _ := filepath.Abs(targetPath)
			if !strings.HasPrefix(absTarget, absDst) {
				continue
			}
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(targetPath, 0755)
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// 解压文件
		dstFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		srcFile, err := f.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// findSkillMD 查找 SKILL.md 文件
func findSkillMD(dir string) (string, string, error) {
	var skillMDPath string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "SKILL.md" {
			skillMDPath = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", "", err
	}

	if skillMDPath == "" {
		return "", "", errors.New("SKILL.md not found in the ZIP file")
	}

	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read SKILL.md: %w", err)
	}

	return skillMDPath, string(content), nil
}

// SkillMetadata 技能元数据
type SkillMetadata struct {
	Name        string
	Description string
	Version     string
	Author      string
}

// parseSkillMetadataFromContent 从 SKILL.md 内容解析技能元数据
func parseSkillMetadataFromContent(content string) *SkillMetadata {
	metadata := &SkillMetadata{
		Version: "1.0.0",
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return metadata
	}

	// 检查是否有 YAML front matter（以 --- 开始和结束）
	if strings.TrimSpace(lines[0]) == "---" {
		// 查找结束的 ---
		endIndex := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				endIndex = i
				break
			}
		}

		if endIndex > 1 {
			// 解析 YAML front matter
			for i := 1; i < endIndex; i++ {
				line := lines[i]
				colonIdx := strings.Index(line, ":")
				if colonIdx == -1 {
					continue
				}

				key := strings.TrimSpace(line[:colonIdx])
				value := strings.TrimSpace(line[colonIdx+1:])

				// 移除引号
				if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') {
					value = strings.Trim(value, "\"'")
				}

				switch key {
				case "name":
					metadata.Name = value
				case "description":
					metadata.Description = value
				case "version":
					if value != "" {
						metadata.Version = value
					}
				case "author":
					metadata.Author = value
				}
			}

			// 如果描述为空，设置默认值
			if metadata.Description == "" {
				metadata.Description = "No description provided"
			}

			return metadata
		}
	}

	// 回退到旧格式解析（第一行 # 名称）
	firstLine := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLine, "# ") {
		metadata.Name = strings.TrimSpace(strings.TrimPrefix(firstLine, "# "))
	}

	// 解析描述（第二行开始到空行之前）
	var descLines []string
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			break
		}
		descLines = append(descLines, line)
	}
	if len(descLines) > 0 {
		metadata.Description = strings.Join(descLines, " ")
	} else {
		metadata.Description = "No description provided"
	}

	return metadata
}

// parseSkillMetadata 解析技能元数据（保留用于兼容性，现在调用 parseSkillMetadataFromContent）
func parseSkillMetadata(content, skillDir string) *SkillMetadata {
	metadata := parseSkillMetadataFromContent(content)

	// 如果名称为空，尝试从目录名提取
	if metadata.Name == "" {
		metadata.Name = extractSkillNameFromDir(skillDir)
	}

	return metadata
}

// extractSkillNameFromDir 从目录路径中提取技能名称，排除临时目录名
func extractSkillNameFromDir(skillDir string) string {
	dirName := filepath.Base(skillDir)

	// 排除临时目录名
	tempNames := map[string]bool{
		"extracted": true,
		"temp":      true,
		"tmp":       true,
		"upload":    true,
	}

	if tempNames[strings.ToLower(dirName)] {
		// 如果是临时目录名，返回默认名称
		return "unnamed-skill"
	}

	return dirName
}

// extractJSONString 从 JSON 字节中提取字符串字段值
func extractJSONString(data []byte, field string) string {
	// 简单的 JSON 字符串提取
	// 查找 "field": "value" 模式
	str := string(data)
	// 构建搜索模式
	patterns := []string{
		fmt.Sprintf(`"%s"\s*:\s*"`, field),
		fmt.Sprintf(`'%s'\s*:\s*'`, field),
	}

	for _, pattern := range patterns {
		idx := strings.Index(str, pattern)
		if idx == -1 {
			continue
		}

		// 找到值的起始位置
		valueStart := idx + len(pattern)
		// 找到值的结束引号
		quoteChar := pattern[len(pattern)-1]
		valueEnd := strings.IndexByte(str[valueStart:], quoteChar)
		if valueEnd == -1 {
			continue
		}

		return str[valueStart : valueStart+valueEnd]
	}

	return ""
}

// sanitizeName 清理名称，用于生成目录名
func sanitizeName(name string) string {
	// 替换空格和特殊字符
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// 只保留字母、数字和连字符
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	sanitized := result.String()
	if sanitized == "" {
		sanitized = "skill"
	}

	return sanitized
}

// copyDir 复制目录
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
