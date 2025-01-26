package model_test

import (
	"context"
	_ "embed"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
	"strings"
	"testing"
	"time"
)

const (
	fundHistory = "test_history"
	source      = "testdata"
	stockSymbol = "HD"
	fundSymbol  = "USAIX"
)

func TestSymbolInformationSet_MutualFund(t *testing.T) {
	pgxConn, err := connectToPostgres()
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if err = truncateTransactions(pgxConn); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	ls := model.LoadLookupSet("1", string(csvLookupData))
	if err := model.TransactionSetLoadToDB(pgxConn, ls, transactionTable, testTransactionsAll); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	ds := model.NewHistoricalDataSet(pgxConn, fundHistory)
	if err := ds.LoadSet(string(histUsaix), source, fundSymbol); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	for m := 1; m < 13; m++ {
		sd := model.NewSymbolDetail(fundHistory, fundSymbol, 2023, m)

		if err := sd.SetNumberOfShares(pgxConn); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		if err := sd.SetDividends(pgxConn); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		if err := sd.SetPrice(); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		t.Log(sd.String())
	}
}

func TestSymbolInformation_Stock(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}

	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	ls := model.LoadLookupSet("1", string(csvLookupData))
	if err := model.TransactionSetLoadToDB(pgxConn, ls, transactionTable, testTransactionsAll); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	for m := 1; m < 13; m++ {
		sd := model.NewSymbolDetail(fundHistory, stockSymbol, 2023, m)
		if err := sd.SetNumberOfShares(pgxConn); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		if err := sd.SetDividends(pgxConn); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		if err := sd.SetPrice(); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		t.Log(sd.String())
	}
}

func TestNewSymbolDetailSet(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}

	pgxConn, err := connectToPostgres()
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if err = truncateTransactions(pgxConn); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	ls := model.LoadLookupSet("1", string(csvLookupData))
	if err := model.TransactionSetLoadToDB(pgxConn, ls, transactionTable, testTransactionsAll); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	date := time.Date(2024, time.Month(1), 1, 00, 00, 00, 00, time.UTC)
	set := model.NewSymbolDetailSet(pgxConn, stockSymbol, fundHistory)
	if err := set.Create(date, 12); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	t.Log(set.String())
}
