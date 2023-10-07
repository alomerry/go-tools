package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func HandleSgsTools(group *gin.RouterGroup) {
	group.POST("/merge-excels/files", mergeExcel)
}

func mergeExcel(c *gin.Context) {
	// Multipart form
	form, _ := c.MultipartForm()
	files := form.File["upload"]
	log.Println(files)
	for _, file := range files {
		log.Println(file.Filename)

		// 上传文件至指定目录
		err := c.SaveUploadedFile(file, "/Users/alomerry/workspace/go-tools/web/static/xxx.xx")
		fmt.Println(err)
	}
	c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
}
