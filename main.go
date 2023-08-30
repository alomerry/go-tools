package main

import (
	"github.com/alomerry/sgs-tools/delay"
	_ "net/http"

	_ "github.com/gin-gonic/gin"
)

func main() {
	// delay.DoDelayReason()
	delay.DoDelaySummary()
}

// r := gin.Default()
// r.GET("/ping", func(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "pong",
// 	})
// })
// r.Run()
