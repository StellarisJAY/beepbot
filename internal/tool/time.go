package tool

import (
	"context"
	"time"
)

type TimeTool struct{}

func (t *TimeTool) Name() string {
	return "system_time"
}

func (t *TimeTool) Description() string {
	return "Get current system time (local zone) of the following format: YYYY-MM-DD HH:mm:ss"
}

func (t *TimeTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
		"required":   []string{},
	}
}

func (t *TimeTool) Execute(ctx context.Context, _ map[string]any) (string, error) {
	return time.Now().Local().Format(time.DateTime), nil
}
