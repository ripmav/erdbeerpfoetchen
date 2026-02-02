-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA "pfoetchen";
CREATE TABLE "pfoetchen"."lamimi_collection" (
    "collection_key" varchar PRIMARY KEY,
    "collection_user" varchar PRIMARY KEY,
    "collection_json" pg_catalog.json NOT NULL
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "pfoetchen"."lamimi_collection";
DROP SCHEMA "pfoetchen";
-- +goose StatementEnd
