package dividend_analysis_test

import (
	"context"
	_ "embed"
	"log"

	businessdays "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/dividend_analysis"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/sirupsen/logrus"

	"testing"
	"time"
)

const workSheetFileName = "DividendAnalysis.xlsx"
const workSheetName = "Dividend Analysis"

var (
	//go:embed testdata/lookups.csv
	lookupCSV string

	//go:embed testdata/transactions-2026-03-07.csv
	transactions []byte

	ls *lookups.LookUpSet
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	ls = lookups.LoadLookupSet("1", lookupCSV)

	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		logrus.Fatal(err)
	}

	testSet := transaction.NewTransactionSet()
	if err := testSet.LoadToDB(pgxConn, ls, transaction.TransactionTable, transactions); err != nil {
		log.Fatal(err)
	}

	m.Run()
}

/*
func TestWorkSheet_DividendAnalysis(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	w := worksheets.New(pgxConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()

	w.Lookups = ls
	start := businessdays.GetBusinessDay(time.Date(2026, 03, 8, 00, 00, 00, 00, time.UTC))

	d := dividend_analysis.New(w)

	if err = d.DividendAnalysis(context.Background(),workSheetName, start, 24); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.Save(workSheetFileName); err != nil {
		t.Error(err.Error())
		return
	}
}



func TestWorksheet_YearOverYearDividend(t *testing.T) {
	w := worksheets.New(testApp.PGXConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()
	w.Lookups = lookups.LoadLookupSet("1", string(lookups2))
	// w.DividendCache = testApp.DividendCache
	w.StockCache = testApp.StockCache
	if err := w.YearOverYearDividend(workSheetName, "TEST", 57, 2024, 6); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.Save(workSheetFileName); err != nil {
		t.Error(err.Error())
		return
	}
}
*/

func TestWorkSheet_DividendSheets(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	w := worksheets.New(pgxConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()

	w.Lookups = ls
	start := businessdays.GetBusinessDay(time.Date(2026, 03, 8, 00, 00, 00, 00, time.UTC))

	d := dividend_analysis.New(w)
	if err = d.DividendSheets(context.Background(), workSheetName, start, 24); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.Save("DividedSheet.xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}
