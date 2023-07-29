package app_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"iex-indicators/cmd/internal/app"
	"iex-indicators/model"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const julDate = "2022001"
const databaseName = "something"

func TestLoadPortfolioValue(t *testing.T) {
	// t.Log("Starting TestLoadPortfolioValue")
	a := app.App{
		Srv:       nil,
		LookupSet: nil,
	}

	a.LookupSet = model.LoadLookupSet("1", string(csvLookupData))
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(csvPortfolioValueData))
	q := c.Request.URL.Query()
	q.Add("juldate", julDate)
	q.Add("database", databaseName)
	c.Request.URL.RawQuery = q.Encode()
	a.LoadPortfolioValueHandler(c)
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
		func(t *testing.T, a app.App, tc lookupTest) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = tc.Params
			c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			q := c.Request.URL.Query()
			q.Add("juldate", julDate)
			q.Add("database", databaseName)
			c.Request.URL.RawQuery = q.Encode()
			a.GetPortfolioValueHandler(c)
			assert.Equal(t, tc.StatusCode, w.Code)
			if w.Code == http.StatusOK {
				responseData, err := io.ReadAll(w.Body)
				assert.Equal(t, nil, err)
				// t.Log(string(responseData))
				var status model.PortfolioValueDatabaseRecord
				err = json.Unmarshal(responseData, &status)
				assert.Equal(t, nil, err)
				assert.Equal(t, tc.Params[0].Value, status.PV.Symbol)
				// t.Log(status.PV.Name)
				// t.Log("record:", status)
			}
		}(t, a, tc)
	}
}
