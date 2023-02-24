package util

import (
	"log"
	"os"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/internal/db"
)

// RequireEnvVar returns the variable specified, or exits if environment variable not defined
func RequireEnvVar(varName string) string {
	env := os.Getenv(varName)
	if env == "" {
		log.Fatalf("\nERROR: %s environment variable must be set.", varName)
	}
	return env
}

// GetLndClient returns a lndclient for a given node
func GetLndClient(node db.Node) (*lndclient.LightningClient, error) {

	tlsPath := RequireEnvVar("LND_TLS_CERT_PATH") // todo - handle certs for each node

	config := &lndclient.LndServicesConfig{
		LndAddress:            node.URL,
		Network:               lndclient.NetworkMainnet,
		CustomMacaroonHex:     node.Macaroon,
		TLSPath:               tlsPath,
		Insecure:              false,
		BlockUntilChainSynced: false,
		BlockUntilUnlocked:    false,
	}

	// connect to node via grpc
	services, err := lndclient.NewLndServices(config)
	if err != nil {
		return nil, err
	}
	return &services.LndServices.Client, nil
}
