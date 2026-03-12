package ticker_test

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
	//go:embed testdata/aapl.csv
	applTransactions []byte

	//go:embed testdata/transactions.csv
	testTransactionsAll []byte

	//go:embed testdata/bti.csv
	btiTransactions []byte

	//go:embed testdata/bond_data.csv
	bondData []byte
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

func TestTickerAapl(t *testing.T) {
	testSet1 := transaction.NewTransactionSet()
	if err := testSet1.Load(applTransactions); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	Tickers := make(map[string]*ticker.Ticker)
	for _, tr := range testSet1.TransactionRows {
		e, err := ticker.NewEntityFromTransaction(tr)

		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}

		tckr := Tickers[e.Symbol]
		if tckr == nil {
			tckr = ticker.NewTicker(e.Symbol)
			Tickers[e.Symbol] = tckr
		}
		tckr.AddEntity(e)
	}

	t.Log("Length>", len(Tickers))
	for _, tckr := range Tickers {
		t.Log(tckr.Symbol)
		t.Log(tckr.NumberOfShares())
		t.Log(tckr.DividendsPaid())
		t.Log(tckr.FirstBought())
		t.Log(tckr.NetCost())
		t.Log(tckr.AveragePrice())
	}
}

func TestTicker_GetAccount(t *testing.T) {
	testSet := transaction.NewTransactionSet()
	if err := testSet.Load(testTransactionsAll); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	Tickers := make(map[string]*ticker.Ticker)
	for _, tr := range testSet.TransactionRows {
		e, err := ticker.NewEntityFromTransaction(tr)

		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}

		tckr := Tickers[e.Symbol]
		if tckr == nil {
			tckr = ticker.NewTicker(e.Symbol)
			Tickers[e.Symbol] = tckr
		}
		tckr.AddEntity(e)
	}

	ticker, ok := Tickers["HD"]
	if !ok {
		t.Error("Ticker not found")
		return
	}

	acct := ticker.GetAccount("HD ESPP")
	if acct == nil {
		t.Error("Account not found")
		return
	}

	t.Log(acct.DividendsPaid())
	t.Log(len(acct.Entities))
}

func TestTickerBTI(t *testing.T) {
	//
	testSet1 := transaction.NewTransactionSet()
	if err := testSet1.Load(btiTransactions); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	Tickers := make(map[string]*ticker.Ticker)
	for _, tr := range testSet1.TransactionRows {
		e, err := ticker.NewEntityFromTransaction(tr)

		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}

		if e.Symbol == "" {
			continue
		}

		tckr := Tickers[e.Symbol]

		if tckr == nil {
			tckr = ticker.NewTicker(e.Symbol)
			Tickers[e.Symbol] = tckr
		}
		tckr.AddEntity(e)
	}

	t.Log("Length>", len(Tickers))
	for _, tckr := range Tickers {
		t.Log(tckr.Symbol)
		t.Log(tckr.NumberOfShares())

		if tckr.NumberOfShares() != 0.00 {
			t.Error("Expecting 0.00 Shares, Found:", tckr.NumberOfShares())
		}
		t.Log(tckr.DividendsPaid())
		t.Log(tckr.FirstBought())
		t.Log(tckr.NetCost())
		t.Log(tckr.AveragePrice())
	}

	if len(Tickers) != 1 {
		t.Error("Expected 1 ticker, got", len(Tickers))
	}

}
