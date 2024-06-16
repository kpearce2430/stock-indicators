package worksheets_test

import (
	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/xuri/excelize/v2"
	"testing"
	"time"
)

func TestWorkSheets_StockAnalysis(t *testing.T) {

	w := worksheets.NewWorkSheet(excelize.NewFile(), testApp.PGXConn)
	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	w.DividendCache = testApp.DividendCache
	w.StockCache = testApp.StockCache

	jDate := utils.JulDateFromTime(business_days.GetBusinessDay(time.Date(2023, 12, 31, 00, 00, 00, 00, time.UTC)))
	t.Log("jDate:", jDate)
	if err := w.StockAnalysis("Stock Analysis", jDate); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if err := w.File.DeleteSheet("Sheet1"); err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	if err := w.File.SaveAs("StockAnalysis.xlsx"); err != nil {
		t.Log(err)
		t.Fail()
	}
}
