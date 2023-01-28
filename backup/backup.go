// Package backup saves LND static channel backups to a PostgreSQL database
package backup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/mvpratt/nodewatcher/db"
	"github.com/uptrace/bun"
)

// SaveChannelBackups ...
func SaveChannelBackups(statusPollInterval time.Duration, node *db.Node, client lnrpc.LightningClient, depotDB *bun.DB) {
	for {
		fmt.Println("\nSaving channel backups ...")
		response, err := GetChannels(client)
		if err != nil {
			log.Print(err.Error())
		}
		for _, item := range response.Channels {
			err := db.InsertChannel(item, node.Pubkey, depotDB)
			if err != nil {
				log.Print(err.Error())
			}
		}

		// static channel backup
		chanBackups, err := GetChannelBackups(client)
		if err != nil {
			log.Print(err.Error())
		}

		for _, item := range chanBackups.SingleChanBackups.ChanBackups {
			err := db.InsertChannelBackup(item, depotDB)
			if err != nil {
				log.Print(err.Error())
			}
		}

		// mulitchannel backup
		err = db.InsertMultiChannelBackup(chanBackups.MultiChanBackup, node.Pubkey, depotDB)
		if err != nil {
			log.Print(err.Error())
		}

		// WIP
		// get backup from db
		// multiBackup, err := db.FindMultiChannelBackupByPubkey(node.Pubkey, depotDB)
		// if err != nil {
		// 	log.Print(err.Error())
		// }

		time.Sleep(statusPollInterval * time.Second)
	}
}

// GetInfo ...
func GetInfo(client lnrpc.LightningClient) (*lnrpc.GetInfoResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
}

// GetChannels ...
func GetChannels(client lnrpc.LightningClient) (*lnrpc.ListChannelsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})
}

// GetChannelBackups ...
func GetChannelBackups(client lnrpc.LightningClient) (*lnrpc.ChanBackupSnapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return client.ExportAllChannelBackups(ctx, &lnrpc.ChanBackupExportRequest{})

}

// VerifyBackup ...
func VerifyBackup(client lnrpc.LightningClient, snapshot lnrpc.ChanBackupSnapshot) (*lnrpc.VerifyChanBackupResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return client.VerifyChanBackup(ctx, &snapshot)
}
