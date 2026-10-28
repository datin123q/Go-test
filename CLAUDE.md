# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ESJP (EnergyShares Japan) backend — a Go monorepo with four applications deployed on AWS ECS:
- **admin-api**: Admin management API
- **user-api**: End-user API
- **jobs-scheduler**: Background job scheduler
- **invest-worker**: Investment processing worker

**Stack**: Go 1.25.1, Echo v4, XORM (MySQL), Kafka, uber/fx (DI), OpenAPI 3.1.0

## Workflow Orchestration

### 1. Plan Node Default
- Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions)
- If something goes sideways, STOP and re-plan immediately – don't keep pushing
- Use plan mode for verification steps, not just building
- Write detailed specs upfront to reduce ambiguity

### 2. Subagent Strategy
- Use subagents liberally to keep main context window clean
- Offload research, exploration, and parallel analysis to subagents
- For complex problems, throw more compute at it via subagents
- One task per subagent for focused execution

### 3. Self-Improvement Loop
- After ANY correction from the user: update `tasks/lessons.md` with the pattern
- Write rules for yourself that prevent the same mistake
- Ruthlessly iterate on these lessons until mistake rate drops
- Review lessons at session start for relevant project

### 4. Verification Before Done
- Never mark a task complete without proving it works
- Diff behavior between main and your changes when relevant
- Ask yourself: "Would a staff engineer approve this?"
- Run tests, check logs, demonstrate correctness

### 5. Demand Elegance (Balanced)
- For non-trivial changes: pause and ask "is there a more elegant way?"
- If a fix feels hacky: "Knowing everything I know now, implement the elegant solution"
- Skip this for simple, obvious fixes – don't over-engineer
- Challenge your own work before presenting it

### 6. Autonomous Bug Fixing
- When given a bug report: just fix it. Don't ask for hand-holding
- Point at logs, errors, failing tests – then resolve them
- Zero context switching required from the user
- Go fix failing CI tests without being told how

## Task Management

1. **Plan First**: Write plan to `tasks/todo.md` with checkable items
2. **Verify Plan**: Check in before starting implementation
3. **Track Progress**: Mark items complete as you go
4. **Explain Changes**: High-level summary at each step
5. **Document Results**: Add review section to `tasks/todo.md`
6. **Capture Lessons**: Update `tasks/lessons.md` after corrections

## Core Principles

- **Simplicity First**: Make every change as simple as possible. Impact minimal code.
- **No Laziness**: Find root causes. No temporary fixes. Senior developer standards.
- **Minimal Impact**: Changes should only touch what's necessary. Avoid introducing bugs.

## ローカルDBのデータは `make db-setup` で全部消える

**`db/sql/version/energyshares.base.schema.sql` は `DROP TABLE` を含むスキーマ定義**なので、
`make db-setup` を流すと**110テーブルが作り直され、投入したデータがすべて失われる。**
環境を立て直すたびに、手で作った会員・ファンド・出資のデータが消える。

```bash
make db-setup     # スキーマを流し直す + マイグレーション → データが全部消える
make db-migrate   # マイグレーションだけ → データは残る
```

**既存のデータを残したいときは `db-migrate` だけを流すこと。**マイグレーションの追加を取り込むだけなら
こちらで足りる。`db-setup` が要るのは、テーブル定義そのものを作り直すときだけ。

**2026-09-08 に踏んだ。**前日に `db-setup` を流した結果、会員もファンドも全部消えており、
画面を開いて初めて気付いた。**ボリューム（`energy-shares-backend-repo_energysharesdb-data`）は
残っていても、中身が作り直されているので復元できない。**

**消えたあとの復旧。**

```bash
make seed-sample-data   # db/sample_data/*.sql を投入する
                        # user / private_profile / user_account / user_transaction /
                        # user_virtual_account / va_deposit / sign_up_request / ekyc_request_log
```

**ファンド（`project`）と出資（`share`）は sample_data に含まれていない。**画面から作るか、
自分で SQL を用意することになる。**残高計算（`user_asset_view`）が `share` /
`withdrawal` / `deposit_request` を集計するため、整合の取れないデータを入れると
出資処理や残高表示でエラーになる。**


## Common Commands

