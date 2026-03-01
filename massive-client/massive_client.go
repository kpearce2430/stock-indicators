package massive_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/utils"
	massive "github.com/massive-com/client-go/v2/rest"
	"github.com/massive-com/client-go/v2/rest/models"
	"github.com/sirupsen/logrus"
)

var (
	errTooManyArguments  = errors.New("too many arguments")
	errInvalidJulianDate = errors.New("invalid julian date format")
)

type MassiveClient struct {
	Client    *massive.Client
	resources map[string]any
}

type MassiveRSIRequest struct {
	RequestDate time.Time `json:"requestDate,omitempty"`
	TimeSpan    string    `json:"timeSpan,omitempty"`
	Adjusted    bool      `json:"adjusted,omitempty"`
	Window      int       `json:"window,omitempty"`
	Order       string    `json:"order,omitempty"`
}

func New() *MassiveClient {
	mc := &MassiveClient{}
	_ = mc.readRC()
	key := "None"
	value, ok := mc.resources["MassiveAPI"]
	if ok {
		key = value.(string)
	}

	apiKey := utils.GetEnv("MASSIVE_API", key)
	mc.Client = massive.New(apiKey)
	return mc
}

func (m *MassiveClient) getRequestDateFromJulian(args ...string) (*time.Time, error) {
	var reqDate time.Time
	switch len(args) {
	case 0:
		reqDate = time.Now()
	case 1:
		if len(args[0]) != 7 {
			logrus.Error(errInvalidJulianDate.Error())
			return nil, errInvalidJulianDate
		}
		year, err := strconv.ParseInt(args[0][0:4], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid year: %w", err)
		}

		julian, err := strconv.ParseInt(args[0][4:], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid days: %w", err)
		}
		logrus.Debug(year, ":", julian)
		reqDate = time.Date(int(year), 01, 01, 00, 00, 00, 00, time.UTC).Add(time.Duration(julian-1) * (24 * time.Hour))
	default:
		logrus.Error(errTooManyArguments.Error())
		return nil, errTooManyArguments
	}

	logrus.Debug("request date:", reqDate)
	return &reqDate, nil
}

func (m *MassiveClient) GetData(symbol string, args ...string) ([]byte, error) {
	//
	return m.GetDailyOpenCloseAgg(symbol, args...)
}

func (m *MassiveClient) GetDataSet(ticker string, args ...string) ([]byte, error) {
	start := time.Now()
	divDate := time.Date(start.Year()-1, start.Month(), start.Day(), 00, 00, 00, 00, time.UTC)
	params := models.ListDividendsParams{}.WithTicker(models.EQ, ticker).WithDeclarationDate(models.GT, models.Date(divDate))
	iter := m.Client.ListDividends(context.Background(), params)

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

func (m *MassiveClient) GetIndicator(indicator, ticker string, data []byte) ([]byte, error) {
	switch indicator {
	case "daily":
		return m.GetDailyOpenCloseAgg(ticker)
	case "rsi":
		return m.GetRSI(ticker, data)
	}
	return []byte{}, fmt.Errorf("bad request type")
}

/* For Future Use...
func (m *MassiveClient) CallRSI(symbol string, request *MassiveRSIRequest) (*models.GetRSIResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	results, err := m.GetRSI(symbol, data)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	var rsiResponse models.GetRSIResponse
	err = json.Unmarshal(results, &rsiResponse)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	return &rsiResponse, nil
}

*/

func (m *MassiveClient) GetRSI(symbol string, data []byte) ([]byte, error) {
	var request MassiveRSIRequest
	err := json.Unmarshal(data, &request)
	if err != nil {
		logrus.Error(err.Error())
		return []byte(""), err
	}

	logrus.Debug("request date:", request.RequestDate)
	params := models.GetRSIParams{
		Ticker: symbol,
	}

	params.WithTimespan(models.Day).WithAdjusted(request.Adjusted).WithWindow(request.Window).WithSeriesType(models.Close).WithOrder(models.Desc).WithTimestamp(models.EQ, models.Millis(request.RequestDate))
	resp, err := m.Client.GetRSI(context.Background(), &params)
	if err != nil {
		logrus.Error(err.Error())
		return []byte("{}"), err
	}
	return json.Marshal(resp)
}

func (m *MassiveClient) GetDailyOpenCloseAgg(symbol string, args ...string) ([]byte, error) {
	reqDate, err := m.getRequestDateFromJulian(args...)
	if err != nil {
		logrus.Error("GetDailyOpenCloseAgg:", err.Error())
		return []byte("{}"), err
	}

	businessDate := business_days.GetBusinessDay(*reqDate)
	logrus.Debug("request date:", businessDate)
	params := models.GetDailyOpenCloseAggParams{
		Ticker: symbol,
		Date:   models.Date(businessDate),
	}

	resp, err := m.Client.GetDailyOpenCloseAgg(context.Background(), params.WithAdjusted(true))
	if err != nil {
		logrus.Error(err.Error())
		return []byte("{}"), err
	}
	return json.Marshal(resp)
}

func (m *MassiveClient) GetPreviousClose(symbol string) ([]byte, error) {
	params := models.GetPreviousCloseAggParams{
		Ticker: symbol,
	}

	resp, err := m.Client.GetPreviousCloseAgg(context.Background(), params.WithAdjusted(true))
	if err != nil {
		logrus.Error(err.Error())
		return []byte("{}"), err
	}
	return json.Marshal(resp)
}

func (m *MassiveClient) readRC() error {
	homeDir, ok := os.LookupEnv("HOME")
	if !ok {
		logrus.Info("HOME not set")
		return fmt.Errorf("HOME not set")
	}

	filePath := homeDir + "/.massive-client.rc"
	m.resources = make(map[string]any)
	content, err := os.ReadFile(filePath)

	if err != nil {
		logrus.Error("error reading resources file:", err.Error())
		return err
	}

	err = json.Unmarshal(content, &m.resources)
	if err != nil {
		logrus.Error(err.Error())
	}
	return err
}
