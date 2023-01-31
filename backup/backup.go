// Package backup saves LND static channel backups to a PostgreSQL database
package backup

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/mvpratt/nodewatcher/db"
	"github.com/uptrace/bun"
)

// SaveChannelBackups ...
func SaveChannelBackups(statusPollInterval time.Duration, node *db.Node, client lnrpc.LightningClient, depotDB *bun.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	// todo - does ctx need macaroon auth?
	backupStream, err := client.SubscribeChannelBackups(ctx, &lnrpc.ChannelBackupSubscription{})
	if err != nil {
		log.Print(err.Error())
	}

	var wg sync.WaitGroup
	backupUpdates := make(chan *lnrpc.ChanBackupSnapshot)
	streamErr := make(chan error, 1)

	// Backups updates goroutine.
	wg.Add(1)
	go func() {
		log.Println("getting snapshots from stream ...")
		defer wg.Done()
		for {
			snapshot, err2 := backupStream.Recv() // seems like this is timing out  - maybe exeed deadline?
			if err != nil {
				log.Print("error getting snapshot")
				log.Print(err2.Error())
				streamErr <- err2
				return
			}

			select {
			case backupUpdates <- snapshot:
			case <-ctx.Done():
				log.Print("ctx.Done()")
				return
			}
		}
	}()

	go func() {
		log.Println("waiting for snaps ...")
		for {
			snap, ok := <-backupUpdates
			log.Printf("got a snap %#v, ok:%t", snap, ok)
		}
	}()

	//go routine
	//go log.Printf("\ngrpc stream:\n%v\n", backupStream)

	for {
		log.Println("\nSaving channel backups ...")
		// 	response, err := GetChannels(client)
		// 	if err != nil {
		// 		log.Print(err.Error())
		// 	}
		// 	for _, item := range response.Channels {
		// 		err := db.InsertChannel(item, node.Pubkey, depotDB)
		// 		if err != nil {
		// 			log.Print(err.Error())
		// 		}
		// 	}

		// 	// static channel backup
		// 	chanBackups, err := GetChannelBackups(client)
		// 	if err != nil {
		// 		log.Print(err.Error())
		// 	}

		// 	for _, item := range chanBackups.SingleChanBackups.ChanBackups {
		// 		err := db.InsertChannelBackup(item, depotDB)
		// 		if err != nil {
		// 			log.Print(err.Error())
		// 		}
		// 	}

		// 	// multichannel backup
		// 	err = db.InsertMultiChannelBackup(chanBackups.MultiChanBackup, node.Pubkey, depotDB)
		// 	if err != nil {
		// 		log.Print(err.Error())
		// 	}

		// 	// todo - testing
		// 	multiBackup, err := db.FindMultiChannelBackupByPubkey(node.Pubkey, depotDB)
		// 	if err != nil {
		// 		log.Print(err.Error())
		// 	}
		// 	log.Printf("\n\nmulti backup: %v", multiBackup)

		// 	// todo .. add single channel backups from the db?
		// 	// resp, err := VerifyBackup(client, multiBackup)
		// 	// log.Printf("\n\nverify backup response: %v", resp)

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
func VerifyBackup(client lnrpc.LightningClient, snapshot *lnrpc.ChanBackupSnapshot) (*lnrpc.VerifyChanBackupResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return client.VerifyChanBackup(ctx, snapshot)
}

// SubscribeChannelBackups gets new backup whenever channel state changes
// func SubscribeChannelBackups(client lnrpc.LightningClient) (lnrpc.Lightning_SubscribeChannelBackupsClient, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()

// 	//var sub lnrpc.ChannelBackupSubscription
// 	//var stream lnrpc.Lightning_SubscribeChannelBackupsClient

// 	stream, err := client.SubscribeChannelBackups(ctx, &lnrpc.ChannelBackupSubscription{})
// 	return stream, err
// }
