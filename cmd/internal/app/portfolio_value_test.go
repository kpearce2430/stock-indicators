package app_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const julDate = "2022001"
const databaseName = "something"

func TestLoadPortfolioValue(t *testing.T) {
	t.Setenv("PV_COUCHDB_DATABASE", databaseName)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(csvPortfolioValueData))
	q := c.Request.URL.Query()
	q.Add("juldate", julDate)
	q.Add("database", databaseName)
	c.Request.URL.RawQuery = q.Encode()
	testApp.LoadPortfolioValueHandler(c)
	assert.Equal(t, http.StatusOK, w.Code)

	type lookupTest struct {
		Name          string
		Params        []gin.Param
		StatusCode    int
		ErrorResponse string
	}

	testCases := []lookupTest{
		{
			Name: "Happy Path",
			Params: []gin.Param{
				{Key: "symbol", Value: "AAPL"},
			},
			StatusCode:    http.StatusOK,
			ErrorResponse: "",
		},
		{
			Name: "Bad Symbol",
			Params: []gin.Param{
				{Key: "symbol", Value: "MISSING"},
			},
			StatusCode:    http.StatusNotFound,
			ErrorResponse: "",
		},
		{
			Name: "Coke Not Found",
			Params: []gin.Param{
				{Key: "symbol", Value: "COKE"},
			},
			StatusCode:    http.StatusNotFound,
			ErrorResponse: "",
		},
		{
			Name: "KO Found",
			Params: []gin.Param{
				{Key: "symbol", Value: "KO"},
			},
			StatusCode:    http.StatusOK,
			ErrorResponse: "",
		},
		{
			Name: "NSC Found",
			Params: []gin.Param{
				{Key: "symbol", Value: "NSC"},
			},
			StatusCode:    http.StatusOK,
			ErrorResponse: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = tc.Params
			c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			q := c.Request.URL.Query()
			q.Add("juldate", julDate)
			q.Add("database", databaseName)
			c.Request.URL.RawQuery = q.Encode()
			testApp.GetPortfolioValueHandler(c)
			if w.Code == http.StatusOK {
				responseData, err := io.ReadAll(w.Body)
				if tc.StatusCode == http.StatusNotFound && string(responseData) == "null" {
					return
				}
				if err != nil {
					t.Log(err.Error())
					t.Fail()
					return
				}

				var status model.PortfolioValueDatabaseRecord
				err = json.Unmarshal(responseData, &status)
				if err != nil {
					t.Log(err.Error())
					t.Fail()
					return
				}

				if status.PV == nil {
					t.Log("PV nil")
					t.Fail()
					return
				}
				if tc.Params[0].Value != status.PV.Symbol {
					t.Log("Symbols dont match")
					t.Fail()
					return
				}
			}
		})
	}
}
