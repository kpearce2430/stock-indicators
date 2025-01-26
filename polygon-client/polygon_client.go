package polygon_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kpearce2430/keputils/utils"
	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type PolygonClient struct {
	Client *polygon.Client
}

func NewPolygonClient(key string) *PolygonClient {
	apiKey := utils.GetEnv("POLYGON_API", key)
	return &PolygonClient{
		Client: polygon.New(apiKey),
	}
}

func (p *PolygonClient) GetIndicator(indicator, symbol string, args []byte) ([]byte, error) {
	switch indicator {
	case "daily":
		return p.GetDailyOpenCloseAgg(symbol)
	case "rsi":
		return p.GetRSI(symbol, args)
	}
	return []byte{}, fmt.Errorf("bad request type")
}

func (p *PolygonClient) GetData(symbol string, args ...string) ([]byte, error) {
	return p.GetDailyOpenCloseAgg(symbol, args...)
}

func (p *PolygonClient) GetDataSet(symbol string, args ...string) ([]byte, error) {
	start := time.Now()
	divDate := time.Date(start.Year()-1, start.Month(), start.Day(), 00, 00, 00, 00, time.UTC)
	params := models.ListDividendsParams{}.WithTicker(models.EQ, symbol).WithDeclarationDate(models.GT, models.Date(divDate))
	iter := p.Client.ListDividends(context.Background(), params)

	var dividends []models.Dividend

	for iter.Next() {
		div := iter.Item()
		dividends = append(dividends, div)
	}

	if iter.Err() != nil {
		return []byte("{}"), iter.Err()
	}

	return json.Marshal(dividends)
}

func (p *PolygonClient) GetPreviousClose(symbol string) ([]byte, error) {
	params := models.GetPreviousCloseAggParams{
		Ticker: symbol,
	}

	resp, err := p.Client.GetPreviousCloseAgg(context.Background(), params.WithAdjusted(true))

	if err != nil {
		logrus.Error(err.Error())
		return []byte("{}"), err
	}

	return json.Marshal(resp)

}

func (p *PolygonClient) GetDailyOpenCloseAgg(symbol string, args ...string) ([]byte, error) {
	var reqDate time.Time
	switch len(args) {
	case 0:
		reqDate = time.Now()
	case 1:
		// Expecting YYYYJJJ
		// year := args[0][0:4]
		year, err := strconv.ParseInt(args[0][0:4], 10, 32)
		if err != nil {
			return []byte{}, nil
		}

		julian, err := strconv.ParseInt(args[0][4:], 10, 32)
		if err != nil {
			return []byte{}, nil
		}
		logrus.Info(year, ":", julian)
		reqDate = time.Date(int(year), 01, 01, 00, 00, 00, 00, time.UTC).Add(time.Duration(julian-1) * (24 * time.Hour))
	}

	logrus.Debug("request date:", reqDate)
	params := models.GetDailyOpenCloseAggParams{
		Ticker: symbol,
		Date:   models.Date(reqDate),
	}

	resp, err := p.Client.GetDailyOpenCloseAgg(context.Background(), params.WithAdjusted(true))
	if err != nil {
		logrus.Error(err.Error())
		return []byte("{}"), err
	}
	return json.Marshal(resp)
}

type PolygonRSIRequest struct {
	RequestDate time.Time `json:"requestDate,omitempty"`
	TimeSpan    string    `json:"timeSpan,omitempty"`
	Adjusted    bool      `json:"adjusted,omitempty"`
	Window      int       `json:"window,omitempty"`
	Order       string    `json:"order,omitempty"`
}

func (p *PolygonClient) CallRSI(symbol string, client *PolygonRSIRequest) (*models.GetRSIResponse, error) {
	data, err := json.Marshal(client)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	results, err := p.GetRSI(symbol, data)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	var rsiResponse models.GetRSIResponse
	err = json.Unmarshal(results, rsiResponse)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	return &rsiResponse, nil
}

func (p *PolygonClient) GetRSI(symbol string, data []byte) ([]byte, error) {
	// https://api.polygon.io/v1/indicators/rsi/AAPL?
	//	timespan=day&
	//	adjusted=true&
	//	window=14&
	//	series_type=close&
	//	order=desc&
	//	apiKey=xxxxxxx
	//
	// reqDate := p.GetBusinessDay(time.Now())

	var request PolygonRSIRequest
	err := json.Unmarshal(data, &request)
	if err != nil {
		logrus.Error(err.Error())
		return []byte(""), err
	}

	logrus.Debug("request date:", request.RequestDate)
	// params := models.
	params := models.GetRSIParams{
		Ticker: symbol,
	}

	params.WithTimespan(models.Day).WithAdjusted(request.Adjusted).WithWindow(request.Window).WithSeriesType(models.Close).WithOrder(models.Desc).WithTimestamp(models.EQ, models.Millis(request.RequestDate))
	resp, err := p.Client.GetRSI(context.Background(), &params)
	if err != nil {
		logrus.Error(err.Error())
		return []byte("{}"), err
	}
	return json.Marshal(resp)
}
