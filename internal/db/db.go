// Package db implements database models, insert and select functions for a postgres database using
// the "bun" ORM (github.com/uptrace/bun)
package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lightninglabs/lndclient"
)

// NodeIF is an interface
type NodeIF interface {
	Create(id int64, url string, alias string, pubkey string, macaroon string)
}

// NodeImpl is a node
type NodeImpl struct {
	ID       int64
	URL      string
	Alias    string
	Pubkey   string
	Macaroon string
}

// Create is a function to create a new node
func (n *NodeImpl) Create(id int64, url string, alias string, pubkey string, macaroon string) {
	fmt.Printf("I have a node: %s\n", n.Alias)
	panic("oops!")
}

// InsertNode adds a lightning node to the database
func (n *NodewatcherDB) InsertNode(node *Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := n.db.NewInsert().
		Model(node).
		On("conflict (\"pubkey\") do nothing").
		Exec(ctx)

	return err
}

// FindNodeByPubkey gets node from the db
func (n *NodewatcherDB) FindNodeByPubkey(pubkey string) (Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var node Node
	err := n.db.NewSelect().
		Model(&node).
		Where("pubkey = ?", pubkey).
		Scan(ctx, &node)

	return node, err
}

// InsertChannel adds a channel to the db
func (n *NodewatcherDB) InsertChannel(channel lndclient.ChannelInfo, pubkey string) error {
	//log.Printf("\npubkey: %s", pubkey)
	//log.Printf("\nchannelpoint: %s", channel.ChannelPoint)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	nodeFromDB, err := n.FindNodeByPubkey(pubkey)
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

	_, err = n.db.NewInsert().
		Model(mychan).
		On("conflict (\"funding_txid\",\"output_index\") do nothing").
		Exec(ctx)

	return err
}

// FindChannelByNodeID gets channel from the db
func (n *NodewatcherDB) FindChannelByNodeID(id int64) (Channel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var c Channel
	err := n.db.NewSelect().
		Model(&c).
		Where("node_id = ?", id).
		Scan(ctx, &c)

	return c, err
}

// InsertMultiChannelBackup adds a static channel backup of all channels to the database
func (n *NodewatcherDB) InsertMultiChannelBackup(backup string, pubkey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	nodeFromDB, err := n.FindNodeByPubkey(pubkey)
	if err != nil {
		return err
	}

	multiBackup := &MultiChannelBackup{
		ID:        0,
		Backup:    backup,
		NodeID:    nodeFromDB.ID,
		CreatedAt: time.Now(),
	}
	_, err = n.db.NewInsert().
		Model(multiBackup).
		Exec(ctx)

	return err
}

// FindMultiChannelBackupByPubkey gets the most recent multi-channel backup from the db
func (n *NodewatcherDB) FindMultiChannelBackupByPubkey(pubkey string) (MultiChannelBackup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var mc MultiChannelBackup

	nodeFromDB, err := n.FindNodeByPubkey(pubkey)
	if err != nil {
		return mc, err
	}

	err = n.db.NewSelect().
		Model(&mc).
		Where("node_id = ?", nodeFromDB.ID).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx, &mc)

	return mc, err
}
