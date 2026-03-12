package model_test

/*
import (
	_ "embed"
)

var (
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

	//go:embed transaction/testdata/transactions-2023-12-09.csv
	testTransactions3 []byte

	//go:embed transaction/testdata/transactions-2023-12-16.csv
	testTransactions4 []byte

	//go:embed testdata/usaix_hist.csv
	testHistoricalData []byte

	////go:embed dividends/testdata/dividends.json
	//testDividendsData []byte

	//go:embed testdata/trans_2023_1.csv
	testTrans20231 []byte
)

const (
	historicalTable = "historical"
	stockCache      = "cache"
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

	// dbs := []string{historicalTable, "pv", symbol_details.fundHistory, "cache"}
	dbs := []string{historicalTable, "pv", "cache"}
	for _, db := range dbs {
		databaseStore := couch_database.New[portfolio_value.PortfolioValueDatabaseRecord](db, url, "admin", "password")
		if databaseStore.DatabaseCreate() != true {
			logrus.Fatal("Error creating a database")
		}
	}

	_ = os.Setenv("COUCHDB_URL", url)
	_ = os.Setenv("COUCHDB_USER", "admin")
	_ = os.Setenv("COUCHDB_PASSWORD", "password")
	_ = os.Setenv("CACHE_COUCHDB_DATABASE", stockCache)

	os.Exit(m.Run())
}

*/
