package main

import (
	wrap "github.com/alomerry/go-tools/web/api"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.Static("/sgs", "static/sgs")

	api := router.Group("/api")
	wrap.HandleAPI(api)

	router.Run(":8089")
}
