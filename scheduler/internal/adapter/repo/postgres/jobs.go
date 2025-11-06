package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/port/repo"
	uuid2 "github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"time"
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
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database)
}

func (r *JobsRepo) Create(ctx context.Context, job *repo.JobDTO) error {
	ib := sqlbuilder.PostgreSQL.NewInsertBuilder()

	var payloadValue interface{} = nil
	if job.Payload != nil {
		jsonBytes, err := json.Marshal(job.Payload)
		if err != nil {
			return err
		}
		payloadValue = string(jsonBytes)
	}

	var intervalValue interface{} = nil
	if job.Interval != nil {
		intervalValue = *job.Interval
	}

	ib.InsertInto("job").
		Cols("id", "inter", "payload", "status", "created_at").
		Values(job.Id, intervalValue, payloadValue, job.Status, time.UnixMilli(job.CreatedAt))

	sql, args := ib.Build()

	_, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}

	return nil
}

func (r *JobsRepo) Read(ctx context.Context, jobID uuid2.UUID) (*repo.JobDTO, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	sb.Select("id", "inter", "payload", "status", "created_at", "last_finished_at").
		From("job").
		Where(sb.Equal("id", jobID))

	sql, args := sb.Build()

	var (
		id             uuid2.UUID
		interval       *string
		payloadData    []byte
		status         string
		createdAt      time.Time
		lastFinishedAt *time.Time
	)

	err := r.pool.QueryRow(ctx, sql, args...).Scan(
		&id, &interval, &payloadData, &status, &createdAt, &lastFinishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read job: %w", err)
	}

	var payload map[string]interface{}
	if payloadData != nil {
		if err := json.Unmarshal(payloadData, &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	var lastFinishedAtInt64 int64
	if lastFinishedAt != nil {
		lastFinishedAtInt64 = lastFinishedAt.UnixMilli()
	}

	res := &repo.JobDTO{
		Id:             jobID,
		Interval:       interval,
		Payload:        payload,
		Status:         status,
		CreatedAt:      createdAt.UnixMilli(),
		LastFinishedAt: lastFinishedAtInt64,
	}

	if interval == nil {
		once := "once"
		res.Once = &once
	}

	return res, nil
}

func (r *JobsRepo) Update(ctx context.Context, job *repo.JobDTO) error {
	panic("not implemented")
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

func (r *JobsRepo) List(ctx context.Context, status string) ([]*repo.JobDTO, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	sb.Select("id", "inter", "payload", "status", "created_at", "last_finished_at").
		From("job").Where(sb.Equal("status", status))

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
			interval       *string
			payloadData    []byte
			status         string
			createdAt      time.Time
			lastFinishedAt *time.Time
		)

		err := rows.Scan(&id, &interval, &payloadData, &status, &createdAt, &lastFinishedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job row: %w", err)
		}

		var payload map[string]interface{}
		if payloadData != nil {
			if err := json.Unmarshal(payloadData, &payload); err != nil {
				return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
			}
		}

		var lastFinishedAtInt64 int64
		if lastFinishedAt != nil {
			lastFinishedAtInt64 = lastFinishedAt.UnixMilli()
		}

		job := &repo.JobDTO{
			Id:             id,
			Interval:       interval,
			Payload:        payload,
			Status:         status,
			CreatedAt:      createdAt.UnixMilli(),
			LastFinishedAt: lastFinishedAtInt64,
		}

		if interval == nil {
			once := "once"
			job.Once = &once
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating job rows: %w", err)
	}

	return jobs, nil
}
