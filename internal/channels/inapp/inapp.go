// Package inapp delivers notifications to in-app inboxes: rows in the database that the app
// reads through the Inbox API.
package inapp

import (
	"context"

	"github.com/gofreego/comms-proxy/internal/channels"
	"github.com/gofreego/comms-proxy/internal/models/dao"
	"github.com/google/uuid"
)

// Store saves inbox rows.
type Store interface {
	// InsertInAppNotifications stores rows, skipping any whose (app, profile, dedup key) is
	// already stored, and returns how many were stored.
	InsertInAppNotifications(ctx context.Context, rows []*dao.InAppNotification) (int, error)
}

type Channel struct {
	store Store
}

func New(store Store) *Channel {
	return &Channel{store: store}
}

// Send stores one inbox row per recipient with a profile id. Recipients without one, repeats
// within the request and already-sent dedup keys are skipped.
func (c *Channel) Send(ctx context.Context, n *dao.Notification, recipients []dao.Recipient) channels.Result {
	seen := make(map[int64]bool, len(recipients))
	rows := make([]*dao.InAppNotification, 0, len(recipients))
	for _, r := range recipients {
		if r.ProfileID <= 0 || seen[r.ProfileID] {
			continue
		}
		seen[r.ProfileID] = true
		rows = append(rows, &dao.InAppNotification{
			ID:        uuid.NewString(),
			App:       n.App,
			ProfileID: r.ProfileID,
			Category:  n.Category,
			Title:     n.Title,
			Body:      n.Body,
			Link:      n.Link,
			ImageURL:  n.ImageURL,
			Data:      n.Data,
			DedupKey:  n.DedupKey,
			CreatedAt: n.CreatedAt,
		})
	}
	skipped := len(recipients) - len(rows)
	if len(rows) == 0 {
		return channels.Result{Skipped: skipped}
	}

	stored, err := c.store.InsertInAppNotifications(ctx, rows)
	if err != nil {
		return channels.Result{Skipped: skipped, Failed: len(rows), Err: err}
	}
	return channels.Result{Sent: stored, Skipped: skipped + len(rows) - stored}
}
