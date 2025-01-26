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

// Account is the intermediary structure that holds a set of Entity values in the Entities list for an account.
type Account struct {
	Name     string    `json:"name,omitempty"`
	Entities []*Entity `json:"entities,omitempty"`
	Pending  []*Entity `json:"pending,omitempty"`
}

// NewAccount creates a new Account and setting the name.
func NewAccount(name string) *Account {
	return &Account{
		Name: name,
	}
}

// AddEntity add an Entity to the account as a transaction.
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
	logrus.Debug("Selling Bonds:", e.Symbol)
	for _, entry := range a.Entities {
		if entry.Type == "Buy Bonds" {
			numShares := entry.Shares
			if entry.SellShares(numShares, 0.00) != 0.00 {
				panic("Number of shares remaining!!!")
			}
		}
	}
}

// RemoveShares will remove the Entity shares from the account.
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

// SellShares will remove the Entity shares from the account from a sell.
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

// SplitShares will execute a split on the shares for the Entities in the account.
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

// NumberOfShares returns the total shares of the Entities in the account.
func (a *Account) NumberOfShares() float64 {
	total := 0.00
	for _, e := range a.Entities {
		total += e.RemainingShares
	}
	return total
}

// NumberOfPending returns the sum of the RemainingShares of the Pending entities in the account.
func (a *Account) NumberOfPending() float64 {
	total := 0.00
	for _, e := range a.Pending {
		total += math.Abs(e.RemainingShares)
	}
	return total
}

// Dividends returns the sum of the dividends of the Entities in the account.
func (a *Account) Dividends() float64 {
	amt := 0.00
	for _, e := range a.Entities {
		amt += e.Dividends()
	}
	return amt
}

// DividendsPaid returns the sum of the dividends paid of the Entities in the account.
func (a *Account) DividendsPaid() float64 {
	amt := 0.00
	for _, e := range a.Entities {
		amt += e.DividendsPaid()
	}
	return amt
}

// InterestIncome returns the sum of the interest paid of the Entities in the account.
func (a *Account) InterestIncome() float64 {
	amt := 0.00
	for _, e := range a.Entities {
		amt += e.InterestIncome()
	}
	return amt
}

// NetCost returns the sum of the costs of the Entities in the account
func (a *Account) NetCost() float64 {
	amt := 0.00
	for _, e := range a.Entities {
		amt += e.NetCost()
	}
	return amt
}

// FirstBought returns the data of the oldest Entity in the account.
func (a *Account) FirstBought() time.Time {
	theDate := time.Now()
	for _, e := range a.Entities {
		if utils.Contains(BuyTransactions, string(e.Type)) {
			if e.RemainingShares > 0.1 {
				if theDate.Unix() > e.Date.Unix() {
					theDate = e.Date
				}
			}
		}
	}
	return theDate
}

// AverageCost returns the average (NetCost / NumberOfShares) cost of the entities in the account.
func (a *Account) AverageCost() float64 {
	return a.NetCost() / a.NumberOfShares()
}

// String returns the string representation of the Account and it's Entity's
func (a *Account) String() string {
	bytes, err := json.Marshal(a)
	if err != nil {
		return fmt.Sprintf("%v", err.Error())
	}
	return fmt.Sprintf(string(bytes))
}
