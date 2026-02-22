-- name: UpsertCollection :exec
INSERT INTO "pfoetchen"."collection" ("key", "streamer", "viewer", "json")
VALUES ($1, $2, $3, $4)
ON CONFLICT ("key", "streamer", "viewer")
DO UPDATE SET "json" = $4;

-- name: GetCollection :one
SELECT * FROM "pfoetchen"."collection"
WHERE "key" = sqlc.arg(key)
AND "streamer" = sqlc.arg(streamer)
AND "viewer" = sqlc.arg(user_hash);

-- name: GetUser :one
SELECT * FROM "pfoetchen"."user"
WHERE "user_name" = sqlc.arg(user_name);

-- name: GetUserId :one
SELECT id FROM "pfoetchen"."user"
WHERE "user_name" = sqlc.arg(user_name);

-- name: GetUserApiToken :one
SELECT api_token FROM "pfoetchen"."user"
WHERE "user_name" = sqlc.arg(user_name);