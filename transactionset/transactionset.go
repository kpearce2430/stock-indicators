package transactionset

import (
	"encoding/csv"
	"fmt"
	"iex-indicators/model"
	"io"
	"strings"
	"time"
)

//type TransactionType string
//
//type TransactionRow struct {
//	Date          time.Time       `json:"date,omitempty"`
//	Type          TransactionType `json:"type,omitempty"`
//	Security      string          `json:"security,omitempty"`
//	Symbol        string          `json:"symbol,omitempty"`
//	SecurityPayee string          `json:"security_payee,omitempty"`
//	Description   string          `json:"description,omitempty"`
//	Shares        float64         `json:"shares,omitempty"`
//	Amount        float64         `json:"amount,omitempty"`
//	Account       string          `json:"account,omitempty"`
//}

var transactionHeaders = []string{
	"Split",
	"Date",
	"Type",
	"Security",
	"Symbol",
	"Security/Payee",
	"Description/Category",
	"Shares",
	"Invest Amt",
	"Amount",
	"Account",
}

type TransactionSet struct {
	TransactionRows []*model.Entity
	Date            time.Time
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			fmt.Println("contains ", str, " in ", s)
			return true
		}
	}
	return false
}

func NewTransactionSet() *TransactionSet {
	t := TransactionSet{}
	return &t
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
			// fmt.Println(len(headers), foundHeader)
			continue
		}

		r, err := model.NewEntity(headers, record[1:])
		if err != nil {
			fmt.Println("Error:", err.Error())
			return fmt.Errorf("TR Load %s", err.Error())
		}
		t.TransactionRows = append(t.TransactionRows, r)
	}
	return nil
}
