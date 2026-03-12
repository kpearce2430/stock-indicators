package dividends_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/model/dividends"
	"github.com/sirupsen/logrus"
)

const dividendsTable = "dividends"

var (
	//go:embed testdata/dividends.json
	testDividendsData []byte
)

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

func TestDividendsSet_ToDB(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	ds, err := dividends.NewDividendsSetFromJSON(testDividendsData)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(len(ds.Dividends))

	err = ds.ToDB(context.Background(), pgxConn, dividendsTable)
	if err != nil {
		t.Error(err)
		return
	}

	responseDS := dividends.DividendsSet{}
	err = responseDS.FromDBbySymbol(context.Background(), pgxConn, dividendsTable, "CSX")
	if err != nil {
		t.Error(err)
		return
	}

	b, err := json.MarshalIndent(&responseDS, "", " ")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(b))

	if len(ds.Dividends) != len(responseDS.Dividends) {
		t.Error("Number of dividends does not match number of dividends")
	}
}
