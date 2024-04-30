package impl

import (
	"fmt"
	open "github.com/tickstep/aliyunpan-api/aliyunpan_open"
	"github.com/tickstep/aliyunpan-api/aliyunpan_open/openapi"
)

type AliyunPan struct {
	openPanClient *open.OpenPanClient
}

func (a *AliyunPan) Init() error {
	openPanClient := open.NewOpenPanClient(openapi.ApiConfig{
		TicketId:     "",
		UserId:       "",
		ClientId:     "",
		ClientSecret: "",
	}, openapi.ApiToken{
		AccessToken: "eyJraWQiOiJLcU8iLC...jIUeqP9mZGZDrFLN--h1utcyVc",
		ExpiredAt:   1709527182,
	}, nil)

	// get user info
	ui, err := openPanClient.GetUserInfo()
	if err != nil {
		fmt.Println("get user info error")
		return err
	}
	fmt.Println("当前登录用户：" + ui.Nickname)

	// do some file operation
	fi, _ := openPanClient.FileInfoByPath(ui.FileDriveId, "/我的文档")
	fmt.Println("\n我的文档 信息：")
	fmt.Println(fi)

	//openPanClient.CreateUploadFile()
	return nil
}

func (a *AliyunPan) Push(filePath string, remotePath string) error {
	//a.openPanClient.CompleteUploadFile()
	return nil
}
