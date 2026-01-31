package postgres

import (
	"context"
	sql2 "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/port/repo"
	uuid2 "github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var _ repo.Jobs = (*JobsRepo)(nil)

type JobsRepo struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewJobsRepo(ctx context.Context, cfg config.StorageConfig, logger *zap.Logger) *JobsRepo {
	pl, err := pgxpool.New(ctx, getConnString(cfg))
	if err != nil {
		panic(err)
	}
	return &JobsRepo{
		pool:   pl,
		logger: logger,
	}
}

func getConnString(cfg config.StorageConfig) string {
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		hostPort,
		cfg.Database)
}

func (r *JobsRepo) Create(ctx context.Context, job *repo.JobDTO) error {
	ib := sqlbuilder.PostgreSQL.NewInsertBuilder()

	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	ib.InsertInto("job").
		Cols("id", "kind", "once", "interval_seconds", "payload", "status").
		Values(
			job.ID,
			job.Kind,
			sql2.NullInt64{
				Valid: job.Once != nil,
				Int64: func() int64 {
					if job.Once != nil {
						return *job.Once
					}
					return 0
				}(),
			},
			job.Interval,
			payloadJSON,
			job.Status)

	r.logger.Info(
		"Job created",
		zap.String("job_id", job.ID.String()),
		zap.Int("job_kind", job.Kind),
		zap.Any("payload", job.Payload),
	)

	sql, args := ib.Build()

	_, err = r.pool.Exec(ctx, sql, args...)

	if err != nil {
		r.logger.Error("Failed to insert", zap.Error(err))
		return fmt.Errorf("failed to insert job: %w", err)
	}

	return nil
}

func (r *JobsRepo) Read(ctx context.Context, jobID uuid2.UUID) (*repo.JobDTO, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	sb.Select("id", "kind", "interval_seconds", "once", "payload", "status", "last_finished_at").
		From("job").
		Where(sb.Equal("id", jobID))

	sql, args := sb.Build()

	var (
		id             uuid2.UUID
		kind           int
		interval       int64
		once           *int64
		payloadJSON    []byte
		status         string
		lastFinishedAt int64
	)

	err := r.pool.QueryRow(ctx, sql, args...).Scan(
		&id, &kind, &interval, &once, &payloadJSON, &status, &lastFinishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrJobIDNotFound
		}
		return nil, fmt.Errorf("failed to read job: %w", err)
	}

	var payload map[string]any
	if len(payloadJSON) > 0 {
		if err := json.Unmarshal(payloadJSON, &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	res := &repo.JobDTO{
		ID:             jobID,
		Kind:           kind,
		Interval:       interval,
		Payload:        payload,
		Status:         status,
		LastFinishedAt: lastFinishedAt,
		Once:           once,
	}

	return res, nil
}

func (r *JobsRepo) Upsert(ctx context.Context, job []*repo.JobDTO) error {
	ib := sqlbuilder.NewInsertBuilder()

	for _, j := range job {
		payloadJSON, err := json.Marshal(j.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal job payload: %w", err)
		}

		query := `
			INSERT INTO job (id, kind, status, interval_seconds, once, last_finished_at, payload)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				status = EXCLUDED.status,
				interval_seconds = EXCLUDED.interval_seconds,
				last_finished_at = EXCLUDED.last_finished_at
		`

		// ib.
		//	InsertInto("job").
		//	Cols("id", "kind", "once", "interval_seconds", "payload", "status", "last_finished_at").
		//	Values(
		//		j.ID,
		//		j.Kind,
		//		sql2.NullInt64{
		//			Valid: j.Once != nil,
		//			Int64: func() int64 {
		//				if j.Once != nil {
		//					return *j.Once
		//				}
		//				return 0
		//			}(),
		//		},
		//		j.Interval,
		//		payloadJSON,
		//		j.Status,
		//		j.LastFinishedAt).
		//	SQL("ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, interval_seconds = EXCLUDED.interval_seconds, last_finished_at = EXCLUDED.last_finished_at;")
		//
		// sql, args := ib.Build()

		_, err = r.pool.Exec(ctx, query,
			j.ID,
			j.Kind,
			j.Status,
			j.Interval,
			sql2.NullInt64{
				Valid: j.Once != nil,
				Int64: func() int64 {
					if j.Once != nil {
						return *j.Once
					}
					return 0
				}(),
			},
			j.LastFinishedAt,
			payloadJSON,
		)
		if err != nil {
			r.logger.Error("Failed to upsert", zap.Error(err),
				zap.String("query", ib.String()))
			return fmt.Errorf("failed to upsert job: %w", err)
		}
	}

	return nil
}

func (r *JobsRepo) Delete(ctx context.Context, jobID uuid2.UUID) error {
	db := sqlbuilder.PostgreSQL.NewDeleteBuilder()

	db.DeleteFrom("job").
		Where(db.Equal("id", jobID))

	sql, args := db.Build()

	result, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("job with ID %s not found", jobID)
	}

	return nil
}

func (r *JobsRepo) List(ctx context.Context) ([]*repo.JobDTO, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	sb.Select("id", "kind", "interval_seconds", "once", "payload", "status", "last_finished_at").
		From("job")

	sql, args := sb.Build()

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*repo.JobDTO

	for rows.Next() {
		var (
			id             uuid2.UUID
			kind           int
			interval       int64
			once           *int64
			payloadJSON    []byte
			status         string
			lastFinishedAt int64
		)

		err := rows.Scan(
			&id, &kind, &interval, &once, &payloadJSON, &status, &lastFinishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read job: %w", err)
		}

		var payload map[string]any
		if len(payloadJSON) > 0 {
			if err := json.Unmarshal(payloadJSON, &payload); err != nil {
				return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
			}
		}

		job := &repo.JobDTO{
			ID:             id,
			Kind:           kind,
			Interval:       interval,
			Payload:        payload,
			Status:         status,
			LastFinishedAt: lastFinishedAt,
			Once:           once,
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating job rows: %w", err)
	}

	return jobs, nil
}
