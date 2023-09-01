package main

import (
	"github.com/alomerry/sgs-tools/delay"
	_ "github.com/gin-gonic/gin"
	_ "net/http"
)

func main() {
	// delay.DoDelayReason()
	delay.DoDelaySummaryMulti()
}

// r := gin.Default()
// r.GET("/ping", func(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "pong",
// 	})
// })
// r.Run()
