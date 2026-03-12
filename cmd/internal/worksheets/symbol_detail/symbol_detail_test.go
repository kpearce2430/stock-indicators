package symbol_detail_test

import (
	"context"
	_ "embed"
	"log"
	"time"

	business_days "github.com/kpearce2430/keputils/business-days"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/symbol_detail"
	mock_client "github.com/kpearce2430/stock-tools/mock-client"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/massive-com/client-go/v2/rest/models"
	"github.com/sirupsen/logrus"

	"testing"

	// "github.com/kpearce2430/stock-tools/cmd/internal/app"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
)

const (
	stockcache                     = "quotes"
	fundHistory                    = "test_history"
	symbolDetailsWorkSheetFileName = "SymbolsDetails.xlsx"
)

var (
	//go:embed testdata/lookups.csv
	lookupCSV string

	//go:embed testdata/transactions-2026-03-07.csv
	transactions []byte

	ls *lookups.LookUpSet
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	logrus.SetReportCaller(true)
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

func TestWorkSheet_SymbolsDetails(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	w := worksheets.New(pgxConn)

	quoteConfig := couch_database.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", stockcache),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}

	w.StockCache, err = stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&quoteConfig, mock_client.New())
	if !w.StockCache.CouchDBUp() {
		t.Fatal("couchdb not up")
		return
	}

	_, err = w.StockCache.DatabaseExists()
	if err != nil {
		if w.StockCache.DatabaseCreate() == false {
			t.Fatal("unable to create cache database")
			return
		}
	}

	_, err = w.StockCache.DatabaseExists()

	if err != nil {
		t.Fatal("unable to create cache database")
	}

	sd := symbol_detail.New(w)

	start := business_days.GetBusinessDay(time.Date(2023, 12, 31, 00, 00, 00, 00, time.UTC))

	if err = sd.SymbolsDetails("TickerInfo", "HD", fundHistory, start, 24); err != nil {
		t.Error(err)
		return
	}

	if err = w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Log(err.Error())
	}

	if err = w.StockFile.Save(symbolDetailsWorkSheetFileName); err != nil {
		t.Error(err)
		return
	}
	t.Log("completed: ", time.Now().Sub(start))
}
