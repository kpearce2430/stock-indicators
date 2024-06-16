package app_test

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/app"
	"github.com/kpearce2430/stock-tools/model"
	polygonclient "github.com/kpearce2430/stock-tools/polygon-client"
	"github.com/kpearce2430/stock-tools/postgres"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/polygon-io/client-go/rest/models"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"testing"
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
		LookupSet: model.LoadLookupSet("1", string(csvLookupData)),
	}

	a.PGXConn, err = pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	if err := model.TransactionSetLoadToDB(a.PGXConn, a.LookupSet, app.TransactionTable, testTransactions); err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	if err := model.LoadPortfolioValues(app.PortfolioValueDB, string(csvPortfolioValueData), utils.JulDate(), a.LookupSet); err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	divConfig := couch_database.DatabaseConfig{
		DatabaseName: utils.GetEnv("DIV_COUCHDB_DATABASE", "dividends"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	a.DividendCache, err = stock_cache.NewCache[models.Dividend](&divConfig, polygonclient.NewPolygonClient(""))
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

	cdbConfig := couch_database.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "cache"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	a.StockCache, err = stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&cdbConfig, polygonclient.NewPolygonClient(""))
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
	couchDBServer, _ := couch_database.CreateCouchDBServer(ctx)
	defer func() {
		_ = couchDBServer.Terminate(ctx)
	}()

	postgresDBServer, _ := postgres.CreatePostgresTestServer(ctx)
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
	_ = os.Setenv("POLYGON_API", "YVaauGHjGDYf8W_sQLMejJ3W15Y1aiV1") // TODO: Read from environment

	pgIP, err := postgresDBServer.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	pgMappedPort, err := postgresDBServer.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatal(err)
	}

	// postgres://postgres:postgres@localhost:5432/postgres
	pgURL := fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres", pgIP, pgMappedPort.Port())
	_ = os.Setenv("PG_DATABASE_URL", pgURL)

	testApp, err = createTestApp()
	if err != nil {
		log.Fatal(err)
	}

	logrus.Debug("Starting tests")
	m.Run()
}

func TestNewApp(t *testing.T) {
	t.Parallel()
	status, err := testApp.PostgresCheck()
	switch {
	case err != nil:
		t.Log(err.Error())
		t.FailNow()
		return
	case status == false:
		t.Log("Postgres:", status)
		t.FailNow()
		return
	}

	status = testApp.CouchDBCheck()
	if status != true {
		t.Log("CouchDB:", status)
		t.FailNow()
	}
}
