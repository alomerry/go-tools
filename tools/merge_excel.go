package tools

import (
	"encoding/csv"
	"fmt"
	"github.com/alomerry/sgs-tools/utils"
	"github.com/tealeg/xlsx/v3"
	"io"
	"os"
	"strings"
)

var (
	root_path = "/Users/alomerry/workspace/sgs-tools/output"
)

func DoMergeExcelSheets() {
	var (
		sheets []*xlsx.Sheet
		files  []*os.File
	)
	entries, err := os.ReadDir(root_path)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := fmt.Sprintf("%s/%s", root_path, entry.Name())
		if strings.HasSuffix(entry.Name(), "xlsx") {
			f, err := xlsx.OpenFile(fileName)
			if err != nil {
				panic(err)
			}
			sheets = append(sheets, f.Sheets[0])
		} else {
			f, err := os.Open(fileName)
			if err != nil {
				panic(err)
			}
			files = append(files, f)
		}
	}

	if len(sheets) > 0 {
		genNewMergeResult(sheets)
	} else {
		genNewMergeCSVResult(files)
	}

}

func genNewMergeResult(sheets []*xlsx.Sheet) {
	var (
		file *xlsx.File
	)
	file = xlsx.NewFile()
	sheet, _ := file.AddSheet("result")
	for i := range sheets {
		row := sheet.AddRow()
		sheets[i].ForEachRow(func(r *xlsx.Row) error {
			if i > 0 && r.GetCoordinate() == 0 {
				return nil
			}
			r.ForEachCell(func(c *xlsx.Cell) error {
				utils.SetCellValueToSheet(c, row.AddCell(), nil)
				return nil
			})
			return nil
		})
	}

	err := file.Save(fmt.Sprintf("%s/%s", root_path, "合并结果.xlsx"))
	if err != nil {
		panic(err)
	}
}

func genNewMergeCSVResult(files []*os.File) {
	var (
		file   *os.File
		writer *csv.Writer
	)
	file, err := os.Create(fmt.Sprintf("%s/%s", root_path, "合并结果.csv"))
	if err != nil {
		panic(err)
	}
	writer = csv.NewWriter(file)
	for i := range files {
		writer.WriteAll(getCSVSheet(files[i], i == 0))
	}

	writer.Flush()

	err = file.Sync()
	if err != nil {
		panic(err)
	}
}

func getCSVSheet(file *os.File, needFirstRow bool) [][]string {
	var (
		reader = csv.NewReader(file)
		idx    = 0
		result [][]string
	)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if needFirstRow || idx != 0 {
			result = append(result, record)
		}

		idx++
	}
	return result
}
