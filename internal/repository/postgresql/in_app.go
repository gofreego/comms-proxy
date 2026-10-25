package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gofreego/comms-proxy/internal/models/dao"
	"github.com/lib/pq"
)

const inAppColumns = `id, app, profile_id, category, title, body, link, image_url, data, dedup_key, created_at, read_at`

// InsertInAppNotifications stores rows in one statement. A row whose (app, profile, dedup key)
// is already stored is skipped; it returns how many were stored.
func (r *Repository) InsertInAppNotifications(ctx context.Context, rows []*dao.InAppNotification) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	n := len(rows)
	ids, apps, categories, titles := make([]string, n), make([]string, n), make([]string, n), make([]string, n)
	bodies, links, images, datas, dedups := make([]string, n), make([]string, n), make([]string, n), make([]string, n), make([]string, n)
	profiles, created := make([]int64, n), make([]int64, n)
	for i, row := range rows {
		data, err := json.Marshal(row.Data)
		if err != nil {
			return 0, fmt.Errorf("failed to encode notification data: %w", err)
		}
		if row.Data == nil {
			data = []byte("{}")
		}
		ids[i], apps[i], profiles[i], categories[i], titles[i] = row.ID, row.App, row.ProfileID, row.Category, row.Title
		bodies[i], links[i], images[i], datas[i], dedups[i], created[i] = row.Body, row.Link, row.ImageURL, string(data), row.DedupKey, row.CreatedAt
	}

	res, err := r.conn.ExecContext(ctx, `
		INSERT INTO in_app_notifications (id, app, profile_id, category, title, body, link, image_url, data, dedup_key, created_at)
		SELECT id::uuid, app, profile_id, category, title, body, link, image_url, data::jsonb, dedup_key, created_at
		FROM unnest($1::text[], $2::text[], $3::bigint[], $4::text[], $5::text[], $6::text[], $7::text[],
			$8::text[], $9::text[], $10::text[], $11::bigint[])
			AS t(id, app, profile_id, category, title, body, link, image_url, data, dedup_key, created_at)
		ON CONFLICT (app, profile_id, dedup_key) WHERE dedup_key <> '' DO NOTHING`,
		pq.Array(ids), pq.Array(apps), pq.Array(profiles), pq.Array(categories), pq.Array(titles), pq.Array(bodies),
		pq.Array(links), pq.Array(images), pq.Array(datas), pq.Array(dedups), pq.Array(created))
	if err != nil {
		return 0, fmt.Errorf("failed to insert in-app notifications: %w", err)
	}
	stored, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to count inserted notifications: %w", err)
	}
	return int(stored), nil
}

func (r *Repository) ListInAppNotifications(ctx context.Context, app string, profileID int64, unreadOnly bool, limit, offset int) ([]*dao.InAppNotification, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT `+inAppColumns+` FROM in_app_notifications
		WHERE app = $1 AND profile_id = $2 AND (NOT $3 OR read_at = 0)
		ORDER BY created_at DESC, id
		LIMIT $4 OFFSET $5`, app, profileID, unreadOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list in-app notifications: %w", err)
	}
	defer rows.Close()

	var out []*dao.InAppNotification
	for rows.Next() {
		var n dao.InAppNotification
		var data []byte
		if err := rows.Scan(&n.ID, &n.App, &n.ProfileID, &n.Category, &n.Title, &n.Body, &n.Link,
			&n.ImageURL, &data, &n.DedupKey, &n.CreatedAt, &n.ReadAt); err != nil {
			return nil, fmt.Errorf("failed to decode in-app notification: %w", err)
		}
		if err := json.Unmarshal(data, &n.Data); err != nil {
			return nil, fmt.Errorf("failed to decode notification data: %w", err)
		}
		out = append(out, &n)
	}
	return out, rows.Err()
}

func (r *Repository) CountUnreadInAppNotifications(ctx context.Context, app string, profileID int64) (int32, error) {
	var count int32
	if err := r.conn.QueryRowContext(ctx,
		`SELECT count(*) FROM in_app_notifications WHERE app = $1 AND profile_id = $2 AND read_at = 0`,
		app, profileID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}
	return count, nil
}

// MarkInAppNotificationRead marks it read unless it already is (keeping the first read time).
func (r *Repository) MarkInAppNotificationRead(ctx context.Context, id string, profileID, readAt int64) (string, bool, error) {
	var app string
	err := r.conn.QueryRowContext(ctx, `
		UPDATE in_app_notifications SET read_at = CASE WHEN read_at = 0 THEN $3 ELSE read_at END
		WHERE id = $1 AND profile_id = $2
		RETURNING app`, id, profileID, readAt).Scan(&app)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to mark notification read: %w", err)
	}
	return app, true, nil
}

func (r *Repository) MarkAllInAppNotificationsRead(ctx context.Context, app string, profileID, readAt int64) error {
	if _, err := r.conn.ExecContext(ctx,
		`UPDATE in_app_notifications SET read_at = $3 WHERE app = $1 AND profile_id = $2 AND read_at = 0`,
		app, profileID, readAt); err != nil {
		return fmt.Errorf("failed to mark notifications read: %w", err)
	}
	return nil
}
