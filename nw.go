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

// request to REST api endpoint
func call(url, method string) (http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return *req.Response, fmt.Errorf("Got error %s", err.Error())
	}

	// set credentials to access node
	macaroon := os.Getenv("MACAROON_HEADER")
	req.Header.Set("grpc-metadata-macaroon", macaroon)
	req.Header.Set("user-agent", "nodewatcher")

	response, err := client.Do(req)
	if err != nil {
		return *req.Response, fmt.Errorf("Got error %s", err.Error())
	}

	return *response, nil
}

// send a text message
func send_text(twilio_client *twilio.RestClient, msg string, to string, from string) {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(from)
	params.SetBody(msg)

	_, err := twilio_client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("SMS sent successfully!")
	}
}

func main() {

	fmt.Println("\n Getting node status ...")

	// get environment variables
	sms_enable := os.Getenv("SMS_ENABLE")
	sms_to := os.Getenv("TO_PHONE_NUMBER")
	sms_from := os.Getenv("TWILIO_PHONE_NUMBER")
	twilio_account_sid := os.Getenv("TWILIO_ACCOUNT_SID")
	twilio_auth_token := os.Getenv("TWILIO_AUTH_TOKEN")
	node_url := os.Getenv("LN_NODE_URL")

	// Note: Twilio credentials must be defined as environment variables for text messaging to work.
	if sms_enable != "TRUE" {
		fmt.Println("\nWARNING: Text messages disabled" +
			"set environment variable SMS_ENABLE to TRUE to enable SMS status updates")
	} else if sms_enable == "TRUE" && (twilio_account_sid == "" || twilio_auth_token == "") {
		fmt.Println("\nERROR: Twilio credentials not set." +
			"TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN must be set as environment variables")
		return
	} else if sms_to == "" || sms_from == "" {
		fmt.Println("\nERROR: Twilio phone numbers not set." +
			"TWILIO_PHONE_NUMBER and TO_PHONE_NUMBER must be set as environment variables")
		return
	}

	// Twilio client
	twilio_client := twilio.NewRestClient()

	// request status from the node
	response, err := call(node_url, "GET")
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

	// marshall data into JSON
	status_json, err := json.MarshalIndent(data, " ", "    ")
	if err != nil {
		log.Fatalf(err.Error())
	}

	// check how long since last block
	// convert unix time string into base10, 64-bit int
	lastblktime, err := strconv.ParseInt(data.BestHeaderTimestamp, 10, 64)
	t := time.Unix(lastblktime, 0)
	timesincelastblk := time.Now().Sub(t)

	var status_string string
	status_string = string(status_json)
	var text_msg string
	// detect if lightning node believes it is synced
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

	// send SMS with status
	if sms_enable == "TRUE" {
		send_text(twilio_client, text_msg, sms_to, sms_from)
	}
	fmt.Println(text_msg)
}
