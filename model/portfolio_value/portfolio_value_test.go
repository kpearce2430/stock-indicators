package portfolio_value_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
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

func initPortfolioValuePostgres(t *testing.T) (*pgxpool.Pool, int) {
	t.Helper()
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return nil, -1
	}
	err = postgres.TruncateTable(pgxConn, portfolio_value.PortfolioValueTable)
	if err != nil {
		t.Error(err.Error())
		return nil, -1
	}

	count, err := portfolio_value.LoadDB(pgxConn, portfolio_value.PortfolioValueTable, string(testPortfolioValues), "", ls)
	if err != nil {
		t.Error(err.Error())
		return nil, -1
	}
	return pgxConn, count
}

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

	err = postgres.TruncateTable(pgxConn, portfolio_value.PortfolioValueTable)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	ls = lookups.LoadLookupSet("1", string(csvLookupData))
	m.Run()
}

func TestLoadPortfolioValues(t *testing.T) {
	_, count := initPortfolioValuePostgres(t)
	t.Log("Count:", count)
}

func TestLoadDBPortfolioValues(t *testing.T) {
	pgxConn, rc := initPortfolioValuePostgres(t)
	if pgxConn == nil {
		t.Error("pgxConn is nil")
		return
	}

	var count int
	sql := fmt.Sprintf("select count(*) from %s", portfolio_value.PortfolioValueTable)
	if err := pgxConn.QueryRow(context.Background(), sql).Scan(&count); err != nil {
		t.Fatal(err)
	}
	t.Log("Count:", count)
	if count != rc {
		t.Error("Counts don'hist_usaix.csv match")
		return
	}

	types, err := portfolio_value.GetTypes(pgxConn, portfolio_value.PortfolioValueTable)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	for k, v := range types {
		t.Run(k+":"+v, func(t *testing.T) {
			myType, err := portfolio_value.GetSymbolType(pgxConn, portfolio_value.PortfolioValueTable, k)
			if err != nil {
				t.Error(err.Error())
				return
			}
			if myType != v {
				t.Error(err.Error())
				return
			}
			var pv portfolio_value.PortfolioValueRecord
			if err = pv.GetLastDB(pgxConn, k, portfolio_value.PortfolioValueTable); err != nil {
				t.Error(err.Error())
				return
			}
			t.Log(pv)
		})
	}
}

func TestPortfolioValueSet_GetSymbolYearMonth(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = postgres.TruncateTable(pgxConn, portfolio_value.PortfolioValueTable)
	if err != nil {
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

type TestPV struct {
	Symbol string
	Year   int
	Month  int
}

func TestPortfolioValueSet_GetLastBefore(t *testing.T) {

	tests := []TestPV{
		{
			Symbol: "USAIX", Year: 2026, Month: 1,
		},
		{
			Symbol: "USAIX", Year: 2025, Month: 12,
		},
		{
			Symbol: "USAIX", Year: 2025, Month: 11,
		},
		{
			Symbol: "HD", Year: 2025, Month: 11,
		},
	}

	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = postgres.TruncateTable(pgxConn, portfolio_value.PortfolioValueTable)
	if err != nil {
		t.Error(err.Error())
		return
	}

	ctx := context.Background()
	err = postgres.LoadTableWithHeaders(ctx, pgxConn, portfolio_value.PortfolioValueTable, "./testdata/postgres_portfolio_value.csv")
	if err != nil {
		t.Error(err.Error())
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%d_%02d", test.Symbol, test.Year, test.Month), func(t *testing.T) {
			ps := portfolio_value.NewSet(pgxConn, portfolio_value.PortfolioValueTable, test.Symbol)
			if err = ps.GetLastBefore(test.Year, test.Month); err != nil {
				t.Error(err.Error())
				return
			}
			t.Log(ps)
		})
	}

}
