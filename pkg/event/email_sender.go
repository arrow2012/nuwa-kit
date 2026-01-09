package event

import (
	"context"

	"github.com/arrow2012/nuwa-kit/pkg/email"
)

// EmailSenderWrapper implements email.Sender using Event Bus
type EmailSenderWrapper struct {
	bus Bus
}

// NewEmailEventSender creates a new Sender that publishes events
func NewEmailEventSender(bus Bus) email.Sender {
	return &EmailSenderWrapper{bus: bus}
}

func (s *EmailSenderWrapper) Send(ctx context.Context, to, subject, body string) error {
	payload := map[string]interface{}{
		"to":      to,
		"subject": subject,
		"body":    body,
	}
	return s.bus.Publish(ctx, "notification.email", payload, nil)
}

func (s *EmailSenderWrapper) Close() {
	// Nothing to close
}
