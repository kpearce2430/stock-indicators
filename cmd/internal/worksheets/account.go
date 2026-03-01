package worksheets

import (
	"context"
	"fmt"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/sirupsen/logrus"
	"time"
)

func (w *WorkSheet) AccountDividends(worksheetName string, start time.Time, monthsAgo int) error {
	sheet, err := w.StockFile.NewSheet(worksheetName)
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

	colInfoSymbol, err := column_info.New(w.StockFile.GetFile(), "Date", worksheetName, 1)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	sheet.AddColumn(colInfoSymbol)

	col := 2
	for _, account := range accounts {
		if account[0] == 'z' {
			continue
		}
		ci, err := column_info.New(w.StockFile.GetFile(), account, worksheetName, col)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}
		sheet.AddColumn(ci)
		col++
	}

	totalColumn, err := column_info.New(w.StockFile.GetFile(), "Total", worksheetName, col)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	totalColumn.SetFormula(true)
	sheet.AddColumn(totalColumn)

	row := 1
	for _, ci := range sheet.Columns {
		_ = ci.WriteHeader(row, w.StockFile.Styles.Header)
		ci.SetSize(12.0)
	}

	row++
	month := int(start.Month())
	year := start.Year()
	col = 1
	colInfo, ok := sheet.GetColumn(0)
	if !ok || colInfo == nil {
		logrus.Fatal("Unable to Get Column 0")
		return err
	}
	for range monthsAgo {
		//
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
		for j, colInfo = range sheet.Columns {
			switch j {
			case 0:
				if monthString == "Jan" || monthString == "Dec" {
					dateStr := fmt.Sprintf("%s/%02d", monthString, year)
					_ = colInfo.WriteCell(row, dateStr, w.StockFile.Styles.TextStyle(row))
				} else {
					_ = colInfo.WriteCell(row, monthString, w.StockFile.Styles.TextStyle(row))
				}
			case len(sheet.Columns) - 1:
				ci, _ := sheet.GetColumn(1)
				startCol := ci.ColumnID
				ci, _ = sheet.GetColumn(len(sheet.Columns) - 2)
				endCol := ci.ColumnID
				formula := fmt.Sprintf("=sum(%s%d:%s%d)", startCol, row, endCol, row)
				_ = colInfo.WriteCell(row, formula, w.StockFile.Styles.CurrencyStyle(row))

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
				_ = colInfo.WriteCell(row, paid, w.StockFile.Styles.CurrencyStyle(row))
			}
		}
		month--
		if month < 1 {
			month = 12
			year--
		}
		row++
	}

	for _, colInfo = range sheet.Columns {
		_ = colInfo.SetColumnSize()
	}

	return nil
}
