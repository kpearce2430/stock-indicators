package worksheets_test

import (
	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/xuri/excelize/v2"
	"testing"
	"time"
)

func TestAccountDividends(t *testing.T) {

	w := worksheets.NewWorkSheet(excelize.NewFile(), testApp.PGXConn)
	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	start := business_days.GetBusinessDay(time.Date(2024, 01, 15, 00, 00, 00, 00, time.UTC))
	err := w.AccountDividends("account-dividends", start, 36)
	if err != nil {
		t.Error(err)
	}
	if err := w.File.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.File.SaveAs("Accounts.xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}
