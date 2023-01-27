package main

import (
	"log"
	"os"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/backup"
	"github.com/mvpratt/nodewatcher/db"
	"github.com/mvpratt/nodewatcher/health"
)

// Exit if environment variable not defined
func requireEnvVar(varName string) string {
	env := os.Getenv(varName)
	if env == "" {
		log.Fatalf("\nERROR: %s environment variable must be set.", varName)
	}
	return env
}

func main() {
	macaroon := requireEnvVar("MACAROON_HEADER")

	// connect to database
	dbParams := &db.ConnectionParams{
		Host:         requireEnvVar("POSTGRES_HOST"),
		Port:         requireEnvVar("POSTGRES_PORT"),
		User:         requireEnvVar("POSTGRES_USER"),
		Password:     requireEnvVar("POSTGRES_PASSWORD"),
		DatabaseName: requireEnvVar("POSTGRES_DB"),
	}

	depotDB := db.ConnectToDB(dbParams)
	db.EnableDebugLogs(depotDB)
	db.RunMigrations(depotDB)

	// connect to node via grpc
	var (
		lnHost        = requireEnvVar("LN_NODE_URL")
		tlsPath       = requireEnvVar("LND_TLS_CERT_PATH")
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
