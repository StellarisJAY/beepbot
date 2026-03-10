package skill

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
)

// Manager 技能管理器
// 从数据库读取智能体关联的技能，生成技能提示词
type Manager struct {
	repo      repository.SkillRepository // 技能仓储
	agentRepo repository.AgentRepository // 智能体仓储（获取关联）
	agentID   string                     // 智能体ID
	dataDir   string                     // 数据目录
}

// NewManager 创建技能管理器
func NewManager(
	repo repository.SkillRepository,
	agentRepo repository.AgentRepository,
	agentID string,
	dataDir string,
) *Manager {
	return &Manager{
		repo:      repo,
		agentRepo: agentRepo,
		agentID:   agentID,
		dataDir:   dataDir,
	}
}

// GetAgentSkills 获取智能体关联的所有可用技能
// 返回：关联且状态为 active 的技能列表
func (m *Manager) GetAgentSkills() ([]types.Skill, error) {
	// 1. 获取智能体关联的技能ID列表
	skillIDs, err := m.agentRepo.GetAgentSkills(m.agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent skills: %w", err)
	}

	if len(skillIDs) == 0 {
		return []types.Skill{}, nil
	}

	// 2. 获取启用的技能
	skills, err := m.repo.GetActiveByIDs(skillIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get active skills: %w", err)
	}

	return skills, nil
}

// GenerateSkillsPrompt 生成技能提示词
// 每次调用都会重新查询数据库，确保获取最新的技能状态
func (m *Manager) GenerateSkillsPrompt() (string, error) {
	skills, err := m.GetAgentSkills()
	if err != nil {
		return "", err
	}

	if len(skills) == 0 {
		return "<skills>\n  <message>当前没有可用的技能。</message>\n</skills>", nil
	}

	var sb strings.Builder
	sb.WriteString("<skills>\n")
	sb.WriteString("  <description>以下技能可用。每个技能有一个 SKILL.md 文件包含详细指令。\n")
	sb.WriteString("要使用技能，首先使用 read_file 工具读取 SKILL.md 文件，然后按照指令执行。</description>\n\n")

	for _, skill := range skills {
		// 构建技能的 SKILL.md 完整路径
		skillMDPath := filepath.Join(m.dataDir, "skills", skill.Path, "SKILL.md")
		sb.WriteString("  <skill>\n")
		sb.WriteString(fmt.Sprintf("    <name>%s</name>\n", escapeXML(skill.Name)))
		sb.WriteString(fmt.Sprintf("    <description>%s</description>\n", escapeXML(skill.Description)))
		sb.WriteString(fmt.Sprintf("    <path>%s</path>\n", escapeXML(skillMDPath)))
		sb.WriteString("  </skill>\n")
	}

	sb.WriteString("\n  <usage>\n")
	sb.WriteString("    1. 从上面的列表中识别你需要的技能\n")
	sb.WriteString("    2. 使用 read_file 工具读取技能的 path 获取详细指令\n")
	sb.WriteString("    3. 按照 SKILL.md 中的指令完成任务\n")
	sb.WriteString("  </usage>\n")
	sb.WriteString("</skills>")

	return sb.String(), nil
}

// escapeXML 转义 XML 特殊字符
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "\u0026amp;")
	s = strings.ReplaceAll(s, "<", "\u0026lt;")
	s = strings.ReplaceAll(s, ">", "\u0026gt;")
	s = strings.ReplaceAll(s, "\"", "\u0026quot;")
	s = strings.ReplaceAll(s, "'", "\u0026apos;")
	return s
}
