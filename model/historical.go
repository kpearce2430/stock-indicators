package model

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	cdb "github.com/kpearce2430/keputils/couch-database"
	"github.com/sirupsen/logrus"
	"io"
	"strconv"
	"strings"
	"time"
)

const historicDBFields = "source,symbol,date, open,high, low, close, adj_close,volume"

var (
	errPGXConnectionNil = fmt.Errorf("pgx connection pool is null")
)

type HistoricalDataSet struct {
	pgxConn      *pgxpool.Pool
	historyTable string
}

func NewHistoricalDataSet(p *pgxpool.Pool, table string) *HistoricalDataSet {
	return &HistoricalDataSet{
		pgxConn:      p,
		historyTable: table,
	}
}

type Historical struct {
	Symbol   string    `json:"symbol"`
	Date     time.Time `json:"date"`
	Open     float64   `json:"open"`
	High     float64   `json:"high"`
	Low      float64   `json:"low"`
	Close    float64   `json:"close"`
	AdjClose float64   `json:"adjClose"`
	Volume   float64   `json:"volume"`
	Source   string    `json:"source"`
}

func (h *Historical) String() string {
	b, err := json.Marshal(h)
	if err != nil {
		logrus.Error(err.Error())
		return err.Error()
	}
	return string(b)
}

type HistoricalDatabaseRecord struct {
	Id         string      `json:"_id"`
	Rev        string      `json:"_rev,omitempty"`
	Historical *Historical `json:"historical,omitempty"`
	Symbol     string      `json:"symbol"`
}

func (h *HistoricalDatabaseRecord) String() string {
	b, err := json.Marshal(h)
	if err != nil {
		logrus.Error(err.Error())
		return err.Error()
	}
	return string(b)
}

const (
	HistoricalDate     = "Date"
	HistoricalOpen     = "Open"
	HistoricalHigh     = "High"
	HistoricalLow      = "Low"
	HistoricalClose    = "Close"
	HistoricalAdjClose = "Adj Close"
	HistoricalVolume   = "Volume"
)

func NewHistorical(symbol, source string, headers, record []string) (*Historical, error) {
	hist := Historical{
		Symbol: symbol,
		Source: source,
	}

	var err error
	for i, h := range headers {
		switch h {
		case HistoricalDate:
			// hist.Date = record[i]
			hist.Date, err = time.Parse("2006-01-02", record[i])
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
		case HistoricalOpen:
			hist.Open, err = strconv.ParseFloat(record[i], 64)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
		case HistoricalHigh:
			hist.High, err = strconv.ParseFloat(record[i], 64)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
		case HistoricalLow:
			hist.Low, err = strconv.ParseFloat(record[i], 64)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
		case HistoricalClose:
			hist.Close, err = strconv.ParseFloat(record[i], 64)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
		case HistoricalAdjClose:
			hist.AdjClose, err = strconv.ParseFloat(record[i], 64)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
		case HistoricalVolume:
			hist.AdjClose, err = strconv.ParseFloat(record[i], 64)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
		}
	}
	return &hist, nil
}

// HistoricalSetLoadCouch Load a set of historical records into CouchDB.  This will create the database if it does not exist.
func HistoricalSetLoadCouch(databaseName, rawData, source, symbol string) error {
	r := csv.NewReader(strings.NewReader(rawData))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1
	foundHeader := false
	numRows := 0
	var headers []string

	historicalDB, err := cdb.GetDataStoreByDatabaseName[HistoricalDatabaseRecord](databaseName)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	_, err = historicalDB.DatabaseExists()
	if err != nil {
		if historicalDB.DatabaseCreate() == false {
			logrus.Error(err.Error())
			return err
		}
		logrus.Debug("Database Created")
	}

	for {
		record, err := r.Read()
		numRows++
		if err == io.EOF {
			break
		}
		if err != nil {
			logrus.Error(err)
			return err
		}

		// Skip the first row
		if foundHeader == false {
			for _, r := range record {
				headers = append(headers, r)
			}

			logrus.Info("number headers:", len(headers))
			foundHeader = true
			continue
		}

		hist, err := NewHistorical(symbol, source, headers, record)
		if err != nil {
			logrus.Error(err)
			return err
		}
		var julDate string
		julDate = hist.Date.Format("2006002")
		key := fmt.Sprintf("%s:%s", julDate, symbol)

		histRecord := HistoricalDatabaseRecord{
			Id:         key,
			Historical: hist,
			Symbol:     symbol,
		}
		rev, err := historicalDB.DocumentCreate(key, &histRecord)
		if err != nil {
			logrus.Error(err)
			return err
		}
		logrus.Info("inserted ", histRecord.String(), " rev:", rev)
	}
	return nil
}

