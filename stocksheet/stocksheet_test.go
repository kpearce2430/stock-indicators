package stocksheet_test

import (
	"fmt"
	"github.com/kpearce2430/stock-tools/stocksheet"
	ci "github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"math/rand/v2"
	"testing"
	"time"
)

func TestStockSheet_New(t *testing.T) {
	s := stocksheet.New()
	if s == nil {
		t.Error("stocksheet.New() returned nil")
	}
}

func fillRandomStockSheet(sFile *stocksheet.StockFile, sheetName string) error {
	row := 1
	colInfoName, err := ci.New(sFile.GetFile(), "Name", sheetName, 1)
	if err != nil {
		return err
	}

	if err = colInfoName.WriteHeader(row, sFile.Styles.Header); err != nil {
		return err
	}

	colInfoValue, err := ci.New(sFile.GetFile(), "Value", sheetName, 2)
	if err != nil {
		return err
	}

	if err = colInfoValue.WriteHeader(row, sFile.Styles.Header); err != nil {
		return err
	}

	for i := range 12 {
		row++
		_ = colInfoName.WriteCell(row, time.Month(i+1), sFile.Styles.TextStyle(row))
		_ = colInfoValue.WriteCell(row, rand.Float32(), sFile.Styles.NumberStyle(row))
	}
	row++
	_ = colInfoName.WriteCell(row, "Total", sFile.Styles.TextStyle(row))
	colInfoValue.SetFormula(true)
	_ = colInfoValue.WriteCell(row, "=sum(b2:b13)", sFile.Styles.NumberStyle(row))
	return nil
}

// TestStockFile_Sheet1 files out Sheet1 with random data.
func TestStockFile_Sheet1(t *testing.T) {
	sFile := stocksheet.New()
	if sFile == nil {
		t.Error("stocksheet.New() returned nil")
		return
	}

	err := fillRandomStockSheet(sFile, "Sheet1")
	if err != nil {
		t.Error(err.Error())
		return
	}

	if err = sFile.Save("Sheet1.xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestStockFile_NewSheet(t *testing.T) {
	sFile := stocksheet.New()
	if sFile == nil {
		t.Error("stocksheet.New() returned nil")
		return
	}

	sSheet, err := sFile.NewSheet("Page 1")
	if err != nil {
		t.Error(err)
		return
	}
	if sSheet == nil {
		t.Error("stocksheet.NewSheet() returned nil")
		return
	}

	err = fillRandomStockSheet(sFile, sSheet.Name())
	if err != nil {
		t.Error(err.Error())
		return
	}

	if err = sFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err = sFile.Save(sSheet.Name() + ".xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestStockFile_MultipleSheets(t *testing.T) {
	sFile := stocksheet.New()
	if sFile == nil {
		t.Error("stocksheet.New() returned nil")
		return
	}

	for i := range 3 {
		sheetName := fmt.Sprintf("Sheet %d", i)

		sSheet, err := sFile.NewSheet(sheetName)
		if err != nil {
			t.Error(err)
			return
		}
		if sSheet == nil {
			t.Error("stocksheet.NewSheet() returned nil")
			return
		}

		err = fillRandomStockSheet(sFile, sSheet.Name())
		if err != nil {
			t.Error(err.Error())
			return
		}
	}

	if err := sFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := sFile.Save("Multisheet.xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}
