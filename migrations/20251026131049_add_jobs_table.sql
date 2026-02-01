-- +goose Up
-- +goose StatementBegin
create table if not exists job(
    id uuid primary key,
    kind int not null,
    status text not null,
    interval_seconds bigint not null default 0,
    once bigint,
    last_finished_at bigint not null default 0,
    payload jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists job;
-- +goose StatementEnd
