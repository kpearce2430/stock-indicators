package app_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/stock-tools/cmd/internal/app"
	"github.com/stretchr/testify/assert"
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

func TestApp_LoadTransactionsHandler(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(testTransactions))
	q := c.Request.URL.Query()
	q.Add("database", app.TransactionAllTable)
	c.Request.URL.RawQuery = q.Encode()

	testApp.LoadTransactionsHandler(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	var count int
	if err := testApp.PGXConn.QueryRow(context.Background(), fmt.Sprintf("select count(*) from %s", app.TransactionAllTable)).Scan(&count); err != nil {
		t.Fatal(err.Error())
	}
	t.Log("Count:", count)
}

func TestBuySellSBUX(t *testing.T) {
	t.Skip("is this needed")
	t.Parallel()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(testSBUXTransactions))
	testApp.LoadTransactionsHandler(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be
	responseData, _ := io.ReadAll(w.Body)
	t.Log(string(responseData))
}

func TestBuySellTHD(t *testing.T) {
	t.Skip("is this needed")
	t.Parallel()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(testTHDTransaction))
	testApp.LoadTransactionsHandler(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	for _, ticker := range testApp.Tickers {
		fmt.Println(ticker)
	}
}
