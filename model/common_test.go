package model_test

import (
	"context"
	_ "embed"
	"fmt"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/stock-tools/postgres"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"testing"
)

var (
	//go:embed testdata/aapl.csv
	applTransactions []byte

	//go:embed testdata/bond_data.csv
	bondTransactions []byte

	//go:embed testdata/hist_usaix.csv
	histUsaix []byte

	//go:embed testdata/msft.csv
	msftTransactions []byte

	//go:embed testdata/usaix.csv
	usaixTransactions []byte

	//go:embed testdata/transactions.csv
	testTransactionsAll []byte

	//go:embed testdata/portfolio_value.csv
	testPortfolioValues []byte

	//go:embed testdata/transactions-2023-12-09.csv
	testTransactions3 []byte

	//go:embed testdata/transactions-2023-12-16.csv
	testTransactions4 []byte

	//go:embed testdata/usaix_hist.csv
	testHistoricalData []byte
)

const (
	allTransactionsTable = "all_transactions"
	transactionTable     = "transactions"
	historicalTable      = "historical"
	stockCache           = "cache"
)

func TestMain(m *testing.M) {

	ctx := context.Background()
	postgresDBServer, _ := postgres.CreatePostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

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

	couchDBServer, _ := couch_database.CreateCouchDBServer(ctx)
	defer func() {
		_ = couchDBServer.Terminate(ctx)
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
	_ = os.Setenv("CACHE_COUCHDB_DATABASE", stockCache)
	_ = os.Setenv("POLYGON_API", "YVaauGHjGDYf8W_sQLMejJ3W15Y1aiV1") // TODO: Read from environment

	os.Exit(m.Run())
}
