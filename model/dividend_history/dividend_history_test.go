package dividend_history_test

import (
	"context"
	_ "embed"
	"log"
	"testing"

	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/model/dividend_history"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/sirupsen/logrus"
)

var (
	//go:embed testdata/lookups.csv
	csvLookupData []byte

	//go:embed testdata/trans_2023_1.csv
	testTrans20231 []byte
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	pgxConn, err := postgres.ConnectToPostgres()
	ls := lookups.LoadLookupSet("1", string(csvLookupData))
	ts := transaction.NewTransactionSet()
	if err = ts.LoadToDB(pgxConn, ls, transaction.TransactionTable, testTrans20231); err != nil {
		logrus.Error(err.Error())
		logrus.Error("Failed to load test data")
		return
	}
	m.Run()
}

func TestDividendHistory_Sum(t *testing.T) {
	ctx := context.Background()
	toPostgres, err := postgres.ConnectToPostgres()
	if err != nil {
		return
	}

	dh := dividend_history.NewDividendHistory(toPostgres, "USAIX")

	err = dh.GetYear(ctx, 2023)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(dh.String())
	t.Log("Sum:", dh.Sum())
}

/*
func TestDividendHistoryFromDB(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	today := time.Now()
	if err != nil {
		t.Fatal(err)
		return
	}
	type DividendHistoryTest struct {
		description string
		symbol      string
		month       int
		year        int
		expectedErr error
	}

	tests := []DividendHistoryTest{
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
			dh, err := dividend_history.GetDividendEntryForYearMonth(context.Background(), pgxConn, test.symbol, test.year, test.month)
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



func TestDividendHistory_Sum(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err)
		return
	}
	err = postgres.TruncateTable(pgxConn, transaction.TransactionTable)
	if err != nil {
		t.Error(err)
		return
	}

	ls := model.LoadLookupSet("1", string(csvLookupData))
	ts := transaction.NewTransactionSet()
	if err = ts.LoadToDB(pgxConn, ls, transaction.TransactionTable, testTrans20231); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	dh := dividend_history.NewDividendHistory("USAIX")
	for i := 1; i <= 12; i++ {
		d, err := dividend_history.GetDividendEntryForYearMonth(context.Background(), pgxConn, "USAIX", 2023, i)
		if err != nil {
			t.Error(err.Error())
			return
		}
		dh.DividendEntries = append(dh.DividendEntries, d)
	}

	t.Log(dh.String())
	t.Log("Sum:", dh.Sum())

	dh2, err := dividend_history.DividendHistoryFromDB(context.Background(), pgxConn, "USAIX", 2023, 0)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(dh2.String())
	t.Log("Sum:", dh2.Sum())
}
*/
