package dividend_analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model/account_info"
	"github.com/kpearce2430/stock-tools/model/dividend_history"
	"github.com/kpearce2430/stock-tools/stocksheet"
	"github.com/kpearce2430/stock-tools/stocksheet/chart_builder"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type DividendAnalysis struct {
	worksheets.WorksheetInterface
}

func New(w worksheets.WorksheetInterface) *DividendAnalysis {
	return &DividendAnalysis{w}
}

var (
	errNoSymbolsFound = errors.New("no symbols found")
)

// reverse reverses the order of a slice of strings.
func reverse(cells []string) []string {
	for i := 0; i < len(cells)/2; i++ {
		j := len(cells) - i - 1
		cells[i], cells[j] = cells[j], cells[i]
	}
	return cells
}

// accountInfoGet gets all the Account Information for the list of symbols.
func (d *DividendAnalysis) accountInfoGet(ctx context.Context, symbols []string) (map[string]*account_info.AccountInfo, error) {
	// Pull the account info data
	acctInfoMap := make(map[string]*account_info.AccountInfo)
	acctInfoChannel := make(chan []byte)
	for _, symbol := range symbols {
		go d.accountInfoWithContext(ctx, acctInfoChannel, symbol)
	}

	for {
		var acctInfo account_info.AccountInfo
		data, ok := <-acctInfoChannel
		if err := json.Unmarshal(data, &acctInfo); err != nil {
			logrus.Error("Error:", err.Error())
			return acctInfoMap, err
		}

		if ok == false {
			break
		}
		acctInfoMap[acctInfo.Symbol] = &acctInfo
		if len(symbols) == len(acctInfoMap) {
			break
		}
	}
	return acctInfoMap, nil
}

// accountInfoWithContext retrieves the Account Information for `symbol` and sends it in `aChan` to be read.
func (d *DividendAnalysis) accountInfoWithContext(ctx context.Context, aChan chan []byte, symbol string) {
	acctInfo, err := account_info.AccountInfoGet(ctx, d.GetPGXConn(), symbol)
	if err != nil {
		logrus.Error("Error:", err.Error())
		// panic(err.Error())
		aChan <- []byte("errors")
	}
	data, err := json.Marshal(acctInfo)
	if err != nil {
		aChan <- []byte("errors")
	}
	aChan <- data
}

// accountInfo calls accountInfoWithContext with the background context.
func (d *DividendAnalysis) accountInfo(aChan chan []byte, symbol string) {
	d.accountInfoWithContext(context.Background(), aChan, symbol)
}

// getSortedSymbolsWithContext returns a sorted list of symbols, the map of symbols to their name.  It will
// return an error if one is encountered.
func (d *DividendAnalysis) getSortedSymbolsWithContext(ctx context.Context) ([]string, map[string]string, error) {
	var sortedSymbols []string
	symbolList, err := account_info.SymbolList(ctx, d.GetPGXConn(), d.GetLookups())
	if err != nil {
		logrus.Error("Error:", err.Error())
		return sortedSymbols, symbolList, err
	}

	for k, _ := range symbolList {
		// logrus.Debug("k>", k, " v>", v)
		if k != "" {
			sortedSymbols = append(sortedSymbols, k)
		}
	}
	sort.Strings(sortedSymbols)
	return sortedSymbols, symbolList, nil
}

// getSortedSymbols calls getSortedSymbolsWithContext with a background context.
func (d *DividendAnalysis) getSortedSymbols() ([]string, map[string]string, error) {
	return d.getSortedSymbolsWithContext(context.Background())
}

// dividendTicker is a background function that will create a model.DividendHistory struct and return it through the `dchan`.
func (d *DividendAnalysis) dividendTicker(ctx context.Context, dchan chan []byte, symbol string, monthsAgo int) {
	start := time.Now()
	year := start.Year()
	month := int(start.Month())
	tickerHistory := dividend_history.NewDividendHistory(d.GetPGXConn(), symbol)

	for i := 0; i < monthsAgo; i++ {
		logrus.Debug("Doing:", symbol, ",", year, ",", month)
		err := tickerHistory.GetYearMonth(ctx, year, month)
		if err != nil {
			logrus.Error(err.Error())
		}

		month = month - 1
		if month < 1 {
			year--
			month = 12
		}
	}

	data, err := json.Marshal(tickerHistory)
	if err != nil {
		dchan <- []byte("errors")
	}
	dchan <- data
}

