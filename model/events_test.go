package model_test

import (
	"github.com/kpearce2430/stock-tools/model"
	"testing"
)

func eventDriver(t *testing.T, testSet *model.TransactionSet) bool {
	t.Helper()

	tickerMap := make(map[string]*model.Ticker)

	for _, tr := range testSet.TransactionRows {

		if tr.Symbol == "" {
			t.Log("Skipping Transaction:", tr)
			continue
		}

		v := tickerMap[tr.Symbol]
		if v == nil {
			t.Log("Adding ticker:", tr.Symbol)
			v = model.NewTicker(tr.Symbol)
			tickerMap[tr.Symbol] = v
		}

		ent, err := model.NewEntityFromTransaction(tr)
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
	testSet := model.NewTransactionSet()
	if err := testSet.Load(msftTransactions); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if !eventDriver(t, testSet) {
		t.Fail()
	}
}

func TestEvents_USAIX(t *testing.T) {
	testSet := model.NewTransactionSet()
	if err := testSet.Load(usaixTransactions); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if !eventDriver(t, testSet) {
		t.Fail()
	}
}
