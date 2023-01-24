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

// ChannelBackup is ...
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

// RunMigrations gets all *.up.sql files from /migrations and runs the SQL queries
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

// ConnectToDB connects to a Postgre database with the credentials provided
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

// InsertNode adds a lightning node to the database
func InsertNode(node *Node, db *bun.DB) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := db.NewInsert().
		Model(node).
		On("conflict (\"pubkey\") do nothing").
		Exec(dbctx)

	if err != nil {
		log.Print(err.Error())
	}
}

// FindNodeByURL gets node from the db
func FindNodeByURL(nodeURL string, db *bun.DB) (Node, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var node Node
	err := db.NewSelect().
		Model(&node).
		Where("url = ?", nodeURL).
		Scan(dbctx, &node)

	if err != nil {
		log.Print(err.Error())
	}
	return node, err
}

// FindNodeByPubkey gets node from the db
func FindNodeByPubkey(pubkey string, db *bun.DB) (Node, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var node Node
	err := db.NewSelect().
		Model(&node).
		Where("pubkey = ?", pubkey).
		Scan(dbctx, &node)

	if err != nil {
		log.Print(err.Error())
	}
	return node, err
}

// InsertChannel adds a channel to the db
func InsertChannel(channel *lnrpc.Channel, pubkey string, db *bun.DB) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
		Exec(dbctx)

	if err != nil {
		log.Print(err.Error())
	}
}

// FindChannelByNodeID gets channel from the db
func FindChannelByNodeID(id int64, db *bun.DB) (Channel, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var c Channel
	err := db.NewSelect().
		Model(&c).
		Where("node_id = ?", id).
		Scan(dbctx, &c)

	if err != nil {
		log.Print(err.Error())
	}
	return c, err
}

// InsertChannelBackup adds a static channel backup to the database
func InsertChannelBackup(backup *lnrpc.ChannelBackup, db *bun.DB) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
		Exec(dbctx)

	if err != nil {
		log.Print(err.Error())
	}
}

// InsertMultiChannelBackup adds a static channel backup of all channels to the database
func InsertMultiChannelBackup(backup *lnrpc.MultiChanBackup, pubkey string, db *bun.DB) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
		Exec(dbctx)

	if err != nil {
		log.Print(err.Error())
	}
}

// FindMultiChannelBackupByPubkey gets the most recent multi channel backup from the db
func FindMultiChannelBackupByPubkey(pubkey string, db *bun.DB) (MultiChannelBackup, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
		Scan(dbctx, &mc)

	if err != nil {
		log.Print(err.Error())
	}
	return mc, err
}
