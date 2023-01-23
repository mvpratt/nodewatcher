package db

import (
	"context"
	"database/sql"
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

	ID       int64  `bun:"id,pk,autoincrement"`
	URL      string `bun:"url,unique"`
	Alias    string `bun:"alias"`
	Pubkey   string `bun:"pubkey"`
	Macaroon string `bund:"macaroon"`
}

// Channel is a Lightning Channel
type Channel struct {
	bun.BaseModel `bun:"table:channels"`

	ID          int64  `bun:"id,pk,autoincrement"`
	FundingTxid string `bun:"funding_txid"`
	OutputIndex int64  `bun:"output_index"`
	NodeID      int64  `bun:node_id`
}

// ChannelBackup is a Lightning Channel
type ChannelBackup struct {
	bun.BaseModel `bun:"table:channel_backups"`

	ID               int64     `bun:"id,pk,autoincrement"`
	FundingTxidBytes string    `bun:"funding_txid_bytes"`
	OutputIndex      int64     `bun:"output_index"`
	Backup           string    `bun:"backup"`
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	ChannelID        int64     `bun:"channel_id"`
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

// ConnectToDB ...
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

// FindNodeByURL gets node from the db
func FindNodeByURL(nodeURL string, depotDB *bun.DB) (Node, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var node Node
	err := depotDB.NewSelect().
		Model(&node).
		Where("url = ?", nodeURL).
		Scan(dbctx, &node)

	if err != nil {
		log.Print(err.Error())
	}
	return node, err
}

// InsertChannel adds a channel to the db
func InsertChannel(channel *lnrpc.Channel, nodeID int64, depotDB *bun.DB) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	splits := strings.Split(channel.ChannelPoint, ":")
	txid := splits[0]
	output, err := strconv.ParseInt(splits[1], 10, 32)

	mychan := &Channel{
		ID:          0,
		FundingTxid: txid,
		OutputIndex: output,
		NodeID:      nodeID,
	}

	_, err = depotDB.NewInsert().
		Model(mychan).
		On("conflict (\"funding_txid\",\"output_index\") do nothing").
		Exec(dbctx)

	if err != nil {
		log.Print(err.Error())
	}

}

// FindChannelByNodeID gets channel from the db
func FindChannelByNodeID(nodeID int64, db *bun.DB) (Channel, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var c Channel
	err := db.NewSelect().
		Model(&c).
		Where("node_id = ?", nodeID).
		Scan(dbctx, &c)

	if err != nil {
		log.Print(err.Error())
	}
	return c, err
}

// InsertChannelBackup blah
func InsertChannelBackup(backup *lnrpc.ChannelBackup, channelID int64, depotDB *bun.DB) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelBackup := &ChannelBackup{
		ID:               0,
		FundingTxidBytes: "placeholder",
		OutputIndex:      int64(backup.ChanPoint.OutputIndex),
		Backup:           string(backup.ChanBackup[:]),
		CreatedAt:        time.Now(),
		ChannelID:        channelID,
	}

	_, err := depotDB.NewInsert().
		Model(channelBackup).
		Exec(dbctx)

	if err != nil {
		log.Print(err.Error())
	}
}

// FindChannelBackupByChannelID gets backup from the db
func FindChannelBackupByChannelID(channelID int64, db *bun.DB) (ChannelBackup, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var cb ChannelBackup

	err := db.NewSelect().
		Model(&cb).
		Where("channel_id = ?", channelID).
		Scan(dbctx, &cb)

	if err != nil {
		log.Print(err.Error())
	}
	return cb, err
}
