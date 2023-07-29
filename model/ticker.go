package model

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
)

type Ticker struct {
	Symbol      string              `json:"symbol,omitempty"`
	Accounts    map[string]*Account `json:"accounts,omitempty"`
	TotalShares float64             `json:"totalShares,omitempty"`
}

type TickerDatabase struct {
	Id     string  `json:"_id"`
	Rev    string  `json:"_rev,omitempty"`
	Ticker *Ticker `json:"ticker,omitempty"`
	Key    string  `json:"key"`
}

func NewTicker(symbol string) *Ticker {
	return &Ticker{
		Symbol:   symbol,
		Accounts: make(map[string]*Account),
	}
}

func (t *Ticker) AddEntity(en *Entity) {
	acct, ok := t.Accounts[en.Account]
	if !ok {
		acct = NewAccount(en.Account)
		t.Accounts[en.Account] = acct
	}
	acct.AddEntity(en)
	logrus.Debug("len of entities:", len(acct.Entities))
}

func (t *Ticker) NumberOfShares() float64 {
	total := 0.00
	for _, acct := range t.Accounts {
		total += acct.NumberOfShares()
	}
	return total
}

func (t *Ticker) String() string {
	bytes, err := json.Marshal(t)
	t.TotalShares = t.NumberOfShares()
	if err != nil {
		return fmt.Sprintf("%v", err.Error())
	}
	return fmt.Sprintf(string(bytes))
}
