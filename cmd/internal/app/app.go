package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/handlers/indicators"
	"github.com/kpearce2430/stock-tools/cmd/internal/handlers/symbollist"
	"github.com/kpearce2430/stock-tools/model"
	polygonclient "github.com/kpearce2430/stock-tools/polygon-client"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/polygon-io/client-go/rest/models"
	"github.com/sirupsen/logrus"
	"net/http"
)

type App struct {
	Srv           *http.Server
	LookupSet     *model.LookUpSet
	PGXConn       *pgxpool.Pool
	Tickers       map[string]*model.Ticker
	StockCache    *stock_cache.Cache[models.GetDailyOpenCloseAggResponse]
	DividendCache *stock_cache.Cache[models.Dividend]
}

const (
	accountListRoute    = "/accountlist"
	dividendRoute       = "/dividend/:symbol"
	dividendCache       = "dividends"
	historicalDB        = "historical"
	historicalLoadRoute = "/historical"
	// historicalDeleteRoute = "/historical/:key"
	lookupsRoute         = "/lookups/:id"
	macdRoute            = "/macd"
	PortfolioValueDB     = "portfolio_value"
	PortfolioLoadDBRoute = "/portfoliovalue"
	pvRoute              = "/pv"
	pvSymbolRoute        = "/pv/:symbol"
	rsiRoute             = "/rsi"
	statusRoute          = "/status"
	stockCacheRoute      = "/stockcache/:symbol"
	stockcache           = "quotes"
	symbolListRoute      = "/symbol/list"
	symbolDetail         = "/symbol/detail"
	tickerInfoRoute      = "/tickerinfo/:symbol"
	transactionRoute     = "/transaction"
	TransactionTable     = "transactions"
	TransactionAllTable  = "all_transactions"
	worksheetRoute       = "/worksheet"
)

func (a *App) routes() {
	router := gin.Default()
	s := symbollist.NewSymbolList(a.PGXConn, a.LookupSet)
	router.GET(accountListRoute, s.AccountListGet)
	router.GET(dividendRoute, a.GetDividendCache)
	router.POST(historicalLoadRoute, a.LoadHistoricalData)
	// router.DELETE(historicalDeleteRoute, a.DeleteHistoricalData)
	router.POST(lookupsRoute, a.LoadLookups)
	router.GET(lookupsRoute, a.GetLookups)
	router.GET(macdRoute, indicators.GetMACDRouter)
	router.POST(pvRoute, a.LoadPortfolioValueHandler)
	router.POST(PortfolioLoadDBRoute, a.LoadDBPortfolioValueHandler)
	router.GET(pvSymbolRoute, a.GetPortfolioValueHandler)
	router.GET(rsiRoute, indicators.GetRsiRouter)
	router.GET(statusRoute, a.Status)
	router.GET(stockCacheRoute, a.GetStockCache)
	router.GET(symbolListRoute, s.SymbolListGet)
	router.POST(transactionRoute, a.LoadTransactionsHandler)
	router.GET(tickerInfoRoute, s.TickerInfoGet)
	router.GET(worksheetRoute, a.CreateWorksheetHandler)
	router.GET(symbolDetail, a.CreateSymbolDetailHandler)
	a.Srv.Handler = router
}

func NewApp(port string) *App {
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		panic(err)
	}

	a := App{
		Srv: &http.Server{
			Addr: port,
		},
		LookupSet: nil,
		Tickers:   make(map[string]*model.Ticker),
		PGXConn:   pgxConn,
	}

	status, err := a.PostgresCheck()
	switch {
	case err != nil:
		panic(err.Error())
	case status == false:
		logrus.Fatal("Postgres not ready")
	}

	lookupSet := utils.GetEnv("LOOKUPS_SET", "2")
	a.LookupSet, err = a.getLookupsFromDatabase(lookupSet)
	if err != nil {
		logrus.Error("Error loading lookups:", err.Error())
	}

	quoteConfig := couch_database.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", stockcache),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	a.StockCache, err = stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&quoteConfig, polygonclient.NewPolygonClient(""))
	if err != nil {
		logrus.Fatal("Error Creating Stock Cache:", err.Error())
		return nil
	}

	divConfig := couch_database.DatabaseConfig{
		DatabaseName: utils.GetEnv("DIV_COUCHDB_DATABASE", dividendCache),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	a.DividendCache, err = stock_cache.NewCache[models.Dividend](&divConfig, polygonclient.NewPolygonClient(""))
	if err != nil {
		logrus.Error("Error Creating Dividend Cache:", err.Error())
		panic(err.Error())
	}

	if status := a.CouchDBCheck(); status != true {
		logrus.Fatal("CouchDB Not Up")
		return nil
	}

	a.routes()
	a.setLogging()
	return &a
}

// SetLogging checks to see if the LOG_LEVEL environmental variable is set to
// override the default.
func (a *App) setLogging() {
	logLevel, err := logrus.ParseLevel(utils.GetEnv("LOG_LEVEL", "info"))
	if err != nil {
		defer logrus.Infof("Unknown log level %s, setting to `info`.", logLevel)
	}
	logrus.SetLevel(logLevel)
	fmtr := &logrus.TextFormatter{
		PadLevelText:     true,
		QuoteEmptyFields: true,
	}
	logrus.SetFormatter(fmtr)
	logrus.SetReportCaller(true)
}
