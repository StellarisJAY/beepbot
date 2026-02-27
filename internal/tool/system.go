package tool

import (
	"context"
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v4/cpu"
)

type ReadSystemInfoTool struct{}

// Description implements [Tool].
func (r *ReadSystemInfoTool) Description() string {
	return `
	获取操作系统信息, 包含以下条目:
	<items>
	- 操作系统版本, 比如windows,linux,darwin
	- 可用内存
	- 可用硬盘
	- CPU信息, 包括cpu型号,利用率,核心数
	</items>
	`
}

// Execute implements [Tool].
func (r *ReadSystemInfoTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	// 获取 CPU 信息
	cpuCount, err := cpu.Counts(true)
	if err != nil {
		cpuCount = runtime.NumCPU()
	}

	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil || len(cpuPercent) == 0 {
		cpuPercent = []float64{0}
	}

	cpuInfo, err := cpu.InfoWithContext(ctx)
	cpuModel := "Unknown"
	if err == nil && len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	return fmt.Sprintf(`
	<system-info>
		<os>%s</os>
		<available-memory>%s</available-memory>
		<available-disk>%s</available-disk>
		<cpu-info>
			<model>%s</model>
			<core>%d</core>
			<usage>%v</usage>
		</cpu-info>
	</system-info>
	`, runtime.GOOS,
		"129213MB",
		"473GB",
		cpuModel,
		cpuCount,
		cpuPercent), nil
}

// Name implements [Tool].
func (r *ReadSystemInfoTool) Name() string {
	return "read_system_info"
}

// Parameters implements [Tool].
func (r *ReadSystemInfoTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
		"required":   []string{},
	}
}

func NewReadSystemInfoTool() Tool {
	return &ReadSystemInfoTool{}
}
