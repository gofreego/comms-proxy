package service

import (
	"context"

	"github.com/gofreego/comms-proxy/api/commsproxy_v1"
	"github.com/gofreego/comms-proxy/internal/channels"
	"github.com/gofreego/comms-proxy/internal/channels/inapp"
	"github.com/gofreego/comms-proxy/internal/models/dao"
)

type Config struct {
	// DefaultApp is the inbox an Inbox request reads when it names no app.
	DefaultApp string `yaml:"DefaultApp"`
}

type Repository interface {
	inapp.Store
	// ListInAppNotifications returns a profile's notifications in an app, newest first.
	ListInAppNotifications(ctx context.Context, app string, profileID int64, unreadOnly bool, limit, offset int) ([]*dao.InAppNotification, error)
	CountUnreadInAppNotifications(ctx context.Context, app string, profileID int64) (int32, error)
	// MarkInAppNotificationRead marks one of the profile's notifications read and returns its
	// app, or found=false if the profile has no such notification.
	MarkInAppNotificationRead(ctx context.Context, id string, profileID, readAt int64) (app string, found bool, err error)
	MarkAllInAppNotificationsRead(ctx context.Context, app string, profileID, readAt int64) error
}

// Service implements both APIs: Notifier (internal sends) and Inbox (a user's in-app
// notifications).
type Service struct {
	commsproxy_v1.UnimplementedNotifierServer
	commsproxy_v1.UnimplementedInboxServer

	cfg  *Config
	repo Repository
	// channels are the mediums that can deliver today. A requested channel that's missing
	// here is reported as unavailable.
	channels map[commsproxy_v1.Channel]channels.Channel
}

func NewService(ctx context.Context, cfg *Config, repo Repository) *Service {
	return &Service{
		cfg:  cfg,
		repo: repo,
		channels: map[commsproxy_v1.Channel]channels.Channel{
			commsproxy_v1.Channel_IN_APP: inapp.New(repo),
		},
	}
}

func (s *Service) app(requested string) string {
	if requested != "" {
		return requested
	}
	return s.cfg.DefaultApp
}
