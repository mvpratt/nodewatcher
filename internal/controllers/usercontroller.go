package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mvpratt/nodewatcher/internal/db"
)

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
	// err := nwDB.InsertUser(&user)
	// if err != nil {
	// 	context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	context.Abort()
	// 	return
	// }
	context.JSON(http.StatusCreated, gin.H{"userId": user.ID, "email": user.Email, "password": user.Password})
}
