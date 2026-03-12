package dividend_history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	ticker2 "github.com/kpearce2430/stock-tools/model/ticker"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/sirupsen/logrus"
)

type DividendEntry struct {
	pgxConn *pgxpool.Pool
	Symbol  string  `json:"symbol"`
	Month   int     `json:"month"`
	Year    int     `json:"year"`
	Amount  float64 `json:"amount"`
}

func NewDividendEntry(symbol string, pgxConn *pgxpool.Pool, year, month int) *DividendEntry {
	return &DividendEntry{
		Symbol:  symbol,
		pgxConn: pgxConn,
		Year:    year,
		Month:   month,
		Amount:  0.00,
	}
}

func (de *DividendEntry) String() string {
	b, err := json.Marshal(de)
	if err != nil {
		logrus.Errorf("failed to marshal dividend history: %v", err)
		panic(err)
	}
	return string(b)
}

// GetDividendEntryForYearMonth returns the dividend entry for the given symbol and year/month.
func (de *DividendEntry) GetYearMonth(ctx context.Context) error {
	if de.pgxConn == nil {
		logrus.Error("pgxConn is nil")
		return errInvalidArguments
	}

	today := time.Now()
	requested := time.Date(de.Year, time.Month(de.Month), 1, 0, 0, 0, 0, time.UTC)
	cutOver := time.Date(today.Year()-1, today.Month(), 1, 0, 0, 0, 0, time.UTC)

	if requested.Before(cutOver) {
		// Check the DB
		err := de.FromDB(ctx)
		if err == nil {
			return nil
		}
		if !errors.Is(err, errDividendEntryNotFound) {
			logrus.Error(err.Error())
			return err
		}
		logrus.Debug(err.Error())
	}

	tSet := transaction.NewTransactionSet()
	if err := tSet.ForMonth(context.Background(), de.pgxConn, de.Symbol, de.Year, de.Month); err != nil {
		logrus.Error(err.Error())
		return err
	}

	logrus.Debugf("%s Found %d transactions", de.Symbol, len(tSet.TransactionRows))
	if len(tSet.TransactionRows) <= 0 {
		de.Amount = 0.00
		err := de.ToDB(ctx)
		if err != nil {
			logrus.Error(err.Error())
		}
		return nil
	}

	tickerSet := ticker2.NewTickerSet()
	if err := tickerSet.LoadTickerSet(tSet); err != nil {
		logrus.Error(err.Error())
		de.Amount = 0.00
		return err
	}

	ticker, ok := tickerSet.GetTicker(de.Symbol)
	if !ok {
		err := fmt.Errorf("error locating dividend for %s date[%04d/%02d]", de.Symbol, de.Year, de.Month)
		logrus.Error(err.Error())
		return err
	}

	de.Amount = ticker.Dividends()
	err := de.ToDB(ctx)
	if err != nil {
		logrus.Error(err.Error())
	}
	return err
}

func (de *DividendEntry) FromDB(ctx context.Context) error {
	if de.pgxConn == nil {
		return fmt.Errorf("pgxConn is nil")
	}

	if de.Symbol == "" {
		logrus.Error("symbol is empty")
		return errInvalidArguments
	}

	if utils.IntInRange(de.Year, 1980, time.Now().Year()) == false {
		logrus.Error("invalid year")
		return errInvalidYear
	}

	if utils.IntInRange(de.Month, 1, 12) == false {
		logrus.Error("invalid month")
		return errInvalidMonth
	}

	selectStatement := fmt.Sprintf(
		"SELECT %s From %s WHERE symbol = '%s' and year = '%d' and month = '%d' ",
		dividendHistoryFields, dividendHistoryTable, de.Symbol, de.Year, de.Month)

	rows, err := de.pgxConn.Query(ctx, selectStatement)
	defer rows.Close()

	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	// Iterate through the result set
	num := 0
	for rows.Next() {
		err = rows.Scan(&de.Symbol, &de.Year, &de.Month, &de.Amount)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		num++
	}

	switch num {
	case 0:
		return errDividendEntryNotFound
	case 1:
		return nil
	}
	logrus.Error(fmt.Sprintf("invalid number of dividend history entries found: %d", num))
	return errInvalidNumEntries
}

// ToDB will insert the dividends History into Postgres.
func (de *DividendEntry) ToDB(ctx context.Context) error {
	if de.pgxConn == nil {
		return fmt.Errorf("pgxConn is nil")
	}

	if de.Symbol == "" {
		logrus.Error("symbol is empty")
		return errInvalidArguments
	}

	if utils.IntInRange(de.Year, 1980, time.Now().Year()) == false {
		logrus.Error("invalid year")
		return errInvalidYear
	}

	if utils.IntInRange(de.Month, 1, 12) == false {
		logrus.Error("invalid month")
		return errInvalidMonth
	}

	insertStatement := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES ('%s','%d','%d','%.2f') ON CONFLICT(symbol, year, month) DO UPDATE SET amount = EXCLUDED.amount;",
		dividendHistoryTable, dividendHistoryFields, de.Symbol, de.Year, de.Month, de.Amount)
	rows, err := de.pgxConn.Query(ctx, insertStatement)
	if err != nil {
		return err
	}

	defer rows.Close()

	return nil
}
