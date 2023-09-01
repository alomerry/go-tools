package delay

import (
	"fmt"
	"github.com/alomerry/sgs-tools/utils"
	"github.com/emirpasic/gods/stacks"
	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/spf13/cast"
	xlsx "github.com/tealeg/xlsx/v3"
	"os"
	"strings"
)

var (
	da_path     = "./A.xlsx"
	db_path     = "./B.xlsx"
	out_path    = "./未出数据.xlsx"
	slims_path  = "./starlims delay.xlsx"
	result_path = "./result.xlsx"
)

func DoDelaySummary() {
	da, db, out, slims := getDataSource()
	da.clearHold()
	slims.mergeDa(da.sheet)
	slims.mergeLogic2Report(slims.getLogicS())
	slims.mergeSortedB(db.getSortedB())
	slims.modifyReportAndSheet3()
	slims.mergeOut(out.getOut())
	slims.modifyReport()
	genNewResult(slims.report)
}

func genNewResult(sheet *xlsx.Sheet) {
	var (
		file   *xlsx.File
		result *xlsx.Sheet
	)
	file = xlsx.NewFile()
	file.AddSheet("报表")
	result = file.Sheets[0]
	sheet.ForEachRow(func(r *xlsx.Row) error {
		row := result.AddRow()
		r.ForEachCell(func(c *xlsx.Cell) error {
			utils.SetCellValueToSheet(c, row.AddCell(), sheet)
			return nil
		})
		return nil
	})
	err := file.Save(result_path)
	if err != nil {
		panic(err)
	}
}

type da struct {
	file  *os.File
	sheet *xlsx.Sheet
}

func (da *da) clearHold() {
	var (
		// TODO 增强库为泛型
		del stacks.Stack = arraystack.New()
	)
	da.sheet.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCoordinate() == 0 {
			return nil
		}
		if r.GetCell(xlsx.ColLettersToIndex("I")).Value == "Y" {
			del.Push(r.GetCoordinate())
		}
		return nil
	})

	for !del.Empty() {
		idx, _ := del.Pop()
		da.sheet.RemoveRowAtIndex(idx.(int))
	}

	da.sheet.File.Save(da_path)
	da.file.Sync()
}

type db struct {
	file  *os.File
	sheet *xlsx.Sheet
}

// TODO 去重
func (db *db) getSortedB() []*xlsx.Row {
	var (
		rows []*xlsx.Row
	)
	db.sheet.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCoordinate() != 0 {
			rows = append(rows, r)
		}
		return nil
	})
	sortFunc := func(a, b *xlsx.Row) int {
		cta, err1 := a.GetCell(xlsx.ColLettersToIndex("AA")).GetTime(false)
		ctb, err2 := b.GetCell(xlsx.ColLettersToIndex("AA")).GetTime(false)
		if err1 != nil || err2 != nil {
			switch {
			case err1 == nil && err2 != nil:
				return -1
			case err1 != nil && err2 == nil:
				return 1
			default:
				return 0
			}
		}
		switch {
		case cta.After(ctb):
			return -1
		case cta.Before(ctb):
			return 1
		default:
			return 0
		}
	}
	utils.Sort[*xlsx.Row](rows, sortFunc)
	return rows
}

type out struct {
	file  *os.File
	sheet *xlsx.Sheet
}

func (ot *out) getOut() []*xlsx.Row {
	var (
		rows []*xlsx.Row
	)
	ot.sheet.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCoordinate() != 0 {
			rows = append(rows, r)
		}
		return nil
	})
	return rows
}

type starlims struct {
	file       *os.File
	dataSource *xlsx.Sheet
	logic      *xlsx.Sheet
	report     *xlsx.Sheet
	sheet3     *xlsx.Sheet
	outReason  *xlsx.Sheet
}

func (sl *starlims) mergeDa(da *xlsx.Sheet) {
	da.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCoordinate() == 0 {
			return nil
		}
		row := sl.dataSource.AddRow()
		r.ForEachCell(func(c *xlsx.Cell) error {
			utils.SetCellValueToSheet(c, row.AddCell(), sl.dataSource)
			return nil
		})
		return nil
	})
	sl.dataSource.File.Save(slims_path)
	sl.file.Sync()
}

