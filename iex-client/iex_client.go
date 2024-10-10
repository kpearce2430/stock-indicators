package iex_client

import (
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/kpearce2430/keputils/http-client"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"io"
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

func (iex IEXHttpClient) GetStockQuote() ([]byte, error) {
	//  curl  "https://cloud.iexapis.com/v1/stock/HD/batch?types=quote,stats,dividends,news&token=pk_189dd9a1c5814706a37220a212dc54a0" | jq .

	//url := fmt.Sprintf("https://%s/v1/stock/%s/batch", iex.domainName, iex.stockSymbol)
	//log.Println(url)

	qParams := make(map[string]string)
	qParams["token"] = iex.iexToken
	return iex.callIEX("quote", qParams)
}

func (iex IEXHttpClient) GetDividends() ([]byte, error) {
	qParams := make(map[string]string)
	qParams["token"] = iex.iexToken
	return iex.callIEX("advanced_dividends", qParams)
}

func (iex IEXHttpClient) callIEX(iexType string, quaryParams map[string]string) ([]byte, error) {
	/*
		curl "https://cloud.iexapis.com/v1/stock/HD/batch?types=quote&token=token"
		curl "https://cloud.iexapis.com/v1/data/core/quote,historical_prices,news/HD?token=token&range=5d"
		curl "https://cloud.iexapis.com/v1/data/core/historical_prices/HD?token=token&range=5d"
		curl "https://cloud.iexapis.com/v1/data/core/quote/HD?token=token"
	*/
	myReq := iex.httpClient.R()
	for k, v := range quaryParams {
		myReq.SetQueryParam(k, v)
	}
	url := fmt.Sprintf("https://%s/v1/data/core/%s/%s", iex.domainName, iexType, iex.stockSymbol)

	if iex.devMode {
		filename := fmt.Sprintf("%s_%s_test.out", iex.stockSymbol, iex.stockSymbol)
		myReq.EnableDumpToFile(filename)
	}

	response, err := myReq.Get(url)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			logrus.Error(err.Error())
		}
	}()

	responseData, _ := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	logrus.Debug(">", string(responseData))
	return responseData, nil
}
