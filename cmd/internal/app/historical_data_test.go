package app_test

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"

	// "github.com/kpearce2430/stock-tools/model"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp_LoadHistoricalData(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(csvHistoricalDAta))
	q := c.Request.URL.Query()
	q.Add("database", databaseName)
	q.Add("symbol", "USAIX")
	q.Add("source", "random")
	c.Request.URL.RawQuery = q.Encode()
	testApp.LoadHistoricalData(c)
	if w.Code != http.StatusOK {
		t.Log("error expecting:", http.StatusOK, " got:", w.Code)
	}
	responseData, err := io.ReadAll(w.Body)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	t.Log(string(responseData))
}
