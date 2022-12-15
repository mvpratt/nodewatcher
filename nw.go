package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// GetInfoResponse is the expected Lightning node response from /v1/getinfo
type GetInfoResponse struct {
	Alias               string `json:"alias"`
	IdentityPubkey      string `json:"identity_pubkey"`
	SyncedToChain       bool   `json:"synced_to_chain"`
	SyncedToGraph       bool   `json:"synced_to_graph"`
	BlockHeight         int    `json:"block_height"`
	BestHeaderTimestamp string `json:"best_header_timestamp"`
}

// HTTP request to lightning node - requires macaroon for authentication
func httpNodeRequest(url, method string, macaroon string) (http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return *req.Response, fmt.Errorf("Got error %s", err.Error())
	}

	// Set credentials to access lightning node
	req.Header.Set("grpc-metadata-macaroon", macaroon)
	req.Header.Set("user-agent", "nodewatcher")

	response, err := client.Do(req)
	if err != nil {
		return *req.Response, fmt.Errorf("Got error %s", err.Error())
	}

	return *response, nil
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

func processGetInfoResponse(data GetInfoResponse) string {
	statusJSON, err := json.MarshalIndent(data, " ", "    ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	statusString := string(statusJSON)

	if data.SyncedToChain != true {
		return fmt.Sprintf("\n\nWARNING: Lightning node is not fully synced."+
			"\nDetails: %s", statusString)
	}
	if data.SyncedToGraph != true {
		return fmt.Sprintf("\n\nWARNING: Network graph is not fully synced."+
			"\nDetails: %s", statusString)
	}

	// Check how long since last block. Convert unix time string into base10, 64-bit int
	lastBlockTime, _ := strconv.ParseInt(data.BestHeaderTimestamp, 10, 64)
	timeSinceLastBlock := time.Now().Sub(time.Unix(lastBlockTime, 0))
	return fmt.Sprintf(
		"\n\nGood news, lightning node \"%s\" is fully synced!"+
			"\nLast block received %s ago", data.Alias, timeSinceLastBlock)
}

func main() {
	fmt.Println("\nGetting node status ...")

	macaroon := requireEnvVar("MACAROON_HEADER")
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

	// Request status from the node
	response, err := httpNodeRequest(nodeURL+"/v1/getinfo", "GET", macaroon) // todo: retry x times
	if err != nil {
		log.Fatalf("HTTP error requesting node status: %s", err.Error())
	}
	defer response.Body.Close() // note: this will not close until the end of main()

	var data GetInfoResponse

	errDecoder := json.NewDecoder(response.Body).Decode(&data)
	if errDecoder != nil {
		print(errDecoder)
	}

	textMsg := processGetInfoResponse(data)
	if smsEnable == "TRUE" {
		sendSMS(twilioClient, textMsg, smsTo, smsFrom)
	}

	fmt.Println(textMsg)
}
