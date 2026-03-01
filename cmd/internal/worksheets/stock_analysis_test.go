package worksheets_test

import (
	"testing"
	"time"

	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
)

func TestWorkSheets_StockAnalysis(t *testing.T) {
	//key := utils.GetEnv("POLYGON_API", "None")
	//if strings.Compare(key, "None") == 0 {
	//	t.Skip("No POLYGON_API key")
	//	return
	//}

	w := worksheets.New(testApp.PGXConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()

	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	w.StockCache = testApp.StockCache

	jDate := utils.JulDateFromTime(business_days.GetBusinessDay(time.Date(2023, 12, 31, 00, 00, 00, 00, time.UTC)))
	t.Log("jDate:", jDate)

	if err := w.StockAnalysis("Stock Analysis", jDate); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.Save("StockAnalysis.xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}
