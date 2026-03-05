package types

import (
	"context"
	"time"
)

type MemoryEntry struct {
	ID        string     `json:"id"`
	CreatedAt *time.Time `json:"created_at"`
	Content   string     `json:"content"`
}

type MemoryManager interface {
	// Append 追加短期记忆
	Append(ctx context.Context, memory MemoryEntry) error
	// SemanticSearch 语义搜索记忆
	SemanticSearch(ctx context.Context, query string) ([]MemoryEntry, error)
	// WriteLongTerm 写入长期记忆
	WriteLongTerm(ctx context.Context, memory MemoryEntry) error
	// ReadLongTerm 读取长期记忆
	ReadLongTerm(ctx context.Context) ([]MemoryEntry, error)
}
