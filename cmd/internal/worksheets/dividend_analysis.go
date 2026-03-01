package worksheets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/kpearce2430/stock-tools/model"
	"github.com/kpearce2430/stock-tools/stocksheet"
	"github.com/kpearce2430/stock-tools/stocksheet/chart_builder"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

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
func (w *WorkSheet) accountInfoGet(ctx context.Context, symbols []string) (map[string]*model.AccountInfo, error) {
	// Pull the account info data
	acctInfoMap := make(map[string]*model.AccountInfo)
	acctInfoChannel := make(chan []byte)
	for _, symbol := range symbols {
		go w.accountInfoWithContext(ctx, acctInfoChannel, symbol)
	}

	for {
		var acctInfo model.AccountInfo
		data, ok := <-acctInfoChannel
		// logrus.Debug(string(data))
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
func (w *WorkSheet) accountInfoWithContext(ctx context.Context, aChan chan []byte, symbol string) {
	acctInfo, err := model.AccountInfoGet(ctx, w.PGXConn, symbol)
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
func (w *WorkSheet) accountInfo(aChan chan []byte, symbol string) {
	w.accountInfoWithContext(context.Background(), aChan, symbol)
}

// getSortedSymbolsWithContext returns a sorted list of symbols, the map of symbols to their name.  It will
// return an error if one is encountered.
func (w *WorkSheet) getSortedSymbolsWithContext(ctx context.Context) ([]string, map[string]string, error) {
	var sortedSymbols []string
	symbolList, err := model.SymbolList(ctx, w.PGXConn, w.Lookups)
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
func (w *WorkSheet) getSortedSymbols() ([]string, map[string]string, error) {
	return w.getSortedSymbolsWithContext(context.Background())
}

// dividendTicker is a background function that will create a model.DividendHistory struct and return it through the `dchan`.
func (w *WorkSheet) dividendTicker(ctx context.Context, dchan chan []byte, symbol string, monthsAgo int) {
	start := time.Now()
	year := start.Year()
	month := int(start.Month())
	tickerHistory := model.NewDividendHistory(symbol)

	for i := 0; i < monthsAgo; i++ {
		logrus.Debug("Doing:", symbol, ",", year, ",", month)
		divEntry, err := model.GetDividendEntryForYearMonth(ctx, w.PGXConn, symbol, year, month)
		if err != nil {
			logrus.Error(err.Error())
		}
		tickerHistory.DividendEntries = append(tickerHistory.DividendEntries, divEntry)

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
func (w *WorkSheet) dividendHistoryGet(ctx context.Context, symbols []string, monthsAgo int) (map[string]*model.DividendHistory, error) {
	// Pull the history data
	divChannel := make(chan []byte)
	historyMatrix := make(map[string]*model.DividendHistory)
	for _, symbol := range symbols {
		go w.dividendTicker(ctx, divChannel, symbol, monthsAgo)
	}

	for {
		var divHistory model.DividendHistory
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
func (w *WorkSheet) dividendWorksheetHeaders(worksheet *stocksheet.StockSheet, start time.Time, monthsAgo, row, column int) error {
	colInfoSymbol, err := column_info.New(w.StockFile.GetFile(), Symbol, worksheet.Name(), column)
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

		colMonth, err := column_info.New(w.StockFile.GetFile(), fmt.Sprintf("%s", monthString), worksheet.Name(), col)
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
		_ = colInfo.WriteHeader(row, w.StockFile.Styles.Header)
	}

	return nil
}

// dividendWorksheetRow will create an individual row for the dividend worksheet.
func (w *WorkSheet) dividendWorksheetRow(
	worksheet *stocksheet.StockSheet,
	acctInfo *model.AccountInfo,
	history *model.DividendHistory,
	symbol, security string, row int) error {

	for i, colInfo := range worksheet.Columns {
		switch i {
		case 0:
			_ = colInfo.WriteCell(row, symbol, w.StockFile.Styles.TextStyle(row))
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
			_ = colInfo.WriteCell(row, entry.Amount, w.StockFile.Styles.AccountingStyle(row))
		}
	}
	return nil
}

// dividendWorksheetTotalRow will create the Total row for the dividend worksheet.
func (w *WorkSheet) dividendWorksheetTotalRow(worksheet *stocksheet.StockSheet, row int) (int, error) {
	lastRow := row
	row++
	for i, colInfo := range worksheet.Columns {
		switch i {
		case 0:
			_ = colInfo.WriteCell(row, "Total", w.StockFile.Styles.TextStyle(row))
		default:
			colInfo.SetFormula(true)
			formula := fmt.Sprintf("=sum($%s$2:$%s%d)", colInfo.ColumnID, colInfo.ColumnID, lastRow)
			logrus.Debug(formula)
			_ = colInfo.WriteCell(row, formula, w.StockFile.Styles.AccountingStyle(row))
		}
		if err := colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return row, err
		}
	}
	return row, nil
}

// dividendWorksheetChart will create the big YoY chart for the worksheet.
func (w *WorkSheet) dividendWorksheetChart(worksheetName string, monthsAgo, row int) error {
	numSeries := monthsAgo / 12

	dividendChart := chart_builder.ChartBuilder{
		File:          w.StockFile.GetFile(),
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
func (w *WorkSheet) dividendWorksheetChartYoY(worksheet *stocksheet.StockSheet, start time.Time, numSeries, row int) error {
	// Year over Year Summary
	if len(worksheet.Columns) < 3 {
		return errors.New("not enough columns to create chart")
	}

	summaryRow := row + 2
	colA, _ := worksheet.GetColumn(0)
	colB, _ := worksheet.GetColumn(1)
	for i := 0; i < numSeries; i++ {
		_ = colA.WriteCell(summaryRow+i, fmt.Sprintf("%d", start.Year()-i), w.StockFile.Styles.TextStyle(summaryRow+i))

		startCol, _ := worksheet.GetColumn(1 + (i * 12))
		endCol, _ := worksheet.GetColumn(12 + (i * 12))
		formula := fmt.Sprintf("=sum(%s%d:%s%d)", startCol.ColumnID, row, endCol.ColumnID, row)
		colB.SetFormula(true)
		_ = colB.WriteCell(summaryRow+i, formula, w.StockFile.Styles.CurrencyStyle(summaryRow+i))
	}

	yoyDividendChart := chart_builder.ChartBuilder{
		File:          w.StockFile.GetFile(),
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
func (w *WorkSheet) DividendBonds(
	worksheetName string,
	start time.Time,
	monthsAgo int,
	symbols []string,
	symbolList map[string]string,
	acctInfoMap map[string]*model.AccountInfo,
	dividendHistory map[string]*model.DividendHistory) (int, error) {

	row := 1
	column := 1

	sheet, err := w.StockFile.NewSheet(worksheetName)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return 0, err
	}

	err = w.dividendWorksheetHeaders(sheet, start, monthsAgo, row, column)
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

		err = w.dividendWorksheetRow(sheet, acctInfo, symbolHistory, symbol, security, row)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return row, err
		}
	}

	lastRow, err := w.dividendWorksheetTotalRow(sheet, row)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return row, err
	}

	if err = w.dividendWorksheetChart(worksheetName, monthsAgo, lastRow); err != nil {
		logrus.Error("Error:", err.Error())
	}

	logrus.Debug("lastRow:", lastRow)
	return lastRow, nil
}

// DividendEquities creates a worksheet excluding Bonds and their dividend/interest payments.
func (w *WorkSheet) DividendEquities(
	worksheetName,
	bondsWorksheet string,
	start time.Time,
	monthsAgo int,
	bondsTotalRow int,
	symbols []string,
	symbolList map[string]string,
	acctInfoMap map[string]*model.AccountInfo,
	dividendHistory map[string]*model.DividendHistory) (int, error) {

	row := 1
	column := 1

	sheet, err := w.StockFile.NewSheet(worksheetName)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return 0, err
	}

	err = w.dividendWorksheetHeaders(sheet, start, monthsAgo, row, column)
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

		err = w.dividendWorksheetRow(sheet, acctInfo, symbolHistory, symbol, security, row)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return row, err
		}
	}

	// Bond totals goes here.
	for i, colInfo := range sheet.Columns {
		switch i {
		case 0:
			_ = colInfo.WriteCell(row, "Bond Total", w.StockFile.Styles.TextStyle(row))
		default:
			colInfo.SetFormula(true)
			formula := fmt.Sprintf("='%s'!$%s$%d", bondsWorksheet, colInfo.ColumnID, bondsTotalRow)
			logrus.Debug(formula)
			_ = colInfo.WriteCell(row, formula, w.StockFile.Styles.AccountingStyle(row))
		}
		if err = colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return row, err
		}
	}

	lastRow, err := w.dividendWorksheetTotalRow(sheet, row)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return row, err
	}

	if err = w.dividendWorksheetChart(worksheetName, monthsAgo, lastRow); err != nil {
		logrus.Error("Error:", err.Error())
	}

	if err = w.dividendWorksheetChartYoY(sheet, start, monthsAgo/12, lastRow); err != nil {
		logrus.Error("Error:", err.Error())
	}

	logrus.Debug("lastRow:", lastRow)
	return lastRow, nil
}

// DividendSheets will create the 3 excel worksheets for dividends.  It will create the Bonds worksheet via call to DividendBonds, pass the total row of
// the bond worksheet to the equities worksheet (via DividendEquities) and create it and finally the year-over-year worksheet via YearOverYearDividendNew.
func (w *WorkSheet) DividendSheets(ctx context.Context, worksheetName string, start time.Time, monthsAgo int) error {
	logrus.Debug(worksheetName, ":", fmt.Sprintf("%4d-%02d-%02d", start.Year(), start.Month(), start.Day()))

	sortedSymbols, symbolList, err := w.getSortedSymbolsWithContext(ctx)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	if len(sortedSymbols) == 0 {
		logrus.Error(errNoSymbolsFound.Error())
		return errNoSymbolsFound
	}

	logrus.Debug("SymbolList: ", len(symbolList))

	acctInfoMap, err := w.accountInfoGet(ctx, sortedSymbols)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	logrus.Debug("Account Info Map: ", len(acctInfoMap))

	dividendHistory, err := w.dividendHistoryGet(ctx, sortedSymbols, monthsAgo)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	logrus.Debug("Dividend History Map: ", len(dividendHistory))

	totalRow, err := w.DividendBonds(worksheetName+"-Bonds", start, monthsAgo, sortedSymbols, symbolList, acctInfoMap, dividendHistory)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	logrus.Debug("Bonds Total Row: ", totalRow)

	totalRow, err = w.DividendEquities(worksheetName+"-Equities", worksheetName+"-Bonds", start, monthsAgo, totalRow, sortedSymbols, symbolList, acctInfoMap, dividendHistory)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	logrus.Debug("Equities Total Row: ", totalRow)

	err = w.YearOverYearDividendNew(fmt.Sprintf("%s-YoY", worksheetName), worksheetName+"-Equities", start.Year(), int(start.Month()), totalRow)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}
	return nil
}

