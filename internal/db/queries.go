// Package db implements database models, insert and select functions for a postgres database using
// the "bun" ORM (github.com/uptrace/bun)
package db

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/lightninglabs/lndclient"
)

// InsertNode adds a lightning node to the database
func InsertNode(node *Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	_, err := Instance.NewInsert().
		Model(node).
		On("conflict (\"pubkey\") do nothing").
		Exec(ctx)

	return err
}

// FindNodeByPubkey gets node from the db
func FindNodeByPubkey(pubkey string) (Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	var node Node
	err := Instance.NewSelect().
		Model(&node).
		Where("pubkey = ?", pubkey).
		Scan(ctx, &node)

	return node, err
}

// FindAllNodes gets node from the db
func FindAllNodes(ctx context.Context) ([]Node, error) {
	var nodes []Node
	err := Instance.NewSelect().
		Model(&nodes).
		Scan(ctx, &nodes)

	return nodes, err
}

// InsertChannel adds a channel to the db
func InsertChannel(channel lndclient.ChannelInfo, pubkey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	nodeFromDB, err := FindNodeByPubkey(pubkey)
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

	_, err = Instance.NewInsert().
		Model(mychan).
		On("conflict (\"funding_txid\",\"output_index\") do nothing").
		Exec(ctx)

	return err
}

// FindChannelByNodeID gets channel from the db
func FindChannelByNodeID(id int64) (Channel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	var c Channel
	err := Instance.NewSelect().
		Model(&c).
		Where("node_id = ?", id).
		Scan(ctx, &c)

	return c, err
}

// FindAllChannels gets channel from the db
func FindAllChannels(ctx context.Context) ([]Channel, error) {
	var channels []Channel
	err := Instance.NewSelect().
		Model(&channels).
		Scan(ctx, &channels)

	return channels, err
}

// FindAllMultiChannelBackups gets channel from the db
func FindAllMultiChannelBackups(ctx context.Context) ([]MultiChannelBackup, error) {
	var channels []MultiChannelBackup
	err := Instance.NewSelect().
		Model(&channels).
		Scan(ctx, &channels)

	return channels, err
}

// InsertMultiChannelBackup adds a static channel backup of all channels to the database
func InsertMultiChannelBackup(backup string, pubkey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	nodeFromDB, err := FindNodeByPubkey(pubkey)
	if err != nil {
		return err
	}

	multiBackup := &MultiChannelBackup{
		ID:        0,
		Backup:    backup,
		NodeID:    nodeFromDB.ID,
		CreatedAt: time.Now(),
	}
	_, err = Instance.NewInsert().
		Model(multiBackup).
		Exec(ctx)

	return err
}

// FindMultiChannelBackupByPubkey gets the most recent multi-channel backup from the db
func FindMultiChannelBackupByPubkey(pubkey string) (MultiChannelBackup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	var mc MultiChannelBackup

	nodeFromDB, err := FindNodeByPubkey(pubkey)
	if err != nil {
		return mc, err
	}

	err = Instance.NewSelect().
		Model(&mc).
		Where("node_id = ?", nodeFromDB.ID).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx, &mc)

	return mc, err
}

// InsertUser adds a lightning node to the database
func InsertUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	_, err := Instance.NewInsert().
		Model(user).
		On("conflict (\"email\") do nothing").
		Exec(ctx)

	return err
}

// FindUserByEmail gets user from the db
func FindUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	var user User
	err := Instance.NewSelect().
		Model(&user).
		Where("email = ?", email).
		Scan(ctx, &user)

	return user, err
}

// FindUserByID gets user from the db
func FindUserByID(id int64) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	var user User
	err := Instance.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx, &user)

	return user, err
}

// FindAllUsers gets users from the db
func FindAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := Instance.NewSelect().
		Model(&users).
		Scan(ctx, &users)

	return users, err
}

// UpdateUserLastSent updates user in the db
func UpdateUserLastSent(user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // todo
	defer cancel()

	_, err := Instance.NewInsert().
		Model(&user).
		On("CONFLICT (id) DO UPDATE").
		Set("sms_last_sent = EXCLUDED.sms_last_sent").
		Exec(ctx)

	return err
}
