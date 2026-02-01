package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	upsertOnConflit = `ON CONFLICT (id) DO UPDATE SET
			worker_id = EXCLUDED.worker_id,
			status = EXCLUDED.status,
			started_at = EXCLUDED.started_at,
			finished_at = EXCLUDED.finished_at,
			error = EXCLUDED.error
			`

	tableExecutions = "execution"

	columnID         = "id"
	columnJobID      = "job_id"
	columnWorkerId   = "worker_id"
	columnStatus     = "status"
	columnStartedAt  = "started_at"
	columnFinishedAt = "finished_at"
	columnError      = "error"
	columnQueuedAt   = "queued_at"
)

var (
	executionColumns = []string{
		columnID,
		columnJobID,
		columnWorkerId,
		columnStatus,
		columnStartedAt,
		columnFinishedAt,
		columnError,
		columnQueuedAt,
	}
)

type ExecutionRepo struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewExecutionRepo(pool *pgxpool.Pool, logger *zap.Logger) *ExecutionRepo {
	return &ExecutionRepo{
		pool:   pool,
		logger: logger,
	}
}

func (repo *ExecutionRepo) Upsert(ctx context.Context, exec *entity.Execution) error {
	query, args := sqlbuilder.
		PostgreSQL.
		NewInsertBuilder().
		InsertInto(tableExecutions).
		Cols(executionColumns...).
		Values(
			exec.ID,
			exec.JobID,
			exec.WorkerID,
			exec.Status,
			exec.StartedAt,
			exec.FinishedAt,
			exec.Error,
			exec.QueuedAt,
		).
		SQL(upsertOnConflit).Build()

	_, err := repo.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ExecutionRepo) Read(ctx context.Context, id uuid.UUID) (*entity.Execution, error) {
	// TODO implement me
	panic("implement me")
}

func (repo *ExecutionRepo) List(ctx context.Context, filter *entity.ListExecutionFilter) ([]*entity.Execution, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	query, args := sb.Select(executionColumns...).From(tableExecutions).Where(sb.Equal(columnJobID, filter.JobIDs[0])).Build()

	repo.logger.Info("Sql query", zap.String("query", query))

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			repo.logger.Info("no rows found", zap.String("query", query), zap.String("id", filter.JobIDs[0].String()))
			return []*entity.Execution{}, nil
		}
		repo.logger.Error("failed to execute query", zap.String("query", query), zap.String("id", filter.JobIDs[0].String()))
		return nil, err
	}

	result := make([]*entity.Execution, 0)

	var (
		id         uuid.UUID
		jobID      uuid.UUID
		workerID   string
		status     string
		queuedAt   int64
		startedAt  *int64
		finishedAt *int64
		jobError   *[]byte
	)

	for rows.Next() {
		err := rows.Scan(
			&id, &jobID, &workerID, &status, &startedAt, &finishedAt, &jobError, &queuedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var errorDetails *entity.ExecutionError
		if jobError != nil && len(*jobError) > 0 {
			if err := json.Unmarshal(*jobError, &errorDetails); err != nil {
				return nil, fmt.Errorf("failed to unmarshal jobError: %w", err)
			}
		}

		result = append(result, &entity.Execution{
			ID:         id,
			JobID:      jobID,
			WorkerID:   workerID,
			Status:     entity.ExecutionStatus(status),
			QueuedAt:   queuedAt,
			StartedAt:  startedAt,
			FinishedAt: finishedAt,
			Error:      errorDetails,
		})
	}

	return result, nil
}
