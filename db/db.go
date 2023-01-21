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
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
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
	//NodeID      *Node  `bun:"node_id"`
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
		log.Fatalf(err.Error())
	}
}

// InsertChannels adds channels to the db
func InsertChannels(channels *lnrpc.ListChannelsResponse, depotDB *bun.DB) {
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second) // note - new context
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
			log.Fatalf(err.Error())
		}
	}
}

// InsertChannelBackups blah
func InsertChannelBackups(backups *lnrpc.ChanBackupSnapshot, depotDB *bun.DB) {
	for _, item := range backups.SingleChanBackups.ChanBackups {
		dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// var mychan Channel
		// err := depotDB.NewSelect().
		// 	Table("channels").
		// 	Where("funding_txid = ? and output_index = ?", item.ChanPoint.FundingTxid.FundingTxidBytes, item.ChanPoint.OutputIndex).
		// 	Limit(1).
		// 	Scan(dbctx, &mychan)
		// if err != nil {
		// 	log.Fatalf(err.Error())
		// }

		channelBackup := &ChannelBackup{
			ID:               0,
			FundingTxidBytes: "placeholder", //string(item.ChanPoint.FundingTxid),
			OutputIndex:      int64(item.ChanPoint.OutputIndex),
			Backup:           string(item.ChanBackup),
			CreatedAt:        time.Now(),
		}

		// itemJSON, err := json.MarshalIndent(item, " ", "    ")
		// if err != nil {
		// 	log.Fatalf(err.Error())
		// }
		// fmt.Println(string(itemJSON))

		// fmt.Println(itemJSON.chan_backup)
		//fmt.Println(string(item.ChanBackup))

		_, err := depotDB.NewInsert().
			Model(channelBackup).
			Returning("*").
			Exec(dbctx)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}
}