// dividendHistoryGet will create background process to retrieve the dividend history.
func (d *DividendAnalysis) dividendHistoryGet(ctx context.Context, symbols []string, monthsAgo int) (map[string]*dividend_history.DividendHistory, error) {
	// Pull the history data
	divChannel := make(chan []byte)
	historyMatrix := make(map[string]*dividend_history.DividendHistory)
	for _, symbol := range symbols {
		go d.dividendTicker(ctx, divChannel, symbol, monthsAgo)
	}

	for {
		var divHistory dividend_history.DividendHistory
		data, ok := <-divChannel
		logrus.Debug(string(data))
		if err := json.Unmarshal(data, &divHistory); err != nil {
			logrus.Error("Error:", err.Error())
			return historyMatrix, err
		}

		if ok == false {
			break
		}
		historyMatrix[divHistory.Symbol] = &divHistory
		if len(symbols) == len(historyMatrix) {
			break
		}
	}
	return historyMatrix, nil
}

// dividendWorksheetHeaders will create the columns and headers for the dividend worksheet.
func (d *DividendAnalysis) dividendWorksheetHeaders(worksheet *stocksheet.StockSheet, start time.Time, monthsAgo, row, column int) error {
	colInfoSymbol, err := column_info.New(d.GetExcelizeFile(), "Symbol", worksheet.Name(), column)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	worksheet.AddColumn(colInfoSymbol)

	month := int(start.Month())
	year := start.Year()
	col := column + 1
	for i := 0; i < monthsAgo; i++ {
		// var colMonth *ColumnInfo
		monthString := time.Month(month).String()
		switch month {
		case 1:
			monthString = fmt.Sprintf("%s %d", monthString[:3], year)
		default:
			monthString = monthString[0:3]
		}

		colMonth, err := column_info.New(d.GetExcelizeFile(), fmt.Sprintf("%s", monthString), worksheet.Name(), col)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}
		month--
		if month < 1 {
			month = 12
			year--
		}

		colMonth.SetMaxSize(9)
		worksheet.AddColumn(colMonth)
		col++
	}

	for _, colInfo := range worksheet.Columns {
		_ = colInfo.WriteHeader(row, d.GetStyles().Header)
	}
	return nil
}

// dividendWorksheetRow will create an individual row for the dividend worksheet.
func (d *DividendAnalysis) dividendWorksheetRow(
	worksheet *stocksheet.StockSheet,
	acctInfo *account_info.AccountInfo,
	history *dividend_history.DividendHistory,
	symbol, security string, row int) error {

	for i, colInfo := range worksheet.Columns {
		switch i {
		case 0:
			_ = colInfo.WriteCell(row, symbol, d.GetStyles().TextStyle(row))
			var comments []string
			if acctInfo.NumberOfShares > 1 {
				shares := fmt.Sprintf("%.4f", acctInfo.NumberOfShares)
				for shares[len(shares)-1] == '0' {
					shares = shares[0 : len(shares)-1]
				}

				if shares[len(shares)-1] == '.' {
					shares = shares[0 : len(shares)-1]
				}

				comments = append(comments, security)
				comments = append(comments, "Shares: "+shares)
				comments = append(comments, "Type:"+acctInfo.SecurityType)
				err := colInfo.AddComments(row, "kep", comments)
				if err != nil {
					logrus.Error("Error:", err.Error())
					return err
				}
			} else {
				comments = append(comments, security)
				comments = append(comments, "Shares: None")
				comments = append(comments, "Type:"+acctInfo.SecurityType)
				err := colInfo.AddComments(row, "kep", comments)
				if err != nil {
					logrus.Error("Error:", err.Error())
					return err
				}
			}
		default:
			entry := history.DividendEntries[i-1]
			logrus.Debug(symbol, " > ", entry.Year, "/", entry.Month, " [", entry.Amount, "]")
			_ = colInfo.WriteCell(row, entry.Amount, d.GetStyles().AccountingStyle(row))
		}
	}
	return nil
}

