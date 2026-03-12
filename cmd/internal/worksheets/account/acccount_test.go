package account_test

import (
	"context"
	_ "embed"
	"log"

	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/account"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/sirupsen/logrus"

	"testing"
	"time"
)

var (
	//go:embed testdata/lookups.csv
	lookupCSV string

	//go:embed testdata/transactions-2026-03-07.csv
	transactions []byte

	ls *lookups.LookUpSet
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	ls = lookups.LoadLookupSet("1", lookupCSV)

	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		logrus.Fatal(err)
	}

	testSet := transaction.NewTransactionSet()
	if err := testSet.LoadToDB(pgxConn, ls, transaction.TransactionTable, transactions); err != nil {
		log.Fatal(err)
	}

	m.Run()
}

func TestAccountDividends(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	w := worksheets.New(pgxConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()

	w.Lookups = ls
	start := business_days.GetBusinessDay(time.Date(2026, 03, 8, 00, 00, 00, 00, time.UTC))

	a := account.New(w)

	err = a.AccountDividends("Account Dividends", start, 36)
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
