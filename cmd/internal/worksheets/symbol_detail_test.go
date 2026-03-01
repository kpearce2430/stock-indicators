package worksheets_test

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	business_days "github.com/kpearce2430/keputils/business-days"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	massive_client "github.com/kpearce2430/stock-tools/massive-client"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/massive-com/client-go/v2/rest/models"

	"testing"

	// "github.com/kpearce2430/stock-tools/cmd/internal/app"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
)

const (
	stockcache                     = "quotes"
	fundHistory                    = "test_history"
	symbolDetailsWorkSheetFileName = "SymbolsDetails.xlsx"
)

func TestWorkSheet_SymbolsDetails(t *testing.T) {
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	w := worksheets.New(pgxConn)
	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	quoteConfig := couch_database.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", stockcache),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}

	w.StockCache, err = stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&quoteConfig, massive_client.New())
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
		t.Fatal("unable to create cache database")
	}

	start := business_days.GetBusinessDay(time.Date(2023, 12, 31, 00, 00, 00, 00, time.UTC))

	if err = w.SymbolsDetails("TickerInfo", "HD", fundHistory, start, 24); err != nil {
		t.Error(err)
		return
	}

	if err = w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Log(err.Error())
	}

	if err = w.StockFile.Save(symbolDetailsWorkSheetFileName); err != nil {
		t.Error(err)
		return
	}
	t.Log("completed: ", time.Now().Sub(start))
}
