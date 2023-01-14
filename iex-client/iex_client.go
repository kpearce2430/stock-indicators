package iex_client

import (
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/kpearce2430/keputils/http-client"
	"github.com/kpearce2430/keputils/utils"
	"log"
	"strconv"
)

type IEXHttpClient struct {
	devMode     bool
	httpClient  *req.Client
	iexToken    string
	domainName  string
	stockSymbol string
	period      string
	indicator   bool
}

func New(domain string, timeout int64, devMode bool) *IEXHttpClient {

	client := http_client.GetDefaultClient(timeout, devMode)

	iexClient := IEXHttpClient{
		devMode:     devMode,
		httpClient:  client,
		iexToken:    utils.GetEnv("TOKEN", "Tpk_76c5b627e1d3420dbd0f2621787941ba"),
		domainName:  domain,
		stockSymbol: "",
		period:      "",
		indicator:   true,
	}

	return &iexClient

}

func (iex IEXHttpClient) Symbol(stockSymbol string) *IEXHttpClient {
	iex.stockSymbol = stockSymbol
	return &iex
}

func (iex IEXHttpClient) Period(p string) *IEXHttpClient {
	iex.period = p
	return &iex
}

func (iex IEXHttpClient) Indicator(ind bool) *IEXHttpClient {
	iex.indicator = ind
	return &iex
}

func (iex IEXHttpClient) GetRSI() (*IexIndicatorResponse, error) {

	return iex.GetIndicator("rsi")

}

func (iex IEXHttpClient) GetMacD() (*IexIndicatorResponse, error) {

	return iex.GetIndicator("macd")

}

func (iex IEXHttpClient) GetIndicator(indicatorSymbol string) (*IexIndicatorResponse, error) {

	//
	// curl "https://sandbox.iexapis.com/v1/stock/HD/indicator/rsi?range=6m&indicatorOnly=false&token=Tpk_76c5b627e1d3420dbd0f2621787941ba"

	url := fmt.Sprintf("https://%s/v1/stock/%s/indicator/%s", iex.domainName, iex.stockSymbol, indicatorSymbol)
	log.Println(url)

	//?range=6m&indicatorOnly=false&token=Tpk_76c5b627e1d3420dbd0f2621787941ba"

	response := IexIndicatorResponse{}
	var err error

	myReq := iex.httpClient.R().
		SetResult(&response).
		SetQueryParam("range", "6m").
		SetQueryParam("indicatorOnly", strconv.FormatBool(iex.indicator)).
		SetQueryParam("token", iex.iexToken)

	if iex.period != "" {
		myReq.SetQueryParam("input1", iex.period)
	}

	if iex.devMode {
		filename := fmt.Sprintf("%s_%s_test.out", indicatorSymbol, iex.stockSymbol)
		myReq.EnableDumpToFile(filename)
	}

	_, err = myReq.Get(url)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println(">", response)

	return &response, nil

}
