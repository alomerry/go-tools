package main

import (
	"errors"
	"fmt"
	_ "net/http"
	"strings"

	_ "github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx/v3"
)

func main() {
	mapper := getDelayId2Reason()
	setIdSourceDelayReason(mapper)
}

func setIdSourceDelayReason(mapper map[string]string) {
	idSource, err := xlsx.OpenFile("./单号数据源模版.xlsx")
	if err != nil {
		panic(err)
	}
	idSh := idSource.Sheets[0]
	var (
		idIndex, reasonIndex = 0, 2
		lineIndex            = 0
	)
	idSh.ForEachRow(func(r *xlsx.Row) error {
		if lineIndex != 0 {
			id := r.GetCell(idIndex).Value
			if reason, ok := mapper[id]; ok {
				basicReason, err := getSimpleReason(reason)
				if err != nil {
					fmt.Printf("%v\n单号：%v\n", err.Error(), id)
				}
				r.GetCell(reasonIndex).SetString(basicReason)
			} else {
				r.GetCell(reasonIndex).SetString("非化学delay")
			}
		}

		lineIndex++
		return nil
	})
	idSource.Save("./单号数据源模版.xlsx")
}

func getDelayId2Reason() map[string]string {
	var mapper = make(map[string]string)
	delaySource, err := xlsx.OpenFile("/home/user/workspace/tools-delay-reason/delay数据源.xlsx")
	if err != nil {
		panic(err)
	}
	delaySh := delaySource.Sheets[0]
	var (
		idIndex, reasonIndex = 1, 10
	)
	err = delaySh.ForEachRow(func(r *xlsx.Row) error {
		mapper[r.GetCell(idIndex).Value] = r.GetCell(reasonIndex).Value
		return nil
	})
	return mapper
}

func getSimpleReason(str string) (string, error) {
	if (strings.HasPrefix(str, "由") && (strings.HasSuffix(str, "key数据") || strings.HasSuffix(str, "KEY数据"))) ||
		(strings.ContainsAny(str, "分包")) {
		return "分包晚出", nil
	}

	if strings.ContainsAny(str, "CUTTING 原因") || strings.ContainsAny(str, "CUTTING原因") ||
		strings.ContainsAny(str, "SVHC 制样原因") || strings.ContainsAny(str, "SVHC制样原因") || strings.ContainsAny(str, "SVHC制样 原因") ||
		strings.ContainsAny(str, "仪器故障") || strings.ContainsAny(str, "收样晚") || strings.ContainsAny(str, "晚收") || strings.ContainsAny(str, "未收") {
		return "内部原因", nil
	}
	if strings.ContainsAny(str, "已出") || strings.ContainsAny(str, "已完成") || str == "" {
		return "delay", nil
	}
	if strings.ContainsAny(str, "DL") || strings.ContainsAny(str, "TAT") || strings.ContainsAny(str, "复测") {
		return "DL需顺延", nil
	}
	if strings.ContainsAny(str, "数据确认") || strings.ContainsAny(str, "延单") {
		return "数据确认", nil
	}
	return "", errors.New("按照广州delay.txt  未匹配到")
}

// r := gin.Default()
// r.GET("/ping", func(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "pong",
// 	})
// })
// r.Run()
