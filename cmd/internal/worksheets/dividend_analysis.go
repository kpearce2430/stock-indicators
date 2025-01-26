package worksheets

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"sort"
	"time"
)

var errNoSymbolsFound = fmt.Errorf("no symbols found")

func (w *WorkSheet) getSortedSymbols() ([]string, map[string]string, error) {
	var sortedSymbols []string
	symbolList, err := model.SymbolList(context.Background(), w.PGXConn, w.Lookups)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return sortedSymbols, symbolList, err
	}

	for k, v := range symbolList {
		logrus.Debug("k>", k, " v>", v)
		if k != "" {
			sortedSymbols = append(sortedSymbols, k)
		}
	}
	sort.Strings(sortedSymbols)
	return sortedSymbols, symbolList, nil
}

func (w *WorkSheet) dividendAnalysisForMonth(symbol string, month, year int) (*model.DividendEntry, error) {

	divEntry, err := model.GetDividendEntryForYearMonth(w.PGXConn, symbol, year, month)
	//divEntry := model.DividendEntry{
	//	Symbol: symbol,
	//	Month:  month,
	//	Year:   year,
	//}
	//
	//err := divEntry.GetDividendForYearMonth(w.PGXConn)
	if err != nil {
		logrus.Error(err.Error())
	}
	return divEntry, err
}