// DividendAnalysis is the original version that created the worksheets.  currently used for testing and historical reasons.
func (w *WorkSheet) DividendAnalysis(worksheetName string, start time.Time, monthsAgo int) error {
	//
	logrus.Debug(worksheetName, ":", fmt.Sprintf("%4d-%02d-%02d", start.Year(), start.Month(), start.Day()))
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

	acctInfoMap, err := w.accountInfoGet(context.Background(), sortedSymbols)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	colInfoSymbol, err := column_info.New(w.StockFile.GetFile(), Symbol, worksheetName, 1)
	sheet.AddColumn(colInfoSymbol)

	// Write the headers
	month := int(start.Month())
	col := 2
	for i := 0; i < monthsAgo; i++ {

		monthString := time.Month(month).String()
		monthString = monthString[0:3]
		colMonth, err := column_info.New(w.StockFile.GetFile(), fmt.Sprintf("%s", monthString), worksheetName, col)
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
		File:          w.StockFile.GetFile(),
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
		File:          w.StockFile.GetFile(),
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
}

// YearOverYearDividend - only for testing and historical
func (w *WorkSheet) YearOverYearDividend(worksheetName, divAnalysisWorksheet string, totalsRow, curYear, curMonth int) error {
	//
	sheet, err := w.StockFile.NewSheet(worksheetName)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	colInfoSymbol, err := column_info.New(w.StockFile.GetFile(), "Year", worksheetName, 1)
	sheet.AddColumn(colInfoSymbol)
	for month := 1; month <= 12; month++ {
		monthString := time.Month(month).String()
		monthString = monthString[0:3]
		colMonth, err := column_info.New(w.StockFile.GetFile(), fmt.Sprintf("%s", monthString), worksheetName, month+1)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		colMonth.SetMaxSize(10)
		colMonth.SetFormula(true)
		sheet.AddColumn(colMonth)
	}

	colInfo, err := column_info.New(w.StockFile.GetFile(), "Total", worksheetName, 14)
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
		_ = colInfo.WriteHeader(row, w.StockFile.Styles.Header)
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
		_ = ci.WriteCell(row, wYear, w.StockFile.Styles.TextStyle(row))

		for i := len(yData) - 1; i >= 0; i-- {
			ci, _ = sheet.GetColumn(i + 1)
			_ = ci.WriteCell(row, yData[i], w.StockFile.Styles.CurrencyStyle(row))
		}

		ciTotal, _ := sheet.GetColumn(ciTotalNum)
		_ = ciTotal.WriteCell(row, fmt.Sprintf("=sum(%s%d:%s%d)", sumStart, row, sumEnd, row), w.StockFile.Styles.CurrencyStyle(row))
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
func (w *WorkSheet) YearOverYearDividendNew(worksheetName, divAnalysisWorksheet string, curYear, curMonth, equitiesTotalRow int) error {

	//_, err := w.File.GetSheetIndex(divAnalysisWorksheet)
	//if err != nil {
	//	logrus.Error(err.Error())
	//	return err
	//}

	sheet, err := w.StockFile.NewSheet(worksheetName)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	colInfoSymbol, err := column_info.New(w.StockFile.GetFile(), "Year", worksheetName, 1)
	sheet.AddColumn(colInfoSymbol)
	for month := 1; month <= 12; month++ {
		monthString := time.Month(month).String()
		colMonth, err := column_info.New(w.StockFile.GetFile(), fmt.Sprintf("%s", monthString), worksheetName, month+1)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		colMonth.SetMaxSize(10)
		sheet.AddColumn(colMonth)
	}

	colInfo, err := column_info.New(w.StockFile.GetFile(), "Total", worksheetName, 14)
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
		_ = colInfo.WriteHeader(row, w.StockFile.Styles.Header)
	}

	row++
	for year := 2021; year <= curYear; year++ {
		ci, _ := sheet.GetColumn(0)
		_ = ci.WriteCell(row, fmt.Sprintf("%d", year), w.StockFile.Styles.TextStyle(row))
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
				err = ci.WriteCell(row, formula, w.StockFile.Styles.AccountingStyle(row))
				if err != nil {
					logrus.Error(err.Error())
				}
				break
			}
			dh, err := model.DividendHistoryFromDB(context.Background(), w.PGXConn, "", year, month)
			if err != nil {
				logrus.Error(err.Error())
				return err
			}
			err = ci.WriteCell(row, dh.Sum(), w.StockFile.Styles.AccountingStyle(row))
			if err != nil {
				logrus.Error(err.Error())
				return err
			}
		}
		ciTotal, _ := sheet.GetColumn(ciTotalNum)
		_ = ciTotal.WriteCell(row, fmt.Sprintf("=sum(%s%d:%s%d)", sumStart, row, sumEnd, row), w.StockFile.Styles.CurrencyStyle(row))
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