// dividendWorksheetTotalRow will create the Total row for the dividend worksheet.
func (d *DividendAnalysis) dividendWorksheetTotalRow(worksheet *stocksheet.StockSheet, row int) (int, error) {
	lastRow := row
	row++
	for i, colInfo := range worksheet.Columns {
		switch i {
		case 0:
			_ = colInfo.WriteCell(row, "Total", d.GetStyles().TextStyle(row))
		default:
			colInfo.SetFormula(true)
			formula := fmt.Sprintf("=sum($%s$2:$%s%d)", colInfo.ColumnID, colInfo.ColumnID, lastRow)
			logrus.Debug(formula)
			_ = colInfo.WriteCell(row, formula, d.GetStyles().AccountingStyle(row))
		}
		if err := colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return row, err
		}
	}
	return row, nil
}

// dividendWorksheetChart will create the big YoY chart for the worksheet.
func (d *DividendAnalysis) dividendWorksheetChart(worksheetName string, monthsAgo, row int) error {
	numSeries := monthsAgo / 12

	dividendChart := chart_builder.ChartBuilder{
		File:          d.GetExcelizeFile(),
		WorksheetName: worksheetName,
		Title:         "Dividend Analysis",
		Type:          excelize.Col,
		Height:        800,
		Width:         1000,
		VaryColors:    false,
	}

	for i := 0; i < numSeries; i++ {
		//=SERIES('Dividend Analysis'!$A$6,'Dividend Analysis'!$B$1:$M$1,'Dividend Analysis'!$B$6:$M$6,1)
		seriesStartCol := 2 + (i * 12)
		columnStart, err := excelize.ColumnNumberToName(seriesStartCol)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}

		seriesStopCol := 13 + (i * 12)
		columnEnd, err := excelize.ColumnNumberToName(seriesStopCol)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		dividendChart.AddValueSeries(columnStart, row, columnEnd, row)
		dividendChart.AddCategorySeries(columnStart, 1, columnEnd, 1)
	}

	if err := dividendChart.BuildChart("f4"); err != nil {
		logrus.Error(err.Error())
		return err
	}
	return nil
}

// dividendWorksheetChartYoY will create small YoY chart at the bottom of the worksheet.
func (d *DividendAnalysis) dividendWorksheetChartYoY(worksheet *stocksheet.StockSheet, start time.Time, numSeries, row int) error {
	// Year over Year Summary
	if len(worksheet.Columns) < 3 {
		return errors.New("not enough columns to create chart")
	}

	summaryRow := row + 2
	colA, _ := worksheet.GetColumn(0)
	colB, _ := worksheet.GetColumn(1)
	for i := 0; i < numSeries; i++ {
		_ = colA.WriteCell(summaryRow+i, fmt.Sprintf("%d", start.Year()-i), d.GetStyles().TextStyle(summaryRow+i))

		startCol, _ := worksheet.GetColumn(1 + (i * 12))
		endCol, _ := worksheet.GetColumn(12 + (i * 12))
		formula := fmt.Sprintf("=sum(%s%d:%s%d)", startCol.ColumnID, row, endCol.ColumnID, row)
		colB.SetFormula(true)
		_ = colB.WriteCell(summaryRow+i, formula, d.GetStyles().CurrencyStyle(summaryRow+i))
	}

	yoyDividendChart := chart_builder.ChartBuilder{
		File:          d.GetExcelizeFile(),
		WorksheetName: worksheet.Name(),
		Title:         "Year Over Year Dividends",
		Type:          excelize.Col,
		Height:        250,
		Width:         450,
		VaryColors:    true,
	}
	yoyDividendChart.AddValueSeries(colB.ColumnID, summaryRow, colB.ColumnID, summaryRow+numSeries-1)
	yoyDividendChart.AddCategorySeries(colA.ColumnID, summaryRow, colA.ColumnID, summaryRow+numSeries-1)
	col3, _ := worksheet.GetColumn(3)
	if err := yoyDividendChart.BuildChart(fmt.Sprintf("%s%d", col3.ColumnID, summaryRow+numSeries)); err != nil {
		logrus.Error(err.Error())
		return err
	}
	return nil
}

