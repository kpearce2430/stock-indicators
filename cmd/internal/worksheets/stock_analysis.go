package worksheets

import (
	"context"
	"fmt"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/polygon-io/client-go/rest/models"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"math"
	"sort"
	"time"
)

var numberSymbols = 0

const (
	AnnualReturn           = "Annual Return"
	AveragePrice           = "Average Price"
	CAGR                   = "CAGR"
	CurrentDividend        = "Current Dividend"
	DaysAgo                = "Days Ago"
	DividendsReceived      = "Dividends Received"
	DividendYield          = "Dividend Yield"
	FirstBought            = "First Bought"
	InterestIncome         = "Interest Income"
	LatestEarningsPerShare = "Latest EPS"
	LatestPrice            = "Latest Price"
	Name                   = "Name"
	Net                    = "Net"
	PercentageOfPortfolio  = "Percentage Portfolio"
	ProjectedDividends     = "Projected Dividends"
	ReturnOnInvestment     = "ROI"
	Symbol                 = "Symbol"
	TotalCost              = "Total Cost"
	TotalShares            = "Total Shares"
	TotalValue             = "Total Value"
	Type                   = "Type"
	YearlyDividend         = "Yearly Dividend"
)

func (w *WorkSheet) writeStockAnalysisDetailRow(row int, columnInfo []*ColumnInfo, tickerInfo *model.AccountInfo, julDate string) error {
	if tickerInfo.Symbol == "JENSX" {
		logrus.Debug("DetailRow:", tickerInfo.Symbol)
	}
	var (
		err                          error
		currentDividendRowCol        string
		daysOwnedColRow              string
		lastPriceColRow              string
		projectedDividendsCol        string
		projectedDividendsRowCol     string
		totalCostColRow              string
		totalDividendsReceivedColRow string
		totalInterestIncomeColRow    string
		totalSharesColRow            string
		totalValueCol                string
		totalValueColRow             string
		yearlyDividendColRow         string

		dividendInfo *models.Dividend
		stockInfo    *models.GetDailyOpenCloseAggResponse
	)

	if len(tickerInfo.Symbol) < 5 {
		dividendInfo, err = w.DividendCache.GetCacheSet(tickerInfo.Symbol)
		if err != nil {
			logrus.Error(tickerInfo.Symbol, " error ", err.Error())
			// return err
		}
	}

	if len(tickerInfo.Symbol) < 5 {
		stockInfo, err = w.StockCache.GetCache(tickerInfo.Symbol, julDate)
		if err != nil {
			logrus.Error(tickerInfo.Symbol, " error ", err.Error())
			return err
		}
	}

	for _, colInfo := range columnInfo {
		logrus.Debug("Working on :", colInfo.Name)
		switch colInfo.Name {
		case Name:
			err = colInfo.WriteCell(row, tickerInfo.Security, w.styles.TextStyle(row))

		case Symbol:
			err = colInfo.WriteCell(row, tickerInfo.Symbol, w.styles.TextStyle(row))

		case Type:
			err = colInfo.WriteCell(row, tickerInfo.SecurityType, w.styles.TextStyle(row))

		case TotalShares:
			err = colInfo.WriteCell(row, tickerInfo.NumberOfShares, w.styles.GeneralStyle(row))
			totalSharesColRow = colInfo.GetColRow(row)

		case LatestPrice:
			switch tickerInfo.SecurityType {
			case "Stock", "Other":
				if stockInfo == nil {
					logrus.Error("No stock info for ", tickerInfo.Symbol)
					return fmt.Errorf("no stock info for %s", tickerInfo.Symbol)
				}
				logrus.Debug("Latest price for ", tickerInfo.Symbol, " is $", stockInfo.Close)
				err = colInfo.WriteCell(row, stockInfo.Close, w.styles.CurrencyStyle(row))

			default: // Bond, Mutual Fund
				err = colInfo.WriteCell(row, tickerInfo.LatestPrice, w.styles.CurrencyStyle(row))
			}
			lastPriceColRow = colInfo.GetColRow(row)
			logrus.Debugln("lastPriceColRow:", lastPriceColRow)

		case TotalValue:
			formula := fmt.Sprintf("=%s*%s", totalSharesColRow, lastPriceColRow)
			logrus.Debugln("Formula>>", formula)
			err = colInfo.WriteCell(row, formula, w.styles.CurrencyStyle(row))
			totalValueColRow = colInfo.GetColRow(row)
			totalValueCol = colInfo.ColumnID

		case DividendsReceived:
			err = colInfo.WriteCell(row, tickerInfo.DividendsReceived, w.styles.CurrencyStyle(row))
			totalDividendsReceivedColRow = colInfo.GetColRow(row)

		case InterestIncome:
			err = colInfo.WriteCell(row, tickerInfo.InterestIncome, w.styles.CurrencyStyle(row))
			totalInterestIncomeColRow = colInfo.GetColRow(row)

		case TotalCost:
			err = colInfo.WriteCell(row, tickerInfo.NetCost, w.styles.CurrencyStyle(row))
			totalCostColRow = colInfo.GetColRow(row)

		case AveragePrice:
			err = colInfo.WriteCell(row, tickerInfo.AveragePrice, w.styles.CurrencyStyle(row))

		case CurrentDividend:
			var value float64
			if dividendInfo != nil {
				value = dividendInfo.CashAmount
			}
			currentDividendRowCol = colInfo.GetColRow(row)
			err = colInfo.WriteCell(row, value, w.styles.CurrencyStyle(row))

		case YearlyDividend:
			var value float64
			if dividendInfo != nil {
				value = dividendInfo.CashAmount * float64(dividendInfo.Frequency)
			}
			err = colInfo.WriteCell(row, value, w.styles.CurrencyStyle(row))
			yearlyDividendColRow = colInfo.GetColRow(row)

		case Net:
			formula := fmt.Sprintf("=(%s - %s) + (%s + %s)", totalValueColRow, totalCostColRow, totalDividendsReceivedColRow, totalInterestIncomeColRow)
			err = colInfo.WriteCell(row, formula, w.styles.CurrencyStyle(row))

		case FirstBought:
			firstBoughtStr := tickerInfo.FirstBought.Format("01-02-2006")
			err = colInfo.WriteCell(row, firstBoughtStr, w.styles.TextStyle(row))

		case DaysAgo:
			now := time.Now()
			diff := now.Sub(tickerInfo.FirstBought).Hours()
			diff = math.Floor(diff / 24.0)
			err = colInfo.WriteCell(row, diff, w.styles.TextStyle(row))
			daysOwnedColRow = colInfo.GetColRow(row)

		case LatestEarningsPerShare:
			fEps := 0.00
			// TODO: Find EPS
			//if stockInfo != nil {
			//	fEps = stockInfo.PeRatio
			//}
			err = colInfo.WriteCell(row, fEps, w.styles.CurrencyStyle(row))

		case ProjectedDividends:
			// =IF(R2> 0,R2*K2, (N2/V2) * 365) where
			// R2 - Yearly Dividend
			// K2 - Total Shares
			// N2 - Dividends Received
			// V2 - Days Owned
			formula := fmt.Sprintf("=IF(%s > 0, %s * %s, ((%s+%s) / %s ) * 365)",
				yearlyDividendColRow, yearlyDividendColRow, totalSharesColRow,
				totalDividendsReceivedColRow, totalInterestIncomeColRow, daysOwnedColRow)
			err = colInfo.WriteCell(row, formula, w.styles.CurrencyStyle(row))
			projectedDividendsRowCol = colInfo.GetColRow(row)
			projectedDividendsCol = colInfo.ColumnID
			if row == 2 {
				formula := fmt.Sprintf("=sum(%s2:%s%d)", projectedDividendsCol, projectedDividendsCol, numberSymbols+1)
				err = colInfo.WriteCell(numberSymbols+2, formula, w.styles.AccountingStyle(numberSymbols+2))
				logrus.Info("Projected Div Form:", formula, " location:", numberSymbols+2)
			}

		case DividendYield:
			// =IF(R2 > 0,R2/K2,IF((M2+N2) > 0, IF( Q2 = 0, (W2/J2)/K2,0),0))
			// =IF(R2 > 0,R2/L2,IF( N2 > 0, IF(Q2 = 0, (W2/K2) / L2,0),0)) where
			// R2 - Yearly Dividend
			// L2 - Latest Price
			// N2 - Dividends Received
			// Q2 - Current Dividend
			// W2 - Projected Dividends
			// K2 - Total Shares
			// L2 - Latest Price
			formula := fmt.Sprintf("=IF(%s > 0,%s/%s,IF((%s+%s) > 0, IF( %s = 0, (%s/%s)/%s,0),0))",
				yearlyDividendColRow, yearlyDividendColRow, lastPriceColRow,
				totalDividendsReceivedColRow, totalInterestIncomeColRow, currentDividendRowCol,
				projectedDividendsRowCol, totalSharesColRow, lastPriceColRow)
			err = colInfo.WriteCell(row, formula, w.styles.PercentStyle(row))

		case PercentageOfPortfolio:
			// = M2 / (SUM($M$2:$M$36)) where
			// M2 is the total value of the current row
			// SUM($M$2:$M$36) is the total value of all the tickers.
			formula := fmt.Sprintf("= %s / (SUM($%s$2:$%s$%d))",
				totalValueColRow, totalValueCol, totalValueCol, numberSymbols+1)
			err = colInfo.WriteCell(row, formula, w.styles.PercentStyle(row))

		case ReturnOnInvestment:
			// =IF(O2>0,(M2-O2) / O2, 0) where
			// O2 - Total Cost
			// M2 - Total Value
			formula := fmt.Sprintf("=IF(%s>0,(%s-%s)/%s,0)",
				totalCostColRow, totalValueColRow, totalCostColRow, totalCostColRow)
			err = colInfo.WriteCell(row, formula, w.styles.PercentStyle(row))

		case AnnualReturn:
			// =IF(O2>0,POWER((M2/O2),(365 / V2))-1, 0) where
			// O2 - Total Cost
			// M2 - Total Value
			// V2 - Days Owned
			formula := fmt.Sprintf("=IF(%s>0,POWER((%s/%s),(365/%s))-1,0)",
				totalCostColRow, totalValueColRow, totalCostColRow, daysOwnedColRow)
			err = colInfo.WriteCell(row, formula, w.styles.PercentStyle(row))

		case CAGR:
			//  =IF(O2>0,POWER(((M2+N2)/O2),(365 / V2))-1,0) where
			// O2 - Total Cost
			// M2 - Total Value
			// N2 - Dividends Received
			// V2 - Days Owned
			formula := fmt.Sprintf("=IF(%s>0,POWER(((%s+%s+%s)/%s),(365/%s))-1,0)",
				totalCostColRow, totalValueColRow, totalDividendsReceivedColRow, totalInterestIncomeColRow, totalCostColRow, daysOwnedColRow)
			err = colInfo.WriteCell(row, formula, w.styles.PercentStyle(row))

		default: // Assumed to be one of the accounts
			shares, ok := tickerInfo.Accounts[colInfo.Name]
			if !ok {
				shares = 0
			}
			if shares < 2 {
				shares = 0
			}
			err = colInfo.WriteCell(row, shares, w.styles.GeneralStyle(row))
		}
	}

	return err
}

