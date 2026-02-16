-- +goose Up
-- +goose StatementBegin
ALTER TABLE "pfoetchen"."lamimi_collection"
ADD CONSTRAINT streamer_id_collection_fk
FOREIGN KEY ("collection_streamer")
REFERENCES "pfoetchen"."streamer" ("id")
ON DELETE CASCADE
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "pfoetchen"."lamimi_collection"
DROP CONSTRAINT streamer_id_collection_fk;
-- +goose StatementEnd
