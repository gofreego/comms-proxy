package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gofreego/comms-proxy/api/commsproxy_v1"
	"github.com/gofreego/comms-proxy/internal/models/dao"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// fakeRepo keeps inbox rows in memory; dedup keys behave like the unique index.
type fakeRepo struct {
	Repository
	rows    []*dao.InAppNotification
	failErr error
}

func (f *fakeRepo) InsertInAppNotifications(_ context.Context, rows []*dao.InAppNotification) (int, error) {
	if f.failErr != nil {
		return 0, f.failErr
	}
	stored := 0
	for _, row := range rows {
		dup := false
		for _, have := range f.rows {
			if row.DedupKey != "" && have.App == row.App && have.ProfileID == row.ProfileID && have.DedupKey == row.DedupKey {
				dup = true
			}
		}
		if !dup {
			f.rows = append(f.rows, row)
			stored++
		}
	}
	return stored, nil
}

func sendReq(channels ...commsproxy_v1.Channel) *commsproxy_v1.SendRequest {
	return &commsproxy_v1.SendRequest{
		App:      "zshala",
		Category: "group.quiz_scheduled",
		Title:    "New quiz",
		Recipients: []*commsproxy_v1.Recipient{
			{ProfileId: 1}, {ProfileId: 2}, {ProfileId: 2}, {Phone: "+919812345678"},
		},
		Channels: channels,
		DedupKey: "quiz:abc",
	}
}

func TestSend(t *testing.T) {
	repo := &fakeRepo{}
	s := NewService(context.Background(), &Config{DefaultApp: "zshala"}, repo)

	resp, err := s.Send(context.Background(), sendReq(commsproxy_v1.Channel_IN_APP, commsproxy_v1.Channel_SMS))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("results = %v, want one per channel", resp.Results)
	}
	// In-app reaches the two profiles once each; the repeat and the phone-only recipient are skipped.
	inApp := resp.Results[0]
	if inApp.Sent != 2 || inApp.Skipped != 2 || inApp.Failed != 0 || inApp.Error != "" {
		t.Fatalf("in-app = %+v, want 2 sent, 2 skipped", inApp)
	}
	// SMS isn't wired yet: reported, not an error for the call.
	if sms := resp.Results[1]; sms.Channel != commsproxy_v1.Channel_SMS || sms.Error == "" || sms.Skipped != 4 {
		t.Fatalf("sms = %+v, want unavailable with all skipped", sms)
	}

	// Same dedup key again: nobody gets it twice.
	resp, err = s.Send(context.Background(), sendReq(commsproxy_v1.Channel_IN_APP))
	if err != nil {
		t.Fatal(err)
	}
	if r := resp.Results[0]; r.Sent != 0 || r.Skipped != 4 {
		t.Fatalf("resend = %+v, want all skipped", r)
	}
	if len(repo.rows) != 2 {
		t.Fatalf("stored %d rows, want 2", len(repo.rows))
	}
}

func TestSendReportsChannelFailure(t *testing.T) {
	s := NewService(context.Background(), &Config{}, &fakeRepo{failErr: errors.New("db down")})
	resp, err := s.Send(context.Background(), sendReq(commsproxy_v1.Channel_IN_APP))
	if err != nil {
		t.Fatal(err)
	}
	if r := resp.Results[0]; r.Failed != 2 || r.Error == "" {
		t.Fatalf("result = %+v, want 2 failed with an error", r)
	}
}

func TestSendValidates(t *testing.T) {
	s := NewService(context.Background(), &Config{}, &fakeRepo{})
	req := sendReq(commsproxy_v1.Channel_CHANNEL_UNSPECIFIED)
	if _, err := s.Send(context.Background(), req); err == nil {
		t.Fatal("want an error for an unspecified channel")
	}
}

func TestInboxNeedsOwnedProfile(t *testing.T) {
	s := NewService(context.Background(), &Config{DefaultApp: "zshala"}, &fakeRepo{})
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Profile-Id", "7", "X-Profile-Ids", "8,9"))
	_, err := s.GetUnreadCount(ctx, &commsproxy_v1.GetUnreadCountRequest{})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("err = %v, want PermissionDenied", err)
	}
}
