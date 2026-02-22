-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA "streaming";

CREATE TABLE "streaming"."collection" (
    "key" varchar NOT NULL,
    "streamer" uuid NOT NULL,
    "viewer" varchar NOT NULL,
    "json" jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY ("key", "streamer", "viewer")
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "streaming"."collection";

DROP SCHEMA "streaming";
-- +goose StatementEnd
