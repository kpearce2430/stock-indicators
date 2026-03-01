package worksheets

import (
	"context"
	"fmt"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
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
	TransactionYear             = "Year"
	TransactionMonth            = "Month"
)

func (w *WorkSheet) Transactions(worksheetName, julDate string) error {
	logrus.Debug(worksheetName, ":", julDate)
	sheet, err := w.StockFile.NewSheet(worksheetName)
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
		TransactionYear, TransactionMonth,
	}

	// var allColumns []*ColumnInfo

	i := 1
	row := 1
	dateCol := ""
	for _, h := range headers {
		colTransaction, err := column_info.New(w.StockFile.GetFile(), h, worksheetName, i)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}
		sheet.AddColumn(colTransaction)
		i++
		// allColumns = append(allColumns, colTransaction)
		switch h {
		case TransactionYear, TransactionMonth:
			colTransaction.SetFormula(true)
		case TransactionDate:
			dateCol = colTransaction.ColumnID
		}
		_ = colTransaction.WriteHeader(row, w.StockFile.Styles.Header)
	}

	//
	row++
	for _, tr := range tSet.TransactionRows {
		for _, col := range sheet.Columns {

			switch col.Name {
			case TransactionID:
				_ = col.WriteCell(row, tr.Id, w.StockFile.Styles.NumberStyle(row))
			case TransactionDate:
				_ = col.WriteCell(row, tr.Date, w.StockFile.Styles.DateStyle(row))
			case TransactionType:
				_ = col.WriteCell(row, tr.Type, w.StockFile.Styles.TextStyle(row))
			case TransactionSecurity:
				_ = col.WriteCell(row, tr.Security, w.StockFile.Styles.TextStyle(row))
			case TransactionSecurityPayee:
				_ = col.WriteCell(row, tr.SecurityPayee, w.StockFile.Styles.TextStyle(row))
			case TransactionSymbol:
				_ = col.WriteCell(row, tr.Symbol, w.StockFile.Styles.TextStyle(row))
			case TransactionAccount:
				_ = col.WriteCell(row, tr.Account, w.StockFile.Styles.TextStyle(row))
			case TransactionDescription:
				_ = col.WriteCell(row, tr.Description, w.StockFile.Styles.TextStyle(row))
			case TransactionShares:
				_ = col.WriteCell(row, tr.Shares, w.StockFile.Styles.NumberStyle(row))
			case TransactionInvestmentAmount:
				_ = col.WriteCell(row, tr.InvestmentAmount, w.StockFile.Styles.CurrencyStyle(row))
			case TransactionAmount:
				_ = col.WriteCell(row, tr.Amount, w.StockFile.Styles.CurrencyStyle(row))
			case TransactionYear:
				_ = col.WriteCell(row, fmt.Sprintf("=Year(%s%d)", dateCol, row), w.StockFile.Styles.TextStyle(row))
			case TransactionMonth:
				_ = col.WriteCell(row, fmt.Sprintf("=Month(%s%d)", dateCol, row), w.StockFile.Styles.TextStyle(row))
			default:
				return fmt.Errorf("bad type[%s]", col.Name)
			}
		}
		row++
	}
	return nil
}
