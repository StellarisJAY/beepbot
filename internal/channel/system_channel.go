package channel

import (
	"context"
	"log/slog"
)

type systemChannel struct{}

// ID implements [Channel].
func (s *systemChannel) ID() string {
	return "system"
}

// IsAllowed implements [Channel].
func (s *systemChannel) IsAllowed(senderID string) bool {
	return true
}

// IsAvailable implements [Channel].
func (s *systemChannel) IsAvailable() bool {
	return true
}

// Send implements [Channel].
func (s *systemChannel) Send(ctx context.Context, message OutboundMessage) error {
	slog.Info("system message", "from", message.UserID, "content", message.Content)
	return nil
}

// Start implements [Channel].
func (s *systemChannel) Start(ctx context.Context) error {
	return nil
}

// Stop implements [Channel].
func (s *systemChannel) Stop() {
}

func newSystemChannel() Channel {
	return &systemChannel{}
}
