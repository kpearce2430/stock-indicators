package model_test

import (
	"github.com/kpearce2430/stock-tools/model"
	"testing"
)

func TestTickerAppl(t *testing.T) {

	testSet1 := model.NewTransactionSet()
	if err := testSet1.Load(applTransactions); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	Tickers := make(map[string]*model.Ticker)

	for _, tr := range testSet1.TransactionRows {
		e, err := model.NewEntityFromTransaction(tr)

		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}

		ticker := Tickers[e.Symbol]
		if ticker == nil {
			ticker = model.NewTicker(e.Symbol)
			Tickers[e.Symbol] = ticker
		}
		ticker.AddEntity(e)
	}

	t.Log("Length>", len(Tickers))
	for _, ticker := range Tickers {
		t.Log(ticker.Symbol)
		t.Log(ticker.NumberOfShares())
		t.Log(ticker.DividendsPaid())
		t.Log(ticker.FirstBought())
		t.Log(ticker.NetCost())
		t.Log(ticker.AveragePrice())
	}
}
