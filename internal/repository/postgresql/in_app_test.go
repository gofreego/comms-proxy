package postgresql

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/gofreego/comms-proxy/internal/models/dao"
	_ "github.com/lib/pq"
)

// Runs against a real Postgres when COMMS_PROXY_TEST_DSN is set (a throwaway database: the
// migration is applied and the table dropped afterwards); skipped otherwise.
func TestInAppNotifications(t *testing.T) {
	dsn := os.Getenv("COMMS_PROXY_TEST_DSN")
	if dsn == "" {
		t.Skip("COMMS_PROXY_TEST_DSN not set")
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	up, err := os.ReadFile("../../../resources/migrations/postgresql/001_create_in_app_notifications.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(string(up)); err != nil {
		t.Fatal(err)
	}
	defer conn.Exec(`DROP TABLE in_app_notifications`)

	r := &Repository{conn: conn}
	ctx := context.Background()
	row := func(id string, profile int64, dedup string, at int64) *dao.InAppNotification {
		return &dao.InAppNotification{ID: id, App: "zshala", ProfileID: profile, Category: "c", Title: "t",
			Data: map[string]string{"k": "v"}, DedupKey: dedup, CreatedAt: at}
	}

	stored, err := r.InsertInAppNotifications(ctx, []*dao.InAppNotification{
		row("00000000-0000-0000-0000-000000000001", 7, "quiz:1", 100),
		row("00000000-0000-0000-0000-000000000002", 7, "", 200),
		row("00000000-0000-0000-0000-000000000003", 8, "quiz:1", 100),
	})
	if err != nil || stored != 3 {
		t.Fatalf("stored %d, %v; want 3", stored, err)
	}
	// The same dedup key for profile 7 again is skipped; no key never dedups.
	stored, err = r.InsertInAppNotifications(ctx, []*dao.InAppNotification{
		row("00000000-0000-0000-0000-000000000004", 7, "quiz:1", 300),
		row("00000000-0000-0000-0000-000000000005", 7, "", 400),
	})
	if err != nil || stored != 1 {
		t.Fatalf("stored %d, %v; want 1", stored, err)
	}

	list, err := r.ListInAppNotifications(ctx, "zshala", 7, false, 10, 0)
	if err != nil || len(list) != 3 || list[0].CreatedAt != 400 || list[0].Data["k"] != "v" {
		t.Fatalf("list = %d rows (first %+v), %v; want 3 newest first with data", len(list), list[0], err)
	}
	if n, _ := r.CountUnreadInAppNotifications(ctx, "zshala", 7); n != 3 {
		t.Fatalf("unread = %d, want 3", n)
	}

	// Marking someone else's notification finds nothing.
	if _, found, _ := r.MarkInAppNotificationRead(ctx, "00000000-0000-0000-0000-000000000003", 7, 500); found {
		t.Fatal("marked another profile's notification")
	}
	app, found, err := r.MarkInAppNotificationRead(ctx, "00000000-0000-0000-0000-000000000001", 7, 500)
	if err != nil || !found || app != "zshala" {
		t.Fatalf("mark = %q %v %v", app, found, err)
	}
	unread, _ := r.ListInAppNotifications(ctx, "zshala", 7, true, 10, 0)
	if len(unread) != 2 {
		t.Fatalf("unread list = %d, want 2", len(unread))
	}
	if err := r.MarkAllInAppNotificationsRead(ctx, "zshala", 7, 600); err != nil {
		t.Fatal(err)
	}
	if n, _ := r.CountUnreadInAppNotifications(ctx, "zshala", 7); n != 0 {
		t.Fatalf("unread after mark all = %d, want 0", n)
	}
	if n, _ := r.CountUnreadInAppNotifications(ctx, "zshala", 8); n != 1 {
		t.Fatalf("profile 8 unread = %d, want 1 (untouched)", n)
	}
}
