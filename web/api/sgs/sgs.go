package sgs

import (
	"github.com/alomerry/go-tools/web/utils"
	"github.com/gin-gonic/gin"
	"time"
)

const (
	baseDir    = "./tmp"
	workExpire = time.Minute
)

var (
	status = executeStatus{}
	meg    *gin.RouterGroup
	dr     *gin.RouterGroup
	ds     *gin.RouterGroup
)

type executeStatus struct {
	mergeExcelCtx utils.Work
}

func HandleSgsTools(group *gin.RouterGroup) {
	HandleSgsData(group)
	HandleSgsMergeExcels(group)
	HandleSgsDelayReason(group)
	HandleSgsDelaySummary(group)
}
