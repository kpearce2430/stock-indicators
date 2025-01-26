package worksheets

import (
	"context"
	"fmt"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"time"
)

func (w *WorkSheet) AccountDividends(worksheetName string, start time.Time, monthsAgo int) error {
	_, err := w.File.NewSheet(worksheetName)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	accounts, err := model.AccountList(context.Background(), w.PGXConn)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	symbols, err := model.SymbolList(context.Background(), w.PGXConn, w.Lookups)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	var allColumns []*ColumnInfo
	colInfoSymbol, err := NewColumnInfo(w.File, "Date", worksheetName, 1)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	allColumns = append(allColumns, colInfoSymbol)
	col := 2

	for _, account := range accounts {
		if account[0] == 'z' {
			continue
		}
		colInfoSymbol, err = NewColumnInfo(w.File, account, worksheetName, col)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}
		allColumns = append(allColumns, colInfoSymbol)
		col++
	}

	totalColumn, err := NewColumnInfo(w.File, "Total", worksheetName, col)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	totalColumn.SetFormula(true)
	allColumns = append(allColumns, totalColumn)

	row := 1
	for _, colInfo := range allColumns {
		_ = colInfo.WriteHeader(row, w.styles.Header)
		colInfo.SetSize(12.0)
	}

	row++
	month := int(start.Month())
	year := start.Year()
	col = 1
	colInfo := allColumns[0]
	for i := 0; i < monthsAgo; i++ {

		monthString := time.Month(month).String()
		monthString = monthString[0:3]

		tickerSet := model.NewTickerSet()
		ts := model.NewTransactionSet()
		if err = ts.GetTransactions(context.Background(), w.PGXConn, "", year, month); err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}

		if err = tickerSet.LoadTickerSet(ts); err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}

		j := 0
		for j, colInfo = range allColumns {
			switch j {
			case 0:
				if monthString == "Jan" || monthString == "Dec" {
					dateStr := fmt.Sprintf("%s/%02d", monthString, year)
					_ = colInfo.WriteCell(row, dateStr, w.styles.TextStyle(row))
				} else {
					_ = colInfo.WriteCell(row, monthString, w.styles.TextStyle(row))
				}
			case len(allColumns) - 1:
				startCol := allColumns[1].ColumnID
				endCol := allColumns[len(allColumns)-2].ColumnID
				formula := fmt.Sprintf("=sum(%s%d:%s%d)", startCol, row, endCol, row)
				_ = colInfo.WriteCell(row, formula, w.styles.CurrencyStyle(row))

			default:
				var paid float64
				for symbolKey, _ := range symbols {
					ticker, ok := tickerSet.Set[symbolKey]
					if ok {
						a := ticker.GetAccount(colInfo.Name)
						if a != nil {
							paid = paid + a.Dividends()
						}
					}
				}
				//if paid > 0 {
				//	logrus.Println(colInfo.Name, ": ", monthString, "/", year, " = ", paid)
				//}
				_ = colInfo.WriteCell(row, paid, w.styles.CurrencyStyle(row))

			}
		}
		month--
		if month < 1 {
			month = 12
			year--
		}
		row++
	}

	for _, colInfo = range allColumns {
		_ = colInfo.SetColumnSize()
	}

	return nil
}
