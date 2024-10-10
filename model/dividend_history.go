package model

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type DividendEntry struct {
	Symbol string  `json:"symbol"`
	Month  int     `json:"month"`
	Year   int     `json:"year"`
	Amount float64 `json:"amount"`
}

func NewDividendEntry(symbol string, year, month int) *DividendEntry {
	return &DividendEntry{
		Symbol: symbol,
		Year:   year,
		Month:  month,
	}
}

type DividendHistory struct {
	Symbol          string           `json:"symbol"`
	DividendEntries []*DividendEntry `json:"entries"`
}

func (d *DividendHistory) Sum() float64 {
	amt := 0.00
	for _, dh := range d.DividendEntries {
		amt += dh.Amount
	}
	return amt
}

func NewDividendHistory(symbol string) *DividendHistory {
	return &DividendHistory{
		Symbol: symbol,
	}
}

func (d *DividendEntry) GetDividendForYearMonth(pg *pgxpool.Pool) error {
	tSet := NewTransactionSet()
	if err := tSet.TransactionsForMonth(context.Background(), pg, d.Symbol, d.Year, d.Month); err != nil {
		d.Amount = 0.00
		logrus.Error(err.Error())
		return err
	}

	logrus.Debugf("%s Found %d transactions", d.Symbol, len(tSet.TransactionRows))
	if len(tSet.TransactionRows) <= 0 {
		d.Amount = 0.00
		return nil
	}

	tickerSet := NewTickerSet()
	if err := tickerSet.LoadTickerSet(tSet); err != nil {
		d.Amount = 0.00
		logrus.Error(err.Error())
		return err
	}

	ticker, ok := tickerSet.GetTicker(d.Symbol)
	if !ok {
		d.Amount = 0.00
		err := fmt.Errorf("error locating dividend for %s date[%04d/%02d]", d.Symbol, d.Year, d.Month)
		logrus.Error(err.Error())
		return err
	}

	d.Amount = ticker.Dividends()
	return nil
}
