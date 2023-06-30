package app_test

import (
	"bytes"
	_ "embed"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"iex-indicators/cmd/internal/app"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:embed testdata/transactions.csv
var testTransactions []byte

func TestGetResourceById(t *testing.T) {
	a := app.App{
		Srv:       nil,
		LookupSet: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(testTransactions))
	a.LoadTransactionsHandler(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	//var got gin.H
	//err := json.Unmarshal(w.Body.Bytes(), &got)
	//if err != nil {
	//	t.Fatal(err)
	//}
	// assert.Equal(t, want, got) // want is a gin.H that contains the wanted map.
}

func TestWithData(t *testing.T) {
	a := app.App{
		Srv:       nil,
		LookupSet: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(testTransactions))
	// c.Request.Header.Set("Content-Type", mw.FormDataContentType())
	// c.engine.MaxMultipartMemory = 8 << 20
	a.LoadTransactionsHandler(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	responseData, _ := io.ReadAll(w.Body)

	t.Log(string(responseData))

}
