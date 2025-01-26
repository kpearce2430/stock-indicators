package worksheets_test

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	business_days "github.com/kpearce2430/keputils/business-days"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	polygonclient "github.com/kpearce2430/stock-tools/polygon-client"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/polygon-io/client-go/rest/models"
	"strings"
	"time"

	// "github.com/kpearce2430/stock-tools/cmd/internal/app"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/xuri/excelize/v2"
	"testing"
)

const (
	stockcache                     = "quotes"
	fundHistory                    = "test_history"
	symbolDetailsWorkSheetFileName = "SymbolsDetails.xlsx"
)

func TestWorkSheet_SymbolsDetails(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	w := worksheets.NewWorkSheet(excelize.NewFile(), pgxConn)
	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	quoteConfig := couch_database.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", stockcache),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}

	w.StockCache, err = stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&quoteConfig, polygonclient.NewPolygonClient(""))
	if !w.StockCache.CouchDBUp() {
		t.Fatal("couchdb not up")
		return
	}

	_, err = w.StockCache.DatabaseExists()
	if err != nil {
		if w.StockCache.DatabaseCreate() == false {
			t.Fatal("unable to create cache database")
			return
		}
	}

	_, err = w.StockCache.DatabaseExists()
	if err != nil {
		// if w.DividendCache.DatabaseCreate() == false {
		t.Fatal("unable to create cache database")
		// }
	}

	start := business_days.GetBusinessDay(time.Date(2023, 12, 31, 00, 00, 00, 00, time.UTC))

	if err := w.SymbolsDetails("TickerInfo", "HD", fundHistory, start, 24); err != nil {
		t.Fatal(err.Error())
		return
	}

	if err := w.File.DeleteSheet("Sheet1"); err != nil {
		t.Log(err.Error())
	}

	if err := w.File.SaveAs(symbolDetailsWorkSheetFileName); err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("completed: ", time.Now().Sub(start))
}
