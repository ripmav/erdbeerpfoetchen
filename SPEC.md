# SPEC — erdbeerpfoetchen (Pfötchen API)

## §G Goal

Go REST API: store/retrieve per-viewer JSON collections keyed by streamer+collection_type. PostgreSQL backend. Auth via per-streamer Bearer token (UUID). Rate-limiting per streamer. Binary embeds migrations (goose). Docker Compose one-command setup.

## §C Constraints

- Go 1.26.2+ | PostgreSQL only
- No external migration tooling at runtime (embedded goose)
- Config priority: CLI flags > env vars (`PFOETCHEN_*`) > YAML file > defaults
- `--server.unsafe` required when no TLS configured
- Body limit: 12 MiB per request
- Default rate: 100 req/min per streamer (DB-configurable)
- No sub-agents, no orchestration binaries
- sqlc-generated DB code — do not edit `internal/database/model/` manually

## §I Interfaces

### HTTP Endpoints

**Auth (browser-facing, no Bearer):**

| Method | Path | Description | Response |
|--------|------|-------------|----------|
| GET | `/` | Login page | 200 HTML |
| GET | `/auth/twitch` | Redirect to Twitch OAuth | 307 |
| GET | `/auth/twitch/callback` | Exchange code → create/load user → show api_token | 200 HTML |

**Collection (require `Authorization: Bearer <api_token>`):**

| Method | Path | Headers | Response |
|--------|------|---------|----------|
| POST | `/api/v1/collection/{streamer_name}` | `X-USER-KEY`, `X-COLLECTION-KEY`, `X-COLLECTION-TYPE` (opt, default="default"), JSON body | 202 |
| GET  | `/api/v1/collection/{streamer_name}` | `X-USER-KEY`, `X-COLLECTION-KEY` | 200 + JSON |

### Domain Types

```
Collection     { RawMessage json.RawMessage; CollectionType string }
StreamingUser  { ID uuid; UserName string; ApiToken uuid; RateLimitPerMinute int32; TwitchID sql.NullString; IsAdmin bool }
StreamingCollection { CollectionKey string; Streamer uuid; Viewer string; Json json.RawMessage; CollectionType string }
```

### Key Interfaces

```
TokenValidator      GetUserApiToken(ctx, userName) (uuid.UUID, error)
RateLimitProvider   GetUserRateLimit(ctx, userName) (int32, error)
CollectionService   WriteCollection / ReadCollection
UserService         GetUserApiToken / GetUserByUserName
AuthUserService     GetUserByTwitchId / CreateUser
```

### Config (YAML / env / flags)

| Key | Env | Default |
|-----|-----|---------|
| `debug` | `PFOETCHEN_DEBUG` | false |
| `server.listen` | `PFOETCHEN_SERVER_LISTEN` | `:8080` |
| `server.unsafe` | `PFOETCHEN_SERVER_UNSAFE` | false |
| `server.tls.cert` | `PFOETCHEN_SERVER_TLS_CERT_PATH` | — |
| `server.tls.key` | `PFOETCHEN_SERVER_TLS_KEY_PATH` | — |
| `database.uri` | `PFOETCHEN_DATABASE_URI` | — |
| `twitch.client-id` | `PFOETCHEN_TWITCH_CLIENT_ID` | — |
| `twitch.client-secret` | `PFOETCHEN_TWITCH_CLIENT_SECRET` | — |
| `twitch.redirect-url` | `PFOETCHEN_TWITCH_REDIRECT_URL` | — |
| `twitch.admin-ids` | `PFOETCHEN_TWITCH_ADMIN_IDS` | — (comma-separated Twitch user IDs) |
| `--config` / `-c` | `PFOETCHEN_CONFIG` | — |

### CLI Commands

```
./api api      — start server (auto-migrates on startup)
./api migrate  — run migrations only, then exit
```

## §V Invariants

- Auth middleware validates Bearer token as UUID against DB per streamer — no hardcoded tokens
- Rate limiter uses per-streamer token bucket; falls back to 100 req/min on DB error
- `streaming.collection` PK = (collection_key, streamer, viewer) — upsert on conflict
- `streaming.user.api_token` = UUID, unique per user
- `streaming.user.twitch_id` = VARCHAR UNIQUE (nullable for pre-OAuth users)
- `streaming.user.is_admin` = BOOLEAN, never exposed to the user in the UI
- Migrations embedded in binary via goose; applied before server start
- `X-COLLECTION-TYPE` absent → stored as "default"
- `X-USER-KEY` or `X-COLLECTION-KEY` absent → 400
- Invalid/missing Bearer → 401; token mismatch → 403
- Rate exceeded → 429
- Rate limiter cleanup: buckets idle >3min are evicted
- Twitch OAuth CSRF: state is generated per login attempt, stored in a HttpOnly SameSite=Lax cookie, verified in callback
- First Twitch login creates a `streaming.user`; subsequent logins return the same `api_token`
- Admin status is set at creation from `twitch.admin-ids` config; not visible in the UI

## §T Tasks

| # | task | status | notes |
|---|------|--------|-------|
| 1 | DB token auth middleware | done | uuid comparison |
| 2 | Unit + integration tests | done | testcontainers |
| 3 | CI/CD pipeline | done | GitHub Actions |
| 4 | Multiple collection types | done | X-COLLECTION-TYPE header |
| 5 | Rate-limiting middleware | done | token bucket, per-streamer |
| 6 | OpenAPI docs | done | docs/openapi.yaml |
| 7 | slog context logging | done | |
| 8 | Debug mode DB connections | done | |
| 9 | YAML config support | done | kong resolver |
| 10 | Docker Compose setup | done | migrate svc + api svc |
| 11 | Graceful shutdown | done | SIGINT/SIGTERM |
| 12 | Twitch OAuth login | done | browser flow, first-login creates user, CSRF state cookie |
| 13 | Admin user designation | done | `twitch.admin-ids` config; is_admin stored in DB |
| 14 | User management API | open | no CRUD endpoints yet for users |
| 15 | TLS termination in-process | open | cert/key config exists, not wired |
| 16 | OpenAPI: X-COLLECTION-TYPE doc | done | documented in openapi.yaml |

## §B Bugs

| # | description | status |
|---|-------------|--------|
| — | — | — |
