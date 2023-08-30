package utils

import (
	"github.com/tealeg/xlsx/v3"
	"strings"
)

func SetCellValueToSheet(oldCell, newCell *xlsx.Cell, sheet *xlsx.Sheet) {
	switch oldCell.Type() {
	case xlsx.CellTypeString:
		newCell.SetValue(oldCell.Value)
	case xlsx.CellTypeStringFormula:
		formula := oldCell.Formula()
		if len(strings.Split(formula, "!")) >= 2 {
			x, y, _ := xlsx.GetCoordsFromCellIDString(strings.Split(formula, "!")[1])
			row, _ := sheet.File.Sheet[strings.Split(formula, "!")[0]].Row(y)
			SetCellValueToSheet(row.GetCell(x), newCell, sheet)
		} else {
			panic(formula)
		}
	case xlsx.CellTypeNumeric:
		newCell.SetNumeric(oldCell.Value) // ?
	case xlsx.CellTypeBool:
		newCell.SetBool(oldCell.Bool())
	case xlsx.CellTypeInline:
		fallthrough
	case xlsx.CellTypeDate:
		ct, err := oldCell.GetTime(false)
		if err != nil {
			panic(err)
		}
		newCell.SetDate(ct)
	}
}
