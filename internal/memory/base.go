package memory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/types"
)

func CreateMemoryManager(config config.StandaloneConfig) (types.MemoryManager, error) {
	memoryConfig := config.MemoryConfig
	switch {
	case memoryConfig.Flash != nil:
		slog.Info("Using Flash Memory Manager, memory will be lost if restarted")
		return NewFlashMemoryManager(memoryConfig), nil
	case memoryConfig.Milvus != nil:
		slog.Info("Using Milvus Memory Manager")
		return NewMilvusMemoryManager(), nil
	default:
		slog.Info("no available memory config, using flash memory")
		return NewFlashMemoryManager(memoryConfig), nil
	}
}

// BuildMemoryManager 构建记忆上下文提示词
func BuildMemoryContext(c context.Context, memory types.MemoryManager) string {
	longTermMemories, _ := memory.ReadLongTerm(c)
	builder := strings.Builder{}
	builder.WriteString("<LongTermMemories>\n\n")
	for _, memory := range longTermMemories {
		fmt.Fprintf(&builder, "Time:%s; Content:\"%s\";\n\n", memory.CreatedAt.Local().Format(time.DateTime), memory.Content)
	}
	builder.WriteString("</LongTermMemories>\n\n")
	return builder.String()
}
