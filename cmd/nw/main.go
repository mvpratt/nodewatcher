package main

import (
	"context"
	"time"

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

	monitorParams := health.MonitorParams{
		Interval:   60 * time.Second,
		NotifyTime: 1, // when time = 01:00 UTC
	}

	nodes, _ := db.FindAllNodes(context.Background())

	done := make(chan bool)
	for i := range nodes {
		go health.Monitor(monitorParams, twilioConfig, nodes[i])
	}

	for i := range nodes {
		go backup.SaveChannelBackups(monitorParams.Interval, nodes[i])
	}
	<-done // Block forever
}
