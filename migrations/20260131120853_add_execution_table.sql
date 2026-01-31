-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS execution (
    id uuid primary key,
    job_id uuid references job(id),
    worker_id text not null,
    status text not null,
    queued_at BIGINT not null,
    started_at bigint,
    finished_at bigint,
    error jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS execution;
-- +goose StatementEnd
