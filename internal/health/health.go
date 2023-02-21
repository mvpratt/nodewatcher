// Package health provides functions to check the status of an LND lightning node
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lightninglabs/lndclient"
	"github.com/mvpratt/nodewatcher/internal/util"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type GithubLatestReleaseResponse struct {
	TagName string `json:"tag_name"`
}

// Get latest release tag from Github
func getLatestReleaseTag(org string, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", org, repo)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return "", err
	}

	var release GithubLatestReleaseResponse
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return release.TagName, nil
}

// Compare version strings
// Assumes:
//
//	githubTag is of the form "v0.15.5-beta"
//	lndVersionString is of the form "0.15.5-beta commit=v0.15.5-beta.f1"
func compareVersions(githubTag string, lndVersionString string) bool {
	githubTag = githubTag[1:]                                  // remove leading "v" from github tag
	lndVersionString = strings.Split(lndVersionString, " ")[0] // remove trailing "commit=" from version string
	return githubTag == lndVersionString
}

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
	log.Println("\nSMS sent successfully!")
	return nil
}

func processGetInfoResponse(info *lndclient.Info) (string, error) {
	statusJSON, err := json.MarshalIndent(info, " ", "    ")
	if err != nil {
		log.Print(err.Error())
		return "", err
	}
	statusString := string(statusJSON)

	if !info.SyncedToChain {
		return fmt.Sprintf("\n\nWARNING: Lightning node is not fully synced."+
			"\nDetails: %s", statusString), nil
	}
	if !info.SyncedToGraph {
		return fmt.Sprintf("\n\nWARNING: Network graph is not fully synced."+
			"\nDetails: %s", statusString), nil
	}

	latest, _ := getLatestReleaseTag("lightningnetwork", "lnd")
	if !compareVersions(latest, info.Version) {
		return fmt.Sprintf("\n\nWARNING: Lightning node is not running the latest version."+
			"\nDetails: %s", statusString), nil
	}

	// Check how long since last block. Convert unix time string into base10, 64-bit int
	lastBlockTime := info.BestHeaderTimeStamp
	timeSinceLastBlock := time.Since(lastBlockTime)
	return fmt.Sprintf(
		"\nGood news, lightning node \"%s\" is fully synced!"+
			"\nLast block received %s ago", info.Alias, timeSinceLastBlock), nil
}

// Monitor - Once a day, send a text message with lightning node status if SMS_ENABLE is true
func Monitor(statusPollInterval time.Duration, client lndclient.LightningClient) {
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
		log.Println("\nWARNING: Text messages disabled. " +
			"Set environment variable SMS_ENABLE to TRUE to enable SMS status updates")
	}

	smsAlreadySent := false

	for {
		log.Println("\nChecking node status ...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel() // todo - defer will never run (endless loop)

		nodeInfo, err := client.GetInfo(ctx)
		if err != nil {
			log.Print(err.Error())
			time.Sleep(statusPollInterval * time.Second)
			continue // no point in processing info response
		}

		textMsg, err := processGetInfoResponse(nodeInfo)
		if err != nil {
			log.Print(err.Error())
			time.Sleep(statusPollInterval * time.Second)
			continue // no point in processing info response
		}

		isTimeToSendStatus := (time.Now().Hour() == statusNotifyTime)

		if smsEnable == "TRUE" && isTimeToSendStatus && !smsAlreadySent {
			err := sendSMS(twilioClient, textMsg, smsTo, smsFrom)
			if err != nil {
				log.Print(err.Error())
			} else {
				smsAlreadySent = true
			}
		}

		// if time to send status window has passed, reset the smsAlreadySent boolean
		if !isTimeToSendStatus && smsAlreadySent {
			smsAlreadySent = false
		}
		log.Println(textMsg)
		time.Sleep(statusPollInterval * time.Second)
	}
}
