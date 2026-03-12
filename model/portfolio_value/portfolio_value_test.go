package portfolio_value_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/portfolio_value"
	"github.com/sirupsen/logrus"
)

var (
	ls *lookups.LookUpSet

	//go:embed testdata/lookups.csv
	csvLookupData []byte

	//go:embed testdata/portfolio_value.csv
	testPortfolioValues []byte
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			logrus.Fatal(err.Error())
		}
	}()

	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		logrus.Fatal(err.Error())
		return
	}

	truncateSql := fmt.Sprintf("TRUNCATE %s;", portfolio_value.PortfolioValueTable)
	if _, err = pgxConn.Exec(context.Background(), truncateSql); err != nil {
	}

	ls = lookups.LoadLookupSet("1", string(csvLookupData))
	m.Run()
}

func TestLoadPortfolioValues(t *testing.T) {
	// t.Parallel()
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	count, err := portfolio_value.LoadDB(pgxConn, portfolio_value.PortfolioValueTable, string(testPortfolioValues), "", ls)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	t.Log("Count:", count)
}

func TestLoadDBPortfolioValues(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	truncateSql := fmt.Sprintf("TRUNCATE %s;", portfolio_value.PortfolioValueTable)
	if _, err = pgxConn.Exec(context.Background(), truncateSql); err != nil {
		t.Error(err.Error())
		return
	}

	rc, err := portfolio_value.LoadDB(pgxConn, portfolio_value.PortfolioValueTable, string(testPortfolioValues), "", ls)
	if err != nil {
		t.Error(err.Error())
		return
	}

	var count int
	sql := fmt.Sprintf("select count(*) from %s", portfolio_value.PortfolioValueTable)
	if err := pgxConn.QueryRow(context.Background(), sql).Scan(&count); err != nil {
		t.Fatal(err)
	}
	t.Log("Count:", count)
	if count != rc { // TODO: Get the number actually loaded - len(testSet.TransactionRows) {
		t.Error("Counts don'hist_usaix.csv match")
		return
	}

	types, err := portfolio_value.GetTypes(pgxConn, portfolio_value.PortfolioValueTable)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	for k, v := range types {
		t.Log(k, ":", v)
		myType, err := portfolio_value.GetSymbolType(pgxConn, portfolio_value.PortfolioValueTable, k)
		if err != nil {
			t.Log(err.Error())
			t.FailNow()
		}
		if myType != v {
			t.Log("Types for ", k, " do not match ", v, "/", myType)
			t.FailNow()
		}
	}

	var pv portfolio_value.PortfolioValueRecord
	if err = pv.GetLastDB(pgxConn, "HD", "portfolio_value"); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(pv)
}

func TestPortfolioValueSet_GetSymbolYearMonth(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}
	truncateSql := fmt.Sprintf("TRUNCATE %s;", portfolio_value.PortfolioValueTable)
	if _, err = pgxConn.Exec(context.Background(), truncateSql); err != nil {
		t.Error(err.Error())
		return
	}

	_, err = portfolio_value.LoadDB(pgxConn, portfolio_value.PortfolioValueTable, string(testPortfolioValues), "", ls)
	if err != nil {
		t.Error(err.Error())
		return
	}

	pSet := portfolio_value.NewSet(pgxConn, portfolio_value.PortfolioValueTable, "")
	err = pSet.GetSymbolYearMonth(2023, 10)
	if err != nil {
		t.Error(err.Error())
		return
	}

	data, err := json.MarshalIndent(pSet, "", "  ")
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(string(data))

}
