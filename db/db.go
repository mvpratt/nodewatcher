// Package db implements database models, insert and select functions for a postgres database using
// the "bun" ORM (github.com/uptrace/bun)
package db

import (
	"context"
	"database/sql"
	"encoding/base64"
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

// ConnectionParams include database credentials and network details
type ConnectionParams struct {
	Host         string
	Port         string
	User         string
	Password     string
	DatabaseName string
}

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

// ChannelBackup is an encrypted static channel backup of a single lightning channel
type ChannelBackup struct {
	bun.BaseModel `bun:"table:channel_backups"`

	ID               int64     `bun:"id,pk,autoincrement"`
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	FundingTxidBytes string    `bun:"funding_txid_bytes"`
	OutputIndex      int64     `bun:"output_index"`
	Backup           string    `bun:"backup"`
}

// MultiChannelBackup is an encrypted backup of a lightning channel state
type MultiChannelBackup struct {
	bun.BaseModel `bun:"table:multi_channel_backups"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	Backup    string    `bun:"backup"`
	NodeID    int64     `bun:node_id`
}

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
		fmt.Printf("there are no new migrations to run (database is up to date)\n")
	}
	fmt.Printf("migrated to %s\n", group)
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

// InsertNode adds a lightning node to the database
func InsertNode(node *Node, db *bun.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := db.NewInsert().
		Model(node).
		On("conflict (\"pubkey\") do nothing").
		Exec(ctx)

	return err
}

// FindNodeByPubkey gets node from the db
func FindNodeByPubkey(pubkey string, db *bun.DB) (Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var node Node
	err := db.NewSelect().
		Model(&node).
		Where("pubkey = ?", pubkey).
		Scan(ctx, &node)

	return node, err
}

// InsertChannel adds a channel to the db
func InsertChannel(channel *lnrpc.Channel, pubkey string, db *bun.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	nodeFromDB, err := FindNodeByPubkey(pubkey, db)
	if err != nil {
		log.Fatal(err)
	}

	splits := strings.Split(channel.ChannelPoint, ":")
	txid := splits[0]
	output, err := strconv.ParseInt(splits[1], 10, 32)

	mychan := &Channel{
		ID:          0,
		FundingTxid: txid,
		OutputIndex: output,
		NodeID:      nodeFromDB.ID,
	}

	_, err = db.NewInsert().
		Model(mychan).
		On("conflict (\"funding_txid\",\"output_index\") do nothing").
		Exec(ctx)

	return err
}

// FindChannelByNodeID gets channel from the db
func FindChannelByNodeID(id int64, db *bun.DB) (Channel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var c Channel
	err := db.NewSelect().
		Model(&c).
		Where("node_id = ?", id).
		Scan(ctx, &c)

	return c, err
}

// InsertChannelBackup adds a static channel backup to the database
func InsertChannelBackup(backup *lnrpc.ChannelBackup, db *bun.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	fundingTxidBytes := backup.ChanPoint.FundingTxid.(*lnrpc.ChannelPoint_FundingTxidBytes).FundingTxidBytes

	channelBackup := &ChannelBackup{
		ID:               0,
		FundingTxidBytes: base64.StdEncoding.EncodeToString(fundingTxidBytes),
		OutputIndex:      int64(backup.ChanPoint.OutputIndex),
		Backup:           base64.StdEncoding.EncodeToString(backup.ChanBackup),
		CreatedAt:        time.Now(),
	}

	_, err := db.NewInsert().
		Model(channelBackup).
		Exec(ctx)

	return err
}

// InsertMultiChannelBackup adds a static channel backup of all channels to the database
func InsertMultiChannelBackup(backup *lnrpc.MultiChanBackup, pubkey string, db *bun.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	nodeFromDB, err := FindNodeByPubkey(pubkey, db)
	if err != nil {
		log.Fatal(err)
	}

	multiBackup := &MultiChannelBackup{
		ID:        0,
		Backup:    base64.StdEncoding.EncodeToString(backup.MultiChanBackup),
		NodeID:    nodeFromDB.ID,
		CreatedAt: time.Now(),
	}
	_, err = db.NewInsert().
		Model(multiBackup).
		Exec(ctx)

	return err
}

// FindMultiChannelBackupByPubkey gets the most recent multi channel backup from the db
func FindMultiChannelBackupByPubkey(pubkey string, db *bun.DB) (MultiChannelBackup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	nodeFromDB, err := FindNodeByPubkey(pubkey, db)
	if err != nil {
		log.Fatal(err)
	}

	var mc MultiChannelBackup
	err = db.NewSelect().
		Model(&mc).
		Where("node_id = ?", nodeFromDB.ID).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx, &mc)

	return mc, err
}