// DividendBonds creates a worksheet of only Bonds and their dividend/interest payments.
func (d *DividendAnalysis) DividendBonds(
	worksheetName string,
	start time.Time,
	monthsAgo int,
	symbols []string,
	symbolList map[string]string,
	acctInfoMap map[string]*account_info.AccountInfo,
	dividendHistory map[string]*dividend_history.DividendHistory) (int, error) {

	row := 1
	column := 1

	stockFile := d.GetFile()
	if stockFile == nil {
		logrus.Fatal("Unable to get stock file")
		return -1, errors.New("unable to get stock file")
	}
	sheet, err := stockFile.NewSheet(worksheetName)

	if err != nil {
		logrus.Error("Error:", err.Error())
		return 0, err
	}

	err = d.dividendWorksheetHeaders(sheet, start, monthsAgo, row, column)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return row, err
	}

	for _, symbol := range symbols {
		acctInfo, ok := acctInfoMap[symbol]
		if ok == false {
			logrus.Error("Error: missing symbol:", symbol)
			continue
		}

		if acctInfo.SecurityType != "Bond" {
			logrus.Debug("skipping security type:", acctInfo.SecurityType)
			continue
		}

		symbolHistory := dividendHistory[symbol]
		if symbolHistory.Sum() <= 0 {
			logrus.Debug("Skipping:", symbol)
			continue
		}

		row++
		security, ok := symbolList[symbol]
		if ok == false {
			logrus.Error("Error: missing symbol:", symbol)
			security = "Missing"
		}

		err = d.dividendWorksheetRow(sheet, acctInfo, symbolHistory, symbol, security, row)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return row, err
		}
	}

	lastRow, err := d.dividendWorksheetTotalRow(sheet, row)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return row, err
	}

	if err = d.dividendWorksheetChart(worksheetName, monthsAgo, lastRow); err != nil {
		logrus.Error("Error:", err.Error())
	}

	logrus.Debug("lastRow:", lastRow)
	return lastRow, nil
}

// DividendEquities creates a worksheet excluding Bonds and their dividend/interest payments.
func (d *DividendAnalysis) DividendEquities(
	worksheetName,
	bondsWorksheet string,
	start time.Time,
	monthsAgo int,
	bondsTotalRow int,
	symbols []string,
	symbolList map[string]string,
	acctInfoMap map[string]*account_info.AccountInfo,
	dividendHistory map[string]*dividend_history.DividendHistory) (int, error) {

	row := 1
	column := 1

	stockFile := d.GetFile()
	if stockFile == nil {
		logrus.Fatal("Unable to get stock file")
		return -1, errors.New("unable to get stock file")
	}
	sheet, err := stockFile.NewSheet(worksheetName)

	err = d.dividendWorksheetHeaders(sheet, start, monthsAgo, row, column)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return row, err
	}

	for _, symbol := range symbols {
		acctInfo, ok := acctInfoMap[symbol]
		if ok == false {
			logrus.Error("Error: missing symbol:", symbol)
			// return errors.New("missing symbol in account info")
			continue
		}
		if acctInfo.SecurityType == "Bond" {
			logrus.Debug("skipping security type:", acctInfo.SecurityType)
			continue
		}

		symbolHistory := dividendHistory[symbol]
		if symbolHistory.Sum() <= 0 {
			logrus.Debug("Skipping:", symbol)
			continue
		}

		row++
		security, ok := symbolList[symbol]
		if ok == false {
			logrus.Error("Error: missing symbol:", symbol)
			security = "Missing"
		}

		err = d.dividendWorksheetRow(sheet, acctInfo, symbolHistory, symbol, security, row)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return row, err
		}
	}

	// Bond totals goes here.
	for i, colInfo := range sheet.Columns {
		switch i {
		case 0:
			_ = colInfo.WriteCell(row, "Bond Total", d.GetStyles().TextStyle(row))
		default:
			colInfo.SetFormula(true)
			formula := fmt.Sprintf("='%s'!$%s$%d", bondsWorksheet, colInfo.ColumnID, bondsTotalRow)
			logrus.Debug(formula)
			_ = colInfo.WriteCell(row, formula, d.GetStyles().AccountingStyle(row))
		}
		if err = colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return row, err
		}
	}

	lastRow, err := d.dividendWorksheetTotalRow(sheet, row)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return row, err
	}

	if err = d.dividendWorksheetChart(worksheetName, monthsAgo, lastRow); err != nil {
		logrus.Error("Error:", err.Error())
	}

	if err = d.dividendWorksheetChartYoY(sheet, start, monthsAgo/12, lastRow); err != nil {
		logrus.Error("Error:", err.Error())
	}

	logrus.Debug("lastRow:", lastRow)
	return lastRow, nil
}

