// Package backup saves LND static channel backups to a PostgreSQL database
package backup

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/internal/db"
	"github.com/uptrace/bun"
)

// SaveChannelBackups ...
func SaveChannelBackups(statusPollInterval time.Duration, node *db.Node, client lndclient.LightningClient, depotDB *bun.DB) {
	for {
		fmt.Println("\nSaving channel backups ...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel() // todo - defer will never run (endless loop)

		channels, err := client.ListChannels(ctx, true, false)
		if err != nil {
			log.Print(err.Error())
		}
		for _, item := range channels {
			err := db.InsertChannel(item, node.Pubkey, depotDB)
			if err != nil {
				log.Print(err.Error())
			}
		}

		// static channel backup (multi)
		chanBackups, err := client.ChannelBackups(ctx)
		if err != nil {
			log.Print(err.Error())
		}

		// mulitchannel backup
		err = db.InsertMultiChannelBackup(base64.StdEncoding.EncodeToString(chanBackups), node.Pubkey, depotDB)
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