func (w *WorkSheet) dividendTicker(dchan chan []byte, symbol string, monthsAgo int) {

	start := time.Now()
	year := start.Year()
	month := int(start.Month())
	tickerHistory := model.NewDividendHistory(symbol)

	for i := 0; i < monthsAgo; i++ {
		logrus.Debug("Doing:", symbol, ",", year, ",", month)
		divEntry, err := model.GetDividendEntryForYearMonth(w.PGXConn, symbol, year, month)
		if err != nil {
			logrus.Error(err.Error())
		}
		//divEntry := model.NewDividendEntry(symbol, year, month)
		//if err := divEntry.GetDividendForYearMonth(w.PGXConn); err != nil {
		//	logrus.Error(err.Error())
		//}
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

func (w *WorkSheet) accountInfo(aChan chan []byte, symbol string) {
	acctInfo, err := model.AccountInfoGet(context.Background(), w.PGXConn, symbol)
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

func (w *WorkSheet) DividendAnalysis(worksheetName string, start time.Time, monthsAgo int) error {
	//
	divChannel := make(chan []byte)
	logrus.Debug(worksheetName, ":", fmt.Sprintf("%4d-%02d-%02d", start.Year(), start.Month(), start.Day()))
	_, err := w.File.NewSheet(worksheetName)
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
	historyMatrix := make(map[string]*model.DividendHistory)
	for _, symbol := range sortedSymbols {
		go w.dividendTicker(divChannel, symbol, monthsAgo)
	}

	for {
		var divHistory model.DividendHistory
		data, ok := <-divChannel
		logrus.Debug(string(data))
		err := json.Unmarshal(data, &divHistory)
		if err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}

		if ok == false {
			break
		}
		historyMatrix[divHistory.Symbol] = &divHistory
		if len(sortedSymbols) == len(historyMatrix) {
			break
		}
	}

	// Pull the account info data
	acctInfoMap := make(map[string]*model.AccountInfo)
	acctInfoChannel := make(chan []byte)
	for _, symbol := range sortedSymbols {
		go w.accountInfo(acctInfoChannel, symbol)
	}

	for {
		var acctInfo model.AccountInfo
		data, ok := <-acctInfoChannel
		logrus.Debug(string(data))
		if err := json.Unmarshal(data, &acctInfo); err != nil {
			logrus.Error("Error:", err.Error())
			return err
		}

		if ok == false {
			break
		}
		acctInfoMap[acctInfo.Symbol] = &acctInfo
		if len(sortedSymbols) == len(acctInfoMap) {
			break
		}
	}

	var allColumns []*ColumnInfo
	colInfoSymbol, err := NewColumnInfo(w.File, Symbol, worksheetName, 1)
	allColumns = append(allColumns, colInfoSymbol)

	month := int(start.Month())
	col := 2
	for i := 0; i < monthsAgo; i++ {
		var colMonth *ColumnInfo
		monthString := time.Month(month).String()
		monthString = monthString[0:3]
		colMonth, err = NewColumnInfo(w.File, fmt.Sprintf("%s", monthString), worksheetName, col)
		month--
		if month < 1 {
			month = 12
		}

		colMonth.SetMaxSize(9)
		allColumns = append(allColumns, colMonth)
		col++
	}

	row := 1
	for _, colInfo := range allColumns {
		_ = colInfo.WriteHeader(row, w.styles.Header)
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
		for i, colInfo := range allColumns {
			switch i {
			case 0:
				_ = colInfo.WriteCell(row, symbol, w.styles.TextStyle(row))
				security := symbolList[symbol]
				var msg string
				if acctInfo.NumberOfShares > 1 {
					// strconv.FormatFloat(acctInfo.NumberOfShares, 'f', -1, 64)

					shares := fmt.Sprintf("%.4f", acctInfo.NumberOfShares)
					for shares[len(shares)-1] == '0' {
						shares = shares[0 : len(shares)-1]
					}

					if shares[len(shares)-1] == '.' {
						shares = shares[0 : len(shares)-1]
					}
					var comments []string
					comments = append(comments, security)
					comments = append(comments, "Shares: "+shares)
					err = colInfo.AddComments(row, "kep", comments)
					if err != nil {
						logrus.Error("Error:", err.Error())
						return err
					}
				} else {
					msg = fmt.Sprintf("%s\nShares: None", security)

					err = colInfo.AddComment(row, "kep", "Security: ", msg)
					if err != nil {
						logrus.Error("Error:", err.Error())
						return err
					}
				}
			default:
				entry := symbolHistory.DividendEntries[i-1]
				logrus.Debug(symbol, " > ", entry.Year, "/", entry.Month, " [", entry.Amount, "]")
				_ = colInfo.WriteCell(row, entry.Amount, w.styles.AccountingStyle(row))
			}
		}
	}

	lastRow := row
	row++
	for i, colInfo := range allColumns {
		switch i {
		case 0:
			_ = colInfo.WriteCell(row, "Total", w.styles.TextStyle(row))
		default:
			colInfo.SetFormula(true)
			formula := fmt.Sprintf("=sum($%s$2:$%s%d)", colInfo.ColumnID, colInfo.ColumnID, lastRow)
			logrus.Debug(formula)
			_ = colInfo.WriteCell(row, formula, w.styles.CurrencyStyle(row))
		}
		if err = colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return err
		}
	}

	numSeries := monthsAgo / 12

	dividendChart := ChartBuilder{
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

	if err = dividendChart.BuildChart(w, "e3"); err != nil {
		logrus.Error(err.Error())
		return err
	}

	// Year over Year Summary
	summaryRow := row + 2
	colA := allColumns[0]
	colB := allColumns[1]
	for i := 0; i < numSeries; i++ {
		_ = colA.WriteCell(summaryRow+i, fmt.Sprintf("%d", start.Year()-i), w.styles.TextStyle(summaryRow+i))

		startCol := allColumns[1+(i*12)]
		endCol := allColumns[12+(i*12)]
		formula := fmt.Sprintf("=sum(%s%d:%s%d)", startCol.ColumnID, row, endCol.ColumnID, row)
		colB.SetFormula(true)
		_ = colB.WriteCell(summaryRow+i, formula, w.styles.CurrencyStyle(summaryRow+i))
	}

	yoyDividendChart := ChartBuilder{
		WorksheetName: worksheetName,
		Title:         "Year Over Year Dividends",
		Type:          excelize.Col,
		Height:        250,
		Width:         450,
		VaryColors:    true,
	}
	yoyDividendChart.AddValueSeries(colB.ColumnID, summaryRow, colB.ColumnID, summaryRow+numSeries-1)
	yoyDividendChart.AddCategorySeries(colA.ColumnID, summaryRow, colA.ColumnID, summaryRow+numSeries-1)
	if err = yoyDividendChart.BuildChart(w, fmt.Sprintf("%s%d", allColumns[3].ColumnID, summaryRow+numSeries)); err != nil {
		logrus.Error(err.Error())
		return err
	}

	return w.YearOverYearDividendNew(fmt.Sprintf("%s-YoY", worksheetName), worksheetName, start.Year(), int(start.Month()))
}

func reverse(cells []string) []string {
	for i := 0; i < len(cells)/2; i++ {
		j := len(cells) - i - 1
		cells[i], cells[j] = cells[j], cells[i]
	}
	return cells
}

func (w *WorkSheet) YearOverYearDividend(worksheetName, divAnalysisWorksheet string, totalsRow, curYear, curMonth int) error {

	_, err := w.File.GetSheetIndex(divAnalysisWorksheet)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	_, err = w.File.NewSheet(worksheetName)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	var allColumns []*ColumnInfo
	colInfoSymbol, err := NewColumnInfo(w.File, "Year", worksheetName, 1)
	allColumns = append(allColumns, colInfoSymbol)
	for month := 1; month <= 12; month++ {
		var colMonth *ColumnInfo
		monthString := time.Month(month).String()
		monthString = monthString[0:3]
		colMonth, err = NewColumnInfo(w.File, fmt.Sprintf("%s", monthString), worksheetName, month+1)
		colMonth.SetMaxSize(10)
		colMonth.SetFormula(true)
		allColumns = append(allColumns, colMonth)
	}

	colInfo, err := NewColumnInfo(w.File, "Total", worksheetName, 14)
	allColumns = append(allColumns, colInfo)
	colInfo.SetMaxSize(12)
	colInfo.SetFormula(true)
	ciTotalNum := len(allColumns) - 1

	row := 1
	for _, colInfo := range allColumns {
		_ = colInfo.WriteHeader(row, w.styles.Header)
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

		ci := allColumns[0]
		_ = ci.WriteCell(row, wYear, w.styles.TextStyle(row))

		for i := len(yData) - 1; i >= 0; i-- {
			ci = allColumns[i+1]
			_ = ci.WriteCell(row, yData[i], w.styles.CurrencyStyle(row))
		}

		ciTotal := allColumns[ciTotalNum]
		_ = ciTotal.WriteCell(row, fmt.Sprintf("=sum(%s%d:%s%d)", sumStart, row, sumEnd, row), w.styles.CurrencyStyle(row))
		row++
	}

	for _, colInfo = range allColumns {
		if err = colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return err
		}
	}
	return nil
}

func (w *WorkSheet) YearOverYearDividendNew(worksheetName, divAnalysisWorksheet string, curYear, curMonth int) error {

	_, err := w.File.GetSheetIndex(divAnalysisWorksheet)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	_, err = w.File.NewSheet(worksheetName)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	var allColumns []*ColumnInfo
	colInfoSymbol, err := NewColumnInfo(w.File, "Year", worksheetName, 1)
	allColumns = append(allColumns, colInfoSymbol)
	for month := 1; month <= 12; month++ {
		var colMonth *ColumnInfo
		monthString := time.Month(month).String()
		// monthString = monthString[0:3]
		colMonth, err = NewColumnInfo(w.File, fmt.Sprintf("%s", monthString), worksheetName, month+1)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		colMonth.SetMaxSize(10)
		// colMonth.SetFormula(true)
		allColumns = append(allColumns, colMonth)
	}

	colInfo, err := NewColumnInfo(w.File, "Total", worksheetName, 14)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	allColumns = append(allColumns, colInfo)
	colInfo.SetMaxSize(12)
	colInfo.SetFormula(true)
	ciTotalNum := len(allColumns) - 1

	sumStart, _ := excelize.ColumnNumberToName(2)
	sumEnd, _ := excelize.ColumnNumberToName(13)

	row := 1
	for _, colInfo := range allColumns {
		_ = colInfo.WriteHeader(row, w.styles.Header)
	}

	row++
	for year := 2021; year <= curYear; year++ {
		ci := allColumns[0]
		_ = ci.WriteCell(row, fmt.Sprintf("%d", year), w.styles.TextStyle(row))
		for month := 1; month <= 12; month++ {
			if year == curYear && month > curMonth {
				break
			}
			dh, err := model.DividendHistoryFromDB(context.Background(), w.PGXConn, "", year, month)
			// logrus.Info("dh>", dh.Sum())
			if err != nil {
				logrus.Error(err.Error())
				return err
			}
			ci := allColumns[month]
			err = ci.WriteCell(row, dh.Sum(), w.styles.CurrencyStyle(row))
			if err != nil {
				logrus.Error(err.Error())
				return err
			}
		}
		ciTotal := allColumns[ciTotalNum]
		_ = ciTotal.WriteCell(row, fmt.Sprintf("=sum(%s%d:%s%d)", sumStart, row, sumEnd, row), w.styles.CurrencyStyle(row))
		row++
	}
	for _, colInfo = range allColumns {
		if err = colInfo.SetColumnSize(); err != nil {
			logrus.Error(err.Error())
			return err
		}
	}
	return nil
}
