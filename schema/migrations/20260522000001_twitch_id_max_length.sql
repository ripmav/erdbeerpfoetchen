-- +goose Up
ALTER TABLE "streaming"."user" ALTER COLUMN "twitch_id" TYPE VARCHAR(25);

-- +goose Down
ALTER TABLE "streaming"."user" ALTER COLUMN "twitch_id" TYPE VARCHAR;
