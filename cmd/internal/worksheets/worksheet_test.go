package worksheets_test

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	business_days "github.com/kpearce2430/keputils/business-days"
	couchdatabase "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/app"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/kpearce2430/stock-tools/postgres"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"testing"
	"time"
)

//go:embed testdata/short-trans.csv
var transactionData []byte

//go:embed testdata/hd-trans.csv
var hdTransData []byte

//go:embed testdata/lookup2.csv
var lookups2 []byte

//go:embed testdata/portfolio_value.csv
var testPortfolioValues string

//go:embed testdata/pv-2023-10-14.csv
var testPV2023Oct14 string

//go:embed testdata/transactions-2024-02-10.csv
var transactions20240210 []byte

//go:embed testdata/pv-2024-02-10.csv
var portfolioValue20240210 string

// var lookups *model.LookUpSet

const (
	stockCacheDBName      = "indicators"
	portfolioDatabaseName = "portfolio_value"
	transactionTable      = "transactions"
)

var testApp *app.App

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Set up the containers.
	couchDBServer, _ := couchdatabase.CreateCouchDBServer(ctx)
	defer func() {
		_ = couchDBServer.Terminate(ctx)
	}()

	postgresDBServer, _ := postgres.CreatePostgresTestServer(ctx)
	defer func() {
		_ = postgresDBServer.Terminate(ctx)
	}()

	// Set up the couchdb environment
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
	_ = os.Setenv("PV_COUCHDB_DATABASE", portfolioDatabaseName)
	_ = os.Setenv("CACHE_COUCH_DATABASE", stockCacheDBName)

	// Set up the postgres environment
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
	pgxConn, err := pgxpool.New(context.Background(), pgURL)
	if err != nil {
		log.Fatal(err)
	}

	lookups := model.LoadLookupSet("1", string(lookups2))

	testSet := model.NewTransactionSet()
	if err := testSet.LoadWithLookups(lookups, transactions20240210); err != nil {
		log.Fatal(err)
	}

	for _, tr := range testSet.TransactionRows {
		if err := tr.TransactionToDB(context.Background(), pgxConn, transactionTable); err != nil {
			log.Fatal(err)
		}
	}

	julDate := utils.JulDateFromTime(business_days.GetBusinessDay(time.Date(2024, 02, 10, 00, 00, 00, 00, time.UTC)))
	logrus.Info("julDate:", julDate)
	// Load Portfolio Value
	if err := model.LoadPortfolioValues(pgxConn, portfolioDatabaseName, portfolioValue20240210, julDate, lookups); err != nil {
		log.Fatal(err.Error())
	}

	var count int
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s;", transactionTable)
	if err := pgxConn.QueryRow(ctx, countSql).Scan(&count); err != nil {
		log.Fatal(err.Error())
	}

	testApp = app.NewApp("8888")
	_, err = testApp.DividendCache.DatabaseExists()
	if err != nil {
		if testApp.DividendCache.DatabaseCreate() == false {
			log.Fatal("unable to create dividend cache couch db")
		}
	}

	_, err = testApp.StockCache.DatabaseExists()
	if err != nil {
		if testApp.StockCache.DatabaseCreate() == false {
			log.Fatal("unable to create dividend cache couch db")
		}
	}

	logrus.Info("Starting tests: ", count, " transactions loaded")
	m.Run()
}
