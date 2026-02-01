package postgres

import (
	"context"

	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, cfg config.StorageConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, getConnString(cfg))
	if err != nil {
		return nil, err
	}
	return pool, nil
}
