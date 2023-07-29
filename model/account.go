package model

import (
	"encoding/json"
	"fmt"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"math"
	"strings"
)

var buyTransactions = []string{
	"Buy",
	"Add Shares",
	"Reinvest Dividend",
	"Reinvest Long-term Capital Gain",
	"Reinvest Short-term Capital Gain",
}

type Account struct {
	Name        string    `json:"name,omitempty"`
	Entities    []*Entity `json:"entities,omitempty"`
	Pending     []*Entity `json:"pending,omitempty"`
	TotalShares float64   `json:"totalShares,omitempty"`
}

func NewAccount(name string) *Account {
	return &Account{
		Name: name,
	}
}

func (a *Account) AddEntity(e *Entity) {
	switch {
	case e.Type == "Sell" || e.Type == "Short Sell":
		a.SellShares(e)
		return
	case e.Type == "Stock Split":
		logrus.Info("Stock Split:", e.Symbol)
		a.SplitShares(e)
		return
	case e.Type == "Remove Shares":
		// Remove Shares
		a.RemoveShares(e)
		return
	default:
		a.Entities = append(a.Entities, e)
	}
	logrus.Debug("length of entities:", len(a.Entities))
	logrus.Info("Num Shares:", a.NumberOfShares())
	logrus.Info("Num Pending:", a.NumberOfPending())

	if len(a.Pending) > 0 && a.NumberOfShares() > a.NumberOfPending() {
		for len(a.Pending) > 0 {
			en := a.Pending[0]
			a.Pending = a.Pending[1:]
			a.RemoveShares(en)
		}
	}
}

func (a *Account) RemoveShares(e *Entity) {
	sharesToSell := math.Abs(e.Shares)
	for _, entry := range a.Entities {
		if contains(buyTransactions, string(entry.Type)) {
			sharesToSell = entry.SellShares(sharesToSell, 0.00)
		}
		if sharesToSell <= 0 {
			break
		}
	}
	if sharesToSell > 0.02 {
		logrus.Errorf("Remove Shares: %.02f Shares of %s Remaining to Sell", sharesToSell, e.Symbol)
		a.Pending = append(a.Pending, e)
	}
}

func (a *Account) SellShares(e *Entity) {
	sharesToSell := math.Abs(e.Shares)
	numberOfShares := a.NumberOfShares()

	if numberOfShares >= sharesToSell {
		logrus.Debugf("%s Selling %0.2f Shares, %.2f PPS: %.2f", e.Security, sharesToSell, numberOfShares, e.PricePerShare)
	}

	for _, entry := range a.Entities {
		if contains(buyTransactions, string(entry.Type)) {
			sharesToSell = entry.SellShares(sharesToSell, e.PricePerShare)
		}
		if sharesToSell <= 0 {
			break
		}
	}
	if sharesToSell > 0.02 {
		logrus.Errorf("%.02f Shares of %s Remaining to Sell", sharesToSell, e.Symbol)
		a.Pending = append(a.Pending, e)
	}
}

func (a *Account) SplitShares(e *Entity) {
	/*
	   wordList = myEntry.description().split()
	   #
	   newShares = wordList[0]
	   oldShares = wordList[2]
	   for e in self.entries:
	       e.splitShares(float(oldShares), float(newShares))
	*/
	parts := strings.Split(e.Description, " ")
	newShares, err := utils.FloatParse(parts[0])
	if err != nil {
		panic(err.Error())
	}
	oldShares, err := utils.FloatParse(parts[2])
	if err != nil {
		panic(err.Error())
	}

	logrus.Debug("New Shares:", newShares, " Old Shares:", oldShares)
	for _, entity := range a.Entities {
		entity.SplitShares(newShares, oldShares)
	}
}

func (a *Account) NumberOfShares() float64 {
	total := 0.00
	for _, e := range a.Entities {
		total += e.RemainingShares
	}
	return total
}

func (a *Account) NumberOfPending() float64 {
	total := 0.00
	for _, e := range a.Pending {
		total += math.Abs(e.RemainingShares)
	}
	return total
}

func (a *Account) String() string {
	a.TotalShares = a.NumberOfShares()
	bytes, err := json.Marshal(a)
	if err != nil {
		return fmt.Sprintf("%v", err.Error())
	}
	return fmt.Sprintf(string(bytes))
}
