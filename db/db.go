package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/mvpratt/nodewatcher/db/migrations"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
)

// Node is a Lightning Node
type Node struct {
	bun.BaseModel `bun:"table:nodes"`

	ID       int32  `bun:"id,pk,autoincrement"`
	URL      string `bun:"url,unique"`
	Alias    string `bun:"alias"`
	Pubkey   string `bun:"pubkey"`
	Macaroon string `bund:"macaroon"`
}

// Channel is a Lightning Channel
type Channel struct {
	bun.BaseModel `bun:"table:channels"`

	ID          int32  `bun:"id,pk,autoincrement"`
	FundingTxid string `bun:"funding_txid"`
	OutputIndex int64  `bun:"output_index"`
	// NodeID      *Node  `bun:node_id`
}

// ChannelBackup is a Lightning Channel
type ChannelBackup struct {
	bun.BaseModel `bun:"table:channel_backups"`

	ID               int32     `bun:"id,pk,autoincrement"`
	FundingTxidBytes string    `bun:"funding_txid_bytes"`
	OutputIndex      int64     `bun:"output_index"`
	Backup           string    `bun:"backup"`
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

// RunMigrations ...
func RunMigrations(db *bun.DB) error {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	migrator := migrate.NewMigrator(db, migrations.Migrations)
	migrator.Init(dbctx)

	if err := migrator.Lock(dbctx); err != nil {
		return err
	}
	defer migrator.Unlock(dbctx) //nolint:errcheck

	group, err := migrator.Migrate(dbctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		fmt.Printf("there are no new migrations to run (database is up to date)\n")
		return nil
	}
	fmt.Printf("migrated to %s\n", group)
	return nil
}

// ConnectToDB blah
func ConnectToDB(host string, port string, user string, password string, dbname string) *bun.DB {

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	depotDB := bun.NewDB(sqldb, pgdialect.New())

	// debug: log all queries
	depotDB.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
	return depotDB
}

// InsertNode adds node to the db
func InsertNode(node *Node, depotDB *bun.DB) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := depotDB.NewInsert().
		Model(node).
		On("conflict (\"url\") do nothing").
		Exec(dbctx)

	if err != nil {
		log.Print(err.Error())
	}
}

// FindNode gets node from the db
func FindNode(node *Node, depotDB *bun.DB) error {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return depotDB.NewSelect().
		Model(node).
		Where("id = ?", 1).
		Scan(dbctx)
}

// InsertChannels adds channels to the db
func InsertChannels(channels *lnrpc.ListChannelsResponse, depotDB *bun.DB) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, channel := range channels.Channels {
		splits := strings.Split(channel.ChannelPoint, ":")
		txid := splits[0]
		output, err := strconv.ParseInt(splits[1], 10, 32)

		mychan := &Channel{
			ID:          0,
			FundingTxid: txid,
			OutputIndex: output,
		}

		_, err = depotDB.NewInsert().
			Model(mychan).
			On("conflict (\"funding_txid\",\"output_index\") do nothing").
			Exec(dbctx)

		if err != nil {
			log.Print(err.Error())
		}
	}
}

// FindChannel gets channel from the db
func FindChannel(channel *Channel, db *bun.DB) error {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return db.NewSelect().
		Model(channel).
		Where("id = ?", 1).
		Scan(dbctx)
}

// InsertChannelBackups blah
func InsertChannelBackups(backups *lnrpc.ChanBackupSnapshot, depotDB *bun.DB) {
	for _, item := range backups.SingleChanBackups.ChanBackups {
		dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		channelBackup := &ChannelBackup{
			ID:               0,
			FundingTxidBytes: "placeholder",
			OutputIndex:      int64(item.ChanPoint.OutputIndex),
			Backup:           string(item.ChanBackup[:]),
			CreatedAt:        time.Now(),
		}

		itemJSON, err := json.MarshalIndent(item, " ", "    ")
		if err != nil {
			log.Print(err.Error())
		}
		fmt.Println(string(itemJSON))

		fmt.Println(item.ChanBackup)

		_, err = depotDB.NewInsert().
			Model(channelBackup).
			Returning("*").
			Exec(dbctx)
		if err != nil {
			log.Print(err.Error())
		}
	}
}

// FindChannelBackup gets backup from the db
func FindChannelBackup(backup *ChannelBackup, db *bun.DB) error {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return db.NewSelect().
		Model(backup).
		Where("id = ?", 1).
		Scan(dbctx)
}
