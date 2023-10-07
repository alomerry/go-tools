package sgs

import "github.com/gin-gonic/gin"

func HandleSgsDelayReason(group *gin.RouterGroup) {
	dr = group.Group("/delay-reason")

	dr.POST("/files/data-source")
}
