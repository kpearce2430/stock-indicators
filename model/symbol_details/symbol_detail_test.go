package symbol_details_test

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"testing"

	couchdatabase "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/portfolio_value"
	"github.com/sirupsen/logrus"
)

const (
	fundHistory = "test_history"
	source      = "testdata"
	stockSymbol = "HD"
	fundSymbol  = "USAIX"
	stockCache  = "cache"
)

var (
	//go:embed testdata/lookups.csv
	csvLookupData []byte

	//go:embed testdata/transactions.csv
	testTransactionsAll []byte

	//go:embed testdata/hist_usaix.csv
	histUsaix []byte

	lookupSet *lookups.LookUpSet
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	lookupSet = lookups.LoadLookupSet("1", string(csvLookupData))

	couchDBServer, _ := couchdatabase.CreateCouchDBServer(ctx)
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

	databaseStore := couchdatabase.New[portfolio_value.PortfolioValueDatabaseRecord](stockCache, url, "admin", "password")
	if databaseStore.DatabaseCreate() != true {
		logrus.Fatal("Error creating a database")
	}

	m.Run()
}

/*
func TestSymbolInformationSet_MutualFund(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	if err = postgres.TruncateTable(pgxConn, transaction.TransactionTable); err != nil {
		t.Error(err.Error())
		return
	}

	ts := transaction.NewTransactionSet()
	if err = ts.LoadToDB(pgxConn, lookupSet, transaction.TransactionTable, testTransactionsAll); err != nil {
		t.Error(err.Error())
		return
	}

	ds := historical.New(pgxConn, fundHistory)
	if err := ds.LoadSet(string(histUsaix), source, fundSymbol); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	for m := 1; m < 13; m++ {
		sd := symbol_details.NewSymbolDetail(fundHistory, fundSymbol, 2023, m)

		if err := sd.SetNumberOfShares(pgxConn); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		if err := sd.SetDividends(pgxConn); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		if err := sd.SetPrice(); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		t.Log(sd.String())
	}
}


func TestSymbolInformation_Stock(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}

	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	ts := transaction.NewTransactionSet()
	if err := ts.LoadToDB(pgxConn, lookupSet, transaction.TransactionTable, testTransactionsAll); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	for m := 1; m < 13; m++ {
		sd := symbol_details.NewSymbolDetail(fundHistory, stockSymbol, 2023, m)
		if err := sd.SetNumberOfShares(pgxConn); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		if err := sd.SetDividends(pgxConn); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		if err := sd.SetPrice(); err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		t.Log(sd.String())
	}
}

func TestNewSymbolDetailSet(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	if err = postgres.TruncateTable(pgxConn, transaction.TransactionTable); err != nil {
		t.Error(err.Error())
		return
	}

	ts := transaction.NewTransactionSet()
	if err = ts.LoadToDB(pgxConn, lookupSet, transaction.TransactionTable, testTransactionsAll); err != nil {
		t.Error(err.Error())
		return
	}

	date := time.Date(2024, time.Month(1), 1, 00, 00, 00, 00, time.UTC)
	set := symbol_details.NewSymbolDetailSet(pgxConn, stockSymbol, fundHistory)
	if err = set.Create(date, 12); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	t.Log(set.String())
}

*/
