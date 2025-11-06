-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    uodated_at TIMESTAMP NOT NULL,
    email TEXT UNIQUE
);

-- +goose Down
DROP TABLE users;