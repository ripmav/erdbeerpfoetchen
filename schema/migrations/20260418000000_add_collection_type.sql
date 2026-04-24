-- +goose Up
ALTER TABLE "streaming"."collection"
    ADD COLUMN "collection_type" VARCHAR NOT NULL DEFAULT 'default';

-- +goose Down
ALTER TABLE "streaming"."collection"
    DROP COLUMN "collection_type";
