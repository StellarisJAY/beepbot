package channel

import (
	"context"
	"sync"
)

// ChannelRegistry Channel 注册表，管理运行中的 Channel 实例
// 线程安全，支持并发访问
type ChannelRegistry struct {
	mu       sync.RWMutex
	channels map[string]Channel            // key: botID, value: Channel
	cancels  map[string]context.CancelFunc // key: botID, value: CancelFunc
}

// NewChannelRegistry 创建新的 Channel 注册表
func NewChannelRegistry() *ChannelRegistry {
	return &ChannelRegistry{
		channels: make(map[string]Channel),
		cancels:  make(map[string]context.CancelFunc),
	}
}

// Register 注册 Channel，返回是否新增
// 如果已存在相同 ID 的 Channel，返回 false
func (r *ChannelRegistry) Register(id string, ch Channel, cancel context.CancelFunc) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.channels[id]; exists {
		return false
	}

	r.channels[id] = ch
	r.cancels[id] = cancel
	return true
}

// Unregister 注销 Channel，返回被注销的 Channel 和 CancelFunc
// 如果不存在，返回 nil, nil, false
func (r *ChannelRegistry) Unregister(id string) (Channel, context.CancelFunc, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch, exists := r.channels[id]
	if !exists {
		return nil, nil, false
	}

	cancel := r.cancels[id]
	delete(r.channels, id)
	delete(r.cancels, id)
	return ch, cancel, true
}

// Get 获取 Channel
func (r *ChannelRegistry) Get(id string) (Channel, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ch, exists := r.channels[id]
	return ch, exists
}

// List 列出所有 Channel
func (r *ChannelRegistry) List() []Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	channels := make([]Channel, 0, len(r.channels))
	for _, ch := range r.channels {
		channels = append(channels, ch)
	}
	return channels
}

// ListIDs 列出所有 Channel ID
func (r *ChannelRegistry) ListIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.channels))
	for id := range r.channels {
		ids = append(ids, id)
	}
	return ids
}

// Count 返回注册的 Channel 数量
func (r *ChannelRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.channels)
}
