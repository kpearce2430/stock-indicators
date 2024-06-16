package app_test

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type cacheTests struct {
	Name   string
	Symbol string
	Date   string
	Expect int // true = success; false = failure
}

func TestApp_GetStockCache(t *testing.T) {
	t.Parallel()
	testCases := []cacheTests{
		{
			Name:   "Tuesday Jan 23, 2024",
			Symbol: "HD",
			Date:   "2024023",
			Expect: http.StatusOK,
		},
		{
			Name:   "Saturday Jan 20, 2024",
			Symbol: "HD",
			Date:   "2024020",
			Expect: http.StatusInternalServerError,
		},
		{
			Name:   "Saturday Jan 2, 2023",
			Symbol: "HD",
			Date:   "2023002",
			Expect: http.StatusInternalServerError,
		},
	}

	gin.SetMode(gin.TestMode)
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "symbol", Value: tc.Symbol}}
			c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(csvPortfolioValueData))
			q := c.Request.URL.Query()
			q.Add("date", tc.Date)
			c.Request.URL.RawQuery = q.Encode()
			testApp.GetStockCache(c)
			if w.Code != tc.Expect {
				t.Log("Status Not:", tc.Expect, " Got:", w.Code)
				t.Fail()
				return
			}

			responseData, err := io.ReadAll(w.Body)
			if err != nil {
				t.Log(err.Error())
				t.Fatal()
				return
			}
			t.Log(string(responseData))
		}) // func
	} // for
}
