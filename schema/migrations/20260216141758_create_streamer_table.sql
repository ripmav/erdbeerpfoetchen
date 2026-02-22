-- +goose Up
-- +goose StatementBegin
CREATE TABLE "pfoetchen"."user" (
    "id" uuid NOT NULL DEFAULT uuidv7(),
    "user_name" varchar NOT NULL,
    "api_token" uuid  NOT NULL DEFAULT gen_random_uuid(),
    PRIMARY KEY ("id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "pfoetchen"."user";
-- +goose StatementEnd
