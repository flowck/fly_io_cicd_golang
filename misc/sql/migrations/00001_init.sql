-- +goose Up
-- +goose StatementBegin
CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE movies;
-- +goose StatementEnd
