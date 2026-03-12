package mock_client

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/massive-com/client-go/v2/rest/models"
)

type MockClient struct {
	client *http.Client
}

func New() *MockClient {
	return &MockClient{
		client: &http.Client{},
	}
}

//
//{
//"status":"OK",
//"symbol":"AAPL",
//"from":"2026-02-25",
//"open":271.78,
//"high":274.94,
//"low":271.05,
//"close":274.23,
//"volume":33714342.626628,
//"afterHours":274.2,
//"preMarket":272
//}

func (m *MockClient) GetData(ticker string, args ...string) ([]byte, error) {

	today := time.Now().Format("2006-01-02")

	response := models.GetDailyOpenCloseAggResponse{
		Symbol:     ticker,
		From:       today,
		Open:       271.78,
		High:       274.94,
		Low:        271.05,
		Close:      274.23,
		Volume:     33714342.626628,
		AfterHours: 274.2,
		PreMarket:  272,
	}

	bytes, err := json.Marshal(response)

	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (m *MockClient) GetIndicator(indicator, ticker string, data []byte) ([]byte, error) {
	return []byte{}, nil
}

func (m *MockClient) GetDataSet(ticker string, args ...string) ([]byte, error) {
	return []byte{}, nil
}