func (w *WorkSheet) StockAnalysis(worksheetName, julDate string) error {

	if w.DividendCache == nil {
		return fmt.Errorf("no dividend cache loaded")
	}

	if w.StockCache == nil {
		return fmt.Errorf("no stock cache loaded")
	}

	if _, err := w.File.NewSheet(worksheetName); err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	symbolList, err := model.SymbolList(context.Background(), w.PGXConn, w.Lookups)
	if err != nil {
		return err
	}

	var sortedSymbols []string
	for k, _ := range symbolList {
		if k != "" {
			sortedSymbols = append(sortedSymbols, k)
		}
	}

	// Needed for the percentage of portfolio formula
	sort.Strings(sortedSymbols)
	var sortedAccounts []string
	accountList, err := model.AccountList(context.Background(), w.PGXConn)
	if err != nil {
		return err
	}
	for _, a := range accountList {
		if a[0] == 'z' {
			continue
		}
		sortedAccounts = append(sortedAccounts, a)
	}
	sort.Strings(sortedAccounts)

	var columnNames = []string{Name, Symbol, Type}
	for i := 0; i < len(sortedAccounts); i++ {
		columnNames = append(columnNames, sortedAccounts[i])
	}

	var remainingColumnTitles = []string{
		TotalShares,
		LatestPrice,
		TotalValue,
		DividendsReceived,
		InterestIncome,
		TotalCost,
		AveragePrice,
		CurrentDividend,
		YearlyDividend,
		LatestEarningsPerShare,
		Net,
		FirstBought,
		DaysAgo,
		ProjectedDividends,
		DividendYield,
		PercentageOfPortfolio,
		ReturnOnInvestment,
		AnnualReturn,
		CAGR,
	}
	for i := 0; i < len(remainingColumnTitles); i++ {
		columnNames = append(columnNames, remainingColumnTitles[i])
	}

	// Create Columns
	var allColumns []*ColumnInfo
	column := 1
	for i := 0; i < len(columnNames); i++ {
		colInfoName, err := NewColumnInfo(w.File, columnNames[i], worksheetName, column)
		if err != nil {
			return err
		}

		switch columnNames[i] {
		case Name:
			colInfoName.SetMaxSize(25)
		case TotalCost, TotalValue:
			colInfoName.SetMaxSize(15)
		default:
			colInfoName.SetMaxSize(10)
		}

		if columnNames[i] == TotalValue ||
			columnNames[i] == Net ||
			columnNames[i] == ProjectedDividends ||
			columnNames[i] == DividendYield ||
			columnNames[i] == PercentageOfPortfolio ||
			columnNames[i] == ReturnOnInvestment ||
			columnNames[i] == AnnualReturn ||
			columnNames[i] == CAGR {
			colInfoName.SetFormula(true)
		}
		column += 1
		allColumns = append(allColumns, colInfoName)
	}

	row := 1
	// Write Headers
	for _, ci := range allColumns {
		if err := ci.WriteHeader(row, w.styles.Header); err != nil {
			return err
		}
	}

	/*
	 * Details
	 */
	row += 1
	var activeSymbols []string
	symbolData := make(map[string]*model.AccountInfo)

	for _, symbol := range sortedSymbols {
		tickerInfo, err := model.AccountInfoGet(context.Background(), w.PGXConn, symbol, julDate)
		logrus.Debug("symbol [", symbol, "] shares [", tickerInfo.NumberOfShares, "]")

		if err != nil {
			logrus.Error(err.Error())
			return err
		}

		if tickerInfo.NumberOfShares < 2 {
			continue
		}

		activeSymbols = append(activeSymbols, symbol)
		symbolData[symbol] = tickerInfo
	}

	numberSymbols = len(activeSymbols)
	logrus.Debug("Number of symbols> ", numberSymbols)
	for _, symbol := range activeSymbols {
		tickerInfo := symbolData[symbol]
		// TODO: need juldate here
		if err := w.writeStockAnalysisDetailRow(row, allColumns, tickerInfo, julDate); err != nil {
			logrus.Error(err.Error())
			return err
		}
		row += 1
	}

	/*
	 * Finish Up
	 */
	for _, ci := range allColumns {
		_ = ci.SetColumnSize()
		switch ci.Name {
		case CAGR, AnnualReturn, ReturnOnInvestment:
			rangeRef := fmt.Sprintf("$%s$2:$%s$%d", ci.ColumnID, ci.ColumnID, numberSymbols+1)
			logrus.Debug("Range Ref> ", rangeRef)
			err := w.File.SetConditionalFormat(worksheetName, rangeRef, []excelize.ConditionalFormatOptions{
				//{
				//	Type:     "data_bar",
				//	Criteria: "=",
				//	MinType:  "min",
				//	MaxType:  "max",
				//	BarColor: "#7BC189",
				//	Value:    "0",
				//},
				{
					Type:     "3_color_scale",
					Criteria: "=",
					MinType:  "min",
					MidType:  "percentile",
					MaxType:  "max",
					MinColor: "#F8696B",
					MidColor: "#FFEB84",
					MaxColor: "#63BE7B",
				},
			})
			if err != nil {
				logrus.Error(err.Error())
			}
		case PercentageOfPortfolio:
			rangeRef := fmt.Sprintf("$%s$2:$%s$%d", ci.ColumnID, ci.ColumnID, numberSymbols+1)
			format, err := w.File.NewConditionalStyle(&excelize.Style{
				Font: &excelize.Font{
					Color:  "000000",
					Bold:   true,
					Family: "Calibri",
					Size:   14.0,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#7BC189"},
					Pattern: 1,
				},
			})
			if err != nil {
				logrus.Error(err.Error())
			}
			err = w.File.SetConditionalFormat(worksheetName, rangeRef, []excelize.ConditionalFormatOptions{
				{
					Type:     "top",
					Criteria: "=",
					Format:   format,
					Value:    "10",
					Percent:  true,
				},
			})
			if err != nil {
				logrus.Error(err.Error())
			}
		}
	}
	return nil
}
