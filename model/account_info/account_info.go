package account_info

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/portfolio_value"
	ticker "github.com/kpearce2430/stock-tools/model/ticker"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/sirupsen/logrus"
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
	if err != nil {
		logrus.Error("Error getting account list", err)
		return nil, err
	}
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

func SymbolList(ctx context.Context, pgxConn *pgxpool.Pool, lookups *lookups.LookUpSet) (map[string]string, error) {
	symbolSet := make(map[string]string)
	rows, err := pgxConn.Query(ctx, selectSymbolStatement)
	if err != nil {
		logrus.Error("Error getting symbol list", err)
		return symbolSet, err
	}

	defer rows.Close()
	// Iterate through the result set
	for rows.Next() {
		var symbol, security string
		err = rows.Scan(&symbol, &security)
		if err != nil {
			logrus.Error("Errors rows.scan()", err)
			return symbolSet, err
		}
		if symbol == "" && security == "" {
			continue
		}

		if symbol == "" {
			value, ok := lookups.GetLookUpByName(security)
			switch ok {
			case true:
				symbol = value
			default:
				logrus.Warning("No symbol for [", security, "]")
			}
		}

		value, _ := lookups.GetLookUpByName(security)
		switch {
		case value == "DEAD":
			continue
			//case ok:
			//	security = value
		}
		if symbol == "" {
			logrus.Warning("Security [", security, "] missing SYMBOL")
		}
		symbolSet[symbol] = security

	}
	rows.Close()
	return symbolSet, nil
}

func getLatestPrice(pv *portfolio_value.PortfolioValueRecord) float64 {
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

func AccountInfoGet(ctx context.Context, pgxConn *pgxpool.Pool, acctSymbol string) (*AccountInfo, error) {
	logrus.Debug("Getting Account Info for ", acctSymbol)
	tSet := transaction.NewTransactionSet()
	if err := tSet.FromDBbySymbol(ctx, pgxConn, transactionTable, acctSymbol); err != nil {
		return nil, err
	}

	var securityNames []string
	acctTicker := ticker.NewTicker(acctSymbol)
	for _, tr := range tSet.TransactionRows {
		ent, err := ticker.NewEntityFromTransaction(tr)
		if err != nil {
			return nil, err
		}

		if tr.Symbol != acctSymbol {
			logrus.Warning("Skipping symbol:", tr.Symbol)
			continue
		}

		if tr.Security != "" && !utils.Contains(securityNames, tr.Security) {
			logrus.Debug("Adding SecurityPayee:", tr.Security)
			securityNames = append(securityNames, tr.Security)
		}
		acctTicker.AddEntity(ent)
	}

	if len(securityNames) == 0 {
		logrus.Error("No security names found")
		return nil, errors.New("no security names found")
	}

	acctInfo := AccountInfo{
		Symbol:         acctSymbol,
		Security:       securityNames[len(securityNames)-1], // last one found
		NumberOfShares: acctTicker.NumberOfShares(),
	}

	var pvValue portfolio_value.PortfolioValueRecord
	err := pvValue.GetLastDB(pgxConn, acctTicker.Symbol, "portfolio_value")

	if err != nil {
		logrus.Error("Error Getting PV for ", acctTicker.Symbol, " Shares:", acctInfo.NumberOfShares, ":", err.Error())
	}

	acctInfo.SecurityType = pvValue.Type
	acctInfo.LatestPrice = getLatestPrice(&pvValue)

	if acctInfo.SecurityType == "" {
		logrus.Debug("Security Type is missing for ", acctSymbol)
		switch len(acctSymbol) {
		case 1, 2, 3, 4:
			logrus.Debug(acctSymbol, " Security Type is missing, assuming Stock")
			acctInfo.SecurityType = "Stock"
		case 5:
			logrus.Debug(acctSymbol, " Security Type is missing, assuming Mutual Fund")
			acctInfo.SecurityType = "Mutual Fund"
		default:
			logrus.Debug(acctSymbol, " Security Type is missing, assuming Bond")
			acctInfo.SecurityType = "Bond"
		}
	}

	if acctTicker.NumberOfShares() <= 0 {
		return &acctInfo, nil
	}

	acctInfo.Accounts = make(map[string]float64)
	for _, acct := range acctTicker.Accounts {
		acctInfo.Accounts[acct.Name] = acct.NumberOfShares()
	}
	acctInfo.DividendsReceived = acctTicker.DividendsPaid()
	acctInfo.InterestIncome = acctTicker.InterestIncome()
	acctInfo.NetCost = acctTicker.NetCost()
	acctInfo.FirstBought = acctTicker.FirstBought()

	if acctInfo.NumberOfShares > 2.00 {
		acctInfo.AveragePrice = acctTicker.AveragePrice()
	}
	return &acctInfo, nil
}
