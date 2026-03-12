package dividend_history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
)

const (
	dividendHistoryTable  = "dividend_history"
	dividendHistoryFields = "symbol, year, month, amount"
)

var (
	errInvalidArguments      = errors.New("invalid arguments")
	errInvalidYear           = errors.New("invalid year")
	errInvalidMonth          = errors.New("invalid month")
	errDividendEntryNotFound = errors.New("dividend history entry not found")
	errInvalidNumEntries     = errors.New("invalid number of dividend history entries not found")
)

type DividendHistory struct {
	pgxConn         *pgxpool.Pool
	Symbol          string           `json:"symbol"`
	DividendEntries []*DividendEntry `json:"entries"`
}

func NewDividendHistory(pgxConn *pgxpool.Pool, symbol string) *DividendHistory {
	return &DividendHistory{
		pgxConn: pgxConn,
		Symbol:  symbol,
	}
}

func (dh *DividendHistory) String() string {
	b, err := json.Marshal(dh)
	if err != nil {
		logrus.Errorf("failed to marshal dividend history: %v", err)
		panic(err)
	}
	return string(b)
}

// Sum returns the sum of all the dividend entries in the history.
func (dh *DividendHistory) Sum() float64 {
	amt := 0.00
	for _, d := range dh.DividendEntries {
		amt += d.Amount
	}
	return amt
}

func (dh *DividendHistory) Clear() {
	clear(dh.DividendEntries)
}

// GetYear will retrieve the dividend history for a given year.
func (dh *DividendHistory) GetYear(ctx context.Context, year int) error {
	now := time.Now()
	if !utils.IntInRange(year, 1980, now.Year()) {
		return errInvalidYear
	}

	for month := 1; month <= 12; month++ {
		if err := dh.GetYearMonth(ctx, year, month); err != nil {
			logrus.Error(err.Error())
			return err
		}
	}
	return nil
}

func (dh *DividendHistory) GetYearMonth(ctx context.Context, year, month int) error {
	de := NewDividendEntry(dh.Symbol, dh.pgxConn, year, month)
	err := de.GetYearMonth(ctx)
	if err != nil {
		return err
	}
	dh.DividendEntries = append(dh.DividendEntries, de)
	return nil
}

/*
// FromDB will retrieve the dividend history from Postgres.
func (dh *DividendHistory) FromDB(ctx context.Context, year, month int) error {
	now := time.Now()
	if year != 0 && !intInRange(year, 1980, now.Year()) {
		return errInvalidYear
	}

	if !intInRange(month, 0, 12) {
		return errInvalidMonth
	}

	needAnd := false
	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(dividendHistoryFields)
	sb.WriteString(" FROM ")
	sb.WriteString(dividendHistoryTable)
	sb.WriteString(" WHERE ")
	if dh.Symbol != "" {
		sb.WriteString(" symbol = '")
		sb.WriteString(dh.Symbol)
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

	rows, err := dh.pgxConn.Query(ctx, sb.String())
	defer rows.Close()
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	// Iterate through the result set
	num := 0
	for rows.Next() {
		var d DividendEntry
		err = rows.Scan(&d.Symbol, &d.Year, &d.Month, &d.Amount)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		dh.DividendEntries = append(dh.DividendEntries, &d)
		num++
	}

	return nil
}
*/

func DividendHistoryFromDB(ctx context.Context, pgxConn *pgxpool.Pool, symbol string, year, month int) (*DividendHistory, error) {
	now := time.Now()
	if symbol == "" && !utils.IntInRange(year, 1980, now.Year()) && !utils.IntInRange(month, 1, 12) {
		return nil, errInvalidArguments
	}

	if year != 0 && !utils.IntInRange(year, 1980, now.Year()) {
		return nil, errInvalidYear
	}

	if !utils.IntInRange(month, 0, 12) {
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

	if utils.IntInRange(year, 1980, now.Year()) {
		if needAnd {
			sb.WriteString(" AND ")
		}
		sb.WriteString(" year = ")
		sb.WriteString(fmt.Sprintf("'%d'", year))
		needAnd = true
	}

	if utils.IntInRange(month, 1, 12) {
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

/*
func (dh *DividendHistory) AddEntry(symbol string, year, month int, amt float64) int {
	de := NewDividendEntry(symbol, dh.pgxConn, year, month)
	de.Amount = amt
	dh.DividendEntries = append(dh.DividendEntries, de)
	return len(dh.DividendEntries)
}

*/
