package main

import (
	"context"
	"encoding/hex"
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

	var (
		macaroon = util.RequireEnvVar("MACAROON_HEADER")
		lndHost  = util.RequireEnvVar("LN_NODE_URL")
		tlsPath  = util.RequireEnvVar("LND_TLS_CERT_PATH")
	)

	lndConfig := &lndclient.LndServicesConfig{
		LndAddress:            lndHost,
		Network:               lndclient.NetworkMainnet,
		CustomMacaroonHex:     macaroon,
		TLSPath:               tlsPath,
		Insecure:              false,
		BlockUntilChainSynced: false,
		BlockUntilUnlocked:    false,
	}

	//sim := util.RequireEnvVar("SIM")
	// if simulation ...
	// lndConfig.Insecure = true
	// lndConfig.TLSPath = ""
	// lndConfig.Network = lndclient.NetworkRegtest

	// connect to node via grpc
	lndServices, err := lndclient.NewLndServices(lndConfig)
	lndClient := lndServices.LndServices.Client
	if err != nil {
		log.Fatal(err.Error())
	}

	nodeInfo, err := lndClient.GetInfo(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}

	node := &db.Node{
		ID:       0,
		URL:      lndHost,
		Alias:    nodeInfo.Alias,
		Pubkey:   hex.EncodeToString(nodeInfo.IdentityPubkey[:]),
		Macaroon: macaroon,
	}

	err = db.InsertNode(node)
	if err != nil {
		log.Fatal(err.Error())
	}

	var smsParams health.SmsParams
	smsParams.Enable = util.RequireEnvVar("SMS_ENABLE") == "TRUE"

	if smsParams.Enable {
		smsParams.To = util.RequireEnvVar("TO_PHONE_NUMBER")
		smsParams.From = util.RequireEnvVar("TWILIO_PHONE_NUMBER")
		smsParams.TwilioClient = twilio.NewRestClient()
		smsParams.TwilioAccountSID = util.RequireEnvVar("TWILIO_ACCOUNT_SID")
		smsParams.TwilioAuthToken = util.RequireEnvVar("TWILIO_AUTH_TOKEN")
	} else {
		log.Println("\nWARNING: Text messages disabled. " +
			"Set environment variable SMS_ENABLE to TRUE to enable SMS status updates")
	}

	monitorParams := health.MonitorParams{
		Interval:   60 * time.Second,
		NotifyTime: 1, // when time = 01:00 UTC
	}

	done := make(chan bool)
	go health.Monitor(monitorParams, lndClient)
	go backup.SaveChannelBackups(monitorParams.Interval, node, lndClient)

	<-done // Block forever
}
