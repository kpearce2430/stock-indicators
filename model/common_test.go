package model_test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
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

	//go:embed testdata/dividends.json
	testDividendsData []byte

	//go:embed testdata/trans_2023_1.csv
	testTrans20231 []byte
)

const (
	allTransactionsTable = "all_transactions"
	transactionTable     = "transactions"
	historicalTable      = "historical"
	stockCache           = "cache"
)

func truncateTransactions(pgxConn *pgxpool.Pool) error {
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s;", transactionTable)
	var count int
	if err := pgxConn.QueryRow(context.Background(), countSql).Scan(&count); err != nil {
		return err
	}
	logrus.Info("Found ", count, " Rows")
	if count == 0 {
		return nil
	}

	truncateSql := fmt.Sprintf("TRUNCATE %s;", transactionTable)
	if _, err := pgxConn.Exec(context.Background(), truncateSql); err != nil {
		return err
	}
	return nil
}

func connectToPostgres() (*pgxpool.Pool, error) {
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		return nil, err
	}

	if pgxConn == nil {
		return nil, errors.New("nil connection")
	}
	return pgxConn, nil
}

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
	os.Exit(m.Run())
}
