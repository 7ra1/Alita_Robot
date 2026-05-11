# SQLite Compatibility Plan for Alita Robot

## Goal
Add SQLite as an alternative database backend alongside PostgreSQL for lightweight deployments and local development.

## Approach: Hybrid (AutoMigrate + Optional SQL Migrations)

Use GORM's multi-dialect support:
- **Schema creation**: `db.AutoMigrate()` for both databases
- **Complex operations**: Go-based migrations via `Migration` interface
- **Index verification**: GORM Migrator (database-agnostic)
- **PostgreSQL SQL migrations**: Keep as reference, not mandatory

---

## Implementation Steps

### Step 1: Add SQLite Driver Dependency
```bash
go get gorm.io/driver/sqlite/v2
```

**File**: `go.mod`

---

### Step 2: Extend Configuration
**File**: `alita/config/config.go`

Add fields to `Config` struct:
```go
DBType   string `env:"DB_TYPE" envDefault:"postgres"` // "postgres" or "sqlite"
DBPath   string `env:"DB_PATH" envDefault:"./alita.db"` // SQLite file path
```

Update `ValidateConfig()`:
- `DB_TYPE=postgres` requires `DATABASE_URL`
- `DB_TYPE=sqlite` requires `DB_PATH`

**File**: `sample.env` - Document new variables:
```env
# Database Type (postgres or sqlite)
# DB_TYPE=postgres

# SQLite Database Path (when DB_TYPE=sqlite)
# DB_PATH=./alita.db

# PostgreSQL Connection (when DB_TYPE=postgres)
DATABASE_URL=postgres://postgres:password@localhost:5432/alita_robot?sslmode=disable
```

---

### Step 3: Refactor Database Initialization
**File**: `alita/db/db.go`

Replace hardcoded PostgreSQL with factory pattern:

```go
func init() {
    // ... existing CLI/env checks ...

    var err error
    switch config.AppConfig.DBType {
    case "sqlite":
        DB, err = openSQLite()
    case "postgres":
        DB, err = openPostgreSQL()
    default:
        log.Fatalf("Unsupported DB_TYPE: %s", config.AppConfig.DBType)
    }
    // ... retry logic ...
}

func openSQLite() (*gorm.DB, error) {
    dbPath := config.AppConfig.DBPath
    dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on", dbPath)
    return gorm.Open(sqlite.Open(dsn), &gorm.Config{
        Logger:      gormLogger,
        PrepareStmt: true,
        NowFunc: func() time.Time { return time.Now().UTC() },
    })
}

func openPostgreSQL() (*gorm.DB, error) {
    dsn := config.AppConfig.DatabaseURL
    return gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger:      gormLogger,
        PrepareStmt: true,
        NowFunc: func() time.Time { return time.Now().UTC() },
    })
}
```

**Skip connection pool settings for SQLite** (lines 692-695):
```go
if config.AppConfig.DBType == "postgres" {
    sqlDB, _ := DB.DB()
    sqlDB.SetMaxIdleConns(config.AppConfig.DBMaxIdleConns)
    // ... other pool settings ...
}
```

---

### Step 4: Update Model Tags for JSON/JSONB Compatibility
**File**: `alita/db/db.go`

Remove explicit `type:jsonb` from GORM tags - let each driver use its default:

Before:
```go
Users Int64Array `gorm:"column:users;type:jsonb"`
```

After:
```go
Users Int64Array `gorm:"column:users"` // SQLite→JSON, PostgreSQL→JSONB
```

Affected fields (7 total):
- `Chat.Users` (line 175)
- `Warns.Reasons` (line 214)
- `WelcomeSettings.Button` (line 231)
- `GoodbyeSettings.Button` (line 242)
- `ChatFilters.Buttons` (line 270)
- `ReportChatSettings.BlockedList` (line 360)
- `Notes.Buttons` (line 539)

**Custom types** (`ButtonArray`, `StringArray`, `Int64Array`) already implement `Scanner`/`Valuer` - they work with both databases.

---

### Step 5: Refactor Migration System
**File**: `alita/db/migrations.go`

Replace PostgreSQL-specific catalog queries with GORM Migrator:

```go
func (m *MigrationRunner) verifyIndexes() error {
    migrator := m.db.Migrator()
    
    // Use Migrator.GetIndexes() instead of pg_index queries
    for _, model := range getModels() {
        tableName := model.TableName()
        indexes := migrator.GetIndexes(tableName)
        // ... verify expected indexes exist ...
    }
    return nil
}
```

