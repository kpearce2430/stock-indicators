package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"time"
)

/*
  {
    "cash_amount": 0.12,
    "declaration_date": "2024-05-08",
    "dividend_type": "CD",
    "ex_dividend_date": "2024-05-31",
    "frequency": 4,
    "pay_date": "2024-06-14",
    "record_date": "2024-05-31",
    "ticker": "CSX"
  },
*/

const dividendsTableFields = "ticker, cash_amount, declaration_date, dividend_type, ex_dividend_date, frequency, pay_date, record_date"

type Dividends struct {
	Ticker          string    `json:"ticker"`
	CashAmount      float64   `json:"cash_amount"`
	DeclarationDate StockTime `json:"declaration_date"`
	DividendType    string    `json:"dividend_type"`
	ExDividendDate  StockTime `json:"ex_dividend_date"`
	Frequency       int       `json:"frequency"`
	PayDate         StockTime `json:"pay_date"`
	RecordDate      StockTime `json:"record_date"`
}

type DividendsSet struct {
	Dividends []Dividends `json:"dividends"`
}

func NewDividendsSet(dividends []Dividends) DividendsSet {
	return DividendsSet{Dividends: dividends}
}

func NewDividendsSetFromJSON(jsonBytes []byte) (DividendsSet, error) {
	ds := DividendsSet{}
	err := json.Unmarshal(jsonBytes, &ds)
	if err != nil {
		return DividendsSet{}, err
	}
	return ds, nil
}

func (d *Dividends) String() string {
	bytes, err := json.Marshal(d)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

func (ds *DividendsSet) String() string {
	bytes, err := json.Marshal(ds)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

func (ds *DividendsSet) ToDB(ctx context.Context, pg *pgxpool.Pool, tableName string) error {

	for _, div := range ds.Dividends {
		err := div.ToDB(ctx, pg, tableName)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
	}
	return nil
}

func (d *Dividends) ToDB(ctx context.Context, pg *pgxpool.Pool, tableName string) error {
	insertStatement := fmt.Sprintf(
		"INSERT INTO %s(ticker, cash_amount, declaration_date, dividend_type, ex_dividend_date, frequency, pay_date, record_date)"+
			" VALUES('%s','%f','%s','%s','%s','%d','%s','%s');",
		tableName,
		d.Ticker,
		d.CashAmount,
		d.DeclarationDate.Format(dateToPgLayout),
		d.DividendType,
		d.ExDividendDate.Format(dateToPgLayout),
		d.Frequency,
		d.PayDate.Format(dateToPgLayout),
		d.RecordDate.Format(dateToPgLayout))
	rows, err := pg.Query(ctx, insertStatement)
	defer rows.Close()
	if err != nil {
		return err
	}
	return nil
}

func (ds *DividendsSet) FromDBbySymbol(ctx context.Context, pg *pgxpool.Pool, tableName, symbol string) error {
	return ds.getDividends(ctx, pg, fmt.Sprintf(
		"SELECT %s FROM %s WHERE ticker = '%s' ORDER BY declaration_date DESC;",
		dividendsTableFields, tableName, symbol))
}

func (ds *DividendsSet) getDividends(ctx context.Context, pg *pgxpool.Pool, selectStatement string) error {

	if len(ds.Dividends) > 0 {
		clear(ds.Dividends)
	}

	rows, err := pg.Query(ctx, selectStatement)
	defer rows.Close()
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	// Iterate through the result set
	for rows.Next() {
		var declarationDate, exDividendDate, payDate, recordDate time.Time
		d := Dividends{}
		err = rows.Scan(
			&d.Ticker,
			&d.CashAmount,
			&declarationDate,
			&d.DividendType,
			&exDividendDate,
			&d.Frequency,
			&payDate,
			&recordDate)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}

		d.DeclarationDate.TimeConv(declarationDate)
		d.ExDividendDate.TimeConv(exDividendDate)
		d.PayDate.TimeConv(payDate)
		d.RecordDate.TimeConv(recordDate)

		ds.Dividends = append(ds.Dividends, d)
	}
	return nil
}
