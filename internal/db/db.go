package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mvpratt/nodewatcher/internal/db/migrations"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
)

// Instance is the global database instance
var Instance *bun.DB

// RunMigrations gets all *.sql files from /migrations and runs them to create tables and constraints
func RunMigrations() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	migrator := migrate.NewMigrator(Instance, migrations.Migrations)
	migrator.Init(ctx)

	if err := migrator.Lock(ctx); err != nil {
		return err
	}
	defer migrator.Unlock(ctx) //nolint:errcheck

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		log.Print("there are no new migrations to run (database is up to date)\n")
	}
	log.Printf("migrated to %s\n", group)
	return nil
}

// Connect connects to a Postgres database with the credentials provided
func Connect(params *ConnectionParams) {

	env := os.Getenv("NODEWATCHER_ENV")
	var sslmode string

	if env == "production" {
		sslmode = "require"
	} else {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		params.User,
		params.Password,
		params.Host,
		params.Port,
		params.DatabaseName,
		sslmode,
	)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	Instance = bun.NewDB(sqldb, pgdialect.New())
}

// EnableDebugLogs logs all database queries to the console
func EnableDebugLogs() {
	Instance.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
}
