-- +goose Up
CREATE TABLE IF NOT EXISTS "users" (
    "id" VARCHAR(36) PRIMARY KEY,
    "email" VARCHAR(255),
    "name" VARCHAR(255),
    "created_at" TIMESTAMP WITH TIME ZONE,
    "updated_at" TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE IF EXISTS "users";
