package migrations

import (
	"embed"
	"fmt"

	"github.com/uptrace/bun/migrate"
)

//go:embed *.sql
var sqlMigrations embed.FS

// Migrations are the db migrations
var Migrations = migrate.NewMigrations()

func init() {
	fmt.Println("Running database migrations...")
	if err := Migrations.Discover(sqlMigrations); err != nil {
		panic(err)
	}
}
