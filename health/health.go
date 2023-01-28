// Package health provides functions to check the status of an LND lightning node
package health

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mvpratt/nodewatcher/backup"
	"github.com/mvpratt/nodewatcher/util"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"

	"github.com/lightningnetwork/lnd/lnrpc"
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

// Monitor - Once a day, send a text message with lightning node status if SMS_ENABLE is true
func Monitor(statusPollInterval time.Duration, client lnrpc.LightningClient) {
	const statusNotifyTime = 1 // when time = 01:00 UTC

	smsEnable := util.RequireEnvVar("SMS_ENABLE")
	var smsTo, smsFrom string
	var twilioClient *twilio.RestClient

	if smsEnable == "TRUE" {
		smsTo = util.RequireEnvVar("TO_PHONE_NUMBER")
		smsFrom = util.RequireEnvVar("TWILIO_PHONE_NUMBER")
		_ = util.RequireEnvVar("TWILIO_ACCOUNT_SID")
		_ = util.RequireEnvVar("TWILIO_AUTH_TOKEN")
		twilioClient = twilio.NewRestClient()
	} else {
		fmt.Println("\nWARNING: Text messages disabled. " +
			"Set environment variable SMS_ENABLE to TRUE to enable SMS status updates")
	}

	smsAlreadySent := false

	for true {
		fmt.Println("\nChecking node status ...")

		nodeInfo := backup.GetInfo(client)
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
