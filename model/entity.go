package model

import (
	"fmt"
	"github.com/kpearce2430/keputils/utils"
	"github.com/segmentio/encoding/json"
	"github.com/sirupsen/logrus"
	"math"
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

func (e *Entity) Copy() *Entity {
	n := Entity{
		Date:             e.Date,
		Type:             e.Type,
		Security:         e.Security,
		Symbol:           e.Symbol,
		SecurityPayee:    e.SecurityPayee,
		Description:      e.Description,
		Shares:           e.Shares,
		InvestmentAmount: e.InvestmentAmount,
		Amount:           e.Amount,
		PricePerShare:    e.PricePerShare,
		RemainingShares:  e.RemainingShares,
	}

	for _, l := range e.SoldLots {
		nLot := Lot{
			NumberShares:  l.NumberShares,
			PricePerShare: l.PricePerShare,
			SoldDate:      l.SoldDate,
		}
		n.SoldLots = append(n.SoldLots, &nLot)
	}

	return &n
}

func NewEntityFromTransaction(tr *Transaction) (*Entity, error) {
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
			pps, err := utils.FloatParse(parts[3])
			// e.PricePerShare, err := utils.FloatParse(parts[3])
			if err != nil {
				e.PricePerShare = 0.00
				return &e, err
			}
			e.PricePerShare = pps

		default:
			logrus.Error("Invalid Description for Price Per Share:", e)
			e.PricePerShare = 0.00
			// return &e, errPricePerShare
		}
	}
	return &e, nil
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

func (e *Entity) amountType(incomeType TransactionType) float64 {
	if e.Type == incomeType {
		if e.Amount > 0 {
			return e.Amount
		}
		return e.InvestmentAmount
	}
	return 0.00
}

func (e *Entity) DividendIncome() float64 {
	return e.amountType("Dividend Income") + e.amountType("Reinvest Dividend")
}

func (e *Entity) LongTermCapitalGain() float64 {
	amt := e.amountType("Long-term Capital Gain")
	amt += e.amountType("Reinvest Long-term Capital Gain")
	return amt
}

func (e *Entity) ShortTermCapitalGain() float64 {
	amt := e.amountType("Short-term Capital Gain")
	amt += e.amountType("Reinvest Short-term Capital Gain")
	return amt
}

func (e *Entity) Dividends() float64 {
	return e.DividendIncome() + e.InterestIncome() + e.LongTermCapitalGain() + e.ShortTermCapitalGain()
}

func (e *Entity) DividendsPaid() float64 {
	/*
	   "Dividend Income", *
	   "Reinvest Dividend", *
	   "Interest Income", x
	   "Long-term Capital Gain", *
	   "Short-term Capital Gain", *
	   "Reinvest Long-term Capital Gain",*
	   "Reinvest Short-term Capital Gain", *
	*/
	switch e.Type {
	case "Dividend Income":
		return e.Amount
	case "Return of Capital":
		return e.Amount
	case "Reinvest Dividend":
		return e.InvestmentAmount
	case "Reinvest Long-term Capital Gain":
		return e.InvestmentAmount
	case "Short-term Capital Gain":
		return e.Amount
	case "Reinvest Short-term Capital Gain":
		return e.Amount
	}
	return 0.00
}

func (e *Entity) InterestIncome() float64 {
	switch e.Type {
	case "Interest Income", "Int Inc", "int inc":
		return e.Amount
	}
	return 0.00
}

func (e *Entity) NetCost() float64 {
	/*
	   amt = 0.00
	   type = self.entry.get("entryType")
	   if type in buyTransactions and self.numShares() > 0:
	       amt = abs(self.amount())
	       for lot in self.soldLots:
	           amt = amt - lot.proceeds()

	   # print("netCost Amount: {}".format(amt))
	   return amt

	*/
	amt := 0.00
	if utils.Contains(BuyTransactions, string(e.Type)) {
		if e.RemainingShares > 0.00 {
			amt = math.Abs(e.Amount)
			for _, lot := range e.SoldLots {
				amt -= lot.Proceeds()
			}
		}
	}
	return amt
}
