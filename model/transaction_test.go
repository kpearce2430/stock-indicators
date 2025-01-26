package model_test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
	"time"
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
	pgxConn, err := connectToPostgres()
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if err := truncateTransactions(pgxConn); err != nil {
		t.Error(err.Error())
		return
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
	if err := pgxConn.QueryRow(context.TODO(), sql).Scan(&count); err != nil {
		t.Fatal(err)
	}
	t.Log("Count:", count)
	if count != 4995 { // TODO: Get the number actually loaded - len(testSet.TransactionRows) {
		t.Log("Counts don'hist_usaix.csv match")
		t.Fail()
	}
}

func TestTransactionSet_GetTransactions(t *testing.T) {
	pgxConn, err := connectToPostgres()
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if err := truncateTransactions(pgxConn); err != nil {
		t.Error(err.Error())
		return
	}

	testSet := model.NewTransactionSet()
	if err := testSet.Load(testTrans20231); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	today := time.Now()
	if err != nil {
		t.Fatal(err)
		return
	}
	type TestSet struct {
		description string
		symbol      string
		month       int
		year        int
		expectedErr error
	}

	tests := []TestSet{
		{
			description: "Happy path",
			symbol:      "AAPL",
			month:       1,
			year:        2020,
		},
		{
			description: "Invalid Parameters",
			symbol:      "",
			month:       -1,
			year:        1776,
			expectedErr: errors.New("invalid arguments"),
		},
		{
			description: "No Symbol",
			symbol:      "",
			month:       int(today.Month()),
			year:        today.Year(),
		},
		{
			description: "No Year",
			symbol:      "AAPL",
			month:       int(today.Month()),
		},
		{
			description: "No Month",
			symbol:      "AAPL",
			year:        today.Year(),
		},
		{
			description: "Invalid Month",
			symbol:      "AAPL",
			year:        today.Year(),
			month:       -1,
			expectedErr: errors.New("invalid month"),
		},
		{
			description: "Invalid Year",
			symbol:      "AAPL",
			year:        today.Year() + 1,
			month:       int(today.Month()),
			expectedErr: errors.New("invalid year"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			dh, err := model.DividendHistoryFromDB(context.Background(), pgxConn, test.symbol, test.year, test.month)
			if test.expectedErr != nil {
				if err.Error() != test.expectedErr.Error() {
					t.Error("Expected", test.expectedErr, "got", err)
					return
				}
				return
			}
			if err != nil {
				t.Error(err.Error())
				return
			}
			if dh == nil {
				t.Error("Expected dividendHistoryFromDB to return a dividend history")
				return
			}
		})
	}
}