func (sl *starlims) getLogicJCell(rowIdx int) string {
	var (
		row, _       = sl.logic.Row(rowIdx)
		jCellFormula = row.GetCell(xlsx.ColLettersToIndex("J")).Formula() // IF((I2<0)*AND(I2>-10000),"C","S")
		eCellVal     = sl.getLogicEorCValue(row, "E")                     //  数据源!AB2
		cCellVal     = sl.getLogicEorCValue(row, "C")
		delay        = eCellVal - cCellVal // E2-C2
	)
	if !strings.HasPrefix(jCellFormula, "IF") {
		panic(fmt.Sprintf("invalid formula: [%v]", jCellFormula))
	}

	if delay < 0 && delay > -10000 {
		return "C"
	} else {
		return "S"
	}
}

func (sl *starlims) getLogicEorCValue(row *xlsx.Row, rolX string) float64 {
	return cast.ToFloat64(utils.GetCellValueBySheet(row.GetCell(xlsx.ColLettersToIndex(rolX)), sl.logic))
}

func (sl *starlims) getLogicS() []*xlsx.Row {
	var (
		rows []*xlsx.Row
	)
	sl.logic.ForEachRow(func(r *xlsx.Row) error {
		// IF((I?<0)*AND(I?>-10000),"C","S")
		if r.GetCoordinate() != 0 && sl.getLogicJCell(r.GetCoordinate()) == "S" {
			rows = append(rows, r)
		}
		return nil
	})
	return rows
}

func (sl *starlims) mergeLogic2Report(logicRows []*xlsx.Row) {
	for i := range logicRows {
		logicRow := sl.report.AddRow()
		logicRows[i].ForEachCell(func(c *xlsx.Cell) error {
			if colIdx, _ := c.GetCoordinates(); colIdx <= xlsx.ColLettersToIndex("H") {
				utils.SetCellValueToSheet(c, logicRow.AddCell(), sl.logic)
			}
			return nil
		})
	}
	sl.report.File.Save(slims_path)
	sl.file.Sync()
}

func (sl *starlims) mergeSortedB(rows []*xlsx.Row) {
	for i := range rows {
		row := sl.sheet3.AddRow()
		row.AddCell().SetString(rows[i].GetCell(xlsx.ColLettersToIndex("A")).Value)
		row.AddCell().SetString(rows[i].GetCell(xlsx.ColLettersToIndex("S")).Value)
	}
	sl.sheet3.File.Save(slims_path)
	sl.file.Sync()
}

func (sl *starlims) modifyReportAndSheet3() {
	var (
		s3ToBMapper = make(map[string]string)
		del         = arraystack.New()
	)
	sl.sheet3.ForEachRow(func(sheet3row *xlsx.Row) error {
		if sheet3row.GetCoordinate() != 0 {
			s3ToBMapper[sheet3row.GetCell(xlsx.ColLettersToIndex("A")).Value] = sheet3row.GetCell(xlsx.ColLettersToIndex("B")).Value
		}
		return nil
	})
	sl.report.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCoordinate() != 0 {
			b, ok := s3ToBMapper[r.GetCell(xlsx.ColLettersToIndex("A")).Value]
			if !ok {
				del.Push(r.GetCoordinate())
			} else if b != "" {
				r.GetCell(xlsx.ColLettersToIndex("G")).SetString(b)
			}
		}
		return nil
	})

	for !del.Empty() {
		idx, _ := del.Pop()
		sl.report.RemoveRowAtIndex(idx.(int))
	}

	sl.report.File.Save(slims_path)
	sl.file.Sync()
}

func (sl *starlims) mergeOut(rows []*xlsx.Row) {
	for i := range rows {
		row := sl.outReason.AddRow()
		row.AddCell().SetString(fmt.Sprintf("%s%s", rows[i].GetCell(xlsx.ColLettersToIndex("A")).Value, rows[i].GetCell(xlsx.ColLettersToIndex("E")).Value))
		rows[i].ForEachCell(func(c *xlsx.Cell) error {
			row.AddCell().SetString(c.Value)
			return nil
		})
	}
	sl.outReason.File.Save(slims_path)
	sl.file.Sync()
}

