package account_info_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"testing"

	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model/account_info"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/model/portfolio_value"
	"github.com/kpearce2430/stock-tools/model/transaction"
)

var (
	//go:embed testdata/aapl.csv
	applTransactions []byte

	//go:embed testdata/transactions-2026-03-07.csv
	allTransactions []byte

	//go:embed testdata/lookups.csv
	csvLookupData []byte

	//go:embed testdata/portfolio_value.csv
	testPortfolioValues []byte

	ls *lookups.LookUpSet
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	ls = lookups.LoadLookupSet("1", string(csvLookupData))

	postgresDBServer, _ := postgres.StartPostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	_, err = portfolio_value.LoadDB(pgxConn, portfolio_value.PortfolioValueTable, string(testPortfolioValues), utils.JulDate(), ls)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	m.Run()
}

func Test_AccountInfo(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	if err = postgres.TruncateTable(pgxConn, transaction.TransactionTable); err != nil {
		t.Error(err.Error())
		return
	}

	tSet := transaction.NewTransactionSet()
	if err = tSet.LoadToDB(pgxConn, ls, transaction.TransactionTable, applTransactions); err != nil {
		t.Error(err.Error())
		return
	}

	info, err := account_info.AccountInfoGet(context.Background(), pgxConn, "AAPL")
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

func Test_AccountInfo_Get(t *testing.T) {
	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	if err = postgres.TruncateTable(pgxConn, transaction.TransactionTable); err != nil {
		t.Error(err.Error())
		return
	}

	tSet := transaction.NewTransactionSet()
	if err = tSet.LoadToDB(pgxConn, ls, transaction.TransactionTable, allTransactions); err != nil {
		t.Error(err.Error())
		return
	}

	tests := []struct {
		Symbol       string
		SecurityType string
	}{
		{"AAPL", "Stock"},
		{"BTI", "Stock"},
		{"USAIX", "Mutual Fund"},
	}

	for _, tc := range tests {
		t.Run(tc.Symbol, func(t *testing.T) {
			info, err := account_info.AccountInfoGet(context.Background(), pgxConn, tc.Symbol)
			if err != nil {
				t.Error(err.Error())
				return
			}

			if info.SecurityType != tc.SecurityType {
				t.Error("SecurityType is not stock:", info.SecurityType)
				b, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					t.Error(err.Error())
					return
				}

				t.Log(string(b))
			}
		})
	}
}
