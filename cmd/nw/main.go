package main

import (
	"context"
	"log"
	"time"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/internal/backup"
	"github.com/mvpratt/nodewatcher/internal/db"
	"github.com/mvpratt/nodewatcher/internal/health"
	"github.com/mvpratt/nodewatcher/internal/util"
	"github.com/twilio/twilio-go"
)

// Nodewatcher runs two processes:
//  1. Checks the health of an LND node and sends an SMS once a day with the status
//  2. Saves LND static channel backups to a PostgreSQL database once per minute
func main() {

	dbParams := &db.ConnectionParams{
		Host:         util.RequireEnvVar("POSTGRES_HOST"),
		Port:         util.RequireEnvVar("POSTGRES_PORT"),
		User:         util.RequireEnvVar("POSTGRES_USER"),
		Password:     util.RequireEnvVar("POSTGRES_PASSWORD"),
		DatabaseName: util.RequireEnvVar("POSTGRES_DB"),
	}

	db.Connect(dbParams)
	db.EnableDebugLogs()
	db.RunMigrations()

	twilioConfig := health.TwilioConfig{
		From:             util.RequireEnvVar("TWILIO_PHONE_NUMBER"),
		TwilioClient:     twilio.NewRestClient(),
		TwilioAccountSID: util.RequireEnvVar("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:  util.RequireEnvVar("TWILIO_AUTH_TOKEN"),
	}

	lndClients := make(map[string]*lndclient.LightningClient)

	for {
		nodes, _ := db.FindAllNodes(context.Background())

		for _, node := range nodes {
			client, ok := lndClients[node.Alias]

			if !ok || client == nil {
				newClient, err := util.GetLndClient(node)
				if err != nil {
					log.Printf("Error connecting to LND node %s: %s", node.Alias, err)
					continue
				}
				lndClients[node.Alias] = newClient
			}

			err := health.Check(twilioConfig, node, lndClients[node.Alias])
			if err != nil {
				log.Printf("Error checking health of LND node %s: %s", node.Alias, err)
			}

			err = backup.Save(node, lndClients[node.Alias])
			if err != nil {
				log.Printf("Error saving multi-channel backup for LND node %s: %s", node.Alias, err)
			}
		}
		time.Sleep(60 * time.Second)
	}
}
