-- +goose Up
CREATE TABLE IF NOT EXISTS "todos" (
    "id" VARCHAR(36) PRIMARY KEY,
    "user_id" VARCHAR(255) NOT NULL,
    "title" VARCHAR(255) NOT NULL,
    "completed" BOOLEAN DEFAULT false,
    "created_at" TIMESTAMP WITH TIME ZONE,
    "updated_at" TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE IF EXISTS "todos";
