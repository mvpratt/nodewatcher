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
func http_node_request(url, method string, macaroon string) (http.Response, error) {
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
func send_text(twilio_client *twilio.RestClient, msg string, to string, from string) {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(from)
	params.SetBody(msg)

	_, err := twilio_client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("\nSMS sent successfully!")
	}
}

func main() {

	fmt.Println("\n Getting node status ...")

	// Get environment variables
	sms_enable := os.Getenv("SMS_ENABLE")
	sms_to := os.Getenv("TO_PHONE_NUMBER")
	sms_from := os.Getenv("TWILIO_PHONE_NUMBER")
	twilio_account_sid := os.Getenv("TWILIO_ACCOUNT_SID")
	twilio_auth_token := os.Getenv("TWILIO_AUTH_TOKEN")

	// Note: Twilio credentials must be defined as environment variables for text messaging to work.
	if sms_enable != "TRUE" {
		fmt.Println("\nWARNING: Text messages disabled. " +
			"Set environment variable SMS_ENABLE to TRUE to enable SMS status updates")
	} else if sms_enable == "TRUE" && (twilio_account_sid == "" || twilio_auth_token == "") {
		fmt.Println("\nERROR: Twilio credentials not set. " +
			"TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN must be set as environment variables")
		return
	} else if sms_to == "" || sms_from == "" {
		fmt.Println("\nERROR: Twilio phone numbers not set. " +
			"TWILIO_PHONE_NUMBER and TO_PHONE_NUMBER must be set as environment variables")
		return
	}

	// Twilio client
	twilio_client := twilio.NewRestClient()

	// Get lightning node credentials
	node_url := os.Getenv("LN_NODE_URL")
	macaroon := os.Getenv("MACAROON_HEADER")
	if node_url == "" {
		fmt.Println("\nERROR: LN_NODE_URL environment variable must be set.")
		return

	} else if macaroon == "" {
		fmt.Println("\nERROR: MACAROON_HEADER environment variable must be set.")
		return
	}

	// Request status from the node
	response, err := http_node_request(node_url, "GET", macaroon)
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

	err2 := json.NewDecoder(response.Body).Decode(&data)
	if err2 != nil {
		print(err2)
	}

	// Marshall data into JSON
	status_json, err3 := json.MarshalIndent(data, " ", "    ")
	if err3 != nil {
		log.Fatalf(err3.Error())
	}
	var status_string string
	status_string = string(status_json)

	// Check how long since last block. Convert unix time string into base10, 64-bit int
	lastblktime, _ := strconv.ParseInt(data.BestHeaderTimestamp, 10, 64)
	t := time.Unix(lastblktime, 0)
	timesincelastblk := time.Now().Sub(t)

	// Contents to be sent via SMS
	var text_msg string

	// Detect if lightning node is synced
	if data.SyncedToChain != true {
		text_msg = fmt.Sprintf("WARNING: Lightning node is not fully synced."+
			"\nDetails: %s", status_string)
	} else if data.SyncedToGraph != true {
		text_msg = fmt.Sprintf("WARNING: Network graph is not fully synced."+
			"\nDetails: %s", status_string)
	} else {
		text_msg = fmt.Sprintf(
			"\n\nGood news, lightning node \"%s\" is fully synced!"+
				"\nLast block received %s minutes ago", data.Alias, timesincelastblk)
	}

	// Send SMS with status
	if sms_enable == "TRUE" {
		send_text(twilio_client, text_msg, sms_to, sms_from)
	}

	fmt.Println(text_msg)
}
