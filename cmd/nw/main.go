package main

import (
	"context"
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

	// todo - constraint - require user fields

	for {
		nodes, _ := db.FindAllNodes(context.Background())

		for _, node := range nodes {
			client, ok := lndClients[node.Alias]

			if !ok {
				client = util.GetLndClient(node) // handle error
				lndClients[node.Alias] = client
			}

			health.Check(twilioConfig, node, client) // todo - handle error
			backup.Save(node, client)                // todo - handle error
		}
		time.Sleep(60 * time.Second)
	}
}
