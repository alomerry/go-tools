package sgs

import (
	"context"
	"fmt"
	"github.com/alomerry/go-tools/sgs/tools"
	"github.com/alomerry/go-tools/web/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	mergeExcelWorkDir = ".tmp/merge-excels"
)

func HandleSgsMergeExcels(group *gin.RouterGroup) {
	meg = group.Group("/merge-excels")

	meg.POST("/files", initMergeExcel)
	meg.POST("/execute", mergeExcel)
	meg.GET("/status", mergeExcelStatus)
}

func initMergeExcel(c *gin.Context) {
	utils.ClearDirectory(mergeExcelWorkDir)

	form, _ := c.MultipartForm()
	files := form.File["files"]
	for _, file := range files {
		err := c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", mergeExcelWorkDir, file.Filename))
		fmt.Println(err)
	}
	c.Status(http.StatusOK)
}

func mergeExcel(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), workExpire)
	status.mergeExcelCtx = utils.Work{Ctx: ctx, Cancel: cancel}
	go func() {
		defer cancel()
		tools.DoMergeExcelSheets(mergeExcelWorkDir)
	}()
	c.Status(http.StatusOK)
}

func mergeExcelStatus(c *gin.Context) {
	select {
	case <-status.mergeExcelCtx.Ctx.Done():
		c.String(http.StatusOK, "done")
	default:
		c.String(http.StatusOK, "running")
	}
}
