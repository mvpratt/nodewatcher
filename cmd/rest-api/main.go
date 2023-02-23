package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mvpratt/nodewatcher/internal/controllers"
	"github.com/mvpratt/nodewatcher/internal/db"
	"github.com/mvpratt/nodewatcher/internal/middlewares"
	"github.com/mvpratt/nodewatcher/internal/util"
)

func main() {

	dbParams := &db.ConnectionParams{
		Host:         util.RequireEnvVar("POSTGRES_HOST"),
		Port:         util.RequireEnvVar("POSTGRES_PORT"),
		User:         util.RequireEnvVar("POSTGRES_USER"),
		Password:     util.RequireEnvVar("POSTGRES_PASSWORD"),
		DatabaseName: util.RequireEnvVar("POSTGRES_DB"),
	}

	db.Connect(dbParams)
	db.EnableDebugLogs()
	db.RunMigrations()

	router := initRouter()
	router.Run(":8000")
}

func initRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.POST("/token", controllers.GenerateToken)
		api.POST("/user/register", controllers.RegisterUser)
		secured := api.Group("/secured").Use(middlewares.Auth())
		{
			secured.POST("/user/node", controllers.CreateNode)
			secured.GET("/user/node", controllers.GetNodes)
			secured.GET("/user/node/multi-channel-backup", controllers.GetMultiChannelBackup)
		}
	}
	return router
}
