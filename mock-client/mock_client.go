package mock_client

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
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

	value := rand.Float64() * 100
	diffValue := rand.Float64()
	negPos := rand.IntN(1)
	var chgValue float64
	switch negPos {
	case 0:
		chgValue = value + diffValue
	default:
		value = value - diffValue
	}

	value, err := strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	if err != nil {
		return []byte{}, err
	}

	chgValue, err = strconv.ParseFloat(fmt.Sprintf("%.2f", chgValue), 64)
	if err != nil {
		return []byte{}, err
	}

	response := models.GetDailyOpenCloseAggResponse{
		Symbol:     ticker,
		From:       today,
		Open:       value,
		High:       chgValue,
		Low:        chgValue,
		Close:      chgValue,
		Volume:     rand.Float64() * 1000000000,
		AfterHours: chgValue,
		PreMarket:  value,
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
