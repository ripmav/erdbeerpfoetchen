-- name: UpsertLamimiCollection :exec
INSERT INTO "pfoetchen"."lamimi_collection" ("collection_key", "collection_json")
VALUES ($1, $2)
ON CONFLICT ("collection_key")
DO UPDATE SET "collection_json" = $2;

-- name: GetCollection :one
SELECT * FROM "pfoetchen"."lamimi_collection"
WHERE "collection_key" = sqlc.arg(collection_key);