package repository

import (
	"context"
	"sync"

	"github.com/gofreego/comms-proxy/internal/repository/postgresql"
	"github.com/gofreego/comms-proxy/internal/service"

	"github.com/gofreego/goutils/databases/connections/sql/pgsql"
)

type Config struct {
	Postgres pgsql.Config `yaml:"Postgres"`
}

var (
	instance service.Repository
	once     sync.Once
)

// GetInstance returns the shared repository, connecting on first use.
func GetInstance(ctx context.Context, cfg *Config) service.Repository {
	once.Do(func() {
		repo, err := postgresql.NewRepository(ctx, &cfg.Postgres)
		if err != nil {
			panic("failed to create repository: " + err.Error())
		}
		instance = repo
	})
	return instance
}
