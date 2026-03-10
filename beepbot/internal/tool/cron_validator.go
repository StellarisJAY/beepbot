package tool

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// CronValidator 验证cron表达式的安全性
type CronValidator struct {
	minInterval time.Duration // 最小执行间隔
}

// NewCronValidator 创建一个新的Cron验证器
func NewCronValidator(minInterval time.Duration) *CronValidator {
	return &CronValidator{
		minInterval: minInterval,
	}
}

// Validate 验证cron表达式是否有效且满足最小间隔要求
// 返回解析后的schedule和错误信息
func (v *CronValidator) Validate(cronExpr string) (cron.Schedule, error) {
	// 使用标准解析器解析cron表达式
	// 支持可选的秒字段（6字段格式）
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("无效的cron表达式: %w", err)
	}

	// 检查执行频率是否过快
	if err := v.checkMinInterval(schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

// checkMinInterval 检查cron表达式的执行间隔是否满足最小要求
func (v *CronValidator) checkMinInterval(schedule cron.Schedule) error {
	// 从当前时间开始，计算接下来几次执行的时间
	now := time.Now()

	// 计算接下来5次执行的时间，检查间隔
	prevTime := schedule.Next(now)
	for i := 0; i < 5; i++ {
		nextTime := schedule.Next(prevTime)
		interval := nextTime.Sub(prevTime)

		if interval < v.minInterval {
			return fmt.Errorf("cron表达式执行频率过快，最小间隔为 %v，当前间隔为 %v", v.minInterval, interval)
		}

		prevTime = nextTime
	}

	return nil
}

// ValidateAndExplain 验证cron表达式并返回解释说明
func (v *CronValidator) ValidateAndExplain(cronExpr string) (string, error) {
	schedule, err := v.Validate(cronExpr)
	if err != nil {
		return "", err
	}

	// 生成解释说明
	now := time.Now()
	explanation := fmt.Sprintf("Cron表达式 '%s' 的接下来5次执行时间:\n", cronExpr)
	nextTime := schedule.Next(now)
	for i := 0; i < 5; i++ {
		explanation += fmt.Sprintf("  %d. %s\n", i+1, nextTime.Format("2006-01-02 15:04:05"))
		nextTime = schedule.Next(nextTime)
	}

	return explanation, nil
}

// DefaultCronValidator 返回默认的Cron验证器（最小间隔1分钟）
func DefaultCronValidator() *CronValidator {
	return NewCronValidator(time.Minute)
}
