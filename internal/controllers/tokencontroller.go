package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mvpratt/nodewatcher/internal/auth"
	"github.com/mvpratt/nodewatcher/internal/db"
)

// TokenRequest is the request body for the GenerateToken endpoint
type TokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// GenerateToken returns a JWT that is good for X hours
// borrowed from this tutorial:
// https://codewithmukesh.com/blog/jwt-authentication-in-golang/#Setting_up_the_Database_Migrations
func GenerateToken(context *gin.Context) {
	var request TokenRequest

	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	user, err := db.FindUserByEmail(request.Email)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	credentialError := user.CheckPassword(request.Password)
	if credentialError != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		context.Abort()
		return
	}

	tokenString, err := auth.GenerateJWT(user.Email)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"token": tokenString})
}
