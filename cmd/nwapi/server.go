package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mvpratt/nodewatcher/internal/db"
	"github.com/mvpratt/nodewatcher/internal/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// dbParams := &db.ConnectionParams{
	// 	Host:         util.RequireEnvVar("POSTGRES_HOST"),
	// 	Port:         util.RequireEnvVar("POSTGRES_PORT"),
	// 	User:         util.RequireEnvVar("POSTGRES_USER"),
	// 	Password:     util.RequireEnvVar("POSTGRES_PASSWORD"),
	// 	DatabaseName: util.RequireEnvVar("POSTGRES_DB"),
	// }

	// // connect to database
	// depotDB := db.ConnectToDB(dbParams)
	nodeService := &db.NodeImpl{ID: 1, URL: "hello", Alias: "world", Pubkey: "111", Macaroon: "macaroon"}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Node: nodeService}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
