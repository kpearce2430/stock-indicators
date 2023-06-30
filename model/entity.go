package model

import (
	"fmt"
	"github.com/kpearce2430/keputils/utils"
	"time"
)

type TransactionType string

type Entity struct {
	Date          time.Time       `json:"date,omitempty"`
	Type          TransactionType `json:"type,omitempty"`
	Security      string          `json:"security,omitempty"`
	Symbol        string          `json:"symbol,omitempty"`
	SecurityPayee string          `json:"security_payee,omitempty"`
	Description   string          `json:"description,omitempty"`
	Shares        float64         `json:"shares,omitempty"`
	Amount        float64         `json:"amount,omitempty"`
	Account       string          `json:"account,omitempty"`
}

func NewEntity(headers []string, row []string) (*Entity, error) {
	//
	e := Entity{}
	for i, h := range headers {
		switch h {
		case "Date":
			date, err := time.Parse("1/2/2006", row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Date: %v", err.Error())
			}
			e.Date = date
		case "Type":
			e.Type = TransactionType(row[i])
		case "Security":
			e.Security = row[i]
		case "Symbol":
			e.Symbol = row[i]
		case "Security/Payee":
			e.SecurityPayee = row[i]
		case "Description/Category":
			e.Description = row[i]
		case "Shares":
			shares, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Shares: %v", err.Error())
			}
			e.Shares = shares
		case "Invest Amt":
			iAmt, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Invest Amt: %v", err.Error())
			}
			e.Shares = iAmt
		case "Amount":
			amt, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Invest Amt: %v", err.Error())
			}
			e.Shares = amt
		case "Account":
			e.Account = row[i]
		default:
			if h != "Split" {
				fmt.Println("Skipping ", h)
			}
		}
	}
	return &e, nil
}
