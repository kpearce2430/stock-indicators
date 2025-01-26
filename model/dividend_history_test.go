package model_test

import (
	"context"
	"errors"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
	"time"
)

func TestDividendEntry_ToDB(t *testing.T) {
	pgxConn, err := connectToPostgres()
	if err != nil {
		t.Fatal(err)
		return
	}

	const (
		symbol = "AAPL"
		year   = 2024
		month  = 11
	)

	d := model.NewDividendEntry(symbol, year, month)
	d.Amount = 123.45

	if err := d.ToDB(context.Background(), pgxConn); err != nil {
		t.Error(err.Error())
		return
	}

	d2, err := model.DividendEntryFromDB(context.Background(), pgxConn, symbol, year, month)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if d2.Amount != 123.45 {
		t.Error("Expected 123.45, got ", d.Amount)
		return
	}
	d2.Amount = 223.45
	if err := d2.ToDB(context.Background(), pgxConn); err != nil {
		t.Error(err.Error())
		return
	}

	d3, err := model.DividendEntryFromDB(context.Background(), pgxConn, symbol, year, month)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if d3.Amount != 223.45 {
		t.Error("Expected 223.45, got ", d.Amount)
	}

	d4, err := model.DividendEntryFromDB(context.Background(), pgxConn, "JUNK", year, month)
	if err == nil {
		t.Error("Expected error found:", d4)
		return
	}
	t.Log(err)
}

func TestDividendHistoryFromDB(t *testing.T) {
	pgxConn, err := connectToPostgres()

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

func TestDividendHistory_Sum(t *testing.T) {
	pgxConn, err := connectToPostgres()
	if err != nil {
		t.Error(err)
		return
	}
	err = truncateTransactions(pgxConn)
	if err != nil {
		t.Error(err)
		return
	}

	ls := model.LoadLookupSet("1", string(csvLookupData))

	if err = model.TransactionSetLoadToDB(pgxConn, ls, transactionTable, testTrans20231); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	dh := model.NewDividendHistory("USAIX")
	for i := 1; i <= 12; i++ {
		d, err := model.GetDividendEntryForYearMonth(pgxConn, "USAIX", 2023, i)
		if err != nil {
			t.Error(err.Error())
			return
		}
		dh.DividendEntries = append(dh.DividendEntries, d)
	}

	t.Log(dh.String())
	t.Log("Sum:", dh.Sum())

	dh2, err := model.DividendHistoryFromDB(context.Background(), pgxConn, "USAIX", 2023, 0)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(dh2.String())
	t.Log("Sum:", dh2.Sum())
}
