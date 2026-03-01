package worksheets_test

import (
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"testing"
)

const monteCarloWorkSheetFileName = "MonteCarlo.xlsx"
const monteCarloWorkSheetName = "Monte Carol"

func TestWorkSheet_MonteCarlo(t *testing.T) {
	w := worksheets.New(testApp.PGXConn)

	if err := w.MonteCarlo(monteCarloWorkSheetName); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.Save(monteCarloWorkSheetFileName); err != nil {
		t.Error(err.Error())
		return
	}
}
