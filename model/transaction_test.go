package model_test

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
)

func TestTransactionSet_Load(t *testing.T) {
	t.Log("in TestTransactionSet_Load")
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	ls := model.LoadLookupSet("1", string(csvLookupData))

	if err := model.TransactionSetLoadToDB(pgxConn, ls, allTransactionsTable, testTransactions3); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if err := model.TransactionSetLoadToDB(pgxConn, ls, allTransactionsTable, testTransactions4); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestTransactionFullLoad(t *testing.T) {
	ctx := context.Background()
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	testSet := model.NewTransactionSet()
	if err := testSet.Load(testTransactionsAll); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	ls := model.LoadLookupSet("1", string(csvLookupData))

	if err := model.TransactionSetLoadToDB(pgxConn, ls, transactionTable, testTransactionsAll); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	var count int
	sql := fmt.Sprintf("select count(*) from %s", transactionTable)
	if err := pgxConn.QueryRow(ctx, sql).Scan(&count); err != nil {
		t.Fatal(err)
	}
	t.Log("Count:", count)
	if count != 4995 { // TODO: Get the number actually loaded - len(testSet.TransactionRows) {
		t.Log("Counts don'hist_usaix.csv match")
		t.Fail()
	}
}
