package monte_carlo_test

import (
	"testing"

	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/monte_carlo"
)

const monteCarloWorkSheetFileName = "MonteCarlo.xlsx"
const monteCarloWorkSheetName = "Monte Carol"

func TestMain(m *testing.M) {
	m.Run()
}

func TestWorkSheet_MonteCarlo(t *testing.T) {
	w := worksheets.New(nil)
	m := monte_carlo.New(w)
	if err := m.MonteCarlo(monteCarloWorkSheetName); err != nil {
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
