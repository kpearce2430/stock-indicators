package historical_test

/*
import (
	_ "embed"

	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/model/historical"
	"github.com/sirupsen/logrus"

	"testing"
	"time"
)

const (
	historicalTable = "historical"
)

var (
	//go:embed testdata/usaix_hist.csv
	testHistoricalData []byte
)

func TestMain(m *testing.M) {
	logrus.SetReportCaller(true)
	m.Run()
}

func TestHistorical_LoadHistorical(t *testing.T) {
	t.Skip("Skip historical tests")
	if err := historical.LoadCouchDB(historicalTable, string(testHistoricalData), "Random", "USAIX"); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	type HistoryTest struct {
		Key   string
		Found bool
	}

	tests := []HistoryTest{
		{
			Key:   "2024001:USAIX",
			Found: false,
		},
		{
			Key:   "2024002:USAIX",
			Found: true,
		},
		{
			Key:   "2024003:USAIX",
			Found: true,
		},
		{
			Key:   "2024004:USAIX",
			Found: true,
		},
		{
			Key:   "2024099:USAIX",
			Found: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Key, func(t *testing.T) {
			h, err := historical.HistoricalCacheGet(historicalTable, tc.Key)
			if err != nil {
				t.Log(err.Error())
				if tc.Found == true {
					t.Fail()
				}
				return
			}
			if h != nil {
				t.Log(h)
			}
		})
	}

	for _, tc := range tests {
		t.Run(tc.Key, func(t *testing.T) {
			rev, err := historical.HistoricalCacheDelete(historicalTable, tc.Key)
			if err != nil {
				t.Log(err.Error())
				if tc.Found == true {
					t.Fail()
				}
				return
			}
			t.Log(tc.Key, rev, " deleted")
		})
	}
}

func TestHistorical_LoadHistoricalDB(t *testing.T) {
	t.Skip("Skip historical tests")
	const symbol = "USAIX"
	const source = "testcases"
	const fundHistory = "fund_history"

	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		t.Error(err.Error())
		return
	}

	ds := historical.NewHistoricalDataSet(pgxConn)
	if err := ds.LoadSet(string(testHistoricalData), source, symbol); err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	feb1 := time.Date(2024, 02, 01, 00, 00, 00, 00, time.UTC)
	hist, err := ds.Last(symbol, feb1)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	t.Log("last>>>", hist)
}

*/
