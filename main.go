package main

import (
	"log"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/backup"
	"github.com/mvpratt/nodewatcher/db"
	"github.com/mvpratt/nodewatcher/health"
	"github.com/mvpratt/nodewatcher/util"
)

// Nodewatcher runs two processes:
//  1. Checks the health of an LND node and sends an SMS once a day with the status
//  2. Saves LND static channel backups to a PostgreSQL database once per minute
func main() {
	macaroon := util.RequireEnvVar("MACAROON_HEADER")

	// connect to database
	dbParams := &db.ConnectionParams{
		Host:         util.RequireEnvVar("POSTGRES_HOST"),
		Port:         util.RequireEnvVar("POSTGRES_PORT"),
		User:         util.RequireEnvVar("POSTGRES_USER"),
		Password:     util.RequireEnvVar("POSTGRES_PASSWORD"),
		DatabaseName: util.RequireEnvVar("POSTGRES_DB"),
	}

	depotDB := db.ConnectToDB(dbParams)
	db.EnableDebugLogs(depotDB)
	db.RunMigrations(depotDB)

	// connect to node via grpc
	var (
		lnHost        = util.RequireEnvVar("LN_NODE_URL")
		tlsPath       = util.RequireEnvVar("LND_TLS_CERT_PATH")
		macDir        = ""
		network       = "mainnet"
		macDataOption = lndclient.MacaroonData(macaroon)
	)

	client, err := lndclient.NewBasicClient(lnHost, tlsPath, macDir, network, macDataOption)
	if err != nil {
		log.Fatal(err.Error())
	}

	nodeInfo, err := backup.GetInfo(client)
	if err != nil {
		log.Fatal(err.Error())
	}

	node := &db.Node{
		ID:       0,
		URL:      lnHost,
		Alias:    nodeInfo.Alias,
		Pubkey:   nodeInfo.IdentityPubkey,
		Macaroon: macaroon,
	}

	err = db.InsertNode(node, depotDB)
	if err != nil {
		log.Fatal(err.Error())
	}

	const pollInterval = 60 // 1 minute

	done := make(chan bool)
	go health.Monitor(pollInterval, client)
	go backup.SaveChannelBackups(pollInterval, node, client, depotDB)

	<-done // Block forever
}
