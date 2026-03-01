package worksheets_test

import (
	"context"
	businessdays "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
	"time"
)

const workSheetFileName = "DividendAnalysis.xlsx"
const workSheetName = "Dividend Analysis"

func TestWorkSheet_DividendAnalysis(t *testing.T) {
	w := worksheets.New(testApp.PGXConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()
	w.Lookups = model.LoadLookupSet("1", string(lookups2))

	w.StockCache = testApp.StockCache

	start := businessdays.GetBusinessDay(time.Date(2022, 10, 01, 00, 00, 00, 00, time.UTC))
	if err := w.DividendAnalysis(workSheetName, start, 24); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.Save(workSheetFileName); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestWorksheet_YearOverYearDividend(t *testing.T) {
	w := worksheets.New(testApp.PGXConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()
	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	// w.DividendCache = testApp.DividendCache
	w.StockCache = testApp.StockCache
	if err := w.YearOverYearDividend(workSheetName, "TEST", 57, 2024, 6); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.Save(workSheetFileName); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestWorkSheet_DividendSheets(t *testing.T) {
	w := worksheets.New(testApp.PGXConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()
	w.Lookups = model.LoadLookupSet("1", string(lookups2))

	w.StockCache = testApp.StockCache

	start := businessdays.GetBusinessDay(time.Date(2022, 10, 02, 00, 00, 00, 00, time.UTC))
	if err := w.DividendSheets(context.Background(), workSheetName, start, 24); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.Save("DividedSheet.xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}
