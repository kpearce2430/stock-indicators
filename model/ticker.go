package model

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

// Ticker is the top level storing the Accounts processing Entity records as they are added through AddEntity.
type Ticker struct {
	Symbol          string              `json:"symbol,omitempty"`
	SymbolType      string              `json:"symbolType,omitempty"`
	Accounts        map[string]*Account `json:"accounts,omitempty"`
	pendingEntities map[string]*Entity
	events          []Events
}

type TickerSet struct {
	Set map[string]*Ticker
}

func NewTickerSet() *TickerSet {
	return &TickerSet{
		Set: make(map[string]*Ticker),
	}
}

func (s *TickerSet) LoadTickerSet(ts *TransactionSet) error {
	for _, tr := range ts.TransactionRows {
		en, err := NewEntityFromTransaction(tr)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}

		ticker, ok := s.Set[en.Symbol]
		if !ok {
			ticker = NewTicker(en.Symbol)
			s.Set[en.Symbol] = ticker
		}
		ticker.AddEntity(en)
	}
	return nil
}

func (s *TickerSet) GetTicker(symbol string) (*Ticker, bool) {
	ticker, ok := s.Set[symbol]
	return ticker, ok
}

// TickerDatabase is the CouchDB record for a Ticker.
type TickerDatabase struct {
	Id     string  `json:"_id"`
	Rev    string  `json:"_rev,omitempty"`
	Ticker *Ticker `json:"ticker,omitempty"`
	Key    string  `json:"key"`
}

// NewTicker creates a new Ticker
func NewTicker(symbol string) *Ticker {
	return &Ticker{
		Symbol:   symbol,
		Accounts: make(map[string]*Account),
		events: []Events{
			{
				Date:        time.Date(2020, time.December, 28, 00, 00, 00, 00, time.UTC),
				FromAccount: "z HD Restricted Stock",
				ToAccount:   "HD ML Individual Account",
			},
			{
				Date:        time.Date(2021, time.March, 28, 00, 00, 00, 00, time.UTC),
				FromAccount: "z HD Restricted Stock",
				ToAccount:   "HD ML Individual Account",
			},
			{
				Date:        time.Date(2022, time.March, 23, 00, 00, 00, 00, time.UTC),
				FromAccount: "z HD Restricted Stock",
				ToAccount:   "HD ML Individual Account",
			},
			{
				Date:        time.Date(2023, time.March, 24, 00, 00, 00, 00, time.UTC),
				FromAccount: "z HD Restricted Stock",
				ToAccount:   "HD ML Individual Account",
			},
			{
				Date:        time.Date(2023, time.September, 05, 00, 00, 00, 00, time.UTC),
				FromAccount: "z Ameritrade IRA",
				ToAccount:   "Schwab Rollover IRA Keith",
			},
			{
				Date:        time.Date(2023, time.September, 05, 00, 00, 00, 00, time.UTC),
				FromAccount: "z Jane IRA",
				ToAccount:   "Schwab Contributory IRA Jane",
			},
			{
				Date:        time.Date(2024, time.March, 25, 00, 00, 00, 00, time.UTC),
				FromAccount: "z HD Restricted Stock",
				ToAccount:   "HD ML Individual Account",
			},
		},
		pendingEntities: make(map[string]*Entity),
	}
}

func (t *Ticker) AddEntity(en *Entity) {
	acct, ok := t.Accounts[en.Account]
	if !ok {
		acct = NewAccount(en.Account)
		t.Accounts[en.Account] = acct
	}

	switch en.Type {
	case "Remove Shares":
		for _, ev := range t.events {
			found, _ := ev.IsFromAccount(en.Date, en.Account)
			if found {
				t.pendingEntities[en.Account] = en
				return // not removing the shares
			}
			logrus.Debug("len>>", len(t.pendingEntities))

		}
	case "Add Shares":
		for _, ev := range t.events {
			logrus.Debug(ev.ToAccount, ":", ev.FromAccount)
			found, fromAccount := ev.IsToAccount(en.Date, en.Account)
			if found {
				pendingEntity := t.pendingEntities[fromAccount]
				if pendingEntity == nil {
					logrus.Debug("ERROR Missing pending entity >>>" + fromAccount + " | " + t.Symbol)
					return
				}

				fromAcct := t.Accounts[ev.FromAccount]
				if fromAcct == nil {
					logrus.Error("ERROR Missing from account")
					return
				}

				toAcct := t.Accounts[ev.ToAccount]
				if toAcct == nil {
					toAcct = NewAccount(en.Account)
					t.Accounts[ev.ToAccount] = toAcct
				}

				for _, fromEntity := range fromAcct.Entities {
					if (strings.Compare(string(fromEntity.Type), "Buy") == 0 ||
						strings.Compare(string(fromEntity.Type), "Buy Bonds") == 0 ||
						strings.Compare(string(fromEntity.Type), "Reinvest Dividend") == 0) &&
						fromEntity.RemainingShares == en.RemainingShares {
						addEn := fromEntity.Copy()
						addEn.Account = ev.ToAccount
						fromEntity.RemainingShares = 0
						toAcct.AddEntity(addEn)
						pendingEntity.RemainingShares -= addEn.RemainingShares
						return
					}
				}
			}
		}
	}
	acct.AddEntity(en)
	logrus.Debug("len of entities:", len(acct.Entities))
}

func (t *Ticker) NumberOfShares() float64 {
	return t.TotalShares(false)
}

func (t *Ticker) TotalShares(allAccounts bool) float64 {
	total := 0.00
	for _, acct := range t.Accounts {
		if !allAccounts && acct.Name[0] == 'z' {
			continue
		}
		total += acct.NumberOfShares()
	}
	return total
}

func (t *Ticker) Dividends() float64 {
	amt := 0.00
	for _, a := range t.Accounts {
		amt += a.Dividends()
	}
	return amt
}

func (t *Ticker) DividendsPaid() float64 {
	amt := 0.00
	for _, a := range t.Accounts {
		amt += a.DividendsPaid()
	}
	return amt
}

func (t *Ticker) InterestIncome() float64 {
	amt := 0.00
	for _, a := range t.Accounts {
		amt += a.InterestIncome()
	}
	return amt
}

func (t *Ticker) NetCost() float64 {
	amt := 0.00
	for _, a := range t.Accounts {
		amt += a.NetCost()
	}
	return amt
}

func (t *Ticker) FirstBought() time.Time {
	theDate := time.Now()
	for _, acct := range t.Accounts {
		acctFirstBought := acct.FirstBought()
		if theDate.Unix() > acctFirstBought.Unix() {
			theDate = acctFirstBought
		}
	}
	return theDate
}

func (t *Ticker) String() string {
	bytes, err := json.Marshal(t)
	if err != nil {
		return fmt.Sprintf("%v", err.Error())
	}
	return fmt.Sprintf(string(bytes))
}

func (t *Ticker) AveragePrice() float64 {
	if t.NumberOfShares() <= 0 {
		return 0.00
	}
	return t.NetCost() / t.NumberOfShares()
}

func (t *Ticker) GetAccount(name string) *Account {
	acct, ok := t.Accounts[name]
	if !ok {
		logrus.Error("Account not found")
		return nil
	}
	return acct
}
