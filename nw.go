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

func main() {

	fmt.Println("\n Getting node status ...")

	// Get SMS environment variables
	smsEnable := os.Getenv("SMS_ENABLE")
	smsTo := os.Getenv("TO_PHONE_NUMBER")
	smsFrom := os.Getenv("TWILIO_PHONE_NUMBER")
	twilioAccountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	twilioAuthToken := os.Getenv("TWILIO_AUTH_TOKEN")

	// Note: Twilio credentials must be defined as environment variables for text messaging to work.
	if smsEnable != "TRUE" {
		fmt.Println("\nWARNING: Text messages disabled. " +
			"Set environment variable SMS_ENABLE to TRUE to enable SMS status updates")
	} else if smsEnable == "TRUE" && (twilioAccountSid == "" || twilioAuthToken == "") {
		log.Fatal("\nERROR: Twilio credentials not set. " +
			"TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN must be set as environment variables")
	} else if smsTo == "" || smsFrom == "" {
		log.Fatal("\nERROR: Twilio phone numbers not set. " +
			"TWILIO_PHONE_NUMBER and TO_PHONE_NUMBER must be set as environment variables")
	}

	// Twilio client
	twilioClient := twilio.NewRestClient()

	// Get lightning node credentials
	nodeURL := os.Getenv("LN_NODE_URL")
	macaroon := os.Getenv("MACAROON_HEADER")
	if nodeURL == "" {
		log.Fatal("\nERROR: LN_NODE_URL environment variable must be set.")

	} else if macaroon == "" {
		log.Fatal("\nERROR: MACAROON_HEADER environment variable must be set.")
	}

	// Request status from the node
	nodeURL += "/v1/getinfo"
	response, err := httpNodeRequest(nodeURL, "GET", macaroon)
	if err != nil {
		print(err)
	}
	defer response.Body.Close() // note: this will not close until the end of main()

	var data struct {
		Alias               string `json:"alias"`
		IdentityPubkey      string `json:"identity_pubkey"`
		SyncedToChain       bool   `json:"synced_to_chain"`
		SyncedToGraph       bool   `json:"synced_to_graph"`
		BlockHeight         int    `json:"block_height"`
		BestHeaderTimestamp string `json:"best_header_timestamp"`
	}

	errDecoder := json.NewDecoder(response.Body).Decode(&data)
	if errDecoder != nil {
		print(errDecoder)
	}

	// Marshall data into JSON
	statusJSON, err := json.MarshalIndent(data, " ", "    ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	var statusString string
	statusString = string(statusJSON)

	// Check how long since last block. Convert unix time string into base10, 64-bit int
	lastBlockTime, _ := strconv.ParseInt(data.BestHeaderTimestamp, 10, 64)
	t := time.Unix(lastBlockTime, 0)
	timeSinceLastBlock := time.Now().Sub(t)

	// Contents to be sent via SMS
	var textMsg string

	// Detect if lightning node is synced
	if data.SyncedToChain != true {
		textMsg = fmt.Sprintf("WARNING: Lightning node is not fully synced."+
			"\nDetails: %s", statusString)
	} else if data.SyncedToGraph != true {
		textMsg = fmt.Sprintf("WARNING: Network graph is not fully synced."+
			"\nDetails: %s", statusString)
	} else {
		textMsg = fmt.Sprintf(
			"\n\nGood news, lightning node \"%s\" is fully synced!"+
				"\nLast block received %s minutes ago", data.Alias, timeSinceLastBlock)
	}

	// Send SMS with status
	if smsEnable == "TRUE" {
		sendSMS(twilioClient, textMsg, smsTo, smsFrom)
	}

	fmt.Println(textMsg)
}
