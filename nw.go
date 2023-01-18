package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

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
		log.Fatalf(err.Error())
	}
	statusString := string(statusJSON)

	// todo - test this
	if info.SyncedToChain != true {
		return fmt.Sprintf("\n\nWARNING: Lightning node is not fully synced."+
			"\nDetails: %s", statusString)
	}
	if info.SyncedToGraph != true {
		return fmt.Sprintf("\n\nWARNING: Network graph is not fully synced."+
			"\nDetails: %s", statusString)
	}

	// Check how long since last block. Convert unix time string into base10, 64-bit int
	lastBlockTime := info.BestHeaderTimestamp //strconv.ParseInt(info.BestHeaderTimestamp, 10, 64)
	timeSinceLastBlock := time.Now().Sub(time.Unix(lastBlockTime, 0))
	return fmt.Sprintf(
		"\nGood news, lightning node \"%s\" is fully synced!"+
			"\nLast block received %s ago", info.Alias, timeSinceLastBlock)
}

// Once a day, send a text message with lightning node status if SMS_ENABLE is true,
func main() {
	const statusPollInterval = 60 // 1 minute
	const statusNotifyTime = 1    // when time = 01:00 UTC

	//macaroon := requireEnvVar("MACAROON_HEADER")
	nodeURL := requireEnvVar("LN_NODE_URL")
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

	var (
		options = lndclient.Insecure() // todo - no tls
		lnHost  = nodeURL
		tlsPath = "/Users/mike/projects/bitcoin/mvpratt/voltage-creds"
		macDir  = "/Users/mike/projects/bitcoin/mvpratt/voltage-creds"
	)

	client, err := lndclient.NewBasicClient(
		lnHost,
		tlsPath,
		macDir,
		"regtest",
		options,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	smsAlreadySent := false

	for true {
		// Request status from the node
		fmt.Println("\nGetting node status ...")

		info, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Node info: %s", info)

		textMsg := processGetInfoResponse(info)

		// check to see if desired time
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
