package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	getEngine().Run(":8080")
}

func getEngine() *gin.Engine {
	r := gin.Default()
	r.Use(handleErrors)
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			secret := v1.Group("/secret")
			{
				secret.POST("/", createSecret)
				secret.POST("/encode", encodeSecret)
			}
		}
		api.GET("/ping", ping)
	}

	api.Use()

	return r
}

func handleErrors(c *gin.Context) {
	c.Next()
	errorToPrint := c.Errors.ByType(gin.ErrorTypePublic).Last()
	if errorToPrint != nil {
		log.Error(errorToPrint)
		c.JSON(500, gin.H{
			"status":  500,
			"message": errorToPrint.Error(),
		})
	}
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
