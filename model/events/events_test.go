package events_test

import (
	"context"
	_ "embed"
	"log"
	"testing"

	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/model/ticker"
	"github.com/kpearce2430/stock-tools/model/transaction"
)

var (
	//go:embed testdata/msft.csv
	msftTransactions []byte

	//go:embed testdata/usaix.csv
	usaixTransactions []byte
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()
	m.Run()
}

func eventDriver(t *testing.T, testSet *transaction.TransactionSet) bool {
	t.Helper()

	tickerMap := make(map[string]*ticker.Ticker)

	for _, tr := range testSet.TransactionRows {

		if tr.Symbol == "" {
			t.Log("Skipping Transaction:", tr)
			continue
		}

		v := tickerMap[tr.Symbol]
		if v == nil {
			t.Log("Adding ticker:", tr.Symbol)
			v = ticker.NewTicker(tr.Symbol)
			tickerMap[tr.Symbol] = v
		}

		ent, err := ticker.NewEntityFromTransaction(tr)
		if err != nil {
			t.FailNow()
		}
		v.AddEntity(ent)
	}

	for _, ticker := range tickerMap {
		t.Log("Ticker:", ticker.Symbol)
		for _, acct := range ticker.Accounts {
			t.Log(acct.Name, ":", acct.NumberOfShares(), ":", acct.FirstBought())
		}
	}
	t.Log(len(tickerMap))
	return true
}

func TestEvents(t *testing.T) {
	testSet := transaction.NewTransactionSet()
	if err := testSet.Load(msftTransactions); err != nil {
		t.Error(err.Error())
		return
	}
	if !eventDriver(t, testSet) {
		t.Error("eventDriver Test failed")
		return
	}
}

func TestEvents_USAIX(t *testing.T) {
	testSet := transaction.NewTransactionSet()
	if err := testSet.Load(usaixTransactions); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if !eventDriver(t, testSet) {
		t.Fail()
	}
}
