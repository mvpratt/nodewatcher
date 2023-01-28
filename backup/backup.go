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
	for true {
		fmt.Println("\nSaving channel backups ...")
		response := GetChannels(client)
		for _, item := range response.Channels {
			err := db.InsertChannel(item, node.Pubkey, depotDB)
			if err != nil {
				log.Print(err.Error())
			}
		}

		// static channel backup
		chanBackups := GetChannelBackups(client)
		for _, item := range chanBackups.SingleChanBackups.ChanBackups {
			err := db.InsertChannelBackup(item, depotDB)
			if err != nil {
				log.Print(err.Error())
			}
		}

		// mulitchannel backup
		err := db.InsertMultiChannelBackup(chanBackups.MultiChanBackup, node.Pubkey, depotDB)
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
func GetInfo(client lnrpc.LightningClient) *lnrpc.GetInfoResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	info, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		log.Print(err)
	}
	return info
}

// GetChannels ...
func GetChannels(client lnrpc.LightningClient) *lnrpc.ListChannelsResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channels, err := client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})
	if err != nil {
		log.Print(err)
	}
	return channels
}

// GetChannelBackups ...
func GetChannelBackups(client lnrpc.LightningClient) *lnrpc.ChanBackupSnapshot {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	chanBackups, err := client.ExportAllChannelBackups(ctx, &lnrpc.ChanBackupExportRequest{})
	if err != nil {
		log.Print(err)
	}
	return chanBackups
}

// VerifyBackup ...
func VerifyBackup(client lnrpc.LightningClient, snapshot lnrpc.ChanBackupSnapshot) *lnrpc.VerifyChanBackupResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.VerifyChanBackup(ctx, &snapshot)
	if err != nil {
		log.Print(err)
	}
	return response
}
