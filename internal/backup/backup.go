// Package backup saves LND static channel backups to a PostgreSQL database
package backup

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/internal/db"
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

// Save multi-channel backup to db
func Save(node db.Node, lndClient *lndclient.LightningClient) error {
	fmt.Printf("\nSaving multi-channel backup: %s", node.Alias)

	err := getChannels(node, *lndClient)
	if err != nil {
		return err
	}

	err = getMultiChannelBackups(node, *lndClient)
	if err != nil {
		return err
	}
	return nil
}

// WIP
// get backup from db
// multiBackup, err := db.FindMultiChannelBackupByPubkey(node.Pubkey, depotDB)
// if err != nil {
// 	log.Print(err.Error())
// }
