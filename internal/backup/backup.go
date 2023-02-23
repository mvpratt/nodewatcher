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
	"github.com/mvpratt/nodewatcher/internal/util"
)

func getChannels(node db.Node, client lndclient.LightningClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channels, err := client.ListChannels(ctx, true, false)
	if err != nil {
		return err
	}
	for _, item := range channels {
		err := db.InsertChannel(item, node.Pubkey)
		if err != nil {
			return err
		}
	}
	return nil
}

func getMultiChannelBackups(node db.Node, client lndclient.LightningClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	chanBackups, err := client.ChannelBackups(ctx)
	if err != nil {
		return err
	}
	err = db.InsertMultiChannelBackup(base64.StdEncoding.EncodeToString(chanBackups), node.Pubkey)
	if err != nil {
		return err
	}
	return nil
}

// SaveChannelBackups ...
func SaveChannelBackups(statusPollInterval time.Duration, node db.Node) {
	lndClient := util.GetLndClient(node)
	for {
		fmt.Println("\nSaving channel backups ...")

		err := getChannels(node, lndClient)
		if err != nil {
			log.Print(err.Error())
		}

		err = getMultiChannelBackups(node, lndClient)
		if err != nil {
			log.Print(err.Error())
		}

		time.Sleep(statusPollInterval * time.Second)
	}
}

// WIP
// get backup from db
// multiBackup, err := db.FindMultiChannelBackupByPubkey(node.Pubkey, depotDB)
// if err != nil {
// 	log.Print(err.Error())
// }
