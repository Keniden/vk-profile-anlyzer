-- +goose Up
-- +goose StatementBegin
CREATE TABLE articles(
    id BIGSERIAL PRIMARY KEY NOT NULL,
    name text NOT NULL DEFAULT '',
    rating int not null DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE articles;
-- +goose StatementEnd
