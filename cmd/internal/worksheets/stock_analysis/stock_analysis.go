package stock_analysis

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model/account_info"
	"github.com/kpearce2430/stock-tools/model/dividends"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/kpearce2430/stock-tools/stocksheet"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/massive-com/client-go/v2/rest/models"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

var (
	numberSymbols = 0
)

type StockAnalysisWorksheet struct {
	worksheets.WorksheetInterface
	StockCache *stock_cache.Cache[models.GetDailyOpenCloseAggResponse]
}

const (
	AffectedDividend       = "Affected Dividend"
	AnnualReturn           = "Annual Return"
	AveragePrice           = "Average Price"
	CAGR                   = "CAGR"
	CurrentDividend        = "Current Dividend"
	CurrentAmount          = "Current Amount"
	DaysAgo                = "Days Ago"
	DividendsReceived      = "Dividends Received"
	DividendYield          = "Dividend Yield"
	Fidelity               = "To Fidelity"
	FidelityDividend       = "Fidelity Dividend"
	FirstBought            = "First Bought"
	InterestIncome         = "Interest Income"
	LatestEarningsPerShare = "Latest EPS"
	LatestPrice            = "Latest Price"
	Name                   = "Name"
	Net                    = "Net"
	NetDividendAmount      = "Net Dividend Amt"
	NetDividendChange      = "Net Dividend Change"
	PercentageOfPortfolio  = "Percentage Portfolio"
	ProjectedDividends     = "Projected Dividends"
	ReturnOnInvestment     = "ROI"
	Schwab                 = "To Schwab"
	SchwabDividend         = "Schwab Dividend"
	Symbol                 = "Symbol"
	TotalCost              = "Total Cost"
	TotalShares            = "Total Shares"
	TotalValue             = "Total Value"
	Trigger                = "Div < 3% and CAGR < 10%"
	Type                   = "Type"
	YearlyDividend         = "Yearly Dividend"
)

func New(w worksheets.WorksheetInterface) *StockAnalysisWorksheet {
	return &StockAnalysisWorksheet{
		WorksheetInterface: w,
		StockCache:         nil,
	}
}

func (s *StockAnalysisWorksheet) SetStockCache(cache *stock_cache.Cache[models.GetDailyOpenCloseAggResponse]) {
	s.StockCache = cache
}

