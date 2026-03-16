package agent

import (
	"context"

	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/types"
)

type Runner interface {
	RunWithMessage(ctx context.Context, sess session.Session, message types.Message) error
	StreamWithMessage(ctx context.Context, sess session.Session, message types.Message) error
}
