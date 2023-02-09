package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mvpratt/nodewatcher/internal/auth"
	"github.com/mvpratt/nodewatcher/internal/db"
)

type TokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func GenerateToken(context *gin.Context) {
	log.Println("generate token")
	var request TokenRequest

	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	log.Println("query db")
	// check if email exists and password is correct
	user, err := db.FindUserByEmail(request.Email)
	log.Println(user)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	log.Println("check password")
	credentialError := user.CheckPassword(request.Password)
	if credentialError != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		context.Abort()
		return
	}
	log.Println("generate jwt")
	tokenString, err := auth.GenerateJWT(user.Email)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"token": tokenString})
}
