-- +goose Up
-- +goose StatementBegin
CREATE TABLE "pfoetchen"."streamer" (
    "id" uuid NOT NULL DEFAULT uuidv7(),
    "streamer_name" varchar NOT NULL,
    "api_token" uuid  NOT NULL DEFAULT gen_random_uuid(),
    PRIMARY KEY ("id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "pfoetchen"."streamer";
-- +goose StatementEnd
