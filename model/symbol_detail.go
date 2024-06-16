package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	business_days "github.com/kpearce2430/keputils/business-days"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	polygonclient "github.com/kpearce2430/stock-tools/polygon-client"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/polygon-io/client-go/rest/models"
	"github.com/sirupsen/logrus"
	"time"
)

type SymbolDetail struct {
	FundsTable string  `json:"funds_table,omitempty"`
	Symbol     string  `json:"symbol,omitempty"`
	Month      int     `json:"month,omitempty"`
	Year       int     `json:"year,omitempty"`
	Quantity   float64 `json:"quantity,omitempty"`
	Price      float64 `json:"price,omitempty"`
	Dividends  float64 `json:"dividends,omitempty"`
}

type SymbolDetailSet struct {
	pgxConn    *pgxpool.Pool
	Symbol     string          `json:"symbol,omitempty"`
	FundsTable string          `json:"fundsTable,omitempty"`
	Info       []*SymbolDetail `json:"info,omitempty"`
}

func (set *SymbolDetailSet) String() string {
	b, err := json.Marshal(set)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func NewSymbolDetail(funds, symbol string, year, month int) *SymbolDetail {
	return &SymbolDetail{
		Symbol:     symbol,
		FundsTable: funds,
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

func NewSymbolDetailSet(pgxConn *pgxpool.Pool, symbol, table string) *SymbolDetailSet {
	return &SymbolDetailSet{
		pgxConn:    pgxConn,
		Symbol:     symbol,
		FundsTable: table,
	}
}

func (set *SymbolDetailSet) Create(date time.Time, monthsAgo int) error {
	year := date.Year()
	month := date.Month()
	for m := 0; m < monthsAgo; m++ {
		sd := NewSymbolDetail(set.FundsTable, set.Symbol, year, int(month))
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

	date := time.Date(year, time.Month(month), 01, 00, 00, 00, 00, time.UTC)

	// date := business_days.GetBusinessDay(time.Date(year, time.Month(month), 01, 00, 00, 00, 00, time.UTC).Add(-24 * time.Hour))

	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	defer pgxConn.Close()

	hs := NewHistoricalDataSet(pgxConn, s.FundsTable)
	hr, err := hs.Last(s.Symbol, date)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	s.Price = hr.Close

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
	date = business_days.GetBusinessDay(date)
	jDate := fmt.Sprintf("%d%03d", date.Year(), date.YearDay())
	logrus.Debug("jDate:", jDate)

	var p *models.GetDailyOpenCloseAggResponse
	config := couch_database.DatabaseConfig{
		DatabaseName: utils.GetEnv("CACHE_COUCHDB_DATABASE", "cache"),
		CouchDBUrl:   utils.GetEnv("COUCHDB_URL", "http://localhost:5984"),
		Username:     utils.GetEnv("COUCHDB_USERNAME", "admin"),
		Password:     utils.GetEnv("COUCHDB_PASSWORD", "password"),
	}
	stockCache, err := stock_cache.NewCache[models.GetDailyOpenCloseAggResponse](&config, polygonclient.NewPolygonClient(""))
	if err != nil {
		logrus.Fatal("Error Creating Stock Cache:", err.Error())
		return nil
	}

	if _, err := stockCache.DatabaseExists(); err != nil {
		if ok := stockCache.DatabaseCreate(); !ok {
			err := fmt.Errorf("couchdb error with %s", config.DatabaseName)
			logrus.Error(err)
		}
	}

	logrus.Debug("jDate:", jDate)
	p, err = stockCache.GetCache(s.Symbol, jDate)
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

	tickerSet := NewTickerSet()
	ts := NewTransactionSet()
	if err := ts.TransactionsSymbolGetBeforeDate(context.Background(), pg, s.Symbol, year, month, 01); err != nil {
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

	tickerSet := NewTickerSet()
	ts := NewTransactionSet()
	if err := ts.TransactionsForMonth(context.Background(), pg, s.Symbol, s.Year, s.Month); err != nil {
		logrus.Error(err.Error())
		return err
	}

	if ts.NumberOfTransactions() == 0 {
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

	symbolType, ok := SymbolTypeMap[s.Symbol]
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
