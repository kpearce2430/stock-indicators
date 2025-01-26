package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
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
		Amount: 0.00,
	}
}

type DividendHistory struct {
	Symbol          string           `json:"symbol"`
	DividendEntries []*DividendEntry `json:"entries"`
}

const (
	dividendHistoryTable  = "dividend_history"
	dividendHistoryFields = "symbol, year, month, amount"
)

var (
	errInvalidArguments      = errors.New("invalid arguments")
	errInvalidYear           = errors.New("invalid year")
	errInvalidMonth          = errors.New("invalid month")
	errDividendEntryNotFound = errors.New("dividend history entry not found")
)

// check to see if an int is between two numbers.
func intInRange(num, min, max int) bool {
	if min >= max {
		panic("invalid arguments")
	}

	if num >= min && num <= max {
		// fmt.Println(num, "is in the range")
		return true
	}
	// fmt.Println(num, "is not in the range")
	return false
}

func (d *DividendHistory) String() string {
	b, err := json.Marshal(d)
	if err != nil {
		logrus.Errorf("failed to marshal dividend history: %v", err)
		panic(err)
	}
	return string(b)
}

func (d *DividendHistory) Sum() float64 {
	amt := 0.00
	for _, dh := range d.DividendEntries {
		amt += dh.Amount
	}
	return amt
}

func (d *DividendHistory) ToDB(ctx context.Context, pgxConn *pgxpool.Pool) error {
	for _, entry := range d.DividendEntries {
		err := entry.ToDB(ctx, pgxConn)
		if err != nil {
			return err
		}
	}
	return nil
}

func DividendHistoryFromDB(ctx context.Context, pgxConn *pgxpool.Pool, symbol string, year, month int) (*DividendHistory, error) {

	now := time.Now()
	if symbol == "" && !intInRange(year, 1980, now.Year()) && !intInRange(month, 1, 12) {
		return nil, errInvalidArguments
	}

	if year != 0 && !intInRange(year, 1980, now.Year()) {
		return nil, errInvalidYear
	}

	if !intInRange(month, 0, 12) {
		return nil, errInvalidMonth
	}

	needAnd := false
	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(dividendHistoryFields)
	sb.WriteString(" FROM ")
	sb.WriteString(dividendHistoryTable)
	sb.WriteString(" WHERE ")
	if symbol != "" {
		sb.WriteString(" symbol = '")
		sb.WriteString(symbol)
		needAnd = true
		sb.WriteString("'")
	}

	if intInRange(year, 1980, now.Year()) {
		if needAnd {
			sb.WriteString(" AND ")
		}
		sb.WriteString(" year = ")
		sb.WriteString(fmt.Sprintf("'%d'", year))
		needAnd = true
	}

	if intInRange(month, 1, 12) {
		if needAnd {
			sb.WriteString(" AND ")
		}
		sb.WriteString(" month = ")
		sb.WriteString(fmt.Sprintf("'%d'", month))
	}

	rows, err := pgxConn.Query(ctx, sb.String())
	defer rows.Close()
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	dh := &DividendHistory{
		Symbol: symbol,
	}
	// Iterate through the result set
	num := 0
	for rows.Next() {
		var d DividendEntry
		err = rows.Scan(&d.Symbol, &d.Year, &d.Month, &d.Amount)
		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}
		dh.DividendEntries = append(dh.DividendEntries, &d)
		num++
	}

	return dh, nil
}

func (d *DividendHistory) AddEntry(symbol string, year, month int, amt float64) int {
	dh := NewDividendEntry(symbol, year, month)
	dh.Amount = amt
	d.DividendEntries = append(d.DividendEntries, dh)
	return len(d.DividendEntries)
}

func NewDividendHistory(symbol string) *DividendHistory {
	return &DividendHistory{
		Symbol: symbol,
	}
}

func GetDividendEntryForYearMonth(pg *pgxpool.Pool, symbol string, year, month int) (*DividendEntry, error) {
	today := time.Now()
	requested := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	cutOver := time.Date(today.Year()-1, today.Month(), 1, 0, 0, 0, 0, time.UTC)

	if requested.Before(cutOver) {
		// Check the DB
		d, err := DividendEntryFromDB(context.Background(), pg, symbol, year, month)
		if err == nil {
			return d, nil
		}
		if !errors.Is(err, errDividendEntryNotFound) {
			logrus.Error(err.Error())
			return nil, err
		}
		logrus.Info(err.Error())
	}

	d := NewDividendEntry(symbol, year, month)
	tSet := NewTransactionSet()
	if err := tSet.TransactionsForMonth(context.Background(), pg, d.Symbol, d.Year, d.Month); err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	logrus.Debugf("%s Found %d transactions", d.Symbol, len(tSet.TransactionRows))
	if len(tSet.TransactionRows) <= 0 {
		d.Amount = 0.00
		err := d.ToDB(context.Background(), pg)
		if err != nil {
			logrus.Error(err.Error())
		}
		return d, nil
	}

	tickerSet := NewTickerSet()
	if err := tickerSet.LoadTickerSet(tSet); err != nil {
		d.Amount = 0.00
		logrus.Error(err.Error())
		return nil, err
	}

	ticker, ok := tickerSet.GetTicker(d.Symbol)
	if !ok {
		err := fmt.Errorf("error locating dividend for %s date[%04d/%02d]", d.Symbol, d.Year, d.Month)
		logrus.Error(err.Error())
		return nil, err
	}

	d.Amount = ticker.Dividends()

	err := d.ToDB(context.Background(), pg)
	if err != nil {
		logrus.Error(err.Error())
	}
	return d, err
}

func DividendEntryFromDB(ctx context.Context, pgxConn *pgxpool.Pool, symbol string, year, month int) (*DividendEntry, error) {
	selectStatement := fmt.Sprintf(
		"SELECT %s From %s WHERE symbol = '%s' and year = '%d' and month = '%d' ",
		dividendHistoryFields, dividendHistoryTable, symbol, year, month)

	rows, err := pgxConn.Query(ctx, selectStatement)
	defer rows.Close()
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	var d DividendEntry
	// Iterate through the result set
	num := 0
	for rows.Next() {
		err = rows.Scan(&d.Symbol, &d.Year, &d.Month, &d.Amount)
		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}
		num++
	}

	switch num {
	case 0:
		return nil, errDividendEntryNotFound
	case 1:
		return &d, nil
	}

	return nil, fmt.Errorf("invalid number of dividend history entries not found: %d", num)

}

func (d *DividendEntry) ToDB(ctx context.Context, pgxConn *pgxpool.Pool) error {
	insertStatement := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES ('%s','%d','%d','%.2f') ON CONFLICT(symbol, year, month) DO UPDATE SET amount = EXCLUDED.amount;",
		dividendHistoryTable, dividendHistoryFields, d.Symbol, d.Year, d.Month, d.Amount)
	rows, err := pgxConn.Query(ctx, insertStatement)
	defer rows.Close()
	if err != nil {
		return err
	}
	return nil
}
