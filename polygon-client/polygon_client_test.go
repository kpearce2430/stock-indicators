package polygon_client_test

import (
	"encoding/json"
	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/utils"
	polygon_client "github.com/kpearce2430/stock-tools/polygon-client"
	"github.com/polygon-io/client-go/rest/models"
	"strings"
	"testing"
	"time"
)

/*
https://api.polygon.io/v2/aggs/ticker/AAPL/range/1/day/2023-01-09/2023-01-09?apiKey=<token>
*/

type PolygonTest struct {
	Symbol        string
	ExpectedError bool
}

var (
	tests = []PolygonTest{
		{"HD", false},
		{"AAPL", false},
		{"FAGIX", true},
	}
	p *polygon_client.PolygonClient
)

func TestMain(m *testing.M) {
	p = polygon_client.NewPolygonClient("")
	m.Run()
}

func TestPolygonClient_GetPreviousClose(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}
	t.Parallel()
	for _, tc := range tests {
		t.Run(tc.Symbol, func(t *testing.T) {
			t.Parallel()
			resp, err := p.GetPreviousClose(tc.Symbol)
			if err != nil {
				t.Log(err)
				if tc.ExpectedError != true {
					t.Fail()
				}
				return
			}

			var quote models.GetPreviousCloseAggResponse
			if err := json.Unmarshal(resp, &quote); err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}
			// hist_usaix.csv.Log(string(resp))
			for _, r := range quote.Results {
				t.Log(r.Close)
			}
		})
	}
}

func TestPolygonClient_Dividends(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}
	t.Parallel()
	for _, tc := range tests {
		t.Run(tc.Symbol, func(t *testing.T) {
			t.Parallel()
			var dividends []models.Dividend
			resp, err := p.GetDataSet(tc.Symbol)
			if err != nil {
				t.Log(err)
				if tc.ExpectedError != true {
					t.Fail()
				}
				return
			}

			if err := json.Unmarshal(resp, &dividends); err != nil {
				t.Log(err)
				t.Fail()
				return
			}
			t.Log("Number Dividends", len(dividends))
			t.Log(string(resp))
		})
	}
}

func TestPolygonClient_GetDailyOpenCloseAgg(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}
	t.Parallel()
	for _, tc := range tests {
		t.Run(tc.Symbol, func(t *testing.T) {
			t.Parallel()
			resp, err := p.GetDailyOpenCloseAgg(tc.Symbol)
			if err != nil {
				if tc.ExpectedError != true {
					t.Fail()
				}
				return
			}

			var daily models.GetDailyOpenCloseAggResponse
			if err := json.Unmarshal(resp, &daily); err != nil {
				t.Log(err)
				t.Fail()
				return
			}
			t.Log(daily)
			t.Log(string(resp))
		})
	}
}

func TestPolygonClient_GetRSI(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}

	p = polygon_client.NewPolygonClient("")

	tests := []PolygonTest{
		{"AAPL", false},
		{"HD", false},
		{"blah", true},
	}

	for _, tc := range tests {
		t.Run(tc.Symbol, func(t *testing.T) {
			request := polygon_client.PolygonRSIRequest{
				RequestDate: business_days.GetBusinessDay(time.Now()),
			}

			data, err := json.Marshal(request)
			if err != nil {
				t.Log(err)
				if tc.ExpectedError != true {
					t.Fail()
				}
				return
			}
			resp, err := p.GetRSI(tc.Symbol, data)
			if err != nil {
				t.Log(err)
				if tc.ExpectedError != true {
					t.Fail()
				}
				return
			}

			var rsi models.GetRSIResponse
			t.Log(string(resp))
			if err := json.Unmarshal(resp, &rsi); err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			t.Log(len(rsi.Results.Values))
			for i, r := range rsi.Results.Values {
				// var q models.Millis
				b, _ := json.Marshal(r.Timestamp)
				t.Log(i, ":", string(b), ":", r.Value)
			}
		})
	}

}

func TestPolygonMap(t *testing.T) {

	myMap := make(map[string]string)
	myMap["date"] = "2023-12-28"
	myMap["sympol"] = "HD"

	bytes, err := json.Marshal(myMap)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	t.Log(string(bytes))

	var testMap map[string]string
	if err := json.Unmarshal(bytes, &testMap); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	t.Log(testMap)

}