```bash
# Workspace setup (run once, or after go.work changes)
make work
make work-sync

# Run all tests (requires VPN + MFA session)
make test

# Run tests for a single app
cd apps/admin-api && go test -v ./...
cd apps/user-api && go test -v ./...

# Run a specific test
go test -v -run Test_adminUsecase_GetAll ./apps/admin-api/usecase/admin_user/...

# Generate mocks (requires: brew install mockery)
make mock

# Generate Swagger JSON
make gen-swagger

# Run apps locally
make start-admin-api    # admin-api on local env
make start-user-api     # user-api on local env
make start-jobs-scheduler

# Local MySQL
make mysql_dump          # dump DEV DB
make mysql_run           # start MySQL container
make db-setup            # seed schemas + run migrations
make db-migrate          # run migrations only
make seed-sample-data    # load sample data
make mysql-client        # connect to local MySQL

# DB migration: create new migration file
goose create <migration_name> sql
# Run migration
cd apps/admin-api && go run main.go migrate --env local

# Code generation from YAML definitions
make codegen

# Environment config (requires AWS MFA)
./paramstore.sh get dev       # download dev config
./paramstore.sh get all       # download all envs
./paramstore.sh set dev qa    # upload configs
```

## Architecture

The project follows **Go Clean Architecture** with four layers, wired via `go.uber.org/fx`:

```
Delivery (HTTP/Kafka) → Usecase (business logic) → Service (shared logic) → Repository (data access)
```

### Key directories

| Path | Purpose |
|------|---------|
| `clean-arch-domain/` | Shared interfaces for Repository and Service layers + query option builders |
| `domain/` | Database model structs with XORM tags |
| `repository/{domain}/mysql/` | MySQL repository implementations |
| `service/` | Cross-cutting business services (auth, email, file upload, etc.) |
| `apps/{app}/domain/` | App-specific usecase interfaces + Request/Response DTOs |
| `apps/{app}/usecase/{domain}/` | Usecase implementations |
| `apps/{app}/delivery/{domain}/http/` | HTTP handlers (Echo) |
| `apps/{app}/server/bootstrap.go` | fx dependency injection wiring |
| `pkg/` | Shared utilities (orm wrapper, factory, config, helpers) |

### Layer conventions

- **Repository**: CRUD naming (`Create`, `GetAll`, `GetById`, `GetBy{Field}`, `Update`, `Delete`). All deletes are **soft delete** (sets `DeletedAt`). Uses functional options pattern for query filtering/pagination.
- **Service**: Optional layer for reusable cross-domain logic (auth, permissions, email, S3). Not every domain needs one.
- **Usecase**: 1:1 with Delivery. Accepts `{Domain}{Action}Request`, returns `{Domain}{Action}Response`. All requests include `JWTUser` and `ClientIp` for audit. Manages transactions via `factory.TxManager`.
- **Delivery**: HTTP handlers using Echo. Handles bind/validate, JWT extraction, error mapping to HTTP status codes.

### Error handling

Domain errors in `clean-arch-domain/common.go` (`ErrItemNotFound`, `ErrAlreadyExists`, etc.) are mapped in delivery handlers to HTTP responses via `factory.ErrorTemplate` (`ErrorNotFound`, `ErrorParameter`, `ErrorDB`, etc.).

### Database access

- Two databases: `EnergySharesDB` (main) and `MyNumberDB` (separate)
- DB instances are stored in `context.Context` and retrieved via `factory.DBFactory`
- Custom ORM wrapper at `pkg/orm/orm.go` wraps XORM with chainable query builder
- Transactions via `txManager.WithTx(ctx, func(txCtx) error { ... })`

### Testing

- Table-driven tests with `testify` (assert/mock)
- Mocks generated by `mockery` into `clean-arch-domain/mocks/` and `apps/{app}/domain/mocks/`
- Mock config in `.mockery.yaml` at repo root
- Tests mock at interface boundaries (repositories, services, factories)
- **A merge conflict in a generated mock is not resolvable by picking a side.**
  `mockery` emits methods in an unstable order, so a branch that only regenerates
  the file still produces a large diff (equal insertions and deletions) and
  conflicts in the middle of an unrelated method body. Comparing method *names*
  across the two sides is not enough — a signature change hides inside an
  identical name. Resolve by regenerating (`make mock`) or by taking the side
  whose signatures match the merged interface, then confirm with `go vet` in
  every module that uses the mock. This repo is multi-module: `go vet ./...`
  from the root skips `apps/*`, so run it per module (`cd apps/{name} && go vet
  ./...`) or with an explicit pattern (`go vet ./apps/{name}/...`). A mismatch
  surfaces as `does not implement ... (wrong type for method X)` only at
  test-build time — a package build of `clean-arch-domain` alone will not catch it.

### Ongoing migration

