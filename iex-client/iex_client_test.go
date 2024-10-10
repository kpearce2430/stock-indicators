package iex_client_test

import (
	"encoding/json"
	iex_client "github.com/kpearce2430/stock-tools/iex-client"
	"os"
	"testing"
	"time"
)

var client *iex_client.IEXHttpClient

func TestMain(m *testing.M) {
	os.Setenv("TOKEN", "Tpk_76c5b627e1d3420dbd0f2621787941ba")
	os.Setenv("IEX_URL", "sandbox.iexapis.com")
	client = iex_client.New("sandbox.iexapis.com", 10, false).Symbol("HD")
	m.Run()
}

func TestIEXHttpClient_GetMacD(t *testing.T) {
	t.Skip("Skipping, need replacement with polygon.io endpoint")
	client.Period("")
	response, err := client.GetMacD()
	if err != nil {
		t.Fatal(err)
	}

	if len(response.Indicator) <= 0 {
		t.Errorf("Invalid length of Indicators %d", len(response.Indicator))
		return
	}
}

func TestIEXHttpClient_GetRSI(t *testing.T) {
	t.Skip("Skipping, need replacement with polygon.io endpoint")
	client.Period("14")
	response, err := client.GetRSI()
	if err != nil {
		t.Fatal(err)
	}

	if len(response.Indicator) <= 0 {
		t.Errorf("Invalid length of Indicators %d", len(response.Indicator))
		return
	}
}

func TestIEXHttpClient_GetStockQuote(t *testing.T) {
	t.Skip("Skipping, need replacement with polygon.io endpoint")
	results, err := client.GetStockQuote()
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	var quote []iex_client.IexStockQuoteResponse
	err = json.Unmarshal(results, &quote)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestIEXHttpClient_GetAdvancedDividends(t *testing.T) {
	t.Skip("Skipping, need replacement with polygon.io endpoint")
	results, err := client.GetDividends()
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	var dividends []iex_client.IEXAdvancedDividendsResponse
	err = json.Unmarshal(results, &dividends)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	for _, d := range dividends {
		tm := time.Unix(d.Date/1000, d.Date%1000)
		t.Log(d.Symbol, tm.Format("2006-01-02"), ",", tm.Format(time.RFC822), ".", d.Date%1000, ",", d.Amount, ",", d.Frequency)
	}
}
