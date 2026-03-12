package lookups_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/sirupsen/logrus"
)

//go:embed testdata/lookups.csv
var csvLookupData []byte

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			logrus.Fatal(err.Error())
		}
	}()
	m.Run()
}

func TestLoadLookupSet(t *testing.T) {
	ls := lookups.LoadLookupSet("1", string(csvLookupData))
	if len(ls.LookUps) != 14 {
		t.Error("LookUp Count ", len(ls.LookUps), " does not equal 9")
	}
}

func TestLoadLookupToDB(t *testing.T) {
	const lookupTableName = "lookups"

	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = lookups.LoadLookupFromCSV(context.TODO(), pgxConn, lookupTableName, csvLookupData)
	if err != nil {
		t.Error(err.Error())
		return
	}

	ls, err := lookups.GetLookUpsFromDB(context.TODO(), pgxConn, lookupTableName)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(ls.LookUps) != 14 {
		t.Error("LookUp Count ", len(ls.LookUps), " does not equal 9")
	}
	t.Log(ls)
}
