package model_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
)

func Test_AccountInfo(t *testing.T) {
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	ls := model.LoadLookupSet("1", string(csvLookupData))
	if err = model.LoadPortfolioValues(pgxConn, "pv", string(testPortfolioValues), "", ls); err != nil {
		t.Error(err.Error())
		return
	}
	
	if err = truncateTransactions(pgxConn); err != nil {
		t.Error(err.Error())
		return
	}

	if err = model.TransactionSetLoadToDB(pgxConn, ls, transactionTable, applTransactions); err != nil {
		t.Error(err.Error())
		return
	}

	_, err = model.PortfolioValuesLoadDB(pgxConn, portfolioValueTable, string(testPortfolioValues), utils.JulDate(), ls)
	if err != nil {
		t.Error(err.Error())
		return
	}

	info, err := model.AccountInfoGet(context.Background(), pgxConn, "AAPL")
	if err != nil {
		t.Error(err.Error())
		return
	}

	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(string(b))
	if info.SecurityType != "Stock" {
		t.Error("SecurityType is not stock:", info.SecurityType)
	}
}
