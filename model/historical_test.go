package model_test

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
	"time"
)

func TestHistorical_LoadHistorical(t *testing.T) {
	if err := model.HistoricalSetLoadCouch(historicalTable, string(testHistoricalData), "Random", "USAIX"); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	type HistoryTest struct {
		Key   string
		Found bool
	}

	tests := []HistoryTest{
		{
			Key:   "2024001:USAIX",
			Found: false,
		},
		{
			Key:   "2024002:USAIX",
			Found: true,
		},
		{
			Key:   "2024003:USAIX",
			Found: true,
		},
		{
			Key:   "2024004:USAIX",
			Found: true,
		},
		{
			Key:   "2024099:USAIX",
			Found: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Key, func(t *testing.T) {
			h, err := model.HistoricalCacheGet(historicalTable, tc.Key)
			if err != nil {
				t.Log(err.Error())
				if tc.Found == true {
					t.Fail()
				}
				return
			}
			if h != nil {
				t.Log(h)
			}
		})
	}

	for _, tc := range tests {
		t.Run(tc.Key, func(t *testing.T) {
			rev, err := model.HistoricalCacheDelete(historicalTable, tc.Key)
			if err != nil {
				t.Log(err.Error())
				if tc.Found == true {
					t.Fail()
				}
				return
			}
			t.Log(tc.Key, rev, " deleted")
		})
	}
}

func TestHistorical_LoadHistoricalDB(t *testing.T) {
	const symbol = "USAIX"
	const source = "testcases"
	const fundHistory = "fund_history"

	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	ds := model.NewHistoricalDataSet(pgxConn, fundHistory)
	if err := ds.LoadSet(string(testHistoricalData), source, symbol); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	feb1 := time.Date(2024, 02, 01, 00, 00, 00, 00, time.UTC)
	hist, err := ds.Last(symbol, feb1)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	t.Log("last>>>", hist)
}
