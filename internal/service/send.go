package service

import (
	"context"
	"time"

	"github.com/gofreego/comms-proxy/api/commsproxy_v1"
	"github.com/gofreego/comms-proxy/internal/models/dao"
	"github.com/gofreego/goutils/logger"
)

// Send delivers the notification on each requested channel in turn. Channels are independent:
// one that's unavailable or fails is reported in its result and the others still go out.
func (s *Service) Send(ctx context.Context, req *commsproxy_v1.SendRequest) (*commsproxy_v1.SendResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	n := &dao.Notification{
		App:       req.App,
		Category:  req.Category,
		Title:     req.Title,
		Body:      req.Body,
		Link:      req.Link,
		ImageURL:  req.ImageUrl,
		Data:      req.Data,
		DedupKey:  req.DedupKey,
		CreatedAt: time.Now().UnixMilli(),
	}
	recipients := make([]dao.Recipient, 0, len(req.Recipients))
	for _, r := range req.Recipients {
		recipients = append(recipients, dao.Recipient{ProfileID: r.ProfileId, Phone: r.Phone, Email: r.Email})
	}

	results := make([]*commsproxy_v1.ChannelResult, 0, len(req.Channels))
	done := make(map[commsproxy_v1.Channel]bool, len(req.Channels))
	for _, ch := range req.Channels {
		if done[ch] {
			continue
		}
		done[ch] = true

		channel, ok := s.channels[ch]
		if !ok {
			results = append(results, &commsproxy_v1.ChannelResult{
				Channel: ch,
				Skipped: int32(len(recipients)),
				Error:   "channel not available yet",
			})
			continue
		}
		r := channel.Send(ctx, n, recipients)
		result := &commsproxy_v1.ChannelResult{
			Channel: ch,
			Sent:    int32(r.Sent),
			Skipped: int32(r.Skipped),
			Failed:  int32(r.Failed),
		}
		if r.Err != nil {
			logger.Error(ctx, "Send: %s channel error (app=%s category=%s): %v", ch, req.App, req.Category, r.Err)
			result.Error = "delivery failed"
		}
		results = append(results, result)
	}
	return &commsproxy_v1.SendResponse{Results: results, Message: "success"}, nil
}
