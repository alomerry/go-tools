package sgs

import (
	"context"
	"fmt"
	"github.com/alomerry/go-tools/sgs/delay"
	"github.com/alomerry/go-tools/web/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	delaySummaryWorkDir = ".tmp/delay-summary"
)

func HandleSgsDelaySummary(group *gin.RouterGroup) {
	ds = group.Group("/delay-summary")

	ds.POST("/files", initDelaySummary)
	ds.POST("/execute", delaySummary)
}

func initDelaySummary(c *gin.Context) {
	utils.ClearDirectory(delaySummaryWorkDir)

	form, _ := c.MultipartForm()
	files := form.File["files"]
	for _, file := range files {
		err := c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", delaySummaryWorkDir, file.Filename))
		fmt.Println(err)
	}
	c.Status(http.StatusOK)
}

func delaySummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), workExpire)
	status.mergeExcelCtx = utils.Work{Ctx: ctx, Cancel: cancel}
	go func() {
		defer cancel()
		delay.DoDelaySummaryMulti(delaySummaryWorkDir)
	}()
	c.Status(http.StatusOK)
}
