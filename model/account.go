package model

import (
	"encoding/json"
	"fmt"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"math"
	"strings"
	"time"
)

var BuyTransactions = []string{
	"Buy", "Buy Bonds", "Add Shares", "Reinvest Dividend", "Reinvest Long-term Capital Gain", "Reinvest Short-term Capital Gain", "xxx"}

type Account struct {
	Name     string    `json:"name,omitempty"`
	Entities []*Entity `json:"entities,omitempty"`
	Pending  []*Entity `json:"pending,omitempty"`
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
		logrus.Debug("Stock Split:", e.Symbol)
		a.SplitShares(e)
		return
	case e.Type == "Remove Shares":
		// Remove Shares
		a.RemoveShares(e)
		return
	case e.Type == "Sell Bonds":
		a.SellBonds(e)
		return
	default:
		a.Entities = append(a.Entities, e)
	}

	if len(a.Pending) > 0 && a.NumberOfShares() > a.NumberOfPending() {
		for len(a.Pending) > 0 {
			en := a.Pending[0]
			a.Pending = a.Pending[1:]
			a.RemoveShares(en)
		}
	}
}

// SellBonds will remove all shares from an account.
func (a *Account) SellBonds(e *Entity) {
	logrus.Info("Selling Bonds:", e.Symbol)
	for _, entry := range a.Entities {
		if entry.Type == "Buy Bonds" {
			numShares := entry.Shares
			if entry.SellShares(numShares, 0.00) != 0.00 {
				panic("Number of shares remaining!!!")
			}
		}
	}
}

func (a *Account) RemoveShares(e *Entity) {
	sharesToSell := math.Abs(e.Shares)
	for _, entry := range a.Entities {
		if utils.Contains(BuyTransactions, string(entry.Type)) {
			sharesToSell = entry.SellShares(sharesToSell, 0.00)
		}
		if sharesToSell <= 0 {
			break
		}
	}
	if sharesToSell > 0.02 {
		logrus.Debugf("Remove Shares: %.02f Shares of %s Remaining to Sell", sharesToSell, e.Symbol)
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
		if utils.Contains(BuyTransactions, string(entry.Type)) {
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

func (a *Account) Dividends() float64 {
	amt := 0.00
	for _, e := range a.Entities {
		amt += e.Dividends()
	}
	return amt
}

func (a *Account) DividendsPaid() float64 {
	amt := 0.00
	for _, e := range a.Entities {
		amt += e.DividendsPaid()
	}
	return amt
}

func (a *Account) InterestIncome() float64 {
	amt := 0.00
	for _, e := range a.Entities {
		amt += e.InterestIncome()
	}
	return amt
}

func (a *Account) NetCost() float64 {
	amt := 0.00
	for _, e := range a.Entities {
		amt += e.NetCost()
	}
	return amt
}

func (a *Account) FirstBought() time.Time {
	theDate := time.Now()
	for _, e := range a.Entities {
		if utils.Contains(BuyTransactions, string(e.Type)) {
			if e.RemainingShares > 0.1 {
				if theDate.Unix() > e.Date.Unix() {
					// logrus.Info(">", e)
					theDate = e.Date
				}
			}
		}
	}
	return theDate
}

func (a *Account) AverageCost() float64 {
	return a.NetCost() / a.NumberOfShares()
}

func (a *Account) String() string {
	bytes, err := json.Marshal(a)
	if err != nil {
		return fmt.Sprintf("%v", err.Error())
	}
	return fmt.Sprintf(string(bytes))
}
