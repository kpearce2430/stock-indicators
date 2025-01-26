package worksheets_test

import (
	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/xuri/excelize/v2"
	"testing"
	"time"
)

const workSheetFileName = "DividendAnalysis.xlsx"
const workSheetName = "Dividend Analysis"

func TestWorkSheet_DividendAnalysis(t *testing.T) {
	w := worksheets.NewWorkSheet(excelize.NewFile(), testApp.PGXConn)
	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	// w.DividendCache = testApp.DividendCache
	w.StockCache = testApp.StockCache

	start := business_days.GetBusinessDay(time.Date(2023, 12, 31, 00, 00, 00, 00, time.UTC))
	if err := w.DividendAnalysis(workSheetName, start, 24); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.File.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.File.SaveAs(workSheetFileName); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestWorksheet_YearOverYearDividend(t *testing.T) {
	t.Skip("skipped")
	w := worksheets.NewWorkSheet(excelize.NewFile(), testApp.PGXConn)
	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	// w.DividendCache = testApp.DividendCache
	w.StockCache = testApp.StockCache
	if err := w.YearOverYearDividend(workSheetName, "TEST", 57, 2024, 6); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.File.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.File.SaveAs(workSheetFileName); err != nil {
		t.Error(err.Error())
		return
	}

}
