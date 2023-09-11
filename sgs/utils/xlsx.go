package utils

import (
	"github.com/spf13/cast"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

const (
	VLOOKUP = "VLOOKUP"
)

func SetCellValueToSheet(oldCell, newCell *xlsx.Cell, sheet *xlsx.Sheet) {
	switch oldCell.Type() {
	case xlsx.CellTypeString:
		newCell.SetValue(oldCell.Value)
	case xlsx.CellTypeStringFormula:
		formula := handleFormula(oldCell.Formula())
		switch formula.fType {
		case VLOOKUP:
			x, y, _ := xlsx.GetCoordsFromCellIDString(formula.vLookup.searchKey)
			row, _ := sheet.File.Sheet[formula.vLookup.sheetName].Row(y)
			searchValue := GetCellValueBySheet(row.GetCell(x), sheet)
			foundRowIdx := -1
			sheet.File.Sheet[formula.vLookup.sheetName].ForEachRow(func(r *xlsx.Row) error {
				for _, colIdx := range formula.vLookup.colRange {
					if GetCellValueBySheet(r.GetCell(colIdx), sheet) == searchValue {
						foundRowIdx = r.GetCoordinate()
					}
				}
				return nil
			})
			if foundRowIdx != -1 {
				selectedRow, _ := sheet.File.Sheet[formula.vLookup.sheetName].Row(foundRowIdx)
				SetCellValueToSheet(selectedRow.GetCell(formula.vLookup.resultCol), newCell, sheet)
			}
		default:
			formula := oldCell.Formula()
			if len(strings.Split(formula, "!")) >= 2 {
				x, y, _ := xlsx.GetCoordsFromCellIDString(strings.Split(formula, "!")[1])
				row, _ := sheet.File.Sheet[strings.Split(formula, "!")[0]].Row(y)
				SetCellValueToSheet(row.GetCell(x), newCell, sheet)
			} else {
				panic(formula)
			}
		}
	case xlsx.CellTypeNumeric:
		formula := oldCell.Formula()
		if len(strings.Split(formula, "!")) >= 2 {
			x, y, _ := xlsx.GetCoordsFromCellIDString(strings.Split(formula, "!")[1])
			row, _ := sheet.File.Sheet[strings.Split(formula, "!")[0]].Row(y)
			SetCellValueToSheet(row.GetCell(x), newCell, sheet)
		} else {
			newCell.SetNumeric(oldCell.Value)
		}
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
		// TODO
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

type formula struct {
	fType   string
	vLookup vLookup
}

type vLookup struct {
	searchKey string
	sheetName string
	colRange  []int
	resultCol int
}

func handleFormula(f string) formula {
	var result vLookup
	if strings.HasPrefix(f, VLOOKUP) {
		fs := strings.Split(strings.TrimSuffix(strings.TrimPrefix(f, VLOOKUP+"("), ")"), ",")
		result.searchKey = fs[0]
		result.resultCol = cast.ToInt(fs[2]) - 1
		ran := strings.Split(fs[1], "!")
		result.sheetName = strings.Split(ran[0], "]")[len(strings.Split(ran[0], "]"))-1]
		result.colRange = GetUniqueColRange(xlsx.ColLettersToIndex(strings.Split(ran[1], ":")[0]), xlsx.ColLettersToIndex(strings.Split(ran[1], ":")[1]))
		return formula{fType: VLOOKUP, vLookup: result}
	}
	return formula{}
}
