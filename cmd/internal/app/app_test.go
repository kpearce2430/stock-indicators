package app_test

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	couchdatabase "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/app"
	massiveclient "github.com/kpearce2430/stock-tools/massive-client"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/portfolio_value"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/massive-com/client-go/v2/rest/models"
	"github.com/sirupsen/logrus"
)

//go:embed testdata/lookups.csv
var csvLookupData []byte

//go:embed testdata/portfolio_value.csv
var csvPortfolioValueData []byte

//go:embed testdata/usaix_hist.csv
var csvHistoricalDAta []byte

func createTestApp() (*app.App, error) {
	var err error
	a := app.App{
		Srv:       nil,
		LookupSet: lookups.LoadLookupSet("1", string(csvLookupData)),
	}

	a.PGXConn, err = pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	ts := transaction.NewTransactionSet()
	err = ts.LoadToDB(a.PGXConn, a.LookupSet, transaction.TransactionTable, testTransactions)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	_, err = portfolio_value.LoadDB(a.PGXConn, app.PortfolioValueDB, string(csvPortfolioValueData), utils.JulDate(), a.LookupSet)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	divConfig := couchdatabase.DatabaseConfig{
		DatabaseName: utils.GetEnv("DIV_COUCHDB_DATABASE", "dividends"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	a.DividendCache, err = stock_cache.NewCache[models.Dividend](&divConfig, massiveclient.New())
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	if _, err = a.DividendCache.DatabaseExists(); err != nil {
		if a.DividendCache.DatabaseCreate() == false {
			logrus.Error(err.Error())
			return nil, err
		}
	}

	cdbConfig := couchdatabase.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "cache"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	a.StockCache, err = stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&cdbConfig, massiveclient.New())
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	if _, err = a.StockCache.DatabaseExists(); err != nil {
		if a.StockCache.DatabaseCreate() == false {
			logrus.Error(err.Error())
			return nil, err
		}
	}

	return &a, nil
}

var testApp *app.App

// TestMain
func TestMain(m *testing.M) {
	ctx := context.Background()
	couchDBServer, _ := couchdatabase.CreateCouchDBServer(ctx)
	defer func() {
		_ = couchDBServer.Terminate(ctx)
	}()

	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		_ = postgresDBServer.Terminate(ctx)
	}()

	cdbIP, err := couchDBServer.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cdbMappedPort, err := couchDBServer.MappedPort(ctx, "5984")
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprintf("http://%s:%s", cdbIP, cdbMappedPort.Port())

	logrus.Debugln(url)

	_ = os.Setenv("COUCHDB_URL", url)
	_ = os.Setenv("COUCHDB_USER", "admin")
	_ = os.Setenv("COUCHDB_PASSWORD", "password")
	_ = os.Setenv("COUCHDB_DATABASE", "pv")

	// Note since I'm not reading or writing to the database via the cache, The model isn't relevant.
	for _, db := range []string{"dividends", app.PortfolioValueDB, "cache", "something"} {
		databaseStore := couchdatabase.New[portfolio_value.PortfolioValueDatabaseRecord](db, url, "admin", "password")
		if databaseStore.DatabaseCreate() != true {
			logrus.Fatal("Error creating a database")
		}
	}

	testApp, err = createTestApp()
	if err != nil {
		log.Fatal(err)
	}

	logrus.Info("Starting tests")
	m.Run()
}

func TestNewApp(t *testing.T) {
	t.Parallel()
	status, err := testApp.PostgresCheck()
	switch {
	case err != nil:
		t.Fatal(err.Error())
		return
	case status == false:
		t.Log("Postgres:", status)
		t.FailNow()
		return
	}

	status = testApp.CouchDBCheck()
	if status != true {
		t.Fatal("CouchDB:", status)
	}
}
