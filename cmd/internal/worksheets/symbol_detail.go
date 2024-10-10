package worksheets

import (
	"fmt"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"time"
)

const (
	detailsDate      = "Date"
	detailsPrice     = "Price"
	detailsValue     = "Value"
	detailsQuantity  = "Quantity"
	detailsDividends = "Dividends"
)

func (w *WorkSheet) SymbolsDetails(worksheetName, symbol, table string, date time.Time, monthsAgo int) error {
	startRow := 1
	_, err := w.File.NewSheet(worksheetName)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	sd := model.NewSymbolDetailSet(w.PGXConn, symbol, table)
	if err := sd.Create(date, monthsAgo); err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	headers := []string{detailsDate, detailsPrice, detailsQuantity, detailsValue, detailsDividends}
	var allColumns []*ColumnInfo

	i := 1
	endRow := 1
	for _, h := range headers {
		colTransaction, err := NewColumnInfo(w.File, h, worksheetName, i)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}
		i++
		allColumns = append(allColumns, colTransaction)
		_ = colTransaction.WriteHeader(endRow, w.styles.Header)
	}

	var dateColumn string
	var priceColumn string
	var valuesColumn string
	var quantityColumn string
	var dividendColumn string

	for _, s := range sd.Info {
		endRow++

		for k, col := range allColumns {
			switch col.Name {
			case detailsDate:
				monthStr := time.Month(s.Month).String()
				if s.Month == 1 {
					_ = col.WriteCell(endRow, fmt.Sprintf("%s %4d", monthStr[0:3], s.Year), w.styles.TextStyle(endRow))
				} else {
					_ = col.WriteCell(endRow, fmt.Sprintf("%s", monthStr[0:3]), w.styles.TextStyle(endRow))
				}
				dateColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			case detailsValue:
				_ = col.WriteCell(endRow, s.Value(), w.styles.NumberStyle(endRow))
				valuesColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			case detailsQuantity:
				_ = col.WriteCell(endRow, s.Quantity, w.styles.NumberStyle(endRow))
				quantityColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			case detailsPrice:
				_ = col.WriteCell(endRow, s.Price, w.styles.CurrencyStyle(endRow))
				priceColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			case detailsDividends:
				_ = col.WriteCell(endRow, s.Dividends, w.styles.CurrencyStyle(endRow))
				dividendColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			}
		}
	}
	logrus.Debug(startRow, ":", endRow)

	priceChart := ChartBuilder{
		WorksheetName: worksheetName,
		Title:         "Price",
		Height:        300,
		Width:         500,
		Type:          excelize.Line,
	}
	priceChart.AddValueSeries(priceColumn, startRow+1, priceColumn, endRow)
	priceChart.AddCategorySeries(dateColumn, startRow+1, dateColumn, endRow)
	if err := priceChart.BuildChart(w, "f3"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	valuesChart := ChartBuilder{
		WorksheetName: worksheetName,
		Title:         "Values",
		Height:        300,
		Width:         500,
		Type:          excelize.Line,
	}
	valuesChart.AddValueSeries(valuesColumn, startRow+1, valuesColumn, endRow)
	valuesChart.AddCategorySeries(dateColumn, startRow+1, dateColumn, endRow)
	if err := valuesChart.BuildChart(w, "f20"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	quantityChart := ChartBuilder{
		WorksheetName: worksheetName,
		Title:         "Quantity",
		Height:        300,
		Width:         500,
		Type:          excelize.Line,
	}
	quantityChart.AddValueSeries(quantityColumn, startRow+1, quantityColumn, endRow)
	quantityChart.AddCategorySeries(dateColumn, startRow+1, dateColumn, endRow)
	if err := quantityChart.BuildChart(w, "n3"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	dividendChart := ChartBuilder{
		WorksheetName: worksheetName,
		Title:         "Dividends",
		Height:        300,
		Width:         500,
		Type:          excelize.Col3D,
	}
	dividendChart.AddValueSeries(dividendColumn, startRow+1, dividendColumn, endRow)
	dividendChart.AddCategorySeries(dateColumn, startRow+1, dateColumn, endRow)
	if err := dividendChart.BuildChart(w, "n20"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	return nil
}
