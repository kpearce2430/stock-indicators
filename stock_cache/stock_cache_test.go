package stock_cache_test

import (
	"context"
	"fmt"
	business_days "github.com/kpearce2430/keputils/business-days"
	couchdatabase "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	polygonclient "github.com/kpearce2430/stock-tools/polygon-client"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/polygon-io/client-go/rest/models"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	couchDBServer, _ := couchdatabase.CreateCouchDBServer(ctx)
	defer func() {
		if err := couchDBServer.Terminate(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	ip, err := couchDBServer.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}
	mappedPort, err := couchDBServer.MappedPort(ctx, "5984")
	if err != nil {
		log.Fatal(err)
	}
	url := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	_ = os.Setenv("COUCHDB_DATABASE", "iex-quotes")
	_ = os.Setenv("COUCHDB_URL", url)
	_ = os.Setenv("COUCHDB_USER", "admin")
	_ = os.Setenv("COUCHDB_PASSWORD", "password")

	dataConfig := couchdatabase.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "quotes"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	dataStore := couchdatabase.NewDataStore[models.GetDailyOpenCloseAggResponse](&dataConfig)
	_, err = dataStore.DatabaseExists()
	if err != nil {
		if dataStore.DatabaseCreate() == false {
			panic(err.Error())
			return
		}
		logrus.Info("Database `", dataConfig.DatabaseName, "` created")
	}
	m.Run()
}

func TestNewCache(t *testing.T) {

	quoteConfig := couchdatabase.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "quotes"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}

	client := polygonclient.NewPolygonClient("")

	cache, err := stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&quoteConfig, client)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if !cache.CouchDBUp() {
		t.Log("CouchDB Not Up")
		t.FailNow()
	}

	t.Log(">>", cache.GetConfig())
}

func TestCache_GetStockQuote(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}
	tickers := []string{"HD", "CSX", "AAPL"}

	quoteConfig := couchdatabase.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "quotes"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	poly := polygonclient.NewPolygonClient("")
	cache, err := stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&quoteConfig, poly)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if !cache.CouchDBUp() {
		t.Log("CouchDB Not Up")
		t.FailNow()
	}

	for _, sym := range tickers {
		t.Run(sym, func(t *testing.T) {

			tm := time.Now()
			tm = business_days.GetBusinessDay(tm)
			doc, err := cache.GetCache(sym, utils.JulDateFromTime(tm))
			if err != nil {
				s := err.Error()
				t.Log(s)
				if err.Error() != "{\"error\":\"not_found\",\"reason\":\"missing\"}\n" {
					t.Log(err.Error())
					t.FailNow()
				}
			}
			if doc == nil {
				t.Log("No document found")
			}
			t.Log(doc)

			// tm := time.Date(2024, 01, 22, 19, 00, 00, 00, time.UTC)
			doc, err = cache.DocumentGet(fmt.Sprintf("%s:%s", sym, utils.JulDateFromTime(tm)))
			if err != nil {
				t.Log(err.Error())
				t.FailNow()
			}
			if doc == nil {
				t.Log("Still missing document")
				t.FailNow()
			}
		})
	}
}

func TestCache_GetPastStockQuote(t *testing.T) {
	t.Skip("skipping")
	tickers := []string{"HD", "CSX", "AAPL"}

	quoteConfig := couchdatabase.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "quotes"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	poly := polygonclient.NewPolygonClient("")
	cache, err := stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&quoteConfig, poly)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if !cache.CouchDBUp() {
		t.Log("CouchDB Not Up")
		t.FailNow()
	}

	tm := time.Date(2023, 12, 25, 00, 00, 00, 00, time.UTC)
	jDate := utils.JulDateFromTime(business_days.GetBusinessDay(tm))

	for _, sym := range tickers {
		t.Run(sym, func(t *testing.T) {
			doc, err := cache.GetCache(sym, jDate)
			if err != nil {
				s := err.Error()
				t.Log(s)
				if err.Error() != "{\"error\":\"not_found\",\"reason\":\"missing\"}\n" {
					t.Log(err.Error())
					t.FailNow()
				}
			}
			if doc == nil {
				t.Log("No document found")
			}
			t.Log(doc)

			key := fmt.Sprintf("%s:%s", sym, jDate)
			doc, err = cache.DocumentGet(key)
			if err != nil {
				t.Log(err.Error())
				t.FailNow()
			}
			if doc == nil {
				t.Log("Still missing document")
				t.FailNow()
			}
		})
	}
}

func TestCache_GetStockDividends(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}

	tickers := []string{"HD", "CSX", "AAPL"}

	quoteConfig := couchdatabase.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "quotes"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	poly := polygonclient.NewPolygonClient("")
	cache, err := stock_cache.NewCache[models.Dividend](&quoteConfig, poly)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if !cache.CouchDBUp() {
		t.Log("CouchDB Not Up")
		t.FailNow()
	}

	for _, sym := range tickers {
		t.Run(sym, func(t *testing.T) {
			doc, err := cache.GetCacheSet(sym)
			if err != nil {
				s := err.Error()
				t.Log(s)
				if err.Error() != "{\"error\":\"not_found\",\"reason\":\"missing\"}\n" {
					t.Log(err.Error())
					t.FailNow()
				}
			}
			if doc == nil {
				t.Log("No document found")
			}
			t.Log(doc)

			tm := business_days.GetBusinessDay(time.Now())
			jDate := fmt.Sprintf("%d%03d", tm.Year(), tm.YearDay())
			key := fmt.Sprintf("%s:%s", "HD", jDate)

			doc, err = cache.DocumentGet(key)
			if err != nil {
				t.Log(err.Error())
				t.FailNow()
			}
			if doc == nil {
				t.Log("Still missing document ", key)
				t.FailNow()
			}
		})
	}
}

/*
func TestGetIndicator(t *testing.T) {

	type indicatorTests struct {
		Name          string
		Args          []string
		ExpectedError bool
	}

	tests := []indicatorTests{
		{
			Name:          "HD Today",
			Args:          []string{"HD", "rsi"},
			ExpectedError: false,
		},
		{
			Name:          "HD Dec 28 2023",
			Args:          []string{"HD", "rsi", "2023-12-28"},
			ExpectedError: false,
		},
		{
			Name:          "HD Bad Indicator",
			Args:          []string{"HD", "2023-12-28"},
			ExpectedError: true,
		},
		{
			Name:          "HD Bad Date",
			Args:          []string{"HD", "rsi", "2023-28-28"},
			ExpectedError: true,
		},
	}

	quoteConfig := couchdatabase.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "rsi"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	poly := polygonclient.NewPolygonClient("")
	cache, err := stock_cache.NewCache[models.GetRSIResponse](&quoteConfig, poly)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if !cache.CouchDBUp() {
		t.Log("CouchDB Not Up")
		t.FailNow()
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			rsi, err := cache.GetIndicator(tc.Args...)
			if err != nil {
				t.Log(err.Error())
				if tc.ExpectedError == false {
					t.Fail()
				}
				return
			}

			t.Log(rsi)
		})
	}

}

*/
