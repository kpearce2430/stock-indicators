package dividend_history_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/model/dividend_history"
)

func TestDividendEntry(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	de := dividend_history.NewDividendEntry("USAIX", pgxConn, 2023, 01)
	if err = de.GetYearMonth(context.Background()); err != nil {
		t.Error(err.Error())
		return
	}

	if err = de.ToDB(context.Background()); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(de.String())

	de2 := dividend_history.NewDividendEntry("USAIX", pgxConn, 2023, 01)
	if err = de2.FromDB(context.Background()); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(de2.String())

}

type DeTests struct {
	description string
	pgxConn     *pgxpool.Pool
	symbol      string
	year        int
	month       int
}

func TestDividendEntry_Errors(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	tests := []DeTests{
		{
			description: "month missing",
			pgxConn:     pgxConn,
			symbol:      "USAIX",
			year:        2023,
		},
		{
			description: "symbol missing",
			pgxConn:     pgxConn,
			year:        2023,
			month:       1,
		},
		{
			description: "year missing",
			pgxConn:     pgxConn,
			symbol:      "USAIX",
			month:       2023,
		},
		{
			description: "invalid data",
			pgxConn:     pgxConn,
			symbol:      "USAIX",
			month:       2023,
			year:        1,
		},
		{
			description: "invalid connection",
			symbol:      "USAIX",
			month:       1,
			year:        2023,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			de := dividend_history.NewDividendEntry(test.symbol, test.pgxConn, test.year, test.month)
			if err = de.GetYearMonth(context.Background()); err == nil {
				t.Error("Expected error")
				return
			}
			t.Log(err.Error())
		})
	}

}