func (s *StockAnalysisWorksheet) writeStockAnalysisDetailRow(worksheet *stocksheet.StockSheet, row int, tickerInfo *account_info.AccountInfo, julDate string, fidelityRow, schwabRow int) error {
	//
	var (
		err                          error
		affectedDividendColRow       string
		cagrColRow                   string
		currentDividendRowCol        string
		daysOwnedColRow              string
		dividendYieldColRow          string
		fidelityColRow               string
		fidelityYearlyDividend       string
		fidelityLatestPrice          string
		fidelityDividendColRow       string
		lastPriceColRow              string
		projectedDividendsCol        string
		projectedDividendsRowCol     string
		schwabColRow                 string
		schwabYearlyDividend         string
		schwabDividendColRow         string
		schwabLatestPrice            string
		totalCostColRow              string
		totalDividendsReceivedColRow string
		totalInterestIncomeColRow    string
		totalSharesColRow            string
		totalValueCol                string
		totalValueColRow             string
		triggerColRow                string
		typeColRow                   string
		yearlyDividendColRow         string

		// dividendInfo *models.Dividend
		dividendsSet dividends.DividendsSet
		stockInfo    *models.GetDailyOpenCloseAggResponse
	)

	var fidelityRows []string
	var schwabRows []string

	if len(tickerInfo.Symbol) < 5 {
		_ = dividendsSet.FromDBbySymbol(context.Background(), s.GetPGXConn(), "dividends", tickerInfo.Symbol)
	}

	if len(tickerInfo.Symbol) < 5 {
		stockInfo, err = s.StockCache.GetCache(tickerInfo.Symbol, julDate)
		if err != nil {
			logrus.Error(tickerInfo.Symbol, " error ", err.Error())
			return err
		}
	}

	for _, colInfo := range worksheet.Columns {
		logrus.Debug("Working on :", colInfo.Name)
		switch colInfo.Name {
		case Name:
			err = colInfo.WriteCell(row, tickerInfo.Security, s.GetStyles().TextStyle(row))

		case Symbol:
			err = colInfo.WriteCell(row, tickerInfo.Symbol, s.GetStyles().TextStyle(row))

		case Type:
			err = colInfo.WriteCell(row, tickerInfo.SecurityType, s.GetStyles().TextStyle(row))
			typeColRow = colInfo.GetColRow(row)

		case TotalShares:
			err = colInfo.WriteCell(row, tickerInfo.NumberOfShares, s.GetStyles().GeneralStyle(row))
			totalSharesColRow = colInfo.GetColRow(row)

		case LatestPrice:
			switch tickerInfo.SecurityType {
			case "Stock", "Other":
				if stockInfo == nil {
					logrus.Error("No stock info for ", tickerInfo.Symbol)
					return fmt.Errorf("no stock info for %s", tickerInfo.Symbol)
				}

				logrus.Debug("Latest price for ", tickerInfo.Symbol, " is $", stockInfo.Close)
				// TODO: Fix close being 0 for intraday
				if stockInfo.Close == 0 {
					err = colInfo.WriteCell(row, stockInfo.Open, s.GetStyles().CurrencyStyle(row))
				} else {
					err = colInfo.WriteCell(row, stockInfo.Close, s.GetStyles().CurrencyStyle(row))
				}

			default: // Bond, Mutual Fund
				err = colInfo.WriteCell(row, tickerInfo.LatestPrice, s.GetStyles().CurrencyStyle(row))
			}
			lastPriceColRow = colInfo.GetColRow(row)
			logrus.Debugln("lastPriceColRow:", lastPriceColRow)
			fidelityLatestPrice = fmt.Sprintf("$%s$%d", colInfo.ColumnID, fidelityRow)
			schwabLatestPrice = fmt.Sprintf("$%s$%d", colInfo.ColumnID, schwabRow)

		case TotalValue:
			formula := fmt.Sprintf("=%s*%s", totalSharesColRow, lastPriceColRow)
			// logrus.Debugln("Formula>>", formula)
			err = colInfo.WriteCell(row, formula, s.GetStyles().CurrencyStyle(row))
			totalValueColRow = colInfo.GetColRow(row)
			totalValueCol = colInfo.ColumnID

		case DividendsReceived:
			err = colInfo.WriteCell(row, tickerInfo.DividendsReceived, s.GetStyles().CurrencyStyle(row))
			totalDividendsReceivedColRow = colInfo.GetColRow(row)

		case InterestIncome:
			err = colInfo.WriteCell(row, tickerInfo.InterestIncome, s.GetStyles().CurrencyStyle(row))
			totalInterestIncomeColRow = colInfo.GetColRow(row)

		case TotalCost:
			err = colInfo.WriteCell(row, tickerInfo.NetCost, s.GetStyles().CurrencyStyle(row))
			totalCostColRow = colInfo.GetColRow(row)

		case AveragePrice:
			err = colInfo.WriteCell(row, tickerInfo.AveragePrice, s.GetStyles().CurrencyStyle(row))

		case CurrentDividend:
			var value float64
			if len(dividendsSet.Dividends) > 0 {
				value = dividendsSet.Dividends[0].CashAmount
			}
			currentDividendRowCol = colInfo.GetColRow(row)
			err = colInfo.WriteCell(row, value, s.GetStyles().CurrencyStyle(row))

		case YearlyDividend:
			var value float64
			if len(dividendsSet.Dividends) > 0 {
				value = dividendsSet.Dividends[0].CashAmount * float64(dividendsSet.Dividends[0].Frequency)
			}
			err = colInfo.WriteCell(row, value, s.GetStyles().CurrencyStyle(row))
			yearlyDividendColRow = colInfo.GetColRow(row)
			fidelityYearlyDividend = fmt.Sprintf("$%s$%d", colInfo.ColumnID, fidelityRow)
			schwabYearlyDividend = fmt.Sprintf("$%s$%d", colInfo.ColumnID, schwabRow)

		case Net:
			formula := fmt.Sprintf("=(%s - %s) + (%s + %s)", totalValueColRow, totalCostColRow, totalDividendsReceivedColRow, totalInterestIncomeColRow)
			err = colInfo.WriteCell(row, formula, s.GetStyles().CurrencyStyle(row))

		case FirstBought:
			firstBoughtStr := tickerInfo.FirstBought.Format("01-02-2006")
			err = colInfo.WriteCell(row, firstBoughtStr, s.GetStyles().TextStyle(row))

		case DaysAgo:
			now := time.Now()
			diff := now.Sub(tickerInfo.FirstBought).Hours()
			diff = math.Floor(diff / 24.0)
			err = colInfo.WriteCell(row, diff, s.GetStyles().TextStyle(row))
			daysOwnedColRow = colInfo.GetColRow(row)

		case LatestEarningsPerShare:
			fEps := 0.00
			// TODO: Find EPS
			//if stockInfo != nil {
			//	fEps = stockInfo.PeRatio
			//}
			err = colInfo.WriteCell(row, fEps, s.GetStyles().CurrencyStyle(row))

		case ProjectedDividends:
			// =IF(R2> 0,R2*K2, (N2/V2) * 365) where
			// R2 - Yearly Dividend
			// K2 - Total Shares
			// N2 - Dividends Received
			// V2 - Days Owned
			formula := fmt.Sprintf("=IF(%s > 0, %s * %s, ((%s+%s) / %s ) * 365)",
				yearlyDividendColRow, yearlyDividendColRow, totalSharesColRow,
				totalDividendsReceivedColRow, totalInterestIncomeColRow, daysOwnedColRow)
			err = colInfo.WriteCell(row, formula, s.GetStyles().CurrencyStyle(row))
			projectedDividendsRowCol = colInfo.GetColRow(row)
			projectedDividendsCol = colInfo.ColumnID
			if row == 2 {
				formula := fmt.Sprintf("=sum(%s2:%s%d)", projectedDividendsCol, projectedDividendsCol, numberSymbols+1)
				err = colInfo.WriteCell(numberSymbols+2, formula, s.GetStyles().AccountingStyle(numberSymbols+2))
				logrus.Debug("Projected Div Form:", formula, " location:", numberSymbols+2)
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
			err = colInfo.WriteCell(row, formula, s.GetStyles().PercentStyle(row))
			dividendYieldColRow = colInfo.GetColRow(row)

		case PercentageOfPortfolio:
			// = M2 / (SUM($M$2:$M$36)) where
			// M2 is the total value of the current row
			// SUM($M$2:$M$36) is the total value of all the tickers.
			formula := fmt.Sprintf("= %s / (SUM($%s$2:$%s$%d))",
				totalValueColRow, totalValueCol, totalValueCol, numberSymbols+1)
			err = colInfo.WriteCell(row, formula, s.GetStyles().PercentStyle(row))

		case ReturnOnInvestment:
			// =IF(O2>0,(M2-O2) / O2, 0) where
			// O2 - Total Cost
			// M2 - Total Value
			formula := fmt.Sprintf("=IF(%s>0,(%s-%s)/%s,0)",
				totalCostColRow, totalValueColRow, totalCostColRow, totalCostColRow)
			err = colInfo.WriteCell(row, formula, s.GetStyles().PercentStyle(row))

		case AnnualReturn:
			// =IF(O2>0,POWER((M2/O2),(365 / V2))-1, 0) where
			// O2 - Total Cost
			// M2 - Total Value
			// V2 - Days Owned
			formula := fmt.Sprintf("=IF(%s>0,POWER((%s/%s),(365/%s))-1,0)",
				totalCostColRow, totalValueColRow, totalCostColRow, daysOwnedColRow)
			err = colInfo.WriteCell(row, formula, s.GetStyles().PercentStyle(row))

		case CAGR:
			//  =IF(O2>0,POWER(((M2+N2)/O2),(365 / V2))-1,0) where
			// O2 - Total Cost
			// M2 - Total Value
			// N2 - Dividends Received
			// V2 - Days Owned
			formula := fmt.Sprintf("=IF(%s>0,POWER(((%s+%s+%s)/%s),(365/%s))-1,0)",
				totalCostColRow, totalValueColRow, totalDividendsReceivedColRow, totalInterestIncomeColRow, totalCostColRow, daysOwnedColRow)
			err = colInfo.WriteCell(row, formula, s.GetStyles().PercentStyle(row))
			cagrColRow = colInfo.GetColRow(row)

		case Trigger:
			// =IF(C2<>"Bond",IF(Y2<0.03,IF(AC2<0.1,TRUE,FALSE),FALSE),FALSE)
			formula := fmt.Sprintf("=IF(%s<>\"Bond\",IF(%s<.03,IF(%s<.10,TRUE,FALSE),FALSE),FALSE)",
				typeColRow, dividendYieldColRow, cagrColRow)
			err = colInfo.WriteCell(row, formula, s.GetStyles().TextStyle(row))
			triggerColRow = colInfo.GetColRow(row)

		case CurrentAmount:
			// =IF(AD2=TRUE,N2,0)
			formula := fmt.Sprintf("=IF(%s=TRUE,%s,0)", triggerColRow, totalValueColRow)
			err = colInfo.WriteCell(row, formula, s.GetStyles().AccountingStyle(row))
			currentAmountCol := colInfo.ColumnID
			if row == 2 {
				formula = fmt.Sprintf("=sum(%s2:%s%d)", currentAmountCol, currentAmountCol, numberSymbols+1)
				err = colInfo.WriteCell(numberSymbols+2, formula, s.GetStyles().AccountingStyle(numberSymbols+2))
			}

		case AffectedDividend:
			// =IF(AD2=TRUE,X2,0)
			formula := fmt.Sprintf("=IF(%s=TRUE,%s,0)", triggerColRow, projectedDividendsRowCol)
			err = colInfo.WriteCell(row, formula, s.GetStyles().AccountingStyle(row))
			affectedDividendCol := colInfo.ColumnID
			affectedDividendColRow = colInfo.GetColRow(row)
			if row == 2 {
				formula = fmt.Sprintf("=sum(%s2:%s%d)", affectedDividendCol, affectedDividendCol, numberSymbols+1)
				err = colInfo.WriteCell(numberSymbols+2, formula, s.GetStyles().AccountingStyle(numberSymbols+2))
			}

		case Fidelity:
			// =IF(AD3=TRUE,(SUM(D3)*m3)/$L$15,0)
			if len(fidelityRows) == 0 {
				logrus.Error("No Fidelity Accounts Found")
				return errors.New("no fidelity accounts found")
			}
			fidelitySumRows := s.BuildSumList(fidelityRows, row)

			formula := fmt.Sprintf("=IF(%s=TRUE,(SUM(%s)*%s)/%s,0)",
				triggerColRow, fidelitySumRows, lastPriceColRow, fidelityLatestPrice)
			err = colInfo.WriteCell(row, formula, s.GetStyles().NumberStyle(row))
			fidelityCol := colInfo.ColumnID
			fidelityColRow = colInfo.GetColRow(row)
			if row == 2 {
				logrus.Debug("Fidelity Formula:", formula)
				formula = fmt.Sprintf("=sum(%[1]s2:%[1]s%[2]d)", fidelityCol, numberSymbols+1)
				err = colInfo.WriteCell(numberSymbols+2, formula, s.GetStyles().NumberStyle(numberSymbols+2))
				logrus.Debug("Fidelity Formula:", formula)
			}

		case FidelityDividend:
			// =IF(AD2=TRUE,AG2*$S$15,0)
			formula := fmt.Sprintf("=if(%s=TRUE,(%s*%s),0)", triggerColRow, fidelityColRow, fidelityYearlyDividend)
			err = colInfo.WriteCell(row, formula, s.GetStyles().AccountingStyle(row))
			fidelityDividendCol := colInfo.ColumnID
			fidelityDividendColRow = colInfo.GetColRow(row)
			if row == 2 {
				formula = fmt.Sprintf("=sum(%[1]s2:%[1]s%[2]d)", fidelityDividendCol, numberSymbols+1)
				err = colInfo.WriteCell(numberSymbols+2, formula, s.GetStyles().AccountingStyle(numberSymbols+2))
			}

		case Schwab:
			// =IF(AD3=TRUE,(SUM(D3:F3)*L3)/$L$33,0)
			// ToDo: Columns are static.  Need to add a way determine which accounts are schwab
			if len(schwabRows) == 0 {
				logrus.Error("No Schwab Accounts Found")
				return errors.New("no schwab accounts found")
			}

			sumRows := s.BuildSumList(schwabRows, row)

			formula := fmt.Sprintf("=IF(%s=TRUE,(SUM(%s)*%s)/%s,0)", triggerColRow, sumRows, lastPriceColRow, schwabLatestPrice)
			err = colInfo.WriteCell(row, formula, s.GetStyles().NumberStyle(row))
			schwabCol := colInfo.ColumnID
			schwabColRow = colInfo.GetColRow(row)
			if row == 2 {
				formula = fmt.Sprintf("=sum(%[1]s2:%[1]s%[2]d)", schwabCol, numberSymbols+1)
				err = colInfo.WriteCell(numberSymbols+2, formula, s.GetStyles().NumberStyle(numberSymbols+2))
			}

		case SchwabDividend:
			// =IF(AD2=TRUE,AH2*$S$33,0)
			formula := fmt.Sprintf("=if(%s=TRUE,(%s*%s),0)", triggerColRow, schwabColRow, schwabYearlyDividend)
			err = colInfo.WriteCell(row, formula, s.GetStyles().AccountingStyle(row))
			schwabDividendCol := colInfo.ColumnID
			schwabDividendColRow = colInfo.GetColRow(row)
			if row == 2 {
				formula = fmt.Sprintf("=sum(%[1]s2:%[1]s%[2]d)", schwabDividendCol, numberSymbols+1)
				err = colInfo.WriteCell(numberSymbols+2, formula, s.GetStyles().AccountingStyle(numberSymbols+2))
			}

		case NetDividendAmount:
			formula := fmt.Sprintf("=if(%[1]s=TRUE,%[2]s+%[3]s,0)",
				triggerColRow, fidelityDividendColRow, schwabDividendColRow)
			err = colInfo.WriteCell(row, formula, s.GetStyles().AccountingStyle(row))
			fidelityCol := colInfo.ColumnID
			if row == 2 {
				formula = fmt.Sprintf("=sum(%[1]s2:%[1]s%[2]d)", fidelityCol, numberSymbols+1)
				err = colInfo.WriteCell(numberSymbols+2, formula, s.GetStyles().AccountingStyle(numberSymbols+2))
			}

		case NetDividendChange:
			// =IF(AD32=TRUE,IF(AF32>0,((AH32+AJ32)-AF32)/AF32,0),0)
			formula := fmt.Sprintf("=if(%[1]s=TRUE,if(%[2]s>0,((%[3]s+%[4]s)-%[5]s)/%[5]s,0),0)",
				triggerColRow, affectedDividendColRow, fidelityDividendColRow, schwabDividendColRow, affectedDividendColRow)
			err = colInfo.WriteCell(row, formula, s.GetStyles().PercentStyle(row))

		default: // Assumed to be one of the accounts
			shares, ok := tickerInfo.Accounts[colInfo.Name]
			if !ok {
				shares = 0
			}
			if shares < 2 {
				shares = 0
			}
			if strings.Contains(colInfo.Name, "Fidelity") {
				fidelityRows = append(fidelityRows, colInfo.ColumnID)
			} else {
				schwabRows = append(schwabRows, colInfo.ColumnID)
			}
			err = colInfo.WriteCell(row, shares, s.GetStyles().GeneralStyle(row))
		}
	}
	return err
}

func (s *StockAnalysisWorksheet) StockAnalysis(worksheetName, julDate string) error {
	if s.StockCache == nil {
		return fmt.Errorf("no stock cache loaded")
	}

	sheet, err := s.GetFile().NewSheet(worksheetName)

	if err != nil {
		logrus.Error("Error:", err.Error())
		return err
	}

	symbolList, err := account_info.SymbolList(context.Background(), s.GetPGXConn(), s.GetLookups())
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
	accountList, err := account_info.AccountList(context.Background(), s.GetPGXConn())
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
		Trigger,
		CurrentAmount,
		AffectedDividend,
		Fidelity,
		FidelityDividend,
		Schwab,
		SchwabDividend,
		NetDividendAmount,
		NetDividendChange,
	}
	for i := 0; i < len(remainingColumnTitles); i++ {
		columnNames = append(columnNames, remainingColumnTitles[i])
	}

	column := 1
	for i := 0; i < len(columnNames); i++ {
		colInfoName, err := column_info.New(s.GetExcelizeFile(), columnNames[i], worksheetName, column)
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
			columnNames[i] == CAGR ||
			columnNames[i] == Trigger ||
			columnNames[i] == CurrentAmount ||
			columnNames[i] == AffectedDividend ||
			columnNames[i] == Fidelity ||
			columnNames[i] == Schwab ||
			columnNames[i] == FidelityDividend ||
			columnNames[i] == SchwabDividend ||
			columnNames[i] == NetDividendAmount ||
			columnNames[i] == NetDividendChange {
			colInfoName.SetFormula(true)
		}
		column += 1
		sheet.AddColumn(colInfoName)
	}

	row := 1
	// Write Headers
	for _, ci := range sheet.Columns {
		if err = ci.WriteHeader(row, s.GetStyles().Header); err != nil {
			return err
		}
	}

	/*
	 * Details
	 */
	row += 1
	var activeSymbols []string
	symbolData := make(map[string]*account_info.AccountInfo)

	schwabRow := 0
	fidelRow := 0

	for _, symbol := range sortedSymbols {
		tickerInfo, err := account_info.AccountInfoGet(context.Background(), s.GetPGXConn(), symbol)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		if tickerInfo == nil {
			logrus.Fatal("ticker info is nil")
			return errors.New("ticker info is nil")
		}

		logrus.Debug("symbol [", symbol, "] shares [", tickerInfo.NumberOfShares, "]")

		if tickerInfo.NumberOfShares < 2 {
			continue
		}

		activeSymbols = append(activeSymbols, symbol)
		symbolData[symbol] = tickerInfo

		switch symbol {
		case "FCNTX", "SCYB": //TODO: Need to updated test since it does not have SCYB,
			schwabRow = len(symbolData) + 1
			logrus.Debug("SCYB:", schwabRow)
		case "FAGIX", "FSYD": //TODO: Need to updated test since it does not have FSYD.
			fidelRow = len(symbolData) + 1
			logrus.Debug("FSYD:", fidelRow)
		}
	}

	numberSymbols = len(activeSymbols)
	// logrus.Debug("Number of symbols> ", numberSymbols)
	for _, symbol := range activeSymbols {
		tickerInfo := symbolData[symbol]
		//
		if err = s.writeStockAnalysisDetailRow(sheet, row, tickerInfo, julDate, fidelRow, schwabRow); err != nil {
			logrus.Error(err.Error())
			return err
		}
		row += 1
	}

	/*
	 * Finish Up
	 */
	for _, ci := range sheet.Columns {
		_ = ci.SetColumnSize()
		switch ci.Name {
		case CAGR, AnnualReturn, ReturnOnInvestment, DividendYield:
			rangeRef := fmt.Sprintf("$%s$2:$%s$%d", ci.ColumnID, ci.ColumnID, numberSymbols+1)
			logrus.Debug("Range Ref> ", rangeRef)
			err := s.GetExcelizeFile().SetConditionalFormat(worksheetName, rangeRef, []excelize.ConditionalFormatOptions{
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
			format, err := s.customFormat("000000", "#7BC189", "Calibri", 14.0, true)
			if err != nil {
				logrus.Error(err.Error())
			}
			err = s.GetExcelizeFile().SetConditionalFormat(worksheetName, rangeRef, []excelize.ConditionalFormatOptions{
				{
					Type:     "top",
					Criteria: "=",
					Format:   &format,
					Value:    "10",
					Percent:  true,
				},
			})
			if err != nil {
				logrus.Error(err.Error())
			}
		case CurrentAmount, AffectedDividend, Fidelity, FidelityDividend, Schwab, SchwabDividend, NetDividendAmount, NetDividendChange:
			rangeRef := fmt.Sprintf("$%s$2:$%s$%d", ci.ColumnID, ci.ColumnID, numberSymbols+1)
			format, err := s.customFormat("000000", "#7BC189", "Calibri", 14.0, false)
			if err != nil {
				logrus.Error(err.Error())
			}
			err = s.GetExcelizeFile().SetConditionalFormat(worksheetName, rangeRef,
				[]excelize.ConditionalFormatOptions{
					{
						Type:     "cell",
						Criteria: ">",
						Format:   &format,
						Value:    "0",
					},
				},
			)
		}
	}
	return nil
}

// customFormat
// TODO: cache formats
func (s *StockAnalysisWorksheet) customFormat(fontColor, fillColor, fontFamily string, fontSize float64, bold bool) (int, error) {
	format, err := s.GetExcelizeFile().NewConditionalStyle(
		&excelize.Style{
			Font: &excelize.Font{
				Color:  fontColor,
				Bold:   bold,
				Family: fontFamily,
				Size:   fontSize,
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{fillColor},
				Pattern: 1,
			},
		},
	)
	return format, err
}

func (s *StockAnalysisWorksheet) BuildSumList(list []string, row int) string {
	sb := strings.Builder{}
	for i, v := range list {
		sb.WriteString(fmt.Sprintf("%s%d", v, row))
		if i < len(list)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}
