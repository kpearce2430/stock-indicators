package model_test

import (
	"github.com/kpearce2430/stock-tools/model"
	"testing"
)

func TestAccount_AddEntity(t *testing.T) {
	testSet1 := model.NewTransactionSet()
	if err := testSet1.Load(applTransactions); err != nil {
		t.Log(err.Error())
		t.FailNow()
		return
	}

	Accounts := make(map[string]*model.Account)
	for _, tr := range testSet1.TransactionRows {
		e, err := model.NewEntityFromTransaction(tr)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		var acct *model.Account
		acct = Accounts[tr.Account]
		if acct == nil {
			acct = model.NewAccount(tr.Account)
			Accounts[tr.Account] = acct
		}
		acct.AddEntity(e)
	}

	t.Log("Length>", len(Accounts))
	for _, acct := range Accounts {
		t.Log(acct.Name)
		t.Log(acct.NumberOfShares())
		t.Log(acct.DividendsPaid())
		t.Log(acct.FirstBought())
		t.Log(acct.NetCost())
		t.Log(acct.AverageCost())
	}
}

func TestAccount_SellBonds(t *testing.T) {

	testSet1 := model.NewTransactionSet()
	if err := testSet1.Load(bondTransactions); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	Accounts := make(map[string]*model.Account)
	for _, tr := range testSet1.TransactionRows {
		e, err := model.NewEntityFromTransaction(tr)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		var acct *model.Account
		acct = Accounts[tr.Account]
		if acct == nil {
			acct = model.NewAccount(tr.Account)
			Accounts[tr.Account] = acct
		}
		acct.AddEntity(e)
	}

	t.Log("Length>", len(Accounts))
	for _, acct := range Accounts {
		t.Log(acct.Name, ",", acct.NumberOfShares(), ", $", acct.DividendsPaid(), ", $", acct.InterestIncome())
		if acct.NumberOfShares() != 0.00 {
			t.Error("Expecting 0.00 Shares, Found:", acct.NumberOfShares())
		}
	}

}