// DividendSheets will create the 3 excel worksheets for dividends.  It will create the Bonds worksheet via call to DividendBonds, pass the total row of
// the bond worksheet to the equities worksheet (via DividendEquities) and create it and finally the year-over-year worksheet via YearOverYearDividendNew.
func (d *DividendAnalysis) DividendSheets(ctx context.Context, worksheetName string, start time.Time, monthsAgo int) error {
	logrus.Debug(worksheetName, ":", fmt.Sprintf("%4d-%02d-%02d", start.Year(), start.Month(), start.Day()))

	sortedSymbols, symbolList, err := d.getSortedSymbolsWithContext(ctx)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	if len(sortedSymbols) == 0 {
		logrus.Error(errNoSymbolsFound.Error())
		return errNoSymbolsFound
	}

	logrus.Debug("SymbolList: ", len(symbolList))

	acctInfoMap, err := d.accountInfoGet(ctx, sortedSymbols)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	logrus.Debug("Account Info Map: ", len(acctInfoMap))

	dividendHistory, err := d.dividendHistoryGet(ctx, sortedSymbols, monthsAgo)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	logrus.Debug("Dividend History Map: ", len(dividendHistory))

	totalRow, err := d.DividendBonds(worksheetName+"-Bonds", start, monthsAgo, sortedSymbols, symbolList, acctInfoMap, dividendHistory)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	logrus.Debug("Bonds Total Row: ", totalRow)

	totalRow, err = d.DividendEquities(worksheetName+"-Equities", worksheetName+"-Bonds", start, monthsAgo, totalRow, sortedSymbols, symbolList, acctInfoMap, dividendHistory)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	logrus.Debug("Equities Total Row: ", totalRow)

	err = d.YearOverYearDividendNew(fmt.Sprintf("%s-YoY", worksheetName), worksheetName+"-Equities", start.Year(), int(start.Month()), totalRow)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	return nil
}

