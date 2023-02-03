// Package db implements database models, insert and select functions for a postgres database using
// the "bun" ORM (github.com/uptrace/bun)
package db

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/lightninglabs/lndclient"
	"github.com/uptrace/bun"
)

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
func InsertChannel(channel lndclient.ChannelInfo, pubkey string, db *bun.DB) error {
	//log.Printf("\npubkey: %s", pubkey)
	//log.Printf("\nchannelpoint: %s", channel.ChannelPoint)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	nodeFromDB, err := FindNodeByPubkey(pubkey, db)
	if err != nil {
		return err
	}

	splits := strings.Split(channel.ChannelPoint, ":")
	txid := splits[0]
	output, err := strconv.ParseInt(splits[1], 10, 32)
	if err != nil {
		return err
	}

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

// InsertMultiChannelBackup adds a static channel backup of all channels to the database
func InsertMultiChannelBackup(backup string, pubkey string, db *bun.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	nodeFromDB, err := FindNodeByPubkey(pubkey, db)
	if err != nil {
		return err
	}

	multiBackup := &MultiChannelBackup{
		ID:        0,
		Backup:    backup,
		NodeID:    nodeFromDB.ID,
		CreatedAt: time.Now(),
	}
	_, err = db.NewInsert().
		Model(multiBackup).
		Exec(ctx)

	return err
}

// FindMultiChannelBackupByPubkey gets the most recent multi-channel backup from the db
func FindMultiChannelBackupByPubkey(pubkey string, db *bun.DB) (MultiChannelBackup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var mc MultiChannelBackup

	nodeFromDB, err := FindNodeByPubkey(pubkey, db)
	if err != nil {
		return mc, err
	}

	err = db.NewSelect().
		Model(&mc).
		Where("node_id = ?", nodeFromDB.ID).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx, &mc)

	return mc, err
}
