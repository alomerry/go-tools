package main

import (
	wrap "github.com/alomerry/go-tools/web/api"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Static("/sgs", "web/static/sgs")
	api := router.Group("/api")
	{
		wrap.HandleSgsTools(api.Group("/sgs"))
	}
	router.Run(":8089")
}
