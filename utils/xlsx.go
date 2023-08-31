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

func GetCellValueBySheet(cell *xlsx.Cell, sheet *xlsx.Sheet) any {
	switch cell.Type() {
	case xlsx.CellTypeString:
		return cell.Value
	case xlsx.CellTypeNumeric:
		if cell.Formula() == "" {
			v, err := cell.GeneralNumericWithoutScientific()
			if err != nil {
				panic(err)
			}
			return v
		}
		fallthrough
	case xlsx.CellTypeStringFormula:
		formula := cell.Formula()
		if len(strings.Split(formula, "!")) >= 2 {
			x, y, _ := xlsx.GetCoordsFromCellIDString(strings.Split(formula, "!")[1])
			row, _ := sheet.File.Sheet[strings.Split(formula, "!")[0]].Row(y)
			return GetCellValueBySheet(row.GetCell(x), sheet)
		} else {
			panic(formula)
		}
	case xlsx.CellTypeBool:
		return cell.Bool()
	case xlsx.CellTypeInline:
		fallthrough
	case xlsx.CellTypeDate:
		ct, err := cell.GetTime(false)
		if err != nil {
			panic(err)
		}
		return ct
	}
	panic("empty type")
}
