package main

import (
	"cmd/task_bot/internal/api"
	utils2 "cmd/task_bot/internal/app/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func init() {
	utils2.LoadEnvironmentVariable()
	utils2.ConnectToDb()
}

func main() {
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:4200", "https://app.jaggle.ai"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Authorization"},
	}

	r := gin.Default()
	r.Use(cors.New(config))

	v1 := r.Group("/api/v1/")

	v1.GET("/list-users", api.ListUser)

	err := r.Run(":" + os.Getenv("SERVER_PORT"))

	if err != nil {
		log.Fatal("Error running server on port:" + os.Getenv("PORT"))
	}
}
