-- +goose Up
ALTER TABLE "streaming"."user"
    ADD COLUMN "twitch_id" VARCHAR UNIQUE,
    ADD COLUMN "is_admin" BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE "streaming"."user"
    DROP COLUMN "twitch_id",
    DROP COLUMN "is_admin";