// DividendAnalysis is the original version that created the worksheets.  currently used for testing and historical reasons.
/*
func (d *DividendAnalysis) DividendAnalysis(worksheetName string, start time.Time, monthsAgo int) error {

	logrus.Info(worksheetName, ":", fmt.Sprintf("%4d-%02d-%02d", start.Year(), start.Month(), start.Day()))
	sheet, err := w.StockFile.NewSheet(worksheetName)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	sortedSymbols, symbolList, err := w.getSortedSymbols()
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	if len(sortedSymbols) == 0 {
		logrus.Error(errNoSymbolsFound.Error())
		return errNoSymbolsFound
	}

	// Pull the history data
	historyMatrix, err := w.dividendHistoryGet(context.Background(), sortedSymbols, monthsAgo)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	logrus.Info("History Matrix: ", len(historyMatrix))

	acctInfoMap, err := w.accountInfoGet(context.Background(), sortedSymbols)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	logrus.Info("Account Info: ", len(acctInfoMap))

	colInfoSymbol, err := column_info.New(w.StockFile.GetExcelizeFile(), Symbol, worksheetName, 1)
	sheet.AddColumn(colInfoSymbol)

	// Write the headers
	month := int(start.Month())
	col := 2
	for i := 0; i < monthsAgo; i++ {

		monthString := time.Month(month).String()
		monthString = monthString[0:3]
		colMonth, err := column_info.New(w.StockFile.GetExcelizeFile(), fmt.Sprintf("%s", monthString), worksheetName, col)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}
		month--
		if month < 1 {
			month = 12
		}

		colMonth.SetMaxSize(9)
		sheet.AddColumn(colMonth)
		col++
	}

	row := 1
	for _, colInfo := range sheet.Columns {
		_ = colInfo.WriteHeader(row, w.StockFile.Styles.Header)
	}

	for _, symbol := range sortedSymbols {
		symbolHistory := historyMatrix[symbol]
		if symbolHistory.Sum() <= 0 {
			logrus.Debug("Skipping:", symbol)
			continue
		}

		acctInfo, ok := acctInfoMap[symbol]
		if !ok {
			logrus.Error("Skipping:", symbol, " not found")
			continue
		}

		row++
		for i, colInfo := range sheet.Columns {
			switch i {
			case 0:
				_ = colInfo.WriteCell(row, symbol, w.StockFile.Styles.TextStyle(row))
				security := symbolList[symbol]
				var comments []string
				if acctInfo.NumberOfShares > 1 {
					shares := fmt.Sprintf("%.4f", acctInfo.NumberOfShares)
					for shares[len(shares)-1] == '0' {
						shares = shares[0 : len(shares)-1]
					}

					if shares[len(shares)-1] == '.' {
						shares = shares[0 : len(shares)-1]
					}

					comments = append(comments, security)
					comments = append(comments, "Shares: "+shares)
					comments = append(comments, "Type:"+acctInfo.SecurityType)
					err = colInfo.AddComments(row, "kep", comments)
					if err != nil {
						logrus.Error("Error:", err.Error())
						return err
					}
				} else {
					comments = append(comments, security)
					comments = append(comments, "Shares: None")
					comments = append(comments, "Type:"+acctInfo.SecurityType)
					err = colInfo.AddComments(row, "kep", comments)
					if err != nil {
						logrus.Error("Error:", err.Error())
						return err
					}
				}
			default:
				entry := symbolHistory.DividendEntries[i-1]
				logrus.Debug(symbol, " > ", entry.Year, "/", entry.Month, " [", entry.Amount, "]")
				_ = colInfo.WriteCell(row, entry.Amount, w.StockFile.Styles.AccountingStyle(row))
			}
		}
	}

	lastRow := row
	row++
	for i, colInfo := range sheet.Columns {
		switch i {
		case 0:
			_ = colInfo.WriteCell(row, "Total", w.StockFile.Styles.TextStyle(row))
		default:
			colInfo.SetFormula(true)
			formula := fmt.Sprintf("=sum($%s$2:$%s%d)", colInfo.ColumnID, colInfo.ColumnID, lastRow)
			logrus.Debug(formula)
			_ = colInfo.WriteCell(row, formula, w.StockFile.Styles.CurrencyStyle(row))
		}
		if err = colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return err
		}
	}

	numSeries := monthsAgo / 12

	dividendChart := chart_builder.ChartBuilder{
		File:          w.StockFile.GetExcelizeFile(),
		WorksheetName: worksheetName,
		Title:         "Dividend Analysis",
		Type:          excelize.Col,
		Height:        800,
		Width:         1000,
		VaryColors:    false,
		// xReverse:      true,
	}

	for i := 0; i < numSeries; i++ {
		//=SERIES('Dividend Analysis'!$A$6,'Dividend Analysis'!$B$1:$M$1,'Dividend Analysis'!$B$6:$M$6,1)
		seriesStartCol := 2 + (i * 12)
		columnStart, err := excelize.ColumnNumberToName(seriesStartCol)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}

		seriesStopCol := 13 + (i * 12)
		columnEnd, err := excelize.ColumnNumberToName(seriesStopCol)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		dividendChart.AddValueSeries(columnStart, row, columnEnd, row)
		dividendChart.AddCategorySeries(columnStart, 1, columnEnd, 1)
	}

	if err = dividendChart.BuildChart("f4"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	// Year over Year Summary
	summaryRow := row + 2
	colA, _ := sheet.GetColumn(0)
	colB, _ := sheet.GetColumn(1)
	for i := 0; i < numSeries; i++ {
		_ = colA.WriteCell(summaryRow+i, fmt.Sprintf("%d", start.Year()-i), w.StockFile.Styles.TextStyle(summaryRow+i))

		startCol, _ := sheet.GetColumn(1 + (i * 12))
		endCol, _ := sheet.GetColumn(12 + (i * 12))
		formula := fmt.Sprintf("=sum(%s%d:%s%d)", startCol.ColumnID, row, endCol.ColumnID, row)
		colB.SetFormula(true)
		_ = colB.WriteCell(summaryRow+i, formula, w.StockFile.Styles.CurrencyStyle(summaryRow+i))
	}

	yoyDividendChart := chart_builder.ChartBuilder{
		File:          w.StockFile.GetExcelizeFile(),
		WorksheetName: worksheetName,
		Title:         "Year Over Year Dividends",
		Type:          excelize.Col,
		Height:        250,
		Width:         450,
		VaryColors:    true,
	}
	yoyDividendChart.AddValueSeries(colB.ColumnID, summaryRow, colB.ColumnID, summaryRow+numSeries-1)
	yoyDividendChart.AddCategorySeries(colA.ColumnID, summaryRow, colA.ColumnID, summaryRow+numSeries-1)
	col3, _ := sheet.GetColumn(3)
	if err = yoyDividendChart.BuildChart(fmt.Sprintf("%s%d", col3.ColumnID, summaryRow+numSeries)); err != nil {
		logrus.Error(err.Error())
		return err
	}

	return w.YearOverYearDividendNew(fmt.Sprintf("%s-YoY", worksheetName), worksheetName, start.Year(), int(start.Month()), 0)

	return nil
}

*/

