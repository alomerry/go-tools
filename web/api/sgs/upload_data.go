package sgs

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HandleSgsData(group *gin.RouterGroup) {
	group.POST("/files/starlims", initStarlims)
}

func initStarlims(c *gin.Context) {
	form, _ := c.MultipartForm()
	files := form.File["files"]
	for _, file := range files {
		err := c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", baseDir, file.Filename))
		fmt.Println(err)
	}
	c.Status(http.StatusOK)
}
