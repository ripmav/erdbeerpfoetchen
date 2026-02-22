-- +goose Up
-- +goose StatementBegin
ALTER TABLE "streaming"."collection"
ADD CONSTRAINT user_id_collection_fk
FOREIGN KEY ("streamer")
REFERENCES "streaming"."user" ("id")
ON DELETE CASCADE
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "streaming"."collection"
DROP CONSTRAINT user_id_collection_fk;
-- +goose StatementEnd
