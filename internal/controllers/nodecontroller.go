package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mvpratt/nodewatcher/internal/db"
)

type NodeRequest struct {
	Email string `json:"email"`
}

type MultiChannelBackupRequest struct {
	Email string `json:"email"`
}

func GetNodes(context *gin.Context) {
	var request NodeRequest

	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	nodes, err := db.FindAllNodes(context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	node := nodes[0]
	context.JSON(http.StatusOK, gin.H{
		"id":       node.ID,
		"url":      node.URL,
		"alias":    node.Alias,
		"macaroon": node.Macaroon,
		"pubkey":   node.Pubkey})
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
		"Pubkey":   node.Pubkey,
	})
}

func GetMultiChannelBackups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	backups, err := db.FindAllMultiChannelBackups(context.Background())
	if err != nil {
		log.Print(err)
	}
	json.NewEncoder(w).Encode(backups)
}

func GetMultiChannelBackup(context *gin.Context) { // todo - return most recent backup
	var request MultiChannelBackupRequest

	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	backups, err := db.FindAllMultiChannelBackups(context)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	backup := backups[0]
	context.JSON(http.StatusOK, gin.H{
		"id":         backup.ID,
		"created_at": backup.CreatedAt,
		"backup":     backup.Backup,
		"node_id":    backup.NodeID,
	})
}
