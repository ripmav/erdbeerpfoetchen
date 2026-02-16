-- name: UpsertLamimiCollection :exec
INSERT INTO "pfoetchen"."lamimi_collection" ("collection_key", "collection_streamer", "collection_user", "collection_json")
VALUES ($1, $2, $3, $4)
ON CONFLICT ("collection_key", "collection_streamer", "collection_user")
DO UPDATE SET "collection_json" = $4;

-- name: GetLamimiCollection :one
SELECT * FROM "pfoetchen"."lamimi_collection"
WHERE "collection_key" = sqlc.arg(collection_key)
AND "collection_streamer" = sqlc.arg(collection_streamer)
AND "collection_user" = sqlc.arg(collection_user);

-- name: GetStreamer :one
SELECT * FROM "pfoetchen"."streamer"
WHERE "streamer_name" = sqlc.arg(streamer_name);

-- name: GetStreamerId :one
SELECT id FROM "pfoetchen"."streamer"
WHERE "streamer_name" = sqlc.arg(streamer_name);