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
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// GithubLatestReleaseResponse is the response from the Github API
type GithubLatestReleaseResponse struct {
	TagName string `json:"tag_name"`
}

// SmsParams contains the parameters for sending an SMS message
type SmsParams struct {
	Enable           bool
	To               string
	From             string
	TwilioClient     *twilio.RestClient
	TwilioAccountSID string
	TwilioAuthToken  string
}

// MonitorParams contains the parameters for monitoring an LND node
type MonitorParams struct {
	SMS        SmsParams
	Interval   time.Duration
	NotifyTime int
}

// Get latest release tag from Github
func getLatestReleaseTag(org string, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", org, repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	var release GithubLatestReleaseResponse
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
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
func sendSMS(sms SmsParams, msg string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(sms.To)
	params.SetFrom(sms.From)
	params.SetBody(msg)

	_, err := sms.TwilioClient.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	return nil
}

func warning(warn string, status string) string {
	return fmt.Sprintf("\n\nWARNING: %s\nDetails: %s", warn, status)
}

func isLatestVersion(info *lndclient.Info, status string) (bool, error) {
	latest, err := getLatestReleaseTag("lightningnetwork", "lnd")
	if err != nil {
		return false, err
	}
	if !compareVersions(latest, info.Version) {
		return false, err
	}
	return true, nil
}

func generateStatusMessage(info *lndclient.Info) (string, error) {
	infoJSON, err := json.MarshalIndent(info, " ", "    ")
	if err != nil {
		return "", err
	}

	if !info.SyncedToChain {
		return warning("Lightning node is not fully synced.", string(infoJSON)), nil
	}
	if !info.SyncedToGraph {
		return warning("Network graph is not fully synced.", string(infoJSON)), nil
	}

	isLatest, err := isLatestVersion(info, string(infoJSON))
	if err != nil {
		return "", err
	}
	if !isLatest {
		return warning("Lightning node is not running the latest version", info.Version), nil
	}

	// Check how long since last block. Convert unix time string into base10, 64-bit int
	lastBlockTime := info.BestHeaderTimeStamp
	timeSinceLastBlock := time.Since(lastBlockTime)
	return fmt.Sprintf(
		"\nGood news, lightning node \"%s\" is fully synced!"+
			"\nLast block received %s ago", info.Alias, timeSinceLastBlock), nil
}

// getNodeInfo - Get node info from lnd
func getNodeInfo(client lndclient.LightningClient) (*lndclient.Info, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return client.GetInfo(ctx)
}

// Monitor - Once a day, send a text message with lightning node status if params.SMS.Enable is true
func Monitor(params MonitorParams, client lndclient.LightningClient) {
	alreadySent := false

	for {
		log.Println("\nChecking node status ...")
		time.Sleep(params.Interval)

		nodeInfo, err := getNodeInfo(client)
		if err != nil {
			log.Print(err.Error())
			continue // no point in processing info response
		}

		statusMsg, err := generateStatusMessage(nodeInfo)
		if err != nil {
			log.Print(err.Error())
			continue // no point in processing info response
		}

		sendWindow := time.Now().Hour() == params.NotifyTime // 1-hour notify window

		if sendWindow && params.SMS.Enable && !alreadySent {
			err := sendSMS(params.SMS, statusMsg)
			if err != nil {
				log.Print(err.Error())
			} else {
				log.Println("\nSMS sent successfully!")
				alreadySent = true
			}
		}
		if !sendWindow {
			alreadySent = false
		}
		log.Println(statusMsg)
	}
}