// YearOverYearDividend - only for testing and historical
func (d *DividendAnalysis) YearOverYearDividend(worksheetName, divAnalysisWorksheet string, totalsRow, curYear, curMonth int) error {
	//
	stockFile := d.GetFile()
	if stockFile == nil {
		logrus.Fatal("Unable to get stock file")
		return errors.New("unable to get stock file")
	}
	sheet, err := stockFile.NewSheet(worksheetName)

	colInfoSymbol, err := column_info.New(d.GetExcelizeFile(), "Year", worksheetName, 1)
	sheet.AddColumn(colInfoSymbol)
	for month := 1; month <= 12; month++ {
		monthString := time.Month(month).String()
		monthString = monthString[0:3]
		colMonth, err := column_info.New(d.GetExcelizeFile(), fmt.Sprintf("%s", monthString), worksheetName, month+1)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		colMonth.SetMaxSize(10)
		colMonth.SetFormula(true)
		sheet.AddColumn(colMonth)
	}

	colInfo, err := column_info.New(d.GetExcelizeFile(), "Total", worksheetName, 14)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	sheet.AddColumn(colInfo)
	colInfo.SetMaxSize(12)
	colInfo.SetFormula(true)
	ciTotalNum := len(sheet.Columns) - 1

	row := 1
	for _, colInfo = range sheet.Columns {
		_ = colInfo.WriteHeader(row, d.GetStyles().Header)
	}

	col := 2
	yearMap := make(map[int][]string)

	for wYear := curYear; wYear > curYear-4; wYear-- {
		var data []string
		// ='Dividend Analysis'!B57
		if wYear == curYear {
			for wMonth := curMonth; wMonth > 0; wMonth-- {
				colId, _ := excelize.ColumnNumberToName(col)
				cRef := fmt.Sprintf("='%s'!$%s$%d", divAnalysisWorksheet, colId, totalsRow)
				data = append(data, cRef)
				col++
			}
		} else {
			for wMonth := 12; wMonth > 0; wMonth-- {
				colId, _ := excelize.ColumnNumberToName(col)
				cRef := fmt.Sprintf("='%s'!$%s$%d", divAnalysisWorksheet, colId, totalsRow)
				data = append(data, cRef)
				col++
			}
		}
		yearMap[wYear] = reverse(data)
	}

	sumStart, _ := excelize.ColumnNumberToName(2)
	sumEnd, _ := excelize.ColumnNumberToName(13)

	row = 2
	for wYear := curYear - 3; wYear <= curYear; wYear++ {
		yData, ok := yearMap[wYear]
		if !ok {
			return fmt.Errorf("wYear %d not found in yearMap", wYear)
		}

		ci, _ := sheet.GetColumn(0)
		_ = ci.WriteCell(row, wYear, d.GetStyles().TextStyle(row))

		for i := len(yData) - 1; i >= 0; i-- {
			ci, _ = sheet.GetColumn(i + 1)
			_ = ci.WriteCell(row, yData[i], d.GetStyles().CurrencyStyle(row))
		}

		ciTotal, _ := sheet.GetColumn(ciTotalNum)
		_ = ciTotal.WriteCell(row, fmt.Sprintf("=sum(%s%d:%s%d)", sumStart, row, sumEnd, row), d.GetStyles().CurrencyStyle(row))
		row++
	}

	for _, colInfo = range sheet.Columns {
		if err = colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return err
		}
	}
	return nil
}

