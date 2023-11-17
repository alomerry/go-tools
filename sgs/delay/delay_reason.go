package delay

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tealeg/xlsx/v3"
)

var (
	delay_source = "8月delay数据源.xlsx"
	id_source    = "8月delay单号.xlsx"
)

func DoDelayReason(path string) {
	if path != "" {
		root_path = path
	}

	delaySource, idSource := getSource()
	setIdSourceDelayReason(idSource, getDelayId2Reason(delaySource))
}

func getSource() (delaySource, idSource *xlsx.File) {
	entries, err := os.ReadDir(root_path)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		path := fmt.Sprintf("%s/%s", root_path, entry.Name())
		switch {
		case strings.Contains(entry.Name(), delay_source):
			delaySource, err = xlsx.OpenFile(path)
			if err != nil {
				panic(err)
			}
		case strings.Contains(entry.Name(), id_source):
			idSource, err = xlsx.OpenFile(path)
			if err != nil {
				panic(err)
			}
		}
	}
	return
}

func setIdSourceDelayReason(idSource *xlsx.File, mapper map[string]string) {
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
					fmt.Printf("%v id:[%v] reason:[%v]\n", err.Error(), id, reason)
				}
				r.GetCell(reasonIndex).SetString(basicReason)
			} else {
				r.GetCell(reasonIndex).SetString("非化学delay")
			}
		}

		lineIndex++
		return nil
	})
	err := idSource.Save(fmt.Sprintf("%s/%s", root_path, id_source))
	if err != nil {
		panic(err)
	}
}

func getDelayId2Reason(delaySource *xlsx.File) map[string]string {
	var mapper = make(map[string]string)

	delaySh := delaySource.Sheets[0]
	var (
		idIndex, reasonIndex = 1, 10
	)
	err := delaySh.ForEachRow(func(r *xlsx.Row) error {
		mapper[r.GetCell(idIndex).Value] = r.GetCell(reasonIndex).Value
		return nil
	})

	if err != nil {
		panic(err)
	}

	return mapper
}

// TODO 优化成 reason-fn
func getSimpleReason(str string) (string, error) {
	if (strings.HasPrefix(str, "由") && (strings.HasSuffix(str, "key数据") || strings.HasSuffix(str, "KEY数据"))) ||
		(strings.ContainsAny(str, "分包")) {
		return "分包晚出", nil
	}

	if strings.ContainsAny(str, "CUTTING 原因") || strings.ContainsAny(str, "CUTTING原因") ||
		strings.ContainsAny(str, "SVHC 制样原因") || strings.ContainsAny(str, "SVHC制样原因") || strings.ContainsAny(str, "SVHC制样 原因") ||
		strings.ContainsAny(str, "仪器故障") || strings.ContainsAny(str, "收样晚") || strings.ContainsAny(str, "晚收") || strings.ContainsAny(str, "未收") || strings.ContainsAny(str, "走总镉") || strings.ContainsAny(str, "九楼") {
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
	if strings.ContainsAny(str, "cancel") || strings.ContainsAny(str, "Cancel") {
		return "Cancel", nil
	}
	if strings.ContainsAny(str, "正常流转") {
		return "非化学delay", nil
	}
	return "", errors.New("原因关键词未匹配到")
}