The project is migrating from legacy structure to Clean Architecture. New APIs use `v1` prefix. Existing APIs are refactored domain-by-domain, preserving request/response contracts.

### OpenAPI / Swagger

Routes are defined using `pkg/echoutils/swagger.go` which wraps Echo routing with `swaggest/openapi-go` for simultaneous route + doc generation. Swagger UI is auto-served in dev/qa environments. Path parameters require both `param` (for Echo) and `path` (for OpenAPI) tags.

## Review Checklist

Failure modes this repository has actually hit. Review changes by asking how they
break, not by confirming they work.

### Database

- **Soft-deleted rows.** Unique constraints ignore `deleted`. A deleted row still
  occupies the key, so excluding it from a duplicate check makes the insert fail.
- **Raw SQL subqueries bypass XORM's soft-delete filter.** When an expression is
  written directly into `Table().Select()`, add `deleted IS NULL` yourself.
- **Unqualified column names in JOINs.** Confirm which table a bare column resolves
  to, and what happens when the base table lacks it.
- **Subquery aliases can collide with real columns.** When a subquery is selected as
  `AS some_column` next to `table.*`, and a real column of that name still exists,
  XORM silently takes the real column and discards the subquery result. This bites
  during a migration's intermediate state: the column has not been dropped yet, so
  tests against the final schema pass while production reads the wrong value. List
  the real columns explicitly instead of using `.*`.
- **Calling `Get`/`Find` without `Select` after dropping a column doesn't always fail
  loudly.** XORM builds the SELECT list from the struct's fields when `Select` is
  omitted. Verified against this repo's XORM version (MIR-1163, after
  `investment_waiters.share_id` was dropped): with no `Join`, the missing column
  raises `Unknown column` — an exception you'd notice. But with a `Join` present
  (e.g. `Join("INNER", "user", ...)`) and no column of that name on the joined
  table either, the field is just silently left as its zero value; no error, no
  log. If the joined table happens to have a column with the same name, that
  value likely gets silently pulled in instead (not verified). Always pass an
  explicit `Select` (a column list or `table.*`) for a struct with a field
  mapped to a column that no longer exists, and don't assume the absence of an
  error means the query is safe.
- **`pkg/orm`'s joins fail silently.** `LeftJoin` and `InnerJoin` both route through
  `join()`, which requires the target to implement `TableName()` and `SelectAll()`.
  If either is missing, `toOrmObject` returns false and `join` does nothing — no error,
  no log, and the JOIN clause simply disappears from the generated SQL. Verify the
  emitted SQL, not just that the call compiles.
- **The same calculation written in two places.** List and CSV, screen and batch —
  the expressions drift apart. Extract it instead of copying.
- **`db.Insert` needs a pointer when the caller relies on the returned struct's
  auto-increment `Id`.** `db.Insert(x)` passes a copy; XORM writes the assigned Id
  back into that copy, not into the caller's variable, so a `Create` that later
  `return`s `x` always returns `Id: 0`. This stayed harmless as long as nothing read
  that Id back — MIR-1204 broke that assumption: a later `Update(ctx, execute)`
  keyed on `execute.Id` matched `WHERE id = 0`, updated zero rows, and returned no
  error. The DB row silently stayed on its old status while the API response (built
  from the in-memory struct) reported success. Use `db.Insert(&x)` whenever the
  inserted Id is read afterward.
- **A `time.Time` passed into a raw `Where` condition skips XORM's timezone
  conversion.** `DatabaseTZ` conversion only applies when XORM recognizes the
  struct field being written (`Insert`/`Update`). A condition string like
  `Where("created < ?", t)` is passed straight to the driver in the process's
  local timezone, unconverted. On MySQL this is harmless: `go-sql-driver/mysql`
  re-converts every `time.Time` right before binding it, via `v.In(cfg.Loc)`
  where `cfg.Loc` comes from the DSN's `loc=` parameter — so the value that
  reaches the server is in `loc`'s zone regardless of what zone the caller's
  `time.Time` was in, and it lines up with the column XORM wrote (also
  `loc`'s zone, since `DatabaseTZ` on MySQL is `time.Local`). SQLite has no
  such re-conversion: `mattn/go-sqlite3` binds a `time.Time` verbatim in
  whatever zone it's in, so a condition built from local time (JST on most dev
  machines) is compared against a column XORM wrote in UTC (`DatabaseTZ` is
  forced to `time.UTC` for SQLite) — MIR-1114's `RePublish` age filter matched
  the wrong rows in that sqlite-backed test until `.UTC()` was added on the
  condition side. Calling `.UTC()` is a no-op against MySQL but keeps sqlite
  tests honest, so add it whenever a raw condition compares against a XORM
  timestamp column — the safety margin costs nothing and keeps test and
  production behavior from diverging if a driver or DSN ever changes. A test
  for such a condition needs its fixture rows created far enough in the past
  (or future) by default — a helper that inserts "just now" will fail once any
  age filter is added, and every existing caller breaks at once.
