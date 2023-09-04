package tools

import (
	"encoding/csv"
	"fmt"
	"github.com/alomerry/sgs-tools/utils"
	"github.com/tealeg/xlsx/v3"
	"io"
	"log"
	"os"
	"strings"
)

var (
	root_path = "."
)

func DoMergeExcelSheets() {
	var (
		fileNames []string
		files     []*os.File
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
			fileNames = append(fileNames, fileName)
		} else {
			f, err := os.Open(fileName)
			if err != nil {
				panic(err)
			}
			files = append(files, f)
		}
	}

	if len(fileNames) > 0 {
		iterateAndMerge(fileNames)
	} else {
		genNewMergeCSVResult(files)
	}

}

func iterateAndMerge(fileNames []string) {
	var (
		file *xlsx.File
	)
	file = xlsx.NewFile()
	sheet, _ := file.AddSheet("result")
	for i, fileName := range fileNames {
		f, err := xlsx.OpenFile(fileName)
		if err != nil {
			panic(err)
		}
		s := f.Sheets[0]
		initMaxRow(s, 0)
		s.ForEachRow(func(r *xlsx.Row) error {
			if i > 0 && r.GetCoordinate() == 0 {
				return nil
			}
			row := sheet.AddRow()
			r.ForEachCell(func(c *xlsx.Cell) error {
				utils.SetCellValueToSheet(c, row.AddCell(), nil)
				return nil
			})
			return nil
		})
		fmt.Printf("%s merge rows: [%d] now:[%d]\n", fileName, s.MaxRow, sheet.MaxRow)
	}
	err := file.Save(fmt.Sprintf("%s/%s", root_path, "合并结果.xlsx"))
	if err != nil {
		panic(err)
	}
	log.Default().Println("合并完成！")
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

func initMaxRow(sheet *xlsx.Sheet, noNilCol int) {
	var (
		count = 0
		end   = false
	)
	sheet.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCell(noNilCol).Value != "" {
			count++
		}
		if end {
			return nil
		}
		var (
			last, _ = sheet.Row(r.GetCoordinate() - 1)
			next, _ = sheet.Row(r.GetCoordinate() + 1)
		)
		if last != nil && last.GetCell(noNilCol).Value != "" {
			if next != nil && next.GetCell(noNilCol).Value == "" {
				end = true
			}
		}
		return nil
	})
	sheet.MaxRow = count
}