func (sl *starlims) modifyReport() {
	var (
		outReasonMapper = make(map[string]string)
		del             = arraystack.New()
	)

	sl.outReason.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCoordinate() != 0 {
			outReasonMapper[r.GetCell(xlsx.ColLettersToIndex("B")).Value] = r.GetCell(xlsx.ColLettersToIndex("I")).Value
		}
		return nil
	})

	sl.report.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCoordinate() == 0 {
			return nil
		}

		r.GetCell(xlsx.ColLettersToIndex("J")).SetString(fmt.Sprintf("%s%s", r.GetCell(xlsx.ColLettersToIndex("A")).Value, r.GetCell(xlsx.ColLettersToIndex("G")).Value))
		if r.GetCell(xlsx.ColLettersToIndex("I")).Value == "" {
			reason, ok := outReasonMapper[r.GetCell(xlsx.ColLettersToIndex("A")).Value]
			if !ok {
				del.Push(r.GetCoordinate())
			} else {
				r.GetCell(xlsx.ColLettersToIndex("I")).SetString(reason)
			}
		}
		reason := r.GetCell(xlsx.ColLettersToIndex("I")).Value
		if strings.Contains(reason, "CUTTING 原因") || strings.Contains(reason, "CUTTING原因") {
			r.GetCell(xlsx.ColLettersToIndex("G")).SetString("CUTTING")
		}
		if strings.Contains(reason, "SVHC制样") || strings.Contains(reason, "SVHC 制样") || strings.Contains(reason, "XRF") {
			r.GetCell(xlsx.ColLettersToIndex("G")).SetString("SVHC制样")
		}
		if strings.Contains(reason, "ONHOLD") {
			r.GetCell(xlsx.ColLettersToIndex("G")).SetString("前线原因")
		}
		return nil
	})

	for !del.Empty() {
		idx, _ := del.Pop()
		sl.report.RemoveRowAtIndex(idx.(int))
	}

	sl.report.File.Save(slims_path)
	sl.file.Sync()
}

func getDataSource() (*da, *db, *out, *starlims) {
	return getDA(), getDB(), getOut(), getStarlims()
}

func getDA() *da {
	f, err := os.Open(da_path)
	if err != nil {
		panic(err)
	}
	a, err := xlsx.OpenFile(da_path)
	if err != nil {
		panic(err)
	}
	return &da{f, a.Sheets[0]}
}

func getDB() *db {
	f, err := os.Open(db_path)
	if err != nil {
		panic(err)
	}
	b, err := xlsx.OpenFile(db_path)
	if err != nil {
		panic(err)
	}
	return &db{f, b.Sheets[0]}
}

func getOut() *out {
	f, err := os.Open(out_path)
	if err != nil {
		panic(err)
	}
	o, err := xlsx.OpenFile(out_path)
	if err != nil {
		panic(err)
	}
	return &out{f, o.Sheets[0]}
}

func getStarlims() *starlims {
	f, err := os.Open(slims_path)
	if err != nil {
		panic(err)
	}
	sls, err := xlsx.OpenFile(slims_path)
	if err != nil {
		panic(err)
	}
	res := &starlims{file: f}
	for i := range sls.Sheets {
		switch sls.Sheets[i].Name {
		case "数据源":
			res.dataSource = sls.Sheets[i]
		case "逻辑":
			res.logic = sls.Sheets[i]
		case "报表":
			res.report = sls.Sheets[i]
		case "Sheet3":
			res.sheet3 = sls.Sheets[i]
		case "未出原因":
			res.outReason = sls.Sheets[i]
		}
	}
	clearSheet(res.dataSource)
	initMaxRow(res.logic, 0)
	clearSheet(res.report)
	clearSheet(res.sheet3)
	clearSheet(res.outReason)
	return res
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

func clearSheet(sheet *xlsx.Sheet) {
	var del = arraystack.New()
	sheet.ForEachRow(func(r *xlsx.Row) error {
		if r.GetCoordinate() != 0 {
			del.Push(r.GetCoordinate())
		}
		return nil
	})

	for !del.Empty() {
		idx, _ := del.Pop()
		sheet.RemoveRowAtIndex(idx.(int))
	}
}
