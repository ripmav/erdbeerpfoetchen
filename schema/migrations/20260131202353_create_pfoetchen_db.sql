-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA "pfoetchen";
CREATE TABLE "pfoetchen"."lamimi_collection" (
    "collection_key" varchar,
    "collection_user" varchar,
    "collection_json" pg_catalog.json NOT NULL,
    PRIMARY KEY ("collection_key", "collection_user")
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "pfoetchen"."lamimi_collection";
DROP SCHEMA "pfoetchen";
-- +goose StatementEnd
