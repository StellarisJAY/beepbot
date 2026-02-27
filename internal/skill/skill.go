package skill

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/config"
)

const defaultSkillConfigFile = "./skills.json"

type SkillInfo struct {
	Name        string `json:"name"`        // 技能名称
	Description string `json:"description"` // 技能描述
	Path        string `json:"path"`        // 技能文件路径
	Active      bool   `json:"active"`      // 是否可用
}

type SkillManager struct {
	skillConfigFile string
}

type SkillConfig struct {
	Skills map[string]SkillInfo `json:"skills"`
}

func NewSkillManager(config config.Config) (*SkillManager, error) {
	path := strings.TrimSpace(config.Skills.ConfigFile)
	if path == "" {
		path = defaultSkillConfigFile
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		return nil, errors.New("skills config not exist")
	}
	return &SkillManager{
		skillConfigFile: absPath,
	}, nil
}

func (s *SkillManager) loadConfig() (*SkillConfig, error) {
	data, err := os.ReadFile(s.skillConfigFile)
	if err != nil {
		return nil, fmt.Errorf("load skill config error: %w", err)
	}
	var config SkillConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.New("invalid format for skill config")
	}
	return &config, nil
}

func (s *SkillManager) saveConfig(config SkillConfig) error {
	data, _ := json.Marshal(config)
	return os.WriteFile(s.skillConfigFile, data, 0644)
}

func (s *SkillManager) ListSkills() ([]SkillInfo, error) {
	config, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	skillDict := config.Skills
	skills := make([]SkillInfo, 0, len(skillDict))
	for _, skill := range skillDict {
		skills = append(skills, skill)
	}
	return skills, nil
}
