package worksheets

import (
	"fmt"
)

//func writeCell(f *excelize.File, sheetName string, col, row int, value any) error {
//	colName, err := excelize.ColumnNumberToName(col)
//	if err != nil {
//		return err
//	}
//	colRow := fmt.Sprintf("%s%d", colName, row)
//	f.SetCellValue(sheetName, colRow, value)
//	return nil
//}

func (w *WorkSheet) LookupSheet(worksheetName string) error {
	_, err := w.File.NewSheet(worksheetName)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return err
	}

	colInfoName, err := NewColumnInfo(w.File, "Name", worksheetName, 1)
	if err != nil {
		return err
	}
	colInfoValue, err := NewColumnInfo(w.File, "Value", worksheetName, 2)
	if err != nil {
		return err
	}

	if err := colInfoName.WriteHeader(1, w.styles.Header); err != nil {
		return err
	}
	if err := colInfoValue.WriteHeader(1, w.styles.Header); err != nil {
		return err
	}
	i := 2
	for k, v := range w.Lookups.LookUps {
		if err := colInfoName.WriteCell(i, k, w.styles.TextStyle(i)); err != nil {
			return err
		}
		if err := colInfoValue.WriteCell(i, v, w.styles.TextStyle(i)); err != nil {
			return err
		}
		i++
	}

	if err := colInfoName.SetColumnSize(); err != nil {
		return err
	}
	if err := colInfoValue.SetColumnSize(); err != nil {
		return err
	}
	return nil
}
