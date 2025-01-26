package model

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"strings"
	"time"
)

type LookUpSet struct {
	Id        string            `json:"_id"`
	Rev       string            `json:"_rev,omitempty"`
	Timestamp string            `json:"timestamp"`
	LookUps   map[string]string `json:"look_ups"`
}

func NewLookupSet(id string) *LookUpSet {
	dt := time.Now()
	curTimeStr := dt.Format("2006-01-02 15:04:05")
	lookupSet := LookUpSet{Id: id, Timestamp: curTimeStr}
	lookupSet.LookUps = make(map[string]string)
	return &lookupSet
}

func LoadLookupSet(id string, csvData string) *LookUpSet {
	r := csv.NewReader(strings.NewReader(csvData))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1
	lookupSet := NewLookupSet(id)

	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if len(record) < 2 {
			break
		}
		lookupSet.LookUps[record[0]] = record[1]
	}
	return lookupSet
}

func (l *LookUpSet) GetLookUpByName(name string) (string, bool) {
	if l.LookUps == nil {
		panic("Missing lookups")
	}
	val, ok := l.LookUps[name]
	return val, ok
}

func (l *LookUpSet) GetLookUpBySymbol(symbol string) (string, bool) {
	if l.LookUps == nil {
		panic("Missing lookups")
	}
	for k, v := range l.LookUps {
		if v == symbol {
			return k, true
		}
	}
	return "", false
}

func LoadLookupFromCSV(ctx context.Context, pgxConn *pgxpool.Pool, table string, rawData []byte) error {
	r := csv.NewReader(strings.NewReader(string(rawData)))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1

	count := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			fmt.Println("found end of file. Count:", count)
			break
		}

		if len(record) < 2 {
			continue
		}

		if err != nil {
			fmt.Println("At ", count, " Error >", err.Error())
			return err
		}

		if err = LoadLookupToDB(ctx, pgxConn, table, record[0], record[1]); err != nil {
			fmt.Println("At ", count, " Error >", err.Error())
			return err
		}
		count++
	}
	num, err := countLookups(ctx, pgxConn, table)
	if err != nil {
		return err
	}
	logrus.Infof("Count: %v lookups found", num)
	return nil
}

func LoadLookupToDB(ctx context.Context, pgxConn *pgxpool.Pool, table, security, symbol string) error {
	insertStatement := fmt.Sprintf(
		"INSERT INTO %s ( security, symbol)"+" VALUES('%s','%s');",
		table, security, symbol)
	rows, err := pgxConn.Query(ctx, insertStatement)
	defer rows.Close()
	if err != nil {
		return err
	}
	return nil
}

func countLookups(ctx context.Context, pgxConn *pgxpool.Pool, table string) (int, error) {
	count := 0

	countSql := "SELECT COUNT(*) FROM " + table
	if err := pgxConn.QueryRow(ctx, countSql).Scan(&count); err != nil {
		return -1, err
	}
	logrus.Debug("Count:", count)
	return count, nil
}

func GetLookUpsFromDB(ctx context.Context, pgxConn *pgxpool.Pool, table string) (*LookUpSet, error) {
	l := NewLookupSet(table)

	selectStatement := fmt.Sprintf("SELECT security,symbol  FROM %s", table)

	rows, err := pgxConn.Query(ctx, selectStatement)
	defer rows.Close()
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	// Iterate through the result set
	for rows.Next() {
		var security, symbol string

		err = rows.Scan(&security, &symbol)
		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}
		l.LookUps[security] = symbol
	}
	logrus.Infof("Loaded %v lookups from DB", len(l.LookUps))
	return l, nil
}
