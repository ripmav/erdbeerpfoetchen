-- +goose Up
ALTER TABLE "streaming"."collection"
RENAME COLUMN "key" TO "collection_key";

-- +goose Down
ALTER TABLE "streaming"."collection"
RENAME COLUMN "collection_key" TO "key";
