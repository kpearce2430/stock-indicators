package model_test

import (
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
	"strings"
	"testing"
	"time"
)

func TestNewEntityFromTransaction(t *testing.T) {
	testSet1 := model.NewTransactionSet()
	if err := testSet1.Load(applTransactions); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	var e *model.Entity
	var err error
	for _, tr := range testSet1.TransactionRows {
		switch tr.Type {
		case "Buy":
			e, err = model.NewEntityFromTransaction(tr)
			if err != nil {
				t.Log(err.Error())
				t.FailNow()
			}
			if e.RemainingShares != 100 || e.Shares != 100 {
				t.Log("Invalid Shares")
				t.Fail()
			}
		case "Stock Split":
			parts := strings.Split(tr.Description, " ")
			if e == nil {
				t.Log("No entity")
				t.FailNow()
			}
			newShares, err := utils.FloatParse(parts[0])
			if err != nil {
				t.Log(err.Error())
				t.FailNow()
			}
			oldShares, err := utils.FloatParse(parts[2])
			if err != nil {
				t.Log(err.Error())
				t.FailNow()
			}
			e.SplitShares(newShares, oldShares)
			if e.RemainingShares != 400 {
				t.Log("Number of shares don'hist_usaix.csv match 400:", e.RemainingShares)
				t.Fail()
			}
		}
	}

	e.SellShares(200.0, 100.00)
	if e.RemainingShares != 200 {
		t.Log("Number of shares don'hist_usaix.csv match 200:", e.RemainingShares)
		t.Fail()
	}
}

func TestEntity_BadData(t *testing.T) {
	tr := model.Transaction{
		Id:          1,
		Date:        time.Date(2023, time.August, 01, 00, 00, 00, 00, time.UTC),
		Type:        "Buy",
		Description: "Some description goes here that isn'hist_usaix.csv expected",
	}
	e, err := model.NewEntityFromTransaction(&tr)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	if e.PricePerShare != 0.00 {
		t.Log("Invalid price per share")
		t.Fail()
	}
}
