package api

import (
	"github.com/alomerry/go-tools/web/api/sgs"
	"github.com/gin-gonic/gin"
)

func HandleAPI(api *gin.RouterGroup) {
	sgs.HandleSgsTools(api.Group("/sgs"))
}
