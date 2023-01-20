package main

// todo - switch to readonly macaroon

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

// Node is a Lightning Node
type Node struct {
	bun.BaseModel `bun:"table:nodes"`

	ID       int32  `bun:"id,pk,autoincrement"`
	URL      string `bun:"url,unique"`
	Alias    string `bun:"alias"`
	Pubkey   string `bun:"pubkey"`
	Macaroon string `bund:"macaroon"`
}

// Channel is a Lightning Channel
type Channel struct {
	bun.BaseModel `bun:"table:channels"`

	ID          int32  `bun:"id,pk,autoincrement"`
	FundingTxid string `bun:"funding_txid"`
	OutputIndex int64  `bun:"output_index"`
	NodeID      *Node  `bun:"rel:has-one",join:node_id=id`
}

// ChannelBackup is a Lightning Channel
type ChannelBackup struct {
	bun.BaseModel `bun:"table:channel_backups"`

	ID        int32  `bun:"id,pk,autoincrement"`
	Backup    string `bun:"backup"`
	ChannelID int32  `bun:"channel_id"`
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

func processGetInfoResponse(info *lnrpc.GetInfoResponse) string {
	statusJSON, err := json.MarshalIndent(info, " ", "    ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	statusString := string(statusJSON)
	//fmt.Println(string(statusJSON))

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

	// connect to the database
	const (
		host     = "localhost"
		port     = 5433
		user     = "user"
		password = "password"
		dbname   = "depot"
	)
	dbctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	node := &Node{
		ID:       0,
		URL:      lnHost,
		Alias:    "",
		Pubkey:   "",
		Macaroon: macaroon,
	}

	// insert node in the db
	res, err := db.NewInsert().
		Model(node).
		On("conflict (\"url\") do nothing").
		Exec(dbctx)

	checkError(err)
	fmt.Println(res)

	// connect to node via grpc
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
		// Request status from the node
		fmt.Println("\nGetting node status ...")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		dbctx, cancel := context.WithTimeout(context.Background(), time.Second) // note - new context
		defer cancel()

		info, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
		if err != nil {
			log.Fatal(err)
		}

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

		//get channels
		channels, err := client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})
		if err != nil {
			log.Fatal(err)
		}

		for _, channel := range channels.Channels {
			splits := strings.Split(channel.ChannelPoint, ":")

			txid := splits[0]
			output, err := strconv.ParseInt(splits[1], 10, 32)
			mychan := &Channel{
				ID:          0,
				FundingTxid: txid,
				OutputIndex: output,
			}

			res, err := db.NewInsert().
				Model(mychan).
				On("conflict (\"funding_txid\",\"output_index\") do nothing").
				Exec(dbctx)
			checkError(err)
			fmt.Println(res)
		}

		// static channel backup
		// todo - readable string
		chanBackups, err := client.ExportAllChannelBackups(ctx, &lnrpc.ChanBackupExportRequest{})
		if err != nil {
			log.Fatal(err)
		}

		for _, item := range chanBackups.SingleChanBackups.ChanBackups {
			channelBackup := &ChannelBackup{
				ID:        0,
				Backup:    string(item.ChanBackup),
				ChannelID: 43,
			}

			res, err := db.NewInsert().
				Model(channelBackup).
				//On("conflict (\"url\") do nothing").
				Returning("*").
				Exec(dbctx)
			checkError(err)
			fmt.Println(res)
		}

		time.Sleep(statusPollInterval * time.Second)
	}
}
