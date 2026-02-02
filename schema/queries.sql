-- name: UpsertLamimiCollection :exec
INSERT INTO "pfoetchen"."lamimi_collection" ("collection_key", "collection_user", "collection_json")
VALUES ($1, $2, $3)
ON CONFLICT ("collection_key")
DO UPDATE SET "collection_json" = $3;

-- name: GetCollection :one
SELECT * FROM "pfoetchen"."lamimi_collection"
WHERE "collection_key" = sqlc.arg(collection_key)
AND "collection_user" = sqlc.arg(collection_user);