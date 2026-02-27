package heartbeat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/StellarisJAY/beepbot/internal/config"
)

const defaultHeartBeatFile = "./beepbot/HEARTBEAT.md"

type HeartBeat struct {
	interval      time.Duration
	onHeartBeat   func(ctx context.Context, prompt string)
	ticker        *time.Ticker
	heartBeatFile string
}

func NewHeartBeat(config config.Config, onHeartBeat func(ctx context.Context, prompt string)) (*HeartBeat, error) {
	hb := config.HeartBeat
	interval, err := time.ParseDuration(hb.Interval)
	if err != nil {
		return nil, errors.New("invalid heart beat interval")
	}

	heartBeatFile := strings.TrimSpace(config.HeartBeat.HeartBeatFile)
	if heartBeatFile == "" {
		heartBeatFile = defaultHeartBeatFile
	}
	ticker := time.NewTicker(interval)
	return &HeartBeat{
		ticker:        ticker,
		interval:      interval,
		onHeartBeat:   onHeartBeat,
		heartBeatFile: heartBeatFile,
	}, nil
}

func (h *HeartBeat) Start(ctx context.Context) {
	slog.Info("Heart beat on", "interval", h.interval.String(), "file", h.heartBeatFile)
	for {
		select {
		case <-ctx.Done():
			return
		case <-h.ticker.C:
			slog.Info("Heart beat triggerred...")
			h.onHeartBeat(ctx, h.buildHeartBeatPrompt())
		}
	}
}

func (h *HeartBeat) buildHeartBeatPrompt() string {
	data, err := os.ReadFile(h.heartBeatFile)
	notes := ""
	if err == nil {
		notes = string(data)
	}
	now := time.Now().Format(time.DateTime)
	prompt := fmt.Sprintf(`
	# 心跳检查
	现在的时间是：\"%s\"
	请以主动寻找问题并解决问题的态度, 检查是否有需要执行的任务或动作。
	## notes
	%s
	`, now, notes)
	return prompt
}
