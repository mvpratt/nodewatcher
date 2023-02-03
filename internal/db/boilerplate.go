package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/mvpratt/nodewatcher/internal/db/migrations"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
)

// RunMigrations gets all *.sql files from /migrations and runs them to create tables and constraints
func RunMigrations(db *bun.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	migrator := migrate.NewMigrator(db, migrations.Migrations)
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

// ConnectToDB connects to a Postgres database with the credentials provided
func ConnectToDB(params *ConnectionParams) *bun.DB {

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", params.User, params.Password, params.Host, params.Port, params.DatabaseName)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	depotDB := bun.NewDB(sqldb, pgdialect.New())
	return depotDB
}

// EnableDebugLogs logs all database queries to the console
func EnableDebugLogs(db *bun.DB) {
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
}
