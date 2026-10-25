package service

import (
	"context"
	"time"

	"github.com/gofreego/comms-proxy/api/commsproxy_v1"
	"github.com/gofreego/goutils/customerrors"
	"github.com/gofreego/goutils/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultInboxLimit = 20

// ListNotifications returns the caller's notifications, newest first, with the unread count.
func (s *Service) ListNotifications(ctx context.Context, req *commsproxy_v1.ListNotificationsRequest) (*commsproxy_v1.ListNotificationsResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	profileID, err := getProfileID(ctx)
	if err != nil {
		return nil, err
	}
	app := s.app(req.App)
	limit := int(req.Limit)
	if limit == 0 {
		limit = defaultInboxLimit
	}

	rows, err := s.repo.ListInAppNotifications(ctx, app, profileID, req.UnreadOnly, limit, int(req.Offset))
	if err != nil {
		logger.Error(ctx, "ListNotifications: repo error: %v", err)
		return nil, customerrors.ERROR_INTERNAL_SERVER_ERROR
	}
	unread, err := s.repo.CountUnreadInAppNotifications(ctx, app, profileID)
	if err != nil {
		logger.Error(ctx, "ListNotifications: count error: %v", err)
		return nil, customerrors.ERROR_INTERNAL_SERVER_ERROR
	}

	out := make([]*commsproxy_v1.InAppNotification, 0, len(rows))
	for _, r := range rows {
		out = append(out, &commsproxy_v1.InAppNotification{
			Id:        r.ID,
			Category:  r.Category,
			Title:     r.Title,
			Body:      r.Body,
			Link:      r.Link,
			ImageUrl:  r.ImageURL,
			Data:      r.Data,
			CreatedAt: r.CreatedAt,
			ReadAt:    r.ReadAt,
		})
	}
	return &commsproxy_v1.ListNotificationsResponse{Notifications: out, UnreadCount: unread, Message: "success"}, nil
}

// GetUnreadCount is the caller's unread notification count, for a badge.
func (s *Service) GetUnreadCount(ctx context.Context, req *commsproxy_v1.GetUnreadCountRequest) (*commsproxy_v1.GetUnreadCountResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	profileID, err := getProfileID(ctx)
	if err != nil {
		return nil, err
	}
	unread, err := s.repo.CountUnreadInAppNotifications(ctx, s.app(req.App), profileID)
	if err != nil {
		logger.Error(ctx, "GetUnreadCount: repo error: %v", err)
		return nil, customerrors.ERROR_INTERNAL_SERVER_ERROR
	}
	return &commsproxy_v1.GetUnreadCountResponse{UnreadCount: unread, Message: "success"}, nil
}

// MarkNotificationRead marks one of the caller's notifications read.
func (s *Service) MarkNotificationRead(ctx context.Context, req *commsproxy_v1.MarkNotificationReadRequest) (*commsproxy_v1.MarkReadResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	profileID, err := getProfileID(ctx)
	if err != nil {
		return nil, err
	}
	app, found, err := s.repo.MarkInAppNotificationRead(ctx, req.Id, profileID, time.Now().UnixMilli())
	if err != nil {
		logger.Error(ctx, "MarkNotificationRead: repo error: %v", err)
		return nil, customerrors.ERROR_INTERNAL_SERVER_ERROR
	}
	if !found {
		return nil, status.Error(codes.NotFound, "notification not found")
	}
	return s.markReadResponse(ctx, app, profileID)
}

// MarkAllNotificationsRead marks all the caller's notifications in the app read.
func (s *Service) MarkAllNotificationsRead(ctx context.Context, req *commsproxy_v1.MarkAllNotificationsReadRequest) (*commsproxy_v1.MarkReadResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	profileID, err := getProfileID(ctx)
	if err != nil {
		return nil, err
	}
	app := s.app(req.App)
	if err := s.repo.MarkAllInAppNotificationsRead(ctx, app, profileID, time.Now().UnixMilli()); err != nil {
		logger.Error(ctx, "MarkAllNotificationsRead: repo error: %v", err)
		return nil, customerrors.ERROR_INTERNAL_SERVER_ERROR
	}
	return s.markReadResponse(ctx, app, profileID)
}

func (s *Service) markReadResponse(ctx context.Context, app string, profileID int64) (*commsproxy_v1.MarkReadResponse, error) {
	unread, err := s.repo.CountUnreadInAppNotifications(ctx, app, profileID)
	if err != nil {
		logger.Error(ctx, "markReadResponse: count error: %v", err)
		return nil, customerrors.ERROR_INTERNAL_SERVER_ERROR
	}
	return &commsproxy_v1.MarkReadResponse{UnreadCount: unread, Message: "success"}, nil
}
