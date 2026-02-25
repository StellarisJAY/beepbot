package memory

import (
	"context"
	"strconv"

	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/StellarisJAY/beepbot/internal/types"
)

type FlashMemoryManager struct {
	longTerm      []types.MemoryEntry
	shortTerm     []types.MemoryEntry
	windowSize    int
	semanticIndex any // TODO 语义搜索索引
	nextId        int64
}

func NewFlashMemoryManager(config config.MemoryConfig) types.MemoryManager {
	return &FlashMemoryManager{
		longTerm:   make([]types.MemoryEntry, 0, 64),
		shortTerm:  make([]types.MemoryEntry, 0, config.WindowSize),
		windowSize: config.WindowSize,
		nextId:     1,
	}
}

// Append 追加短期记忆
func (f *FlashMemoryManager) Append(ctx context.Context, memory types.MemoryEntry) error {
	if len(f.shortTerm) == f.windowSize {
		f.shortTerm = f.shortTerm[1:]
	}
	memory.ID = f.getNextID()
	f.shortTerm = append(f.shortTerm, memory)
	return nil
}

// SemanticSearch 语义搜索记忆
func (f *FlashMemoryManager) SemanticSearch(ctx context.Context, query string) ([]types.MemoryEntry, error) {
	panic("not implemented")
}

// WriteLongTerm 写入长期记忆
func (f *FlashMemoryManager) WriteLongTerm(ctx context.Context, memory types.MemoryEntry) error {
	memory.ID = f.getNextID()
	f.longTerm = append(f.longTerm, memory)
	return nil
}

// ReadLongTerm 读取长期记忆
func (f *FlashMemoryManager) ReadLongTerm(ctx context.Context) ([]types.MemoryEntry, error) {
	return f.longTerm, nil
}

// ReadRecent 读取最近记忆
func (f *FlashMemoryManager) ReadRecent(ctx context.Context) ([]types.MemoryEntry, error) {
	return f.shortTerm, nil
}

func (f *FlashMemoryManager) getNextID() string {
	// TODO 是否可能id回绕，考虑换成uuid
	id := strconv.FormatInt(f.nextId, 10)
	f.nextId += 1
	return id
}
