package app_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"iex-indicators/cmd/internal/app"
	"iex-indicators/model"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:embed testdata/transactions.csv
var testTransactions []byte

//go:embed testdata/trans_sbx.csv
var testSBUXTransactions []byte

//go:embed testdata/trans_hd.csv
var testTHDTransaction []byte

func TestGetResourceById(t *testing.T) {
	a := app.App{
		Srv:       nil,
		Tickers:   make(map[string]*model.Ticker),
		LookupSet: model.LoadLookupSet("1", string(csvLookupData)),
		// Symbols:   make(map[string]bool),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(testTransactions))
	a.LoadTransactionsHandler(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

}

func TestBuySellSBUX(t *testing.T) {
	a := app.App{
		Srv:       nil,
		Tickers:   make(map[string]*model.Ticker),
		LookupSet: model.LoadLookupSet("1", string(csvLookupData)),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(testSBUXTransactions))
	a.LoadTransactionsHandler(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	responseData, _ := io.ReadAll(w.Body)
	t.Log(string(responseData))
}

func TestBuySellTHD(t *testing.T) {
	a := app.App{
		Srv:       nil,
		Tickers:   make(map[string]*model.Ticker),
		LookupSet: model.LoadLookupSet("1", string(csvLookupData)),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(testTHDTransaction))
	a.LoadTransactionsHandler(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	//responseData, _ := io.ReadAll(w.Body)
	//t.Log(string(responseData))
	for _, ticker := range a.Tickers {
		fmt.Println(ticker)
		//for _, acct := range ticker.Accounts {
		//	fmt.Println(acct)
		//	for _, entity := range acct.Entities {
		//		fmt.Println(entity)
		//
		//	}
		//}
	}
	//t.Log(a.Tickers["HD"].NumberOfShares())
}
