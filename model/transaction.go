package model

import (
	"encoding/csv"
	"fmt"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"time"
)

// Transaction is an individual transaction read in from the CSV data provided.
type Transaction struct {
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
}

type TransactionSet struct {
	TransactionRows []*Transaction
	Date            time.Time
}

// NewTransaction creates a new transaction record from headers and a CSV row.
func NewTransaction(headers []string, row []string) (*Transaction, error) {
	tr := Transaction{}

	for i, h := range headers {
		switch h {
		case "Date":
			if row[i] == "" {
				logrus.Error("Invalid Row(", len(row), ") ", row)
				return nil, fmt.Errorf("Invalid date in row %d", i)
			}
			date, err := time.Parse("1/2/2006", row[i])
			if err != nil {
				return nil, fmt.Errorf("NewEntity Date[%s]: %v", row[i], err.Error())
			}
			tr.Date = date
		case "Type":
			tr.Type = TransactionType(row[i])
		case "Security":
			tr.Security = row[i]
		case "Symbol":
			tr.Symbol = row[i]
		case "Security/Payee":
			tr.SecurityPayee = row[i]
		case "Description/Category":
			tr.Description = row[i]
		case "Shares":
			shares, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Shares: %v", err.Error())
			}
			tr.Shares = shares
		case "Invest Amt":
			iAmt, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Invest Amt: %v", err.Error())
			}
			tr.InvestmentAmount = iAmt
		case "Amount":
			amt, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Invest Amt: %v", err.Error())
			}
			tr.Amount = amt
		case "Account":
			tr.Account = row[i]
		default:
			if h != "Split" {
				fmt.Println("Skipping ", h)
			}
		}
	}

	return &tr, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			logrus.Debug("contains ", str, " in ", s)
			return true
		}
	}
	return false
}

func NewTransactionSet() *TransactionSet {
	return &TransactionSet{
		Date: time.Now(),
	}
}

func (t *TransactionSet) Load(rawData []byte) error {
	//
	r := csv.NewReader(strings.NewReader(string(rawData)))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1

	fmt.Println("In TransactionSet.Load()", len(rawData))
	foundHeader := false
	var headers []string

	for count := 0; count < 10000; count++ {
		record, err := r.Read()

		if err == io.EOF {
			fmt.Println("found end of file")
			break
		}

		if err != nil {
			fmt.Println("Error>", err.Error())
			break
		}

		if !foundHeader {
			if contains(record, "Date") {
				fmt.Println("Found Header ", record)
				foundHeader = true
				for _, r := range record[1:] {
					headers = append(headers, r)
				}
			}
			continue
		}

		// Need a better way to do this
		if len(record[1:]) != len(headers) {
			logrus.Info("Skipping row(", record[1:], ")")
			continue
		}
		en, err := NewTransaction(headers, record[1:])
		if err != nil {
			fmt.Println("Error:", err.Error())
			return fmt.Errorf("TR Load %s", err.Error())
		}

		t.TransactionRows = append(t.TransactionRows, en)
	}
	return nil
}
