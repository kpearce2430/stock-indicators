package worksheets

import (
	"fmt"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
)

// LookupSheet creates the LookupSheet
func (w *WorkSheet) LookupSheet(worksheetName string) error {
	_, err := w.StockFile.NewSheet(worksheetName)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return err
	}

	colInfoName, err := column_info.New(w.StockFile.GetFile(), "Name", worksheetName, 1)
	if err != nil {
		return err
	}
	colInfoValue, err := column_info.New(w.StockFile.GetFile(), "Value", worksheetName, 2)
	if err != nil {
		return err
	}

	if err = colInfoName.WriteHeader(1, w.StockFile.Styles.Header); err != nil {
		return err
	}
	if err = colInfoValue.WriteHeader(1, w.StockFile.Styles.Header); err != nil {
		return err
	}
	i := 2
	for k, v := range w.Lookups.LookUps {
		if err = colInfoName.WriteCell(i, k, w.StockFile.Styles.TextStyle(i)); err != nil {
			return err
		}
		if err = colInfoValue.WriteCell(i, v, w.StockFile.Styles.TextStyle(i)); err != nil {
			return err
		}
		i++
	}

	if err = colInfoName.SetColumnSize(); err != nil {
		return err
	}
	if err = colInfoValue.SetColumnSize(); err != nil {
		return err
	}
	return nil
}
