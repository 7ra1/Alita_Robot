# Alita Robot - Agent Guide

Go 1.26 Telegram bot (gotgbot, PostgreSQL + GORM, Redis). Module-based architecture in `alita/modules/`.

## Commands

```bash
make run              # go run main.go
make build            # goreleaser snapshot
make lint             # golangci-lint --timeout 10m
make test             # race detection + coverage (10m timeout)
go test -v -run TestName ./alita/db   # single test
go test -count=1 -timeout 10m ./...   # all tests, no cache
make psql-migrate    # requires PSQL_DB_HOST, PSQL_DB_NAME, PSQL_DB_USER, PSQL_DB_PASSWORD
make check-translations
```

## Module System

- Each module: `LoadXxx(dispatcher)` in `alita/modules/`, registered in `alita/main.go:LoadModules()`
- **Load order matters** — help module loads last to collect all registered commands
- `antiflood` and `antispam` use `init()` for background goroutines
- Handler groups: negative (-1) for early interception, 0 for commands, 4-10 for watchers
- Returns: `ext.EndGroups` (stop), `ext.ContinueGroups` (continue)

## Critical Patterns

**DB errors:** Never `_ = db.X()`. Nil returns cause panics. Wrap with `errors.Wrap(err, "context")`.

**Nil sender:** `ctx.EffectiveSender` can be nil for channel messages. Check before `.User`.

**Cache invalidation:** Every DB write must invalidate `alita:{module}:{identifier}` keys. Use `getFromCacheOrLoad()` in `alita/db/cache_helpers.go` for reads.

**Callback data:** Use `callbackcodec.Encode/Decode`, never `strings.Split`. Format: `<namespace>|v1|<url-encoded>`.

**i18n:** Add keys to ALL locale files in `locales/`. Verify with `grep locales/`. YAML double-quote escape sequences. Convert Markdown→HTML via `tgmd2html.MD2HTMLV2()`.

**Schema sync:** Add migration → update struct in `alita/db/` → update optimized queries → update function.

**Double-answer bug:** `RequireUserAdmin(ctx, justCheck=false)` already answers callback. Don't answer again.

**Entity checks:** Check both `msg.Entities` AND `msg.CaptionEntities`.

## Environment

Required: `BOT_TOKEN`, `OWNER_ID`, `MESSAGE_DUMP`

**Database (choose one):**
- PostgreSQL (default): `DATABASE_URL`, `AUTO_MIGRATE=true`
- SQLite (dev/test): `DB_TYPE=sqlite DB_PATH=./alita.db`

Key vars: `HTTP_PORT=8080` (health/metrics/webhook), `USE_WEBHOOKS`

See `sample.env` for full list.

## Database

- **PostgreSQL**: Full support, SQL migrations in `migrations/`
- **SQLite**: Dev/test use, AutoMigrate via `DB_TYPE=sqlite`
- Single connection for SQLite (locking): `DB_PATH` defaults to `./alita.db`
- JSON fields: Custom `Scanner`/`Valuer` types handle both `[]byte` (PG) and `string` (SQLite)

## Pre-commit

`pip install pre-commit && pre-commit install`

Runs: `golangci-lint`, `gofmt -l -w`, `go mod tidy`, trailing-whitespace, check-yaml, detect-private-key.

## Commit Style

`feat:`, `fix:`, `refactor:`, `perf:`, `test:`, `docs:`, `chore:`, `deps:`

PRs: link issues, pass `make test`, run `make lint`, update all locale files.
