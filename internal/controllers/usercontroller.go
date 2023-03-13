package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mvpratt/nodewatcher/internal/db"
)

// RegisterUser adds a user to the database.
func RegisterUser(context *gin.Context) {
	var user db.User
	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	if err := user.HashPassword(user.Password); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	exists, _ := db.FindUserByEmail(user.Email)
	if exists.Email == user.Email {
		context.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		context.Abort()
		return
	}

	user.ID = 0 // set to 0 so the db will auto-increment
	err := db.InsertUser(&user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusCreated, gin.H{"userId": user.ID, "email": user.Email, "password": user.Password})
}
