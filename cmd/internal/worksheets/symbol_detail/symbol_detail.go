package symbol_detail

import (
	"fmt"
	"time"

	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model/symbol_details"
	"github.com/kpearce2430/stock-tools/stocksheet/chart_builder"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

const (
	detailsDate      = "Date"
	detailsPrice     = "Price"
	detailsValue     = "Value"
	detailsQuantity  = "Quantity"
	detailsDividends = "Dividends"
)

type SymbolDetail struct {
	worksheets.WorksheetInterface
}

func New(w worksheets.WorksheetInterface) *SymbolDetail {
	return &SymbolDetail{w}
}

func (s *SymbolDetail) SymbolsDetails(worksheetName, symbol, table string, date time.Time, monthsAgo int) error {
	startRow := 1
	stockFile := s.GetFile()
	if stockFile == nil {
		logrus.Fatal("Unable to get stock file")
		return nil
	}
	sheet, err := stockFile.NewSheet(worksheetName)
	if err != nil {
		logrus.Error("NewSheet Error:", err.Error())
		return err
	}

	sd := symbol_details.NewSymbolDetailSet(s.GetPGXConn(), symbol)
	if err = sd.Create(date, monthsAgo); err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	headers := []string{detailsDate, detailsPrice, detailsQuantity, detailsValue, detailsDividends}

	i := 1
	endRow := 1
	for _, h := range headers {
		colTransaction, err := column_info.New(s.GetExcelizeFile(), h, worksheetName, i)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}
		i++
		sheet.AddColumn(colTransaction)
		_ = colTransaction.WriteHeader(endRow, s.GetStyles().Header)
	}

	var dateColumn string
	var priceColumn string
	var valuesColumn string
	var quantityColumn string
	var dividendColumn string

	for _, sd := range sd.Info {
		endRow++

		for k, col := range sheet.Columns {
			switch col.Name {
			case detailsDate:
				monthStr := time.Month(sd.Month).String()
				if sd.Month == 1 {
					_ = col.WriteCell(endRow, fmt.Sprintf("%s %4d", monthStr[0:3], sd.Year), s.GetStyles().TextStyle(endRow))
				} else {
					_ = col.WriteCell(endRow, fmt.Sprintf("%s", monthStr[0:3]), s.GetStyles().TextStyle(endRow))
				}
				dateColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			case detailsValue:
				_ = col.WriteCell(endRow, sd.Value(), s.GetStyles().NumberStyle(endRow))
				valuesColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			case detailsQuantity:
				_ = col.WriteCell(endRow, sd.Quantity, s.GetStyles().NumberStyle(endRow))
				quantityColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			case detailsPrice:
				_ = col.WriteCell(endRow, sd.Price, s.GetStyles().CurrencyStyle(endRow))
				priceColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			case detailsDividends:
				_ = col.WriteCell(endRow, sd.Dividends, s.GetStyles().CurrencyStyle(endRow))
				dividendColumn, err = excelize.ColumnNumberToName(k + 1)
				if err != nil {
					logrus.Error(err.Error())
					return err
				}
			}
		}
	}
	logrus.Debug(startRow, ":", endRow)

	priceChart := chart_builder.ChartBuilder{
		File:          s.GetExcelizeFile(),
		WorksheetName: worksheetName,
		Title:         "Price",
		Height:        300,
		Width:         500,
		Type:          excelize.Line,
	}
	priceChart.AddValueSeries(priceColumn, startRow+1, priceColumn, endRow)
	priceChart.AddCategorySeries(dateColumn, startRow+1, dateColumn, endRow)
	if err := priceChart.BuildChart("f3"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	valuesChart := chart_builder.ChartBuilder{
		File:          s.GetExcelizeFile(),
		WorksheetName: worksheetName,
		Title:         "Values",
		Height:        300,
		Width:         500,
		Type:          excelize.Line,
	}
	valuesChart.AddValueSeries(valuesColumn, startRow+1, valuesColumn, endRow)
	valuesChart.AddCategorySeries(dateColumn, startRow+1, dateColumn, endRow)
	if err := valuesChart.BuildChart("f20"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	quantityChart := chart_builder.ChartBuilder{
		File:          s.GetExcelizeFile(),
		WorksheetName: worksheetName,
		Title:         "Quantity",
		Height:        300,
		Width:         500,
		Type:          excelize.Line,
	}
	quantityChart.AddValueSeries(quantityColumn, startRow+1, quantityColumn, endRow)
	quantityChart.AddCategorySeries(dateColumn, startRow+1, dateColumn, endRow)
	if err := quantityChart.BuildChart("n3"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	dividendChart := chart_builder.ChartBuilder{
		File:          s.GetExcelizeFile(),
		WorksheetName: worksheetName,
		Title:         "Dividends",
		Height:        300,
		Width:         500,
		Type:          excelize.Col3D,
	}
	dividendChart.AddValueSeries(dividendColumn, startRow+1, dividendColumn, endRow)
	dividendChart.AddCategorySeries(dateColumn, startRow+1, dateColumn, endRow)
	if err := dividendChart.BuildChart("n20"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	return nil
}
