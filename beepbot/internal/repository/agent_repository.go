package repository

import (
	"time"

	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentRepository 智能体仓储接口
type AgentRepository interface {
	Repository[types.Agent]

	// ListWithQuery 带筛选条件的分页查询
	ListWithQuery(page, pageSize int, query *types.AgentQuery) ([]types.Agent, int64, error)

	// GetByName 根据名称获取智能体
	GetByName(name string) (*types.Agent, error)

	// GetByStatus 根据状态获取智能体列表
	GetByStatus(status types.AgentStatus) ([]types.Agent, error)

	// GetActiveAgents 获取所有活跃的智能体
	GetActiveAgents() ([]types.Agent, error)

	// GetWithRelations 获取智能体及其关联数据
	GetWithRelations(id string) (*types.Agent, error)

	// GetAgentSkills 获取智能体关联的技能ID列表
	GetAgentSkills(agentID string) ([]string, error)

	// SetAgentSkills 设置智能体关联的技能（替换现有）
	SetAgentSkills(agentID string, skillIDs []string) error

	// DeleteAgentSkills 删除智能体的所有技能关联
	DeleteAgentSkills(agentID string) error

	// DeleteSkillFromAllAgents 从所有智能体中删除指定技能的关联
	DeleteSkillFromAllAgents(skillID string) error

	// GetAgentTools 获取智能体可用工具名称列表
	GetAgentTools(agentID string) ([]string, error)

	// SetAgentTools 设置智能体可用工具（替换现有）
	SetAgentTools(agentID string, toolNames []string) error

	// DeleteAgentTools 删除智能体的所有工具关联
	DeleteAgentTools(agentID string) error

	// GetCallableAgents 获取所有可作为子智能体调用的智能体
	GetCallableAgents() ([]types.Agent, error)

	// GetUsageStats 获取智能体在指定时间范围内的用量统计
	// groupByHour: true 按小时聚合, false 按天聚合
	GetUsageStats(agentID string, startTime, endTime time.Time, groupByHour bool) ([]types.UsageStatsPoint, error)
}

type agentRepository struct {
	*BaseRepository[types.Agent]
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) AgentRepository {
	return &agentRepository{
		BaseRepository: NewBaseRepository[types.Agent](db),
		db:             db,
	}
}

// ListWithQuery 带筛选条件的分页查询
func (r *agentRepository) ListWithQuery(page, pageSize int, query *types.AgentQuery) ([]types.Agent, int64, error) {
	var agents []types.Agent
	var total int64

	db := r.db.Model(&types.Agent{})

	// 动态拼接筛选条件
	if query != nil {
		if query.Name != "" {
			db = db.Where("name LIKE ?", "%"+query.Name+"%")
		}
		if query.Status != "" {
			db = db.Where("status = ?", query.Status)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&agents).Error; err != nil {
		return nil, 0, err
	}

	return agents, total, nil
}

func (r *agentRepository) GetByName(name string) (*types.Agent, error) {
	var agent types.Agent
	err := r.db.Where("name = ?", name).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *agentRepository) GetByStatus(status types.AgentStatus) ([]types.Agent, error) {
	var agents []types.Agent
	err := r.db.Where("status = ?", status).Find(&agents).Error
	return agents, err
}

func (r *agentRepository) GetActiveAgents() ([]types.Agent, error) {
	return r.GetByStatus(types.AgentStatusActive)
}

func (r *agentRepository) GetWithRelations(id string) (*types.Agent, error) {
	var agent types.Agent
	err := r.db.Where("id = ?", id).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// GetAgentSkills 获取智能体关联的技能ID列表
func (r *agentRepository) GetAgentSkills(agentID string) ([]string, error) {
	var agentSkills []types.AgentSkill
	err := r.db.Where("agent_id = ?", agentID).Find(&agentSkills).Error
	if err != nil {
		return nil, err
	}
	skillIDs := make([]string, len(agentSkills))
	for i, as := range agentSkills {
		skillIDs[i] = as.SkillID
	}
	return skillIDs, nil
}

// SetAgentSkills 设置智能体关联的技能（替换现有）
func (r *agentRepository) SetAgentSkills(agentID string, skillIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有关联
		if err := tx.Where("agent_id = ?", agentID).Delete(&types.AgentSkill{}).Error; err != nil {
			return err
		}

		// 创建新关联
		if len(skillIDs) > 0 {
			now := time.Now()
			agentSkills := make([]types.AgentSkill, len(skillIDs))
			for i, skillID := range skillIDs {
				agentSkills[i] = types.AgentSkill{
					ID:        uuid.NewString(),
					AgentID:   agentID,
					SkillID:   skillID,
					CreatedAt: now,
				}
			}
			if err := tx.Create(&agentSkills).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteAgentSkills 删除智能体的所有技能关联
func (r *agentRepository) DeleteAgentSkills(agentID string) error {
	return r.db.Where("agent_id = ?", agentID).Delete(&types.AgentSkill{}).Error
}

// DeleteSkillFromAllAgents 从所有智能体中删除指定技能的关联
func (r *agentRepository) DeleteSkillFromAllAgents(skillID string) error {
	return r.db.Where("skill_id = ?", skillID).Delete(&types.AgentSkill{}).Error
}

// GetAgentTools 获取智能体可用工具名称列表
func (r *agentRepository) GetAgentTools(agentID string) ([]string, error) {
	var agentTools []types.AgentTool
	err := r.db.Where("agent_id = ?", agentID).Find(&agentTools).Error
	if err != nil {
		return nil, err
	}
	toolNames := make([]string, len(agentTools))
	for i, at := range agentTools {
		toolNames[i] = at.ToolName
	}
	return toolNames, nil
}

// SetAgentTools 设置智能体可用工具（替换现有）
func (r *agentRepository) SetAgentTools(agentID string, toolNames []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有关联
		if err := tx.Where("agent_id = ?", agentID).Delete(&types.AgentTool{}).Error; err != nil {
			return err
		}

		// 创建新关联
		if len(toolNames) > 0 {
			now := time.Now()
			agentTools := make([]types.AgentTool, len(toolNames))
			for i, toolName := range toolNames {
				agentTools[i] = types.AgentTool{
					ID:        uuid.NewString(),
					AgentID:   agentID,
					ToolName:  toolName,
					CreatedAt: now,
				}
			}
			if err := tx.Create(&agentTools).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteAgentTools 删除智能体的所有工具关联
func (r *agentRepository) DeleteAgentTools(agentID string) error {
	return r.db.Where("agent_id = ?", agentID).Delete(&types.AgentTool{}).Error
}

// GetCallableAgents 获取所有可作为子智能体调用的智能体
func (r *agentRepository) GetCallableAgents() ([]types.Agent, error) {
	var agents []types.Agent
	err := r.db.Where("callable = ?", true).Find(&agents).Error
	return agents, err
}

// sessionRow 会话统计行
type sessionStatRow struct {
	ID        string
	CreatedAt time.Time
}

// messageRow 消息统计行
type messageStatRow struct {
	SessionID    string
	CreatedAt    time.Time
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
}

// GetUsageStats 获取智能体在指定时间范围内的用量统计
func (r *agentRepository) GetUsageStats(agentID string, startTime, endTime time.Time, groupByHour bool) ([]types.UsageStatsPoint, error) {
	// 查询会话原始数据
	var sessionRows []sessionStatRow
	if err := r.db.Table("sessions").
		Select("id, created_at").
		Where("agent_id = ? AND created_at >= ? AND created_at < ?", agentID, startTime, endTime).
		Find(&sessionRows).Error; err != nil {
		return nil, err
	}

	// 查询消息原始数据
	var messageRows []messageStatRow
	if err := r.db.Table("session_messages sm").
		Select("sm.session_id, sm.created_at, sm.input_tokens, sm.output_tokens, sm.total_tokens").
		Joins("JOIN sessions s ON s.id = sm.session_id").
		Where("s.agent_id = ? AND sm.created_at >= ? AND sm.created_at < ?", agentID, startTime, endTime).
		Find(&messageRows).Error; err != nil {
		return nil, err
	}

	// 在内存中进行统计
	// 使用 map 存储每个时间桶的统计
	sessionStats := make(map[time.Time]int64)
	messageStats := make(map[time.Time]*struct {
		MessageCount int64
		InputTokens  int64
		OutputTokens int64
		TotalTokens  int64
	})

	// 统计消息
	for _, row := range messageRows {
		bucket := truncateTime(row.CreatedAt, groupByHour)
		if stat, ok := messageStats[bucket]; ok {
			stat.MessageCount++
			stat.InputTokens += row.InputTokens
			stat.OutputTokens += row.OutputTokens
			stat.TotalTokens += row.TotalTokens
		} else {
			messageStats[bucket] = &struct {
				MessageCount int64
				InputTokens  int64
				OutputTokens int64
				TotalTokens  int64
			}{
				MessageCount: 1,
				InputTokens:  row.InputTokens,
				OutputTokens: row.OutputTokens,
				TotalTokens:  row.TotalTokens,
			}
		}
	}

	// 生成完整的时间序列
	var points []types.UsageStatsPoint
	current := startTime
	for current.Before(endTime) {
		bucket := truncateTime(current, groupByHour)

		point := types.UsageStatsPoint{
			Time:         bucket,
			SessionCount: sessionStats[bucket],
		}
		if msgStat, ok := messageStats[bucket]; ok {
			point.MessageCount = msgStat.MessageCount
			point.InputTokens = msgStat.InputTokens
			point.OutputTokens = msgStat.OutputTokens
			point.TotalTokens = msgStat.TotalTokens
		}

		points = append(points, point)

		// 前进到下一个时间点
		if groupByHour {
			current = current.Add(time.Hour)
		} else {
			current = current.AddDate(0, 0, 1)
		}
	}

	return points, nil
}

// truncateTime 截断时间到小时或天
func truncateTime(t time.Time, toHour bool) time.Time {
	if toHour {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