func HistoricalCacheGet(databaseName, key string) (*HistoricalDatabaseRecord, error) {
	historicalDB, err := cdb.GetDataStoreByDatabaseName[HistoricalDatabaseRecord](databaseName)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	return historicalDB.DocumentGet(key)
}

func HistoricalCacheDelete(databaseName, key string) (string, error) {
	historicalDB, err := cdb.GetDataStoreByDatabaseName[HistoricalDatabaseRecord](databaseName)
	if err != nil {
		logrus.Error(err.Error())
		return "", err
	}

	record, err := historicalDB.DocumentGet(key)
	if err != nil {
		logrus.Error(err.Error())
		return "", err
	}

	if record != nil {
		return historicalDB.DocumentDelete(key, record.Rev)
	}

	return "", fmt.Errorf("%s not found", key)
}

// LoadSet loads raw data from source for symbol.
func (h *HistoricalDataSet) LoadSet(rawData, source, symbol string) error {
	r := csv.NewReader(strings.NewReader(rawData))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1
	foundHeader := false
	numRows := 0
	var headers []string

	for {
		record, err := r.Read()
		numRows++
		if err == io.EOF {
			break
		}
		if err != nil {
			logrus.Error(err)
			return err
		}

		// Skip the first row
		if foundHeader == false {
			for _, r := range record {
				headers = append(headers, r)
			}

			logrus.Info("number headers:", len(headers))
			foundHeader = true
			continue
		}

		hist, err := NewHistorical(symbol, source, headers, record)
		if err != nil {
			logrus.Error(err)
			return err
		}

		if err := h.LoadDB(hist); err != nil {
			logrus.Error(err)
			return err
		}
	}
	logrus.Info("Loaded ", numRows, " Rows")

	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s;", h.historyTable)
	var count int
	if err := h.pgxConn.QueryRow(context.Background(), countSql).Scan(&count); err != nil {
		return err
	}
	logrus.Info("Found ", count, " Rows")
	return nil
}

// LoadDB Load the historical record to the postgres table.  The function will not create the table.
func (h *HistoricalDataSet) LoadDB(hist *Historical) error {
	/* Expected structore
	   symbol varchar(10),
	   source varcar(50),
	   date TIMESTAMP,
	   open NUMERIC,
	   high NUMERIC,
	   Low NUMERIC,
	   close NUMERIC,
	   adj_close NUMERIC,
	   volume NUMERIC,
	*/
	if h.pgxConn == nil {
		return errPGXConnectionNil
	}
	insertStatement := fmt.Sprintf(
		"INSERT INTO %s(%s)"+
			" VALUES('%s','%s','%s','%.04f','%.04f','%.04f','%.04f','%.04f','%.04f') "+
			" ON CONFLICT DO NOTHING;",
		h.historyTable, historicDBFields, hist.Source, hist.Symbol, hist.Date.Format("2006-01-02"),
		hist.Open, hist.High, hist.Low, hist.Close, hist.AdjClose, hist.Volume)
	rows, err := h.pgxConn.Query(context.Background(), insertStatement)
	defer rows.Close()
	if err != nil {
		return err
	}
	return nil
}

func (h *HistoricalDataSet) Last(symbol string, date time.Time) (*Historical, error) {
	if h.pgxConn == nil {
		return nil, errPGXConnectionNil
	}
	queryStatement := fmt.Sprintf(
		"SELECT %s FROM %s WHERE symbol='%s' AND date < '%s' ORDER BY DATE DESC LIMIT 1;",
		historicDBFields, h.historyTable, symbol, fmt.Sprintf("%4d-%02d-01", date.Year(), date.Month()))

	rows, err := h.pgxConn.Query(context.Background(), queryStatement)
	defer rows.Close()
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	hist := Historical{}
	// Iterate through the result set
	i := 0
	for rows.Next() {
		err = rows.Scan(&hist.Source, &hist.Symbol, &hist.Date, &hist.Open, &hist.High, &hist.Low, &hist.Close, &hist.AdjClose, &hist.Volume)
		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}
		i++
	}

	switch i {
	case 0:
		return nil, fmt.Errorf("no records found")
	case 1:
		return &hist, nil
	}
	return &hist, fmt.Errorf("retrived %d records", i)
}
