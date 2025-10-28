-- +goose Up
-- +goose StatementBegin
create table if not exists job(
    id uuid primary key,
    inter text,
    payload jsonb,
    status text not null,
    created_at timestamp not null,
    last_finished_at timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists job;
-- +goose StatementEnd
