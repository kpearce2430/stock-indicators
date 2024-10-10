package model_test

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
)

const portfolioValueTable = "portfolio_value"

func TestLoadPortfolioValuesError(t *testing.T) {
	// t.Parallel()
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if err := model.LoadPortfolioValues(pgxConn, "pv", "blah", "2023123", nil); err != nil {
		t.Log(err.Error())
		return
	}
	t.Log("SHOULD HAVE FAILED")
	t.Fail()
}

func TestLoadPortfolioValues(t *testing.T) {
	// t.Parallel()
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	ls := model.LoadLookupSet("1", string(csvLookupData))
	if err := model.LoadPortfolioValues(pgxConn, "pv", string(testPortfolioValues), "", ls); err != nil {
		t.Log(err.Error())
		t.Fail()
	}
}

func TestLoadPortfolioValuesWithJulianDate(t *testing.T) {
	// t.Parallel()
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	ls := model.LoadLookupSet("1", string(csvLookupData))
	if err := model.LoadPortfolioValues(pgxConn, "pv", string(testPortfolioValues), "2023362", ls); err != nil {
		t.Log(err.Error())
		t.Fail()
	}
}

func TestLoadDBPortfolioValues(t *testing.T) {
	// t.Parallel()
	ls := model.LoadLookupSet("1", string(csvLookupData))

	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	rc, err := model.PortfolioValuesLoadDB(pgxConn, portfolioValueTable, string(testPortfolioValues), "", ls)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	var count int
	sql := fmt.Sprintf("select count(*) from %s", portfolioValueTable)
	if err := pgxConn.QueryRow(context.Background(), sql).Scan(&count); err != nil {
		t.Fatal(err)
	}
	t.Log("Count:", count)
	if count != rc { // TODO: Get the number actually loaded - len(testSet.TransactionRows) {
		t.Log("Counts don'hist_usaix.csv match")
		t.Fail()
	}

	types, err := model.PortfolioValueGetTypes(pgxConn, portfolioValueTable)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	for k, v := range types {
		t.Log(k, ":", v)
		myType, err := model.PortfolioValueGetSymbolType(pgxConn, portfolioValueTable, k)
		if err != nil {
			t.Log(err.Error())
			t.FailNow()
		}
		if myType != v {
			t.Log("Types for ", k, " do not match ", v, "/", myType)
			t.FailNow()
		}
	}

	var pv model.PortfolioValueRecord
	if err := pv.GetLastDB(pgxConn, "HD", "portfolio_value"); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(pv)
}
