package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mvpratt/nodewatcher/internal/graph/model"
)

// dbParams := &db.ConnectionParams{
// 	Host:         util.RequireEnvVar("POSTGRES_HOST"),
// 	Port:         util.RequireEnvVar("POSTGRES_PORT"),
// 	User:         util.RequireEnvVar("POSTGRES_USER"),
// 	Password:     util.RequireEnvVar("POSTGRES_PASSWORD"),
// 	DatabaseName: util.RequireEnvVar("POSTGRES_DB"),
// }

// nwDB := db.NodewatcherDB{}
// nwDB.ConnectToDB(dbParams)
// nwDB.EnableDebugLogs()
// nwDB.RunMigrations()

var nodes []model.Node

func startNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

func createNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var node model.Node
	_ = json.NewDecoder(r.Body).Decode(&node)
	nodes = append(nodes, node)
	json.NewEncoder(w).Encode(node)
}

func main() {

	r := mux.NewRouter()
	nodes = append(nodes, model.Node{ID: 1, URL: "url", Alias: "alias", Pubkey: "pubkey", Macaroon: "macaroon"})
	nodes = append(nodes, model.Node{ID: 2, URL: "url2", Alias: "alias2", Pubkey: "pubkey2", Macaroon: "macaroon2"})
	r.HandleFunc("/nodes", startNodes).Methods("GET")
	r.HandleFunc("/nodes", createNode).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", r))
}
