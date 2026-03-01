package massive_client_test

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/massive-client"
	"github.com/massive-com/client-go/v2/rest/models"
)

type MassiveTests struct {
	Symbol        string
	ExpectedError bool
}

var (
	tests = []MassiveTests{
		{"HD", false},
		{"AAPL", false},
		{"FAGIX", true},
	}
	mc *massive_client.MassiveClient
)

func TestMain(m *testing.M) {
	key := utils.GetEnv("MASSIVE_API", "")
	if strings.Compare(key, "None") == 0 {
		slog.Info("Skipping test because MASSIVE_API is not set")
		return
	}
	mc = massive_client.New()
	m.Run()
}

func TestMassiveClient_New(t *testing.T) {
	mClient := massive_client.New()
	if mClient == nil {
		t.Error("massive_client.New() returned nil")
	}
}

func TestMassiveClient_GetPreviousClose(t *testing.T) {
	t.Parallel()
	for _, tc := range tests {
		t.Run(tc.Symbol, func(t *testing.T) {
			t.Parallel()
			resp, err := mc.GetPreviousClose(tc.Symbol)
			if err != nil {
				t.Log(err)
				if tc.ExpectedError != true {
					t.Fail()
				}
				return
			}

			var quote models.GetPreviousCloseAggResponse
			if err = json.Unmarshal(resp, &quote); err != nil {
				t.Error(err.Error())
				return
			}
			// hist_usaix.csv.Log(string(resp))
			for _, r := range quote.Results {
				t.Log(r.Close)
			}
		})
	}
}

func TestMassiveClient_Dividends(t *testing.T) {
	t.Parallel()
	for _, tc := range tests {
		t.Run(tc.Symbol, func(t *testing.T) {
			t.Parallel()
			var dividends []models.Dividend
			resp, err := mc.GetDataSet(tc.Symbol)
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

func callGetDailyOpenCloseAgg(t *testing.T, tc *MassiveTests, args ...string) {
	t.Helper()
	t.Parallel()
	resp, err := mc.GetDailyOpenCloseAgg(tc.Symbol, args...)
	if err != nil {
		if tc.ExpectedError != true {
			t.Fail()
		}
		return
	}

	var daily models.GetDailyOpenCloseAggResponse
	if err = json.Unmarshal(resp, &daily); err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	t.Log(daily)
	t.Log(string(resp))
}

func TestMassiveClient_GetDailyOpenCloseAgg(t *testing.T) {
	t.Parallel()
	args := [][]string{
		{},          // Today...
		{"2026056"}, // 2026-02-25
	}
	for i, arg := range args {
		for j, tc := range tests {
			t.Run(tc.Symbol+"/"+fmt.Sprintf("%d/%d", i, j), func(t *testing.T) {
				callGetDailyOpenCloseAgg(t, &tc, arg...)
			})
		}
	}
}

func TestMassiveClient_GetDailyOpenCloseAggBadArgs(t *testing.T) {
	t.Parallel()
	args := [][]string{
		{"BADARGS"},
		{"2026ZZZ"},
		{"20260"},
		{"2025300", "2026056"},
	}
	tcases := []MassiveTests{
		{"HD", true},
		{"AAPL", true},
	}

	for _, arg := range args {
		for _, tc := range tcases {
			t.Run(tc.Symbol+"/"+arg[0], func(t *testing.T) {
				callGetDailyOpenCloseAgg(t, &tc, arg...)
			})
		}
	}
}

func TestMassiveClient_GetData(t *testing.T) {
	t.Parallel()
	args := [][]string{
		{},          // Today...
		{"2026056"}, // 2026-02-25
	}
	for i, arg := range args {
		for j, tc := range tests {
			t.Run(tc.Symbol+"/"+fmt.Sprintf("%d/%d", i, j), func(t *testing.T) {
				t.Parallel()
				resp, err := mc.GetData(tc.Symbol, arg...)
				if err != nil {
					if tc.ExpectedError != true {
						t.Fail()
					}
					return
				}

				var daily models.GetDailyOpenCloseAggResponse
				if err = json.Unmarshal(resp, &daily); err != nil {
					t.Log(err)
					t.Fail()
					return
				}
				t.Log(daily)
				t.Log(string(resp))
			})
		}
	}
}

func TestClient_GetRSI(t *testing.T) {
	//
	for _, tc := range tests {
		t.Run(tc.Symbol, func(t *testing.T) {
			request := massive_client.MassiveRSIRequest{
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
			resp, err := mc.GetRSI(tc.Symbol, data)
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

func TestClient_GetIndicator(t *testing.T) {
	for _, tc := range tests {
		t.Run(tc.Symbol, func(t *testing.T) {
			request := massive_client.MassiveRSIRequest{
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
			resp, err := mc.GetIndicator("rsi", tc.Symbol, data)
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
func TestMap(t *testing.T) {

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
