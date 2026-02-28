package skill

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Skill 表示一个技能
type Skill struct {
	Name        string `json:"name"`        // 技能名称
	Description string `json:"description"` // 技能简短描述
	Path        string `json:"path"`        // SKILL.md 文件的相对路径
	Source      string `json:"source"`      // 来源: "global" 或 "workspace"
}

// Manager 技能管理器
type Manager struct {
	globalSkillsDir    string // 全局技能库目录
	workspaceSkillsDir string // 工作空间技能库目录
	workingDir         string // 智能体工作目录（用于路径解析）
}

// NewManager 创建技能管理器
func NewManager(dataDir, workingDir string) *Manager {
	// 工作空间技能目录在工作目录下的 skills 子目录
	workspaceSkillsDir := filepath.Join(workingDir, "skills")
	globalSkillsDir := filepath.Join(dataDir, "skills")
	// 确保目录存在
	os.MkdirAll(globalSkillsDir, 0755)
	os.MkdirAll(workspaceSkillsDir, 0755)

	slog.Info("global skills directory", "dir", globalSkillsDir)
	slog.Info("workspace skills directory", "dir", workspaceSkillsDir)

	return &Manager{
		globalSkillsDir:    globalSkillsDir,
		workspaceSkillsDir: workspaceSkillsDir,
		workingDir:         workingDir,
	}
}

// parseSkillFromMD 从 SKILL.md 文件解析技能信息
// 文件格式要求：
// 第一行: # 技能名称
// 第二行及之后直到空行: 技能描述
// 其余内容: 详细指令
func parseSkillFromMD(content, skillPath string) (*Skill, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty skill file: %s", skillPath)
	}

	skill := &Skill{
		Path: skillPath,
	}

	// 解析名称（第一行 # 名称）
	firstLine := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLine, "# ") {
		skill.Name = strings.TrimSpace(strings.TrimPrefix(firstLine, "# "))
	} else {
		skill.Name = filepath.Base(filepath.Dir(skillPath))
	}

	// 解析描述（第二行开始到空行之前）
	descLines := []string{}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			break
		}
		descLines = append(descLines, line)
	}
	skill.Description = strings.Join(descLines, " ")

	// 如果没有描述，使用默认描述
	if skill.Description == "" {
		skill.Description = "No description provided"
	}

	return skill, nil
}

// loadSkillsFromDir 从目录加载所有技能
func (m *Manager) loadSkillsFromDir(dir, source string) ([]Skill, error) {
	skills := []Skill{}

	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return skills, nil
	}

	// 遍历目录
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 检查是否存在 SKILL.md
		skillFile := filepath.Join(dir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); os.IsNotExist(err) {
			continue
		}

		// 读取并解析 SKILL.md
		content, err := os.ReadFile(skillFile)
		if err != nil {
			continue
		}

		skill, err := parseSkillFromMD(string(content), skillFile)
		if err != nil {
			continue
		}

		skill.Source = source

		if source == "global" && m.globalSkillsDir != "" {
			// 全局技能存放在公共目录，需要转成绝对路径
			absPath, _ := filepath.Abs(skillFile)
			skill.Path = absPath
		} else if source == "workspace" {
			// 工作目录了技能转换成相对路径
			relPath, _ := filepath.Rel(m.workingDir, skillFile)
			skill.Path = relPath
		}

		skills = append(skills, *skill)
	}

	return skills, nil
}

// GetAllSkills 获取所有可用技能（全局 + 工作空间）
func (m *Manager) GetAllSkills() ([]Skill, error) {
	allSkills := []Skill{}

	// 加载全局技能
	if m.globalSkillsDir != "" {
		globalSkills, err := m.loadSkillsFromDir(m.globalSkillsDir, "global")
		if err != nil {
			return nil, fmt.Errorf("failed to load global skills: %w", err)
		}
		allSkills = append(allSkills, globalSkills...)
	}

	// 加载工作空间技能
	workspaceSkills, err := m.loadSkillsFromDir(m.workspaceSkillsDir, "workspace")
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace skills: %w", err)
	}
	allSkills = append(allSkills, workspaceSkills...)

	return allSkills, nil
}

// GenerateSkillsPrompt 生成技能提示词
func (m *Manager) GenerateSkillsPrompt() (string, error) {
	skills, err := m.GetAllSkills()
	if err != nil {
		return "", err
	}

	if len(skills) == 0 {
		return "<skills>\n  <message>No skills are currently available.</message>\n</skills>", nil
	}

	var sb strings.Builder
	sb.WriteString("<skills>\n")
	sb.WriteString("  <description>The following skills are available. Each skill has a SKILL.md file that contains detailed instructions.\n")
	sb.WriteString("To use a skill, first use the read_file tool to read the SKILL.md file, then follow the instructions.</description>\n")
	sb.WriteString("\n")

	for _, skill := range skills {
		sb.WriteString("  <skill>\n")
		sb.WriteString(fmt.Sprintf("    <name>%s</name>\n", escapeXML(skill.Name)))
		sb.WriteString(fmt.Sprintf("    <description>%s</description>\n", escapeXML(skill.Description)))
		sb.WriteString(fmt.Sprintf("    <path>%s</path>\n", escapeXML(skill.Path)))
		sb.WriteString("  </skill>\n")
	}

	sb.WriteString("\n")
	sb.WriteString("  <usage>\n")
	sb.WriteString("    1. Identify the skill you need from the list above\n")
	sb.WriteString("    2. Use read_file tool with the skill's path to read detailed instructions\n")
	sb.WriteString("    3. Follow the instructions in SKILL.md to complete the task\n")
	sb.WriteString("  </usage>\n")
	sb.WriteString("</skills>")

	return sb.String(), nil
}

// escapeXML 转义 XML 特殊字符
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// GetGlobalSkillsDir 获取全局技能目录
func (m *Manager) GetGlobalSkillsDir() string {
	return m.globalSkillsDir
}

// GetWorkspaceSkillsDir 获取工作空间技能目录
func (m *Manager) GetWorkspaceSkillsDir() string {
	return m.workspaceSkillsDir
}
