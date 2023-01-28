package main

import (
	"log"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/backup"
	"github.com/mvpratt/nodewatcher/db"
	"github.com/mvpratt/nodewatcher/health"
	"github.com/mvpratt/nodewatcher/util"
)

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
		log.Fatal(err)
	}

	nodeInfo := backup.GetInfo(client)

	node := &db.Node{
		ID:       0,
		URL:      lnHost,
		Alias:    nodeInfo.Alias,
		Pubkey:   nodeInfo.IdentityPubkey,
		Macaroon: macaroon,
	}

	err = db.InsertNode(node, depotDB)
	if err != nil {
		log.Print(err.Error())
	}

	const statusPollInterval = 60 // 1 minute

	done := make(chan bool)
	go health.Monitor(statusPollInterval, client)
	go backup.SaveChannelBackups(statusPollInterval, node, client, depotDB)

	<-done // Block forever
}
