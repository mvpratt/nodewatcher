package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mvpratt/nodewatcher/internal/db"
	"github.com/mvpratt/nodewatcher/internal/graph/model"
	"github.com/mvpratt/nodewatcher/internal/util"
)

func getNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	nodes, err := nwDB.FindAllNodes(context.Background())
	if err != nil {
		log.Print(err)
	}
	json.NewEncoder(w).Encode(nodes)
}

func createNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var node model.Node
	_ = json.NewDecoder(r.Body).Decode(&node)
	dbNode := &db.Node{
		ID:       int64(node.ID),
		URL:      node.URL,
		Alias:    node.Alias,
		Pubkey:   node.Pubkey,
		Macaroon: node.Macaroon,
	}
	nwDB.InsertNode(dbNode)
	json.NewEncoder(w).Encode(node)
}

func getMultiChannelBackups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	backups, err := nwDB.FindAllMultiChannelBackups(context.Background())
	if err != nil {
		log.Print(err)
	}
	json.NewEncoder(w).Encode(backups)
}

var nwDB db.NodewatcherDB

func main() {

	dbParams := &db.ConnectionParams{
		Host:         util.RequireEnvVar("POSTGRES_HOST"),
		Port:         util.RequireEnvVar("POSTGRES_PORT"),
		User:         util.RequireEnvVar("POSTGRES_USER"),
		Password:     util.RequireEnvVar("POSTGRES_PASSWORD"),
		DatabaseName: util.RequireEnvVar("POSTGRES_DB"),
	}

	nwDB = db.NodewatcherDB{}
	nwDB.ConnectToDB(dbParams)
	nwDB.EnableDebugLogs()
	nwDB.RunMigrations()

	r := mux.NewRouter()

	r.HandleFunc("/nodes", getNodes).Methods("GET")
	r.HandleFunc("/multi-channel-backups", getMultiChannelBackups).Methods("GET")
	r.HandleFunc("/nodes", createNode).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", r))
}
