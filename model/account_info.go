package model

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	transactionTable       = "transactions"
	selectAccountStatement = "SELECT DISTINCT account FROM transactions ORDER BY account"
	selectSymbolStatement  = "SELECT DISTINCT symbol, security FROM transactions ORDER BY symbol;"
)

type AccountInfo struct {
	Security          string             `json:"security,omitempty"`
	Symbol            string             `json:"symbol,required"`
	SecurityType      string             `json:"securityType,omitempty"`
	NumberOfShares    float64            `json:"numberOfShares,omitempty"`
	Accounts          map[string]float64 `json:"accounts,omitempty"`
	LatestPrice       float64            `json:"latestPrice,omitempty"` // iex or pv
	DividendsReceived float64            `json:"dividendsReceived,omitempty"`
	InterestIncome    float64            `json:"interestIncome,omitempty"`
	NetCost           float64            `json:"netCost,omitempty"`
	FirstBought       time.Time          `json:"firstBought,omitempty"`
	AveragePrice      float64            `json:"averagePrice,omitempty"`
}

func AccountList(ctx context.Context, pgxConn *pgxpool.Pool) ([]string, error) {
	rows, err := pgxConn.Query(ctx, selectAccountStatement)
	defer rows.Close()
	var accountList []string
	// Iterate through the result set
	for rows.Next() {
		var account string
		err = rows.Scan(&account)
		if err != nil {
			return accountList, err
		}
		accountList = append(accountList, account)
	}
	rows.Close()
	return accountList, nil
}

func SymbolList(ctx context.Context, pgxConn *pgxpool.Pool, lookups *LookUpSet) (map[string]string, error) {
	symbolSet := make(map[string]string)
	rows, err := pgxConn.Query(ctx, selectSymbolStatement)
	defer rows.Close()
	// Iterate through the result set
	for rows.Next() {
		var symbol, security string
		err = rows.Scan(&symbol, &security)
		if symbol == "" && security == "" {
			continue
		}
		if err != nil {
			return symbolSet, err
		}

		value, ok := lookups.GetLookUpByName(security)
		switch {
		case value == "DEAD":
			continue
		case ok:
			security = value
		}
		symbolSet[symbol] = security

	}
	rows.Close()
	return symbolSet, nil
}

func getLatestPrice(pv *PortfolioValueRecord) float64 {
	if pv == nil {
		logrus.Error("pv is nil")
		return 0.00
	}
	switch pv.Type {
	case "Stock":
		logrus.Debug("pv ticker>", pv.Symbol, ":", pv.Quote)
		return pv.Quote
	case "Mutual Fund":
		return pv.Quote
	case "Bond":
		return 100.00
	}
	return 0.00
}

func AccountInfoGet(ctx context.Context, pgxConn *pgxpool.Pool, acctSymbol, julDate string) (*AccountInfo, error) {

	tSet := NewTransactionSet()
	if err := tSet.TransactionSetFromDBbySymbol(ctx, pgxConn, transactionTable, acctSymbol); err != nil {
		return nil, err
	}

	var securityNames []string
	ticker := NewTicker(acctSymbol)
	for _, tr := range tSet.TransactionRows {
		ent, err := NewEntityFromTransaction(tr)
		if err != nil {
			return nil, err
		}

		if tr.Security != "" && !utils.Contains(securityNames, tr.Security) {
			logrus.Debug("Adding SecurityPayee:", tr.Security)
			securityNames = append(securityNames, tr.Security)
		}
		ticker.AddEntity(ent)
	}

	acctInfo := AccountInfo{
		Symbol:         acctSymbol,
		Security:       securityNames[len(securityNames)-1], // last one found
		NumberOfShares: ticker.NumberOfShares(),
	}

	if ticker.NumberOfShares() <= 0 {
		return &acctInfo, nil
	}

	acctInfo.Accounts = make(map[string]float64)
	for _, acct := range ticker.Accounts {
		acctInfo.Accounts[acct.Name] = acct.NumberOfShares()
	}
	acctInfo.DividendsReceived = ticker.DividendsPaid()
	acctInfo.InterestIncome = ticker.InterestIncome()
	acctInfo.NetCost = ticker.NetCost()
	acctInfo.FirstBought = ticker.FirstBought()

	if acctInfo.NumberOfShares > 2.00 {
		acctInfo.AveragePrice = ticker.AveragePrice()
	}

	pvValue, err := GetPortfolioValue(ticker.Symbol, julDate)
	switch {
	case err != nil:
		logrus.Error("Error Getting PV for ", ticker.Symbol, " Shares:", acctInfo.NumberOfShares, ":", err.Error())

	case pvValue != nil:
		acctInfo.SecurityType = pvValue.PV.Type
		acctInfo.LatestPrice = getLatestPrice(pvValue.PV)

	default:
		if acctInfo.NumberOfShares > 2.00 {
			logrus.Error("No PV Value for ", ticker.Symbol, ":", julDate)
		}
		acctInfo.SecurityType = "unknown"
		acctInfo.LatestPrice = 0.00
	}
	return &acctInfo, nil
}
