package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mvpratt/nodewatcher/db"
	"github.com/uptrace/bun"

	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// Send a text message
func sendSMS(twilioClient *twilio.RestClient, msg string, to string, from string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(from)
	params.SetBody(msg)

	_, err := twilioClient.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	fmt.Println("\nSMS sent successfully!")
	return nil
}

// Exit if environment variable not defined
func requireEnvVar(varName string) string {
	env := os.Getenv(varName)
	if env == "" {
		log.Fatalf("\nERROR: %s environment variable must be set.", varName)
	}
	return env
}

func processGetInfoResponse(info *lnrpc.GetInfoResponse) string {
	statusJSON, err := json.MarshalIndent(info, " ", "    ")
	if err != nil {
		log.Print(err.Error())
	}
	statusString := string(statusJSON)

	if info.SyncedToChain != true {
		return fmt.Sprintf("\n\nWARNING: Lightning node is not fully synced."+
			"\nDetails: %s", statusString)
	}
	if info.SyncedToGraph != true {
		return fmt.Sprintf("\n\nWARNING: Network graph is not fully synced."+
			"\nDetails: %s", statusString)
	}

	// Check how long since last block. Convert unix time string into base10, 64-bit int
	lastBlockTime := info.BestHeaderTimestamp
	timeSinceLastBlock := time.Now().Sub(time.Unix(lastBlockTime, 0))
	return fmt.Sprintf(
		"\nGood news, lightning node \"%s\" is fully synced!"+
			"\nLast block received %s ago", info.Alias, timeSinceLastBlock)
}

func saveChannelBackups(node *db.Node, client lnrpc.LightningClient, depotDB *bun.DB) error {
	response := getChannels(client)
	for _, item := range response.Channels {
		err := db.InsertChannel(item, node.Pubkey, depotDB)
		if err != nil {
			return err
		}
	}

	// static channel backup
	chanBackups := getChannelBackups(client)
	for _, item := range chanBackups.SingleChanBackups.ChanBackups {
		err := db.InsertChannelBackup(item, depotDB)
		if err != nil {
			return err
		}
	}

	// mulitchannel backup
	err := db.InsertMultiChannelBackup(chanBackups.MultiChanBackup, node.Pubkey, depotDB)
	if err != nil {
		return err
	}

	return nil
}

func getInfo(client lnrpc.LightningClient) *lnrpc.GetInfoResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	info, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		log.Print(err)
	}
	return info
}

func getChannels(client lnrpc.LightningClient) *lnrpc.ListChannelsResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channels, err := client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})
	if err != nil {
		log.Print(err)
	}
	return channels
}

func getChannelBackups(client lnrpc.LightningClient) *lnrpc.ChanBackupSnapshot {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	chanBackups, err := client.ExportAllChannelBackups(ctx, &lnrpc.ChanBackupExportRequest{})
	if err != nil {
		log.Print(err)
	}
	return chanBackups
}

func verifyBackup(client lnrpc.LightningClient, snapshot lnrpc.ChanBackupSnapshot) *lnrpc.VerifyChanBackupResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.VerifyChanBackup(ctx, &snapshot)
	if err != nil {
		log.Print(err)
	}
	return response
}

// Once a day, send a text message with lightning node status if SMS_ENABLE is true
func healthMonitor(statusPollInterval time.Duration, client lnrpc.LightningClient) {
	const statusNotifyTime = 1 // when time = 01:00 UTC

	smsEnable := requireEnvVar("SMS_ENABLE")
	var smsTo, smsFrom string
	var twilioClient *twilio.RestClient

	if smsEnable == "TRUE" {
		smsTo = requireEnvVar("TO_PHONE_NUMBER")
		smsFrom = requireEnvVar("TWILIO_PHONE_NUMBER")
		_ = requireEnvVar("TWILIO_ACCOUNT_SID")
		_ = requireEnvVar("TWILIO_AUTH_TOKEN")
		twilioClient = twilio.NewRestClient()
	} else {
		fmt.Println("\nWARNING: Text messages disabled. " +
			"Set environment variable SMS_ENABLE to TRUE to enable SMS status updates")
	}

	smsAlreadySent := false

	for true {
		fmt.Println("\nChecking node status ...")

		nodeInfo := getInfo(client)
		textMsg := processGetInfoResponse(nodeInfo)
		isTimeToSendStatus := (time.Now().Hour() == statusNotifyTime)

		if smsEnable == "TRUE" && isTimeToSendStatus == true && smsAlreadySent == false {
			sendSMS(twilioClient, textMsg, smsTo, smsFrom)
			smsAlreadySent = true
		}

		// if time to send status window has passed, reset the smsAlreadySent boolean
		if isTimeToSendStatus == false && smsAlreadySent == true {
			smsAlreadySent = false
		}
		fmt.Println(textMsg)
		time.Sleep(statusPollInterval * time.Second)
	}
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
	//db.EnableDebugLogs(depotDB)
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

	nodeInfo := getInfo(client)

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
	go healthMonitor(statusPollInterval, client)

	for true {
		fmt.Println("\nSaving channel backups ...")
		err := saveChannelBackups(node, client, depotDB)
		if err != nil {
			log.Print(err.Error())
		}

		// WIP
		// get backup from db
		// multiBackup, err := db.FindMultiChannelBackupByPubkey(node.Pubkey, depotDB)
		// if err != nil {
		// 	log.Print(err.Error())
		// }

		time.Sleep(statusPollInterval * time.Second)
	}
}
