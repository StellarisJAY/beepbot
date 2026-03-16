package agent

import (
	"context"
	"errors"
	"io"

	"github.com/StellarisJAY/beepbot/internal/channel"
	"github.com/StellarisJAY/beepbot/internal/session"
	"github.com/StellarisJAY/beepbot/internal/types"
)

type Stream struct {
	outputChan chan types.Message
	errorChan  chan error
	doneChan   chan struct{}
}

var (
	ErrStreamCancelled = errors.New("stream is cancelled")
)

type Runner interface {
	RunWithMessage(ctx context.Context, sess session.Session, message channel.InboundMessage) error
	StreamWithMessage(ctx context.Context, sess session.Session, message channel.InboundMessage) (*Stream, error)
}

func (s *Stream) Consume(ctx context.Context) (types.Message, error) {
	select {
	case <-ctx.Done():
		return types.Message{}, ErrStreamCancelled
	case <-s.doneChan:
		return types.Message{}, io.EOF
	case msg := <-s.outputChan:
		return msg, nil
	case err := <-s.errorChan:
		return types.Message{}, err
	}
}

func (s *Stream) Publish(ctx context.Context, message types.Message) {
	select {
	case <-ctx.Done():
		return
	default:
		s.outputChan <- message
	}
}

func (s *Stream) PublishError(ctx context.Context, err error) {
	select {
	case <-ctx.Done():
		return
	default:
		s.errorChan <- err
	}
}
