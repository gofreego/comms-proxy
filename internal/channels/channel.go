// Package channels holds the mediums a notification can go out on. Each implements Channel;
// adding one (push, SMS, WhatsApp, email) means a new implementation registered in
// service.NewService, with no change for callers.
package channels

import (
	"context"

	"github.com/gofreego/comms-proxy/internal/models/dao"
)

// Channel delivers a notification to recipients over one medium.
type Channel interface {
	// Send delivers n to every recipient this channel can reach. It doesn't return an error:
	// how it went, including a failure, is in the Result.
	Send(ctx context.Context, n *dao.Notification, recipients []dao.Recipient) Result
}

// Result counts each recipient once: sent, skipped (nothing to reach them by on this channel,
// or already sent this one) or failed. Err explains a failure.
type Result struct {
	Sent    int
	Skipped int
	Failed  int
	Err     error
}