Remove or make conditional:
- `pg_class`, `pg_index` queries (lines 630-649) → Skip for SQLite
- `DO $$ BEGIN ... END $$` blocks (lines 448-471) → Use `IF NOT EXISTS`

---

### Step 6: Add Database-Agnostic Migration Runner
**New File**: `alita/db/schema.go`

```go
func RunMigrations(db *gorm.DB) error {
    // Step 1: AutoMigrate creates base schema
    if err := db.AutoMigrate(getAllModels()...); err != nil {
        return fmt.Errorf("automigrate: %w", err)
    }
    
    // Step 2: Run Go-based migrations for complex operations
    if err := runGoMigrations(db); err != nil {
        return fmt.Errorf("go migrations: %w", err)
    }
    
    return nil
}

// AutoMigrate is used for SQLite; PostgreSQL can use SQL migrations instead
func runGoMigrations(db *gorm.DB) error {
    // Add complex data migrations here (idempotent)
    return nil
}
```

Update `db.go` init() (line 705-716):
```go
if config.AppConfig.AutoMigrate {
    if config.AppConfig.DBType == "sqlite" {
        if err := RunMigrations(DB); err != nil {
            // handle error
        }
    } else {
        runner := NewMigrationRunner(DB)
        if err := runner.RunMigrations(); err != nil {
            // handle error
        }
    }
}
```

---

### Step 7: Handle Raw SQL Query Compatibility

**File**: `alita/db/greetings_db.go` (lines 409-419)
- `CASE WHEN` syntax works in both databases - no change needed

**File**: `alita/db/captcha_db.go` (line 322)
- `COALESCE()` works in both databases - no change needed

**File**: `alita/db/captcha_db.go` (lines 248-251)
- `SELECT FOR UPDATE` works in SQLite with WAL mode - no change needed

---

### Step 8: Update CI/CD for SQLite Testing
**File**: `.github/workflows/ci.yml`

Add SQLite to test matrix (optional, can be added later):
```yaml
services:
  # Keep PostgreSQL and Redis for existing tests
  # Add SQLite in-memory for additional coverage
```

Test command for SQLite:
```bash
DB_TYPE=sqlite DB_PATH=/tmp/test.db go test -v -count=1 -timeout 10m ./alita/db
```

---

## Files Modified Summary

| File | Changes |
|------|---------|
| `go.mod` | Add `gorm.io/driver/sqlite/v2` |
| `alita/config/config.go` | Add `DBType`, `DBPath` fields + validation |
| `alita/db/db.go` | Factory pattern, skip pool for SQLite, update migration call |
| `alita/db/db.go` (models) | Remove `type:jsonb` tags (7 fields) |
| `alita/db/migrations.go` | Use GORM Migrator for index verification |
| `alita/db/schema.go` | New file: AutoMigrate + Go-based migrations |
| `sample.env` | Document `DB_TYPE`, `DB_PATH` variables |

---

## Testing Strategy

1. **PostgreSQL** (existing):
   ```bash
   make test
   ```

2. **SQLite** (new):
   ```bash
   DB_TYPE=sqlite DB_PATH=/tmp/alita_test.db go test -v -count=1 -timeout 10m ./alita/db
   ```

3. **Verify JSON fields work** in both databases:
   - Test CRUD operations on models with JSON/JSONB fields
   - Verify custom `Scanner`/`Valuer` types work

4. **Verify migrations**:
   - SQLite: Schema created via AutoMigrate
   - PostgreSQL: Existing SQL migrations still work

---

## Risks and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| JSONB queries break on SQLite | High | Test all JSON queries; use GORM methods |
| CHECK constraints behave differently | Medium | Test constraints with SQLite |
| Concurrent write limits in SQLite | Medium | Document limitation; WAL mode helps |
| Migration drift between databases | High | Use AutoMigrate as single source of truth |
| Connection pool crashes SQLite | Low | Already handled (skip for SQLite) |

---

## Success Criteria

- [ ] Can start bot with `DB_TYPE=sqlite DB_PATH=./alita.db`
- [ ] All models migrate correctly in SQLite
- [ ] JSON/JSONB fields work in both databases
- [ ] Existing PostgreSQL functionality unaffected
- [ ] Tests pass with both database types
- [ ] Documentation updated (`sample.env`, AGENTS.md)
