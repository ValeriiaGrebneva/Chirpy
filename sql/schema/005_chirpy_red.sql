-- +goose Up
ALTER TABLE users
ADD is_chirpy_red BOOLEAN NOT NULL
CONSTRAINT default_user_red DEFAULT FALSE;

-- +goose Down
ALTER TABLE users
DROP COLUMN is_chirpy_red;