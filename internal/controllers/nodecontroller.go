package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mvpratt/nodewatcher/internal/db"
)

// func GetNodes(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	nodes, err := db.FindAllNodes(context.Background())
// 	if err != nil {
// 		log.Print(err)
// 	}
// 	json.NewEncoder(w).Encode(nodes)
// }

type NodeRequest struct {
	Email string `json:"email"`
}

func GetNodes(context *gin.Context) {
	log.Println("get nodes")
	var request NodeRequest

	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	log.Println("query db")
	// check if email exists and password is correct
	nodes, err := db.FindAllNodes(context)
	log.Println(nodes)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	node := nodes[0]
	context.JSON(http.StatusOK, gin.H{
		"userId":   node.ID,
		"url":      node.URL,
		"alias":    node.Alias,
		"Macaroon": node.Macaroon,
		"Pubkey":   node.Pubkey})
}

func CreateNode(context *gin.Context) {
	var node = db.Node{ID: 0}
	if err := context.ShouldBindJSON(&node); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	err := db.InsertNode(&node)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusCreated, gin.H{
		"userId":   node.ID,
		"url":      node.URL,
		"alias":    node.Alias,
		"Macaroon": node.Macaroon,
		"Pubkey":   node.Pubkey})
}

func GetMultiChannelBackups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	backups, err := db.FindAllMultiChannelBackups(context.Background())
	if err != nil {
		log.Print(err)
	}
	json.NewEncoder(w).Encode(backups)
}
