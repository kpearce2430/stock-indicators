package iex_client_test

import (
	iex_client "iex-indicators/iex-client"
	"os"
	"testing"
)

var client *iex_client.IEXHttpClient

func TestMain(m *testing.M) {

	os.Setenv("TOKEN", "Tpk_76c5b627e1d3420dbd0f2621787941ba")
	os.Setenv("IEX_URL", "sandbox.iexapis.com")
	client = iex_client.New("sandbox.iexapis.com", 10, false).Symbol("HD")
	m.Run()

}

func TestIEXHttpClient_GetMacD(t *testing.T) {
	t.Log("starting")
	client.Period("")
	response, err := client.GetMacD()
	if err != nil {
		t.Fatal(err)
	}

	// t.Logf("macd response >> %+v", response)
	if len(response.Indicator) <= 0 {
		t.Errorf("Invalid length of Indicators %d", len(response.Indicator))
		return
	}
	// t.Logf(">> %v\n", len(response.Indicator))
	// indicator := response.Indicator[0]

	// t.Logf(">> %v\n", len(indicator))
	// t.Logf(">> %v\n", len(response.Chart))

}

func TestIEXHttpClient_GetRSI(t *testing.T) {

	t.Log("starting")
	client.Period("14")
	response, err := client.GetRSI()
	if err != nil {
		t.Fatal(err)
	}

	if len(response.Indicator) <= 0 {
		t.Errorf("Invalid length of Indicators %d", len(response.Indicator))
		return
	}
	//t.Logf(">> %v\n", len(response.Indicator))
	//indicator := response.Indicator[0]
	//
	//t.Logf(">> %v\n", len(indicator))
	//t.Logf(">> %v\n", len(response.Chart))

}
