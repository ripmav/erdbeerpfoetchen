-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA "pfoetchen";
CREATE TABLE "pfoetchen"."lamimi_collection" (
    "collection_key" varchar NOT NULL,
    "collection_streamer" uuid NOT NULL,
    "collection_user" varchar NOT NULL,
    "collection_json" jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY ("collection_key", "collection_user")
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "pfoetchen"."lamimi_collection";
DROP SCHEMA "pfoetchen";
-- +goose StatementEnd
