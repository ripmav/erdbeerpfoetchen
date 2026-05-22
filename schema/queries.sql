-- name: UpsertCollection :exec
INSERT INTO "streaming"."collection" ("collection_key", "streamer", "viewer", "json", "collection_type")
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT ("collection_key", "streamer", "viewer")
DO UPDATE SET "json" = $4, "collection_type" = $5;

-- name: GetCollection :one
SELECT * FROM "streaming"."collection"
WHERE "collection_key" = sqlc.arg(key)
AND "streamer" = sqlc.arg(streamer)
AND "viewer" = sqlc.arg(user_hash);

-- name: GetUserByUserName :one
SELECT id, user_name, api_token, rate_limit_per_minute, twitch_id, is_admin FROM "streaming"."user"
WHERE "user_name" = sqlc.arg(user_name);

-- name: GetUserById :one
SELECT id, user_name, api_token, rate_limit_per_minute, twitch_id, is_admin FROM "streaming"."user"
WHERE "id" = sqlc.arg(id);

-- name: GetUserByTwitchId :one
SELECT id, user_name, api_token, rate_limit_per_minute, twitch_id, is_admin FROM "streaming"."user"
WHERE "twitch_id" = sqlc.arg(twitch_id);

-- name: CreateUser :one
INSERT INTO "streaming"."user" ("user_name", "twitch_id", "api_token", "rate_limit_per_minute", "is_admin")
VALUES (sqlc.arg(user_name), sqlc.arg(twitch_id), sqlc.arg(api_token), sqlc.arg(rate_limit_per_minute), sqlc.arg(is_admin))
RETURNING id, user_name, api_token, rate_limit_per_minute, twitch_id, is_admin;

-- name: GetUserId :one
SELECT id FROM "streaming"."user"
WHERE "user_name" = sqlc.arg(user_name);

-- name: GetUserApiToken :one
SELECT api_token FROM "streaming"."user"
WHERE "user_name" = sqlc.arg(user_name);

-- name: GetUserRateLimit :one
SELECT rate_limit_per_minute FROM "streaming"."user"
WHERE "user_name" = sqlc.arg(user_name);
