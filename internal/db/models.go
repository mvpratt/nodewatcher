package db

import (
	"time"

	"github.com/uptrace/bun"
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
	NodeID      int64  `bun:"node_id"`
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
	NodeID    int64     `bun:"node_id"`
}
