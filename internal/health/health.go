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
	"github.com/mvpratt/nodewatcher/internal/db"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// GithubLatestReleaseResponse is the response from the Github API
type GithubLatestReleaseResponse struct {
	TagName string `json:"tag_name"`
}

// TwilioConfig contains the parameters for sending an SMS message
type TwilioConfig struct {
	From             string
	TwilioClient     *twilio.RestClient
	TwilioAccountSID string
	TwilioAuthToken  string
}

// MonitorParams contains the parameters for monitoring an LND node
type MonitorParams struct {
	Interval   time.Duration
	NotifyTime int
}

// SmsDetails contains the parameters for sending an SMS message
type SmsDetails struct {
	To           string
	From         string
	Body         string
	TwilioClient *twilio.RestClient
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
func sendSMS(details SmsDetails) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(details.To)
	params.SetFrom(details.From)
	params.SetBody(details.Body)

	_, err := details.TwilioClient.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	return nil
}

func warning(warn string) string {
	return fmt.Sprintf("\n\nWARNING: %s", warn)
}

func isLatestVersion(info *lndclient.Info) (bool, error) {
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
	if !info.SyncedToChain {
		return warning("Lightning node is not fully synced."), nil
	}
	if !info.SyncedToGraph {
		return warning("Network graph is not fully synced."), nil
	}

	isLatest, err := isLatestVersion(info)
	if err != nil {
		return "", err
	}
	if !isLatest {
		return warning("Lightning node is not running the latest version."), nil
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

// Check node status, send a text message if user has SMS enabled
func Check(twilioConfig TwilioConfig, node db.Node, lndClient *lndclient.LightningClient) error {
	log.Printf("\nChecking node status: %s", node.Alias)

	user, _ := db.FindUserByID(node.UserID)
	if !user.SmsEnabled {
		log.Println("\nWARNING: Text messages disabled for user.")
	}

	nodeInfo, err := getNodeInfo(*lndClient)
	if err != nil {
		return err
	}

	statusMsg, err := generateStatusMessage(nodeInfo)
	if err != nil {
		return err
	}

	sendWindow := time.Now().UTC().Hour() == user.SmsNotifyTime.Hour() // 1-hour notify window
	alreadySent := time.Since(user.SmsLastSent) < time.Hour*24         // only send once per 24 hours

	// todo - check for twilio env vars before trying to send SMS
	if sendWindow && user.SmsEnabled && !alreadySent {
		smsDetails := SmsDetails{
			To:           user.PhoneNumber,
			From:         twilioConfig.From,
			Body:         statusMsg,
			TwilioClient: twilioConfig.TwilioClient,
		}
		err := sendSMS(smsDetails)
		if err != nil {
			return err
		}
		log.Println("\nSMS sent successfully!")
		user.SmsLastSent = time.Now().UTC()
		db.UpdateUserLastSent(user)
	}
	log.Println(statusMsg)
	return nil
}