// YearOverYearDividendNew will create a worksheet by calendar month starting from 2021.
func (d *DividendAnalysis) YearOverYearDividendNew(worksheetName, divAnalysisWorksheet string, curYear, curMonth, equitiesTotalRow int) error {
	stockFile := d.GetFile()
	if stockFile == nil {
		logrus.Fatal("Unable to get stock file")
		return errors.New("unable to get stock file")
	}
	sheet, err := stockFile.NewSheet(worksheetName)

	colInfoSymbol, err := column_info.New(d.GetExcelizeFile(), "Year", worksheetName, 1)
	sheet.AddColumn(colInfoSymbol)
	for month := 1; month <= 12; month++ {
		monthString := time.Month(month).String()
		colMonth, err := column_info.New(d.GetExcelizeFile(), fmt.Sprintf("%s", monthString), worksheetName, month+1)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		colMonth.SetMaxSize(10)
		sheet.AddColumn(colMonth)
	}

	colInfo, err := column_info.New(d.GetExcelizeFile(), "Total", worksheetName, 14)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	sheet.AddColumn(colInfo)
	colInfo.SetMaxSize(12)
	colInfo.SetFormula(true)
	ciTotalNum := len(sheet.Columns) - 1

	sumStart, _ := excelize.ColumnNumberToName(2)
	sumEnd, _ := excelize.ColumnNumberToName(13)

	row := 1
	for _, colInfo = range sheet.Columns {
		_ = colInfo.WriteHeader(row, d.GetStyles().Header)
	}

	row++
	for year := 2021; year <= curYear; year++ {
		ci, _ := sheet.GetColumn(0)
		_ = ci.WriteCell(row, fmt.Sprintf("%d", year), d.GetStyles().TextStyle(row))
		for month := 1; month <= 12; month++ {
			ci, _ = sheet.GetColumn(month)
			if year == curYear && month > curMonth {
				break
			}
			// For the current month, use the total row of the Equities worksheet.
			if year == curYear && month == curMonth {
				// ='Dividend Analysis-Equities'!B47
				ci.SetFormula(true)

				formula := fmt.Sprintf("='%s'!$B$%d", divAnalysisWorksheet, equitiesTotalRow)
				logrus.Debug(curYear, ":", curMonth, ":", row, ">>>", formula)
				err = ci.WriteCell(row, formula, d.GetStyles().AccountingStyle(row))
				if err != nil {
					logrus.Error(err.Error())
				}
				break
			}
			dh, err := dividend_history.DividendHistoryFromDB(context.Background(), d.GetPGXConn(), "", year, month)
			if err != nil {
				logrus.Error(err.Error())
				return err
			}

			err = ci.WriteCell(row, dh.Sum(), d.GetStyles().AccountingStyle(row))
			if err != nil {
				logrus.Error(err.Error())
				return err
			}
		}
		ciTotal, _ := sheet.GetColumn(ciTotalNum)
		_ = ciTotal.WriteCell(row, fmt.Sprintf("=sum(%s%d:%s%d)", sumStart, row, sumEnd, row), d.GetStyles().CurrencyStyle(row))
		row++
	}
	for _, colInfo = range sheet.Columns {
		if err = colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return err
		}
	}

	return nil
}
