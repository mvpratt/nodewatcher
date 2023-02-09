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

func getChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	channels, err := nwDB.FindAllChannels(context.Background())
	if err != nil {
		log.Print(err)
	}
	json.NewEncoder(w).Encode(channels)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users, err := nwDB.FindAllUsers(context.Background())
	if err != nil {
		log.Print(err)
	}
	json.NewEncoder(w).Encode(users)
}

func createNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var node model.Node
	_ = json.NewDecoder(r.Body).Decode(&node)
	dbNode := &db.Node{
		ID:       0,
		URL:      node.URL,
		Alias:    node.Alias,
		Pubkey:   node.Pubkey,
		Macaroon: node.Macaroon,
	}
	nwDB.InsertNode(dbNode)
	json.NewEncoder(w).Encode(node)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user model.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	dbUser := &db.User{
		ID:       0,
		Email:    user.Email,
		Password: user.Password,
	}
	nwDB.InsertUser(dbUser)
	json.NewEncoder(w).Encode(user)
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

	r.HandleFunc("/channels", getChannels).Methods("GET")
	r.HandleFunc("/nodes", getNodes).Methods("GET")
	r.HandleFunc("/nodes", createNode).Methods("POST")
	r.HandleFunc("/multi-channel-backups", getMultiChannelBackups).Methods("GET")
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users", createUser).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", r))
}
