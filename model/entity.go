package model

import (
	"fmt"
	"github.com/kpearce2430/keputils/utils"
	"github.com/segmentio/encoding/json"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type TransactionType string

type Entity struct {
	Date             time.Time       `json:"date,omitempty"`
	Type             TransactionType `json:"type,omitempty"`
	Security         string          `json:"security,omitempty"`
	Symbol           string          `json:"symbol,omitempty"`
	SecurityPayee    string          `json:"security_payee,omitempty"`
	Description      string          `json:"description,omitempty"`
	Shares           float64         `json:"shares,omitempty"`
	InvestmentAmount float64         `json:"investment_amount,omitempty"`
	Amount           float64         `json:"amount,omitempty"`
	Account          string          `json:"account,omitempty"`
	PricePerShare    float64         `json:"pps,omitempty"`
	RemainingShares  float64         `json:"remaining_shares,omitempty"`
	SoldLots         []*Lot          `json:"sold_lots,omitempty"`
}

func NewEntityFromTransaction(tr *Transaction) *Entity {
	e := Entity{
		Date:             tr.Date,
		Type:             tr.Type,
		Security:         tr.Security,
		Symbol:           tr.Symbol,
		SecurityPayee:    tr.SecurityPayee,
		Description:      tr.Description,
		Shares:           tr.Shares,
		InvestmentAmount: tr.InvestmentAmount,
		Amount:           tr.Amount,
		Account:          tr.Account,
		RemainingShares:  tr.Shares,
	}

	if e.Type == "Buy" || e.Type == "Reinvest Dividend" || e.Type == "Sell" {
		parts := strings.Split(e.Description, " ")
		switch len(parts) {
		case 4:
			e.PricePerShare, _ = utils.FloatParse(parts[3])
		default:
			logrus.Error("Invalid Description for Price Per Share:", e)
			e.PricePerShare = 0.00
		}
	}
	return &e
}

func (e *Entity) SellShares(numSharesToSell float64, pps float64) float64 {

	if e.RemainingShares <= 0 {
		return numSharesToSell
	}
	// if there are 50 shares to sell with 100 remaining shares, remove 50
	// return 0 for the number of shares remaining to sell.
	// partial or full sale
	if e.RemainingShares >= numSharesToSell {
		e.RemainingShares = e.RemainingShares - numSharesToSell
		lot := Lot{
			NumberShares:  numSharesToSell,
			PricePerShare: pps,
			SoldDate:      time.Now(), // TODO - Fix this
		}
		e.SoldLots = append(e.SoldLots, &lot)
		return 0.00
	}
	// there are 50 shares remaining and 100 to sell,
	// remove the 50 and return there are 50 more to sell.
	remainingShares := numSharesToSell - e.RemainingShares
	lot := Lot{
		NumberShares:  e.RemainingShares,
		PricePerShare: pps,
		SoldDate:      time.Now(), // TODO - Fix this
	}
	e.SoldLots = append(e.SoldLots, &lot)
	e.RemainingShares = 0.00
	return remainingShares
}

func (e *Entity) SplitShares(newShares, oldShares float64) {
	/*
	   newRemainingShares = (float(self.remainingShares()) / oldShares) * newShares
	   self.entry["entryRemainingShares"] = str(newRemainingShares)
	*/
	if e.RemainingShares <= 0 {
		return
	}
	e.RemainingShares = (e.RemainingShares / oldShares) * newShares
}

func (e *Entity) String() string {
	bytes, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("%v", err.Error())
	}
	return fmt.Sprintf(string(bytes))
}
