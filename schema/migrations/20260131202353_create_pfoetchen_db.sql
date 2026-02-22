-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA "pfoetchen";
CREATE TABLE "pfoetchen"."collection" (
    "key" varchar NOT NULL,
    "streamer" uuid NOT NULL,
    "viewer" varchar NOT NULL,
    "json" jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY ("key", "streamer", "viewer")
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "pfoetchen"."collection";
DROP SCHEMA "pfoetchen";
-- +goose StatementEnd
