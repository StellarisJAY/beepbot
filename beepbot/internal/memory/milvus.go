package memory

import (
	"context"

	"github.com/StellarisJAY/beepbot/internal/types"
)

type MilvusMemoryManager struct {
}

func NewMilvusMemoryManager() types.MemoryManager {
	return &MilvusMemoryManager{}
}

// Append 追加短期记忆
func (f *MilvusMemoryManager) Append(ctx context.Context, memory types.MemoryEntry) error {
	return nil
}

// SemanticSearch 语义搜索记忆
func (f *MilvusMemoryManager) SemanticSearch(ctx context.Context, query string) ([]types.MemoryEntry, error) {
	return nil, nil
}

// WriteLongTerm 写入长期记忆
func (f *MilvusMemoryManager) WriteLongTerm(ctx context.Context, memory types.MemoryEntry) error {
	return nil
}

// ReadLongTerm 读取长期记忆
func (f *MilvusMemoryManager) ReadLongTerm(ctx context.Context) ([]types.MemoryEntry, error) {
	return nil, nil
}

// ReadRecent 读取最近记忆
func (f *MilvusMemoryManager) ReadRecent(ctx context.Context) ([]types.MemoryEntry, error) {
	return nil, nil
}
