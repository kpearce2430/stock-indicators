package stock_analysis_test

import (
	"context"
	_ "embed"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	businessdays "github.com/kpearce2430/keputils/business-days"
	couchdatabase "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/stock_analysis"
	mockclient "github.com/kpearce2430/stock-tools/mock-client"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/portfolio_value"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/massive-com/client-go/v2/rest/models"
	"github.com/sirupsen/logrus"
)

const (
	stockCacheDBName = "indicators"
)

var (
	//go:embed testdata/lookups.csv
	lookupCSV string

	//go:embed testdata/pv-2026-03-07.csv
	portfolioValueCSV string

	//go:embed testdata/transactions-2026-03-07.csv
	transactions []byte

	ls *lookups.LookUpSet

	cache *stock_cache.Cache[models.GetDailyOpenCloseAggResponse]
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresDBServer, err := postgres.StartPostgresTestServer(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	couchDBSerer, err := couchdatabase.StartCouchDBServer(ctx, stockCacheDBName)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer func() {
		_ = couchDBSerer.Terminate(ctx)
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

	portfolio_value.LoadDB(pgxConn, "portfolio_value", portfolioValueCSV, "2026094", ls)

	cdbURL, ok := os.LookupEnv("COUCHDB_URL")
	if !ok {
		logrus.Fatal("Error getting couchdb url")
	}
	user, ok := os.LookupEnv("COUCHDB_USER")
	if !ok {
		logrus.Fatal("Error getting couchdb user")
	}
	pswd, ok := os.LookupEnv("COUCHDB_PASSWORD")
	if !ok {
		logrus.Fatal("Error getting couchdb password")
	}

	// Go ahead and create the database
	databaseStore := couchdatabase.New[models.GetDailyOpenCloseAggResponse](stockCacheDBName, cdbURL, user, pswd)
	if databaseStore.DatabaseCreate() != true {
		logrus.Fatal("Error creating a database")
	}

	//
	config := couchdatabase.DatabaseConfig{
		DatabaseName: stockCacheDBName,
		CouchDBUrl:   cdbURL,
		Username:     user,
		Password:     pswd,
	}

	cache, err = stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&config, mockclient.New())
	if err != nil {
		logrus.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestWorkSheets_StockAnalysis(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		logrus.Error(err)
		return
	}

	w := worksheets.New(pgxConn)
	defer func() {
		_ = w.StockFile.CloseFile()
	}()

	s := stock_analysis.New(w)
	s.SetStockCache(cache)

	w.Lookups = ls

	jDate := utils.JulDateFromTime(businessdays.GetBusinessDay(time.Date(2026, 3, 31, 00, 00, 00, 00, time.UTC)))
	t.Log("jDate:", jDate)

	if err = s.StockAnalysis("Stock Analysis", jDate); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err := w.StockFile.Save("StockAnalysis.xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestWorkSheets_StockAnalysis_BuildSumList(t *testing.T) {
	var list = []string{"a", "b", "c"}
	s := stock_analysis.New(nil)
	result := s.BuildSumList(list, 10)
	if strings.Compare(result, "a10,b10,c10") != 0 {
		t.Error("BuildSumList failed")
		return
	}
	t.Log(result)
	t.Log(s.BuildSumList(list, 11))

	var list2 = []string{"a"}
	result2 := s.BuildSumList(list2, 12)
	if strings.Compare(result2, "a12") != 0 {
		t.Error("BuildSumList failed")
		return
	}
	t.Log(result2)

}
