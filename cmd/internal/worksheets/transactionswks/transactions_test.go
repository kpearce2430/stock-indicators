package transactionswks_test

import (
	"context"
	_ "embed"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/transactionswks"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/sirupsen/logrus"

	"testing"
)

const transactionWorkSheetFileName = "Transactions.xlsx"
const transactionWorkSheetName = "Transactions"

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

func TestWorkSheet_Transactions(t *testing.T) {
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	w := worksheets.New(pgxConn)
	w.Lookups = lookups.LoadLookupSet("1", lookupCSV)

	tr := transactionswks.New(w)

	if err := tr.Transactions(transactionWorkSheetName, utils.JulDate()); err != nil {
		t.Error(err.Error())
		return
	}

	if err = w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err = w.StockFile.Save(transactionWorkSheetFileName); err != nil {
		t.Error(err.Error())
		return
	}
}
