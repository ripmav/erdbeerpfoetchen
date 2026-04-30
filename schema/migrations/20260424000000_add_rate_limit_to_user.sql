-- +goose Up
ALTER TABLE "streaming"."user" ADD COLUMN "rate_limit_per_minute" INTEGER NOT NULL DEFAULT 100;

-- +goose Down
ALTER TABLE "streaming"."user" DROP COLUMN "rate_limit_per_minute";
