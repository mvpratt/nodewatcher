package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mvpratt/nodewatcher/db"

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

func checkError(err error) {
	if err != nil {
		panic(err)
	}
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

// Once a day, send a text message with lightning node status if SMS_ENABLE is true,
func main() {
	const statusPollInterval = 60 // 1 minute
	const statusNotifyTime = 1    // when time = 01:00 UTC

	smsEnable := requireEnvVar("SMS_ENABLE")
	macaroon := requireEnvVar("MACAROON_HEADER")

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

	var (
		lnHost        = requireEnvVar("LN_NODE_URL")
		tlsPath       = ""
		macDir        = ""
		network       = "mainnet"
		tlsOption     = lndclient.Insecure()
		macDataOption = lndclient.MacaroonData(macaroon)
	)

	var (
		host     = requireEnvVar("POSTGRES_HOST")
		port     = requireEnvVar("POSTGRES_PORT")
		user     = requireEnvVar("POSTGRES_USER")
		password = requireEnvVar("POSTGRES_PASSWORD")
		dbname   = requireEnvVar("POSTGRES_DB")
	)

	// connect to database
	depotDB := db.ConnectToDB(host, port, user, password, dbname)
	db.RunMigrations(depotDB)

	// connect to node via grpc
	// todo -- add tls cert for encrypted comms
	client, err := lndclient.NewBasicClient(
		lnHost,
		tlsPath,
		macDir,
		network,
		tlsOption,
		macDataOption,
	)
	if err != nil {
		log.Fatal(err)
	}

	smsAlreadySent := false

	for true {
		fmt.Println("\nGetting node status ...")

		nodeInfo := getInfo(client)

		node := &db.Node{
			ID:       0,
			URL:      lnHost,
			Alias:    nodeInfo.Alias,
			Pubkey:   nodeInfo.IdentityPubkey,
			Macaroon: macaroon,
		}

		db.InsertNode(node, depotDB)

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

		response := getChannels(client)
		for _, item := range response.Channels {
			db.InsertChannel(item, node.Pubkey, depotDB)
		}

		// static channel backup
		chanBackups := getChannelBackups(client)
		for _, item := range chanBackups.SingleChanBackups.ChanBackups {
			db.InsertChannelBackup(item, depotDB)
		}

		// mulitchannel backup
		db.InsertMultiChannelBackup(chanBackups.MultiChanBackup, node.Pubkey, depotDB)

		// get backup from db
		multiBackup, err := db.FindMultiChannelBackupByPubkey(node.Pubkey, depotDB)
		fmt.Println(multiBackup)
		if err != nil {
			log.Print(err.Error())
		}

		time.Sleep(statusPollInterval * time.Second)
	}
}