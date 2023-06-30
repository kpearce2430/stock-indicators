package app_test

import (
	"context"
	_ "embed"
	"fmt"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"testing"
)

//go:embed testdata/lookups.csv
var csvLookupData []byte

//go:embed testdata/portfolio_value.csv
var csvPortfolioValueData []byte

// TestMain
func TestMain(m *testing.M) {
	ctx := context.Background()
	couchDBServer, _ := couch_database.CreateCouchDBServer(ctx)
	defer couchDBServer.Terminate(ctx)

	ip, err := couchDBServer.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := couchDBServer.MappedPort(ctx, "5984")
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	logrus.Debugln(url)

	_ = os.Setenv("COUCHDB_URL", url)
	_ = os.Setenv("COUCHDB_USER", "admin")
	_ = os.Setenv("COUCHDB_PASSWORD", "password")
	_ = os.Setenv("DATABASE_NAME", "pv")
	logrus.Debug("Starting tests")
	m.Run()
}
