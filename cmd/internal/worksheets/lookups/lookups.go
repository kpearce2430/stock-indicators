package lookups

import (
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/sirupsen/logrus"
)

type LookupSheet struct {
	w worksheets.WorksheetInterface
}

func New(w worksheets.WorksheetInterface) *LookupSheet {
	return &LookupSheet{
		w: w,
	}
}

// LookupSheet creates the LookupSheet
func (l *LookupSheet) LookupSheet(worksheetName string) error {
	stockFile := l.w.GetFile()
	if stockFile == nil {
		logrus.Fatal("Unable to get stock file")
		return nil
	}
	_, err := stockFile.NewSheet(worksheetName)

	colInfoName, err := column_info.New(l.w.GetExcelizeFile(), "Name", worksheetName, 1)
	if err != nil {
		return err
	}
	colInfoValue, err := column_info.New(l.w.GetExcelizeFile(), "Value", worksheetName, 2)
	if err != nil {
		return err
	}

	if err = colInfoName.WriteHeader(1, l.w.GetStyles().Header); err != nil {
		return err
	}
	if err = colInfoValue.WriteHeader(1, l.w.GetStyles().Header); err != nil {
		return err
	}
	i := 2
	lookupData := l.w.GetLookups()

	for k, v := range lookupData.LookUps {
		if err = colInfoName.WriteCell(i, k, l.w.GetStyles().TextStyle(i)); err != nil {
			return err
		}
		if err = colInfoValue.WriteCell(i, v, l.w.GetStyles().TextStyle(i)); err != nil {
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
