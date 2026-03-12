package monte_carlo_test

import (
	"testing"

	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/monte_carlo"
	"github.com/kpearce2430/stock-tools/stocksheet"
)

const monteCarloWorkSheetFileName = "MonteCarlo.xlsx"
const monteCarloWorkSheetName = "Monte Carol"

func TestMain(m *testing.M) {
	//ctx := context.Background()
	//postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	//defer func() {
	//	if err := postgresDBServer.Terminate(ctx); err != nil {
	//		log.Fatal(err.Error())
	//	}
	//}()
	m.Run()
}

func TestWorkSheet_MonteCarlo(t *testing.T) {
	//pgxConn, err := postgres.ConnectToPostgres()
	//if err != nil {
	//	t.Error(err.Error())
	//	return
	//}
	w := stocksheet.New()

	m := monte_carlo.NewMonteCarlo(w)

	if err := m.MonteCarlo(monteCarloWorkSheetName); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.Save(monteCarloWorkSheetFileName); err != nil {
		t.Error(err.Error())
		return
	}

}