- **`Update(bean)` silently skips zero-value columns.** XORM's default
  `Update` only includes non-zero fields of the struct in the generated
  UPDATE statement. Assigning `0` / `""` / `nil` to a field and calling
  `Update()` does not write that value to the DB — the column is simply left
  out of the SQL, and the old value stays. Verified with a minimal
  reproduction against this repo's actual XORM version (sqlite, xorm v1.3.9):
  a struct with one non-zero field and one field explicitly set to its zero
  value produced an UPDATE that included only the non-zero field; the
  zero-value field kept its previous DB value. This bit MIR-1223, where
  clearing `ApplicantId`/`Result` back to zero on an unverified webhook
  result silently did nothing on the first attempt. To force a field to
  zero/nil, pass the column name explicitly via `Cols("column_name")` —
  verified this also writes a nil pointer field as SQL NULL, so a separate
  `Nullable(...)` call is not needed once `Cols` is used. `Nullable(...)`
  matters only when neither `Cols` nor `AllCols` is specified.
- **`Cols(...)` restricts the UPDATE to only the named columns — it doesn't just
  force zero values through.** If the struct also has other non-zero fields
  you meant to update in the same call, `Cols("a")` silently drops them; only
  `"a"` reaches the SQL. This bit MIR-1223 a second time, on the same call the
  item above fixed: the fix cleared one zero-value column via `Cols`, but the
  same statement also carried other non-zero fields (`ApplicantId`, a nested
  risk-check result) that got dropped as a result. Use `MustCols("column_name")`
  instead when you want one column forced through as zero/nil while every other
  field keeps XORM's normal auto-detection (non-zero fields still included,
  zero-value ones still skipped). Verified against this repo's XORM version:
  `MustCols` only adds the named column(s) to a separate "always include" set
  checked inside XORM's per-field zero-value test — it does not touch the
  `Cols`-based column filter that `ifAddColUpdate` applies first, so combining
  `MustCols("result")` with no `Cols` call leaves every other field's normal
  update behavior untouched.

### Logging

- **No secrets in logs.** Names, bank account numbers, credentials, tokens.
- **`Info` and `Error` are emitted regardless of the `debug` setting.** `debug=false`
  only disables `BodyDump` and the log level.
- **Do not pass whole structs to `logrus.Fields`.** Every field gets expanded.
- **Do not pass `*http.Response` directly.** `Request.GetBody` is a function value,
  so JSON serialization fails and **the entire log line is dropped**.
- **Leftover `fmt.Println` / debug prints.**

### Existing data

- **Unexpected values.** `NULL`, only one side populated, records that predate a
  migration.
- **Missing configuration.** Does it fail, silently disable itself, or panic?

### External APIs

- **Stop after a validation failure.** Check where `continue` / `return` sits — a
  failed pre-check must not fall through to the real request.
- **Never index local slices with values returned by the API.** Bounds-check first.
- **No unconditional type assertions on errors.** Wrapped errors panic.

### General

- **Would the check still look successful if its condition were inverted?**
  Monitoring and validation code can report "healthy" with reversed logic.
  Write the failure condition, not the success condition.
- **Multi-module layout.** `go build ./...` at the root does not cover `apps/`.
  `cd apps/<name>` and build there. `./domain/...`, `./pkg/...`, etc. are already
  part of the root module and are covered by the bare `./...` — only `apps/*` are
  separate `go.work` modules that need an explicit pattern. An explicit pattern
  for an app module (`go test ./apps/admin-api/...`) does work from the root
  without `cd`, since `go.work` resolves it.
- **SQLite tests hide columns that do not exist in production.** `engine.Sync()`
  creates a real column for every struct field, including `xorm:"<-"` fields that
  production fills only through a JOIN and has no real column for. Calling
  `Get`/`Find` without `.Select(...)` makes XORM build the column list from the
  whole struct, so the query passes against SQLite and fails in production with
  `Unknown column`. Verify against the production schema (`db/sql/version/*.schema.sql`)
  with a raw-SQL test table, or state the columns with
  `.Select(domain.X{}.SelectAll())`.

See `docs/test-checklist.md` for test-specific perspectives.
