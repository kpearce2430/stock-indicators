package transaction_test

import (
	"context"
	_ "embed"
	"log"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/transaction"
)

var (
	//go:embed testdata/lookups.csv
	csvLookupData []byte

	//go:embed testdata/transactions-2023-12-09.csv
	testTransactions3 []byte

	//go:embed testdata/transactions-2023-12-16.csv
	testTransactions4 []byte
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	m.Run()
}

func TestTransaction(t *testing.T) {
	t.Log("test")
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err)
		return
	}

	if pgxConn == nil {
		t.Error("pgxConn is nil")
		return
	}
}

// TestTransactionSet_Load tests loading two sets of transactions created a week apart into the database.
func TestTransactionSet_Load(t *testing.T) {
	ctx := context.Background()
	pgxConn, err := pgxpool.New(ctx, utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	ls := lookups.LoadLookupSet("1", string(csvLookupData))

	sets := [][]byte{testTransactions3, testTransactions4}

	for _, set := range sets {
		ts := transaction.NewTransactionSet()
		if err = ts.LoadToDB(pgxConn, ls, transaction.TransactionTable, set); err != nil {
			t.Error(err.Error())
			return
		}
	}

	newSet := transaction.NewTransactionSet()
	err = newSet.GetAll(ctx, pgxConn)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("Total Transactions", len(newSet.TransactionRows))
}
