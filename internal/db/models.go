package db

import (
	"time"

	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
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
	Macaroon string `bun:"macaroon"`
	UserID   int64  `bun:"user_id"`
}

// User is a
type User struct {
	bun.BaseModel `bun:"table:users"`

	ID          int64  `bun:"id,pk,autoincrement"`
	Email       string `bun:"email,unique"`
	Password    string `bun:"password"`
	PhoneNumber string `bun:"phone_number"`
	SmsEnabled  bool   `bun:"sms_enabled"`
}

func (user *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}

func (user *User) CheckPassword(providedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(providedPassword))
	if err != nil {
		return err
	}
	return nil
}

// Channel is a Lightning Channel
type Channel struct {
	bun.BaseModel `bun:"table:channels"`

	ID          int64  `bun:"id,pk,autoincrement"`
	FundingTxid string `bun:"funding_txid"`
	OutputIndex int64  `bun:"output_index"`
	NodeID      int64  `bun:"node_id"`
}

// todo - remove channel backup - unused
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
