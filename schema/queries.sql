-- name: UpsertCollection :exec
INSERT INTO "streaming"."collection" ("key", "streamer", "viewer", "json")
VALUES ($1, $2, $3, $4)
ON CONFLICT ("key", "streamer", "viewer")
DO UPDATE SET "json" = $4;

-- name: GetCollection :one
SELECT * FROM "streaming"."collection"
WHERE "key" = sqlc.arg(key)
AND "streamer" = sqlc.arg(streamer)
AND "viewer" = sqlc.arg(user_hash);

-- name: GetUserByUserName :one
SELECT * FROM "streaming"."user"
WHERE "user_name" = sqlc.arg(user_name);

-- name: GetUserById :one
SELECT * FROM "streaming"."user"
WHERE "id" = sqlc.arg(id);

-- name: GetUserId :one
SELECT id FROM "streaming"."user"
WHERE "user_name" = sqlc.arg(user_name);

-- name: GetUserApiToken :one
SELECT api_token FROM "streaming"."user"
WHERE "user_name" = sqlc.arg(user_name);