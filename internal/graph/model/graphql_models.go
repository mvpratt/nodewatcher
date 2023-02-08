package model

import "time"

// Node is a Lightning Node
type Node struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"`
	Alias    string `json:"alias"`
	Pubkey   string `json:"pubkey"`
	Macaroon string `json:"macaroon"`
}

// Channel is a Lightning Channel
type Channel struct {
	ID          int64  `json:"id"`
	FundingTxid string `json:"funding_txid"`
	OutputIndex int64  `json:"output_index"`
	NodeID      int64  `json:"node_id"`
}

// MultiChannelBackup is an encrypted backup of a lightning channel state
type MultiChannelBackup struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Backup    string    `json:"backup"`
	NodeID    int64     `json:"node_id"`
}
