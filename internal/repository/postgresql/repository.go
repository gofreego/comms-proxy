package postgresql

import (
	"context"
	"database/sql"

	"github.com/gofreego/goutils/databases/connections/sql/pgsql"
)

type Repository struct {
	conn *sql.DB
}

func NewRepository(ctx context.Context, cfg *pgsql.Config) (*Repository, error) {
	conn, err := pgsql.GetConnection(cfg)
	if err != nil {
		return nil, err
	}
	return &Repository{conn: conn}, nil
}

func (r *Repository) Close(ctx context.Context) error {
	return r.conn.Close()
}
