package main

import (
	"context"
	"encoding/hex"
	"log"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/internal/backup"
	"github.com/mvpratt/nodewatcher/internal/db"
	"github.com/mvpratt/nodewatcher/internal/health"
	"github.com/mvpratt/nodewatcher/internal/util"
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

	nwDB := db.NodewatcherDB{}
	nwDB.ConnectToDB(dbParams)
	nwDB.EnableDebugLogs()
	nwDB.RunMigrations()

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

	err = nwDB.InsertNode(node)
	if err != nil {
		log.Fatal(err.Error())
	}

	const pollInterval = 60 // 1 minute

	done := make(chan bool)
	go health.Monitor(pollInterval, lndClient)
	go backup.SaveChannelBackups(pollInterval, node, lndClient, nwDB)

	<-done // Block forever
}
