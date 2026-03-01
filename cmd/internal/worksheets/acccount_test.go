package worksheets_test

import (
	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
	"time"
)

func TestAccountDividends(t *testing.T) {
	w := worksheets.New(testApp.PGXConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()

	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	start := business_days.GetBusinessDay(time.Date(2024, 01, 15, 00, 00, 00, 00, time.UTC))
	err := w.AccountDividends("Account Dividends", start, 36)
	if err != nil {
		t.Error(err)
	}

	if err = w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err = w.StockFile.Save("Accounts.xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}
