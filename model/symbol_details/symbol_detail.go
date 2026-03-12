package symbol_details

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	businessdays "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/postgres"
	"github.com/kpearce2430/stock-tools/model/portfolio_value"
	"github.com/kpearce2430/stock-tools/model/symbol"
	ticker2 "github.com/kpearce2430/stock-tools/model/ticker"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/massive-com/client-go/v2/rest/models"
	"github.com/sirupsen/logrus"
)

type SymbolDetail struct {
	StockCache *stock_cache.Cache[models.GetDailyOpenCloseAggResponse]
	Symbol     string  `json:"symbol,omitempty"`
	Month      int     `json:"month,omitempty"`
	Year       int     `json:"year,omitempty"`
	Quantity   float64 `json:"quantity,omitempty"`
	Price      float64 `json:"price,omitempty"`
	Dividends  float64 `json:"dividends,omitempty"`
}

type SymbolDetailSet struct {
	pgxConn    *pgxpool.Pool
	StockCache *stock_cache.Cache[models.GetDailyOpenCloseAggResponse]
	Symbol     string
	Info       []*SymbolDetail `json:"info,omitempty"`
}

func (set *SymbolDetailSet) String() string {
	b, err := json.MarshalIndent(set, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func NewSymbolDetail(cache *stock_cache.Cache[models.GetDailyOpenCloseAggResponse], symbol string, year, month int) *SymbolDetail {
	return &SymbolDetail{
		StockCache: cache,
		Symbol:     symbol,
		Year:       year,
		Month:      month,
	}
}

func (s *SymbolDetail) Value() float64 {
	return s.Quantity * s.Price
}

func (s *SymbolDetail) String() string {
	b, err := json.Marshal(s)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func NewSymbolDetailSet(pgxConn *pgxpool.Pool, symbol string) *SymbolDetailSet {
	return &SymbolDetailSet{
		pgxConn: pgxConn,
		Symbol:  symbol,
	}
}

func (set *SymbolDetailSet) Create(date time.Time, monthsAgo int) error {
	year := date.Year()
	month := date.Month()
	for m := 0; m < monthsAgo; m++ {
		sd := NewSymbolDetail(set.StockCache, set.Symbol, year, int(month))
		if err := sd.Set(set.pgxConn); err != nil {
			logrus.Error(err.Error())
			return err
		}
		set.Info = append(set.Info, sd)
		month = month - 1
		if month < 1 {
			month = 12
			year--
		}
	}
	return nil
}

func (s *SymbolDetail) setMutualFundPrice() error {
	month := s.Month + 1
	year := s.Year
	if month > 12 {
		month = 1
		year++
	}

	pgxConn, err := postgres.ConnectToPostgres()
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	defer pgxConn.Close()

	pSet := portfolio_value.NewSet(pgxConn, portfolio_value.PortfolioValueTable, s.Symbol)
	err = pSet.GetSymbolYearMonth(year, month)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	pv, err := pSet.LastInSet()
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	s.Price = pv.Quote

	return nil
}

func (s *SymbolDetail) setStockPrice() error {
	month := s.Month + 1
	year := s.Year
	if month > 12 {
		month = 1
		year++
	}
	date := time.Date(year, time.Month(month), 01, 00, 00, 00, 00, time.UTC).Add(time.Duration(-24) * time.Hour)
	if date.After(time.Now()) {
		logrus.Info("setting date to date")
		date = time.Now()
	}
	logrus.Debugf("start:%d%03d", date.Year(), date.YearDay())
	date = businessdays.GetBusinessDay(date)
	jDate := fmt.Sprintf("%d%03d", date.Year(), date.YearDay())
	logrus.Debug("jDate:", jDate)

	// var p *models.GetDailyOpenCloseAggResponse
	//config := couch_database.DatabaseConfig{
	//	DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "cache"),
	//	CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
	//	Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
	//	Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	//}
	//stockCache, err := stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&config, massive_client.New())
	//if err != nil {
	//	logrus.Fatal("Error Creating Stock Cache:", err.Error())
	//	return nil
	//}
	//
	//if _, err := stockCache.DatabaseExists(); err != nil {
	//	if ok := stockCache.DatabaseCreate(); !ok {
	//		err := fmt.Errorf("couchdb error with %s", config.DatabaseName)
	//		logrus.Error(err)
	//	}
	//}
	// logrus.Debug("jDate:", jDate)

	p, err := s.StockCache.GetCache(s.Symbol, jDate)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	if p == nil {
		err := fmt.Errorf("no response from cache %s:%s", s.Symbol, jDate)
		logrus.Error(err.Error())
		return err
	}
	s.Price = p.Close
	return nil
}

func (s *SymbolDetail) SetNumberOfShares(pg *pgxpool.Pool) error {
	month := s.Month + 1
	year := s.Year
	if month > 12 {
		month = 1
		year++
	}

	tickerSet := ticker2.NewTickerSet()
	ts := transaction.NewTransactionSet()
	if err := ts.SymbolGetBeforeDate(context.Background(), pg, s.Symbol, year, month, 01); err != nil {
		logrus.Error(err.Error())
		return err
	}

	if len(ts.TransactionRows) <= 0 {
		s.Quantity = 0.00
		return nil
	}

	if err := tickerSet.LoadTickerSet(ts); err != nil {
		logrus.Error(err.Error())
		return err
	}

	ticker, ok := tickerSet.GetTicker(s.Symbol)
	if !ok {
		err := fmt.Errorf("error loading ticker %s", s.Symbol)
		logrus.Error(err)
		return err
	}
	s.Quantity = ticker.TotalShares(true)
	return nil
}

func (s *SymbolDetail) SetDividends(pg *pgxpool.Pool) error {

	tickerSet := ticker2.NewTickerSet()
	ts := transaction.NewTransactionSet()
	if err := ts.ForMonth(context.Background(), pg, s.Symbol, s.Year, s.Month); err != nil {
		logrus.Error(err.Error())
		return err
	}

	if ts.Count() == 0 {
		s.Dividends = 0.00
		return nil
	}

	if err := tickerSet.LoadTickerSet(ts); err != nil {
		logrus.Error(err.Error())
		return err
	}
	ticker, ok := tickerSet.GetTicker(s.Symbol)
	if !ok {
		err := fmt.Errorf("error loading ticker %s", s.Symbol)
		logrus.Error(err)
		return err
	}
	s.Dividends = ticker.DividendsPaid() + ticker.InterestIncome()
	return nil
}

func (s *SymbolDetail) SetPrice() error {
	symbolType, ok := symbol.SymbolTypeMap[s.Symbol]
	if !ok {
		err := fmt.Errorf("type for symbol %s", s.Symbol)
		logrus.Error(err)
		return err
	}

	month := s.Month + 1
	year := s.Year
	if month > 12 {
		month = 1
		year++
	}

	switch symbolType {
	case "Stock", "Other":
		return s.setStockPrice()
	case "Mutual Fund":
		return s.setMutualFundPrice()
	case "Bond":
		s.Price = 100.00
	default:
		err := fmt.Errorf("unknown type %s", symbolType)
		logrus.Error(err)
		return err
	}
	return nil
}

func (s *SymbolDetail) Set(pgxConn *pgxpool.Pool) error {
	if err := s.SetNumberOfShares(pgxConn); err != nil {
		logrus.Error(err.Error())
		return err
	}
	if err := s.SetDividends(pgxConn); err != nil {
		logrus.Error(err.Error())
		return err
	}
	if err := s.SetPrice(); err != nil {
		logrus.Error(err.Error())
		return err
	}
	return nil
}
