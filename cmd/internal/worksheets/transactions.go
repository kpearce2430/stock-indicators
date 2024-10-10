package worksheets

import (
	"context"
	"fmt"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
)

const (
	TransactionID               = "ID"
	TransactionDate             = "Date"
	TransactionType             = "Type"
	TransactionSecurity         = "Security"
	TransactionSecurityPayee    = "Security Payee"
	TransactionSymbol           = "Symbol"
	TransactionAccount          = "Account"
	TransactionDescription      = "Description"
	TransactionShares           = "Shares"
	TransactionInvestmentAmount = "Investment Amount"
	TransactionAmount           = "Amount"
)

func (w *WorkSheet) Transactions(worksheetName, julDate string) error {
	logrus.Debug(worksheetName, ":", julDate)
	_, err := w.File.NewSheet(worksheetName)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	//id, date, type, security, security_payee, symbol, account, description, shares, investment_amount,amount
	tSet := model.NewTransactionSet()
	if err := tSet.TransactionsGetAll(context.Background(), w.PGXConn); err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	logrus.Info("Received ", len(tSet.TransactionRows), " transactions")

	headers := []string{
		TransactionID, TransactionDate, TransactionType, TransactionSecurity, TransactionSecurityPayee, TransactionSymbol,
		TransactionAccount, TransactionDescription, TransactionShares, TransactionInvestmentAmount, TransactionAmount,
	}

	var allColumns []*ColumnInfo

	i := 1
	row := 1
	for _, h := range headers {
		colTransaction, err := NewColumnInfo(w.File, h, worksheetName, i)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}
		i++
		allColumns = append(allColumns, colTransaction)
		_ = colTransaction.WriteHeader(row, w.styles.Header)
	}

	//
	row++
	for _, tr := range tSet.TransactionRows {
		for _, col := range allColumns {

			switch col.Name {
			case TransactionID:
				_ = col.WriteCell(row, tr.Id, w.styles.NumberStyle(row))
			case TransactionDate:
				_ = col.WriteCell(row, tr.Date, w.styles.DateStyle(row))
			case TransactionType:
				_ = col.WriteCell(row, tr.Type, w.styles.TextStyle(row))
			case TransactionSecurity:
				_ = col.WriteCell(row, tr.Security, w.styles.TextStyle(row))
			case TransactionSecurityPayee:
				_ = col.WriteCell(row, tr.SecurityPayee, w.styles.TextStyle(row))
			case TransactionSymbol:
				_ = col.WriteCell(row, tr.Symbol, w.styles.TextStyle(row))
			case TransactionAccount:
				_ = col.WriteCell(row, tr.Account, w.styles.TextStyle(row))
			case TransactionDescription:
				_ = col.WriteCell(row, tr.Description, w.styles.TextStyle(row))
			case TransactionShares:
				_ = col.WriteCell(row, tr.Shares, w.styles.TextStyle(row))
			case TransactionInvestmentAmount:
				_ = col.WriteCell(row, tr.InvestmentAmount, w.styles.TextStyle(row))
			case TransactionAmount:
				_ = col.WriteCell(row, tr.Amount, w.styles.TextStyle(row))
			default:
				return fmt.Errorf("bad type[%s]", col.Name)
			}
		}
		row++
	}
	return nil
}
