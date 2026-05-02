package transactionswks

import (
	"context"
	"fmt"

	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/sirupsen/logrus"
)

type TransactionsWorksheet struct {
	w worksheets.WorksheetInterface
}

func New(w worksheets.WorksheetInterface) *TransactionsWorksheet {
	return &TransactionsWorksheet{w: w}
}

func (t *TransactionsWorksheet) Transactions(worksheetName, julDate string) error {
	logrus.Debug(worksheetName, ":", julDate)

	stockFile := t.w.GetFile()
	if stockFile == nil {
		logrus.Fatal("Unable to get stock file")
		return nil
	}
	sheet, err := stockFile.NewSheet(worksheetName)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	//id, date, type, security, security_payee, symbol, account, description, shares, investment_amount,amount
	tSet := transaction.NewTransactionSet()
	if err := tSet.GetAll(context.Background(), t.w.GetPGXConn()); err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	logrus.Debug("Received ", len(tSet.TransactionRows), " transactions")

	headers := []string{
		transaction.TransactionID, transaction.TransactionDate, transaction.TransactionType,
		transaction.TransactionSecurity, transaction.TransactionSecurityPayee, transaction.TransactionSymbol,
		transaction.TransactionAccount, transaction.TransactionDescription, transaction.TransactionShares,
		transaction.TransactionInvestmentAmount, transaction.TransactionAmount,
		transaction.TransactionYear, transaction.TransactionMonth,
	}

	i := 1
	row := 1
	dateCol := ""
	for _, h := range headers {
		colTransaction, err := column_info.New(t.w.GetExcelizeFile(), h, worksheetName, i)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}
		sheet.AddColumn(colTransaction)
		i++
		// allColumns = append(allColumns, colTransaction)
		switch h {
		case transaction.TransactionYear, transaction.TransactionMonth:
			colTransaction.SetFormula(true)
		case transaction.TransactionDate:
			dateCol = colTransaction.ColumnID
		}
		_ = colTransaction.WriteHeader(row, t.w.GetStyles().Header)
	}

	//
	row++
	for _, tr := range tSet.TransactionRows {
		for _, col := range sheet.Columns {

			switch col.Name {
			case transaction.TransactionID:
				_ = col.WriteCell(row, tr.Id, t.w.GetStyles().NumberStyle(row))
			case transaction.TransactionDate:
				_ = col.WriteCell(row, tr.Date, t.w.GetStyles().DateStyle(row))
			case transaction.TransactionType:
				_ = col.WriteCell(row, tr.Type, t.w.GetStyles().TextStyle(row))
			case transaction.TransactionSecurity:
				_ = col.WriteCell(row, tr.Security, t.w.GetStyles().TextStyle(row))
			case transaction.TransactionSecurityPayee:
				_ = col.WriteCell(row, tr.SecurityPayee, t.w.GetStyles().TextStyle(row))
			case transaction.TransactionSymbol:
				_ = col.WriteCell(row, tr.Symbol, t.w.GetStyles().TextStyle(row))
			case transaction.TransactionAccount:
				_ = col.WriteCell(row, tr.Account, t.w.GetStyles().TextStyle(row))
			case transaction.TransactionDescription:
				_ = col.WriteCell(row, tr.Description, t.w.GetStyles().TextStyle(row))
			case transaction.TransactionShares:
				_ = col.WriteCell(row, tr.Shares, t.w.GetStyles().NumberStyle(row))
			case transaction.TransactionInvestmentAmount:
				_ = col.WriteCell(row, tr.InvestmentAmount, t.w.GetStyles().CurrencyStyle(row))
			case transaction.TransactionAmount:
				_ = col.WriteCell(row, tr.Amount, t.w.GetStyles().CurrencyStyle(row))
			case transaction.TransactionYear:
				_ = col.WriteCell(row, fmt.Sprintf("=Year(%s%d)", dateCol, row), t.w.GetStyles().TextStyle(row))
			case transaction.TransactionMonth:
				_ = col.WriteCell(row, fmt.Sprintf("=Month(%s%d)", dateCol, row), t.w.GetStyles().TextStyle(row))
			default:
				return fmt.Errorf("bad type[%s]", col.Name)
			}
		}
		row++
	}
	return nil
}
