package main

import (
	"log"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/health"
	"github.com/mvpratt/nodewatcher/util"
)

// Nodewatcher runs two processes:
//  1. Checks the health of an LND node and sends an SMS once a day with the status
func main() {
	macaroon := util.RequireEnvVar("MACAROON_HEADER")

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

	const pollInterval = 60 // 1 minute

	done := make(chan bool)
	go health.Monitor(pollInterval, client)

	<-done // Block forever
}
