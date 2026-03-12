package portfolio_value

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/sirupsen/logrus"
)

const (
	PortfolioValueTable = "portfolio_value"
	pvTypeStock         = "Stock"
	pvTypeBond          = "Bond"
	pvTypeMutualFund    = "Mutual Fund"
	pvTypeOther         = "Other"
	pvTableFields       = "date, name, symbol, type, quote, pricedaychange, pricedaychangepct, shares, costbasis, marketvalue, averagecostpershare, gainloss12month, gainloss, gaillosspct"
)

var errPortfolioTypeUnknown = errors.New("unknown portfolio type")

type PortfolioValueRecord struct {
	Date                time.Time `json:"date,omitempty"`
	Name                string    `json:"name"`
	Symbol              string    `json:"symbol"`
	Type                string    `json:"type"`
	Quote               float64   `json:"quote"`
	PriceDayChange      float64   `json:"price_day_change"`
	PriceDayChangePct   float64   `json:"price_day_change_pct"`
	Shares              float64   `json:"shares"`
	CostBasis           float64   `json:"cost_basis"`
	MarketValue         float64   `json:"market_value"`
	AverageCostPerShare float64   `json:"avg_cost_per_share"`
	GainLoss12Month     float64   `json:"gain_loss_last_12m"`
	GainLoss            float64   `json:"gain_loss"`
	GainLossPct         float64   `json:"gain_loss_pct"`
}

type PortfolioValueSet struct {
	pgxConn   *pgxpool.Pool
	tableName string
	symbol    string
	Set       []PortfolioValueRecord `json:"portfolio_value"`
}

type PortfolioValueDatabaseRecord struct {
	Id         string                `json:"_id"`
	Rev        string                `json:"_rev,omitempty"`
	PV         *PortfolioValueRecord `json:"portfolio_value,omitempty"`
	Key        string                `json:"key"`
	Symbol     string                `json:"symbol"`
	Julian     string                `json:"julian"`
	IEXHistory string                `json:"iex_history,omitempty"`
}

// NewPortfolioValue create a new PortfolioValueRecord from the headers and row values provided.
func NewPortfolioValue(headers []string, values []string) (*PortfolioValueRecord, error) {
	pv := PortfolioValueRecord{}
	for index, value := range headers {
		switch value {
		case "Name":
			pv.Name = values[index]
		case "Symbol":
			pv.Symbol = values[index]
		case "Type":
			pv.Type = values[index]
		case "Price":
			pv.Quote, _ = utils.FloatParse(values[index])
		case "Quote":
			pv.Quote, _ = utils.FloatParse(values[index])
		case "Price Day Change":
			pv.PriceDayChange, _ = utils.FloatParse(values[index])
		case "Price Day Change (%)":
			pv.PriceDayChangePct, _ = utils.FloatParse(values[index])
		case "Shares":
			pv.Shares, _ = utils.FloatParse(values[index])
		case "Cost Basis":
			pv.CostBasis, _ = utils.FloatParse(values[index])
		case "Market Value":
			pv.MarketValue, _ = utils.FloatParse(values[index])
		case "Average Cost Per Share":
			pv.AverageCostPerShare, _ = utils.FloatParse(values[index])
		case "Gain/Loss 12-Month":
			pv.GainLoss12Month, _ = utils.FloatParse(values[index])
		case "Gain/Loss":
			pv.GainLoss, _ = utils.FloatParse(values[index])
		case "Gain/Loss (%)":
			pv.GainLossPct, _ = utils.FloatParse(values[index])
		} // switch
	} // for

	switch pv.Type {
	case pvTypeBond, pvTypeStock, pvTypeMutualFund:
	case pvTypeOther:
		pv.Type = pvTypeStock
	default:
		logrus.Error(errPortfolioTypeUnknown, ":", pv.Type)
		return &pv, errPortfolioTypeUnknown
	}
	return &pv, nil
}

func LoadDB(pgxConn *pgxpool.Pool, databaseName, rawData, julDate string, lookups *lookups.LookUpSet) (int, error) {
	r := csv.NewReader(strings.NewReader(rawData))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1

	foundHeader := false
	numRows := 2
	count := 0
	var headers []string
	var date time.Time
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
			return -1, err
		}

		if foundHeader == false {

			if julDate == "" && strings.HasPrefix(record[0], "Price and Holdings as of") {

				// if julDate == "" {
				logrus.Debug("found:", record[0])
				parts := strings.Split(record[0], ":")
				str := strings.TrimSpace(parts[1])
				logrus.Debugf("str[%s]", str)

				date, err = time.Parse("2006-01-02", str)
				if err != nil {
					logrus.Error(err)
					return -1, err
				}
				logrus.Info("Date:", date)
				continue
			}

			if len(record) > numRows && utils.Contains(record, "Symbol") {
				record[0] = "Name"
				numRows = len(record)
				foundHeader = true
				headers = record

				if julDate == "" {
					date := time.Now()
					julDate = date.Format("2006002")
				}
			}

		} else {
			record[0] = utils.AsciiString(record[0])
			if strings.Compare(record[0], "Cash") == 0 || strings.Compare(record[0], "Totals") == 0 {
				continue
			}

			if len(record) == numRows {
				pvRec, err := NewPortfolioValue(headers, record)
				if err != nil {
					logrus.Error(err.Error())
					return -1, err
				}

				if pvRec.Symbol == "" {
					Symbol, ok := lookups.GetLookUpByName(pvRec.Name)
					if ok {
						switch Symbol {
						case "DEAD":
							logrus.Debug("Found Dead")
							continue
						case "":
							logrus.Debug("Error Missing Symbol for \"", pvRec, "\"")
							continue
						default:
							pvRec.Symbol = Symbol
						}
					}
				}
				if err = pvRec.LoadDB(pgxConn, date, databaseName); err != nil {
					logrus.Error(err.Error())
					return -1, err
				}
				count++
			}
		}
	}

	if !foundHeader {
		return -1, fmt.Errorf("no portfolio value header found")
	}
	logrus.Info("Loaded ", count, " records")
	return count, nil
}

func GetTypes(pgxConn *pgxpool.Pool, portfolioValueTable string) (map[string]string, error) {
	types := make(map[string]string)
	sql2 := fmt.Sprintf("SELECT DISTINCT symbol, type FROM %s ORDER BY symbol;",
		portfolioValueTable)

	rows, err := pgxConn.Query(context.Background(), sql2)
	if err != nil {
		logrus.Error(err.Error())
		return types, err
	}

	// Iterate through the result set
	i := 1
	for rows.Next() {
		var (
			symbol string
			sType  string
		)
		err = rows.Scan(&symbol, &sType)
		if err != nil {
			rows.Close()
			logrus.Error(err.Error())
			return types, err
		}
		types[symbol] = sType
		i++
	}
	rows.Close()
	return types, nil
}

func (p *PortfolioValueRecord) LoadDB(pgxConn *pgxpool.Pool, date time.Time, tableName string) error {

	if pgxConn == nil {
		return fmt.Errorf("pgxConn is nil")
	}
	insertStatement := fmt.Sprintf(
		"INSERT INTO %s(%s)"+
			" VALUES('%s','%s','%s','%s','%.04f','%.04f','%.04f','%.04f','%.04f','%.04f','%.04f','%.04f','%.04f','%.04f') "+
			" ON CONFLICT DO NOTHING;",
		tableName, pvTableFields,
		date.Format("2006-01-02"), p.Name, p.Symbol, p.Type,
		p.Quote, p.PriceDayChange, p.PriceDayChangePct, p.Shares,
		p.CostBasis, p.MarketValue, p.AverageCostPerShare, p.GainLoss12Month,
		p.GainLoss, p.GainLossPct)

	rows, err := pgxConn.Query(context.Background(), insertStatement)
	defer rows.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *PortfolioValueRecord) GetDB(pgxConn *pgxpool.Pool, symbol, tableName string, date time.Time) error {
	selectStatement := fmt.Sprintf(
		"SELECT %s From %s WHERE symbol = '%s' and date = '%s' ",
		pvTableFields, tableName, symbol, fmt.Sprintf("%4d-%02d-%02d", date.Year(), date.Month(), date.Day()))

	return p.getRecord(pgxConn, selectStatement)
}

func (p *PortfolioValueRecord) GetLastDB(pgxConn *pgxpool.Pool, symbol, tableName string) error {
	selectStatement := fmt.Sprintf(
		"SELECT %s From %s WHERE symbol = '%s' order by date desc limit 1 ",
		pvTableFields, tableName, symbol)

	return p.getRecord(pgxConn, selectStatement)
}

func (p *PortfolioValueRecord) getRecord(pgxConn *pgxpool.Pool, selectStatement string) error {
	rows, err := pgxConn.Query(context.Background(), selectStatement)
	var date time.Time
	// Iterate through the result set
	i := 0
	for rows.Next() {
		err = rows.Scan(
			&date, &p.Name, &p.Symbol, &p.Type,
			&p.Quote, &p.PriceDayChange, &p.PriceDayChangePct, &p.Shares,
			&p.CostBasis, &p.MarketValue, &p.AverageCostPerShare, &p.GainLoss12Month,
			&p.GainLoss, &p.GainLossPct)

		if err != nil {
			rows.Close()
			logrus.Error(err.Error())
			return err
		}
		i++
	}
	rows.Close()

	if i > 1 {
		return fmt.Errorf("more than 1 row returned")
	}
	return nil
}

func GetSymbolType(pgxConn *pgxpool.Pool, portfolioValueTable, symbol string) (string, error) {
	var t string
	sql := fmt.Sprintf("SELECT type FROM %s WHERE symbol = '%s' LIMIT 1;",
		portfolioValueTable, symbol)
	if err := pgxConn.QueryRow(context.Background(), sql).Scan(&t); err != nil {
		logrus.Error(err.Error())
		return t, err
	}
	logrus.Infof("%s : %s", symbol, t)
	return t, nil
}

func NewSet(pgxConn *pgxpool.Pool, table, symbol string) *PortfolioValueSet {
	if pgxConn == nil {
		return nil
	}
	return &PortfolioValueSet{
		tableName: table,
		symbol:    symbol,
		pgxConn:   pgxConn,
	}
}

// LastInSet returns the last record in the PortfolioValueSet.
func (ps *PortfolioValueSet) LastInSet() (*PortfolioValueRecord, error) {
	l := len(ps.Set) - 1
	if l < 0 {
		return nil, errors.New("no records in set")
	}
	return &ps.Set[l], nil
}

func (ps *PortfolioValueSet) GetSymbolYearMonth(year, month int) error {
	// select * from portfolio_value where date < '2026-04-01 00:00:00' and symbol = 'AAPL' order by date desc limit 10;
	var queryStatement string
	symbolQuery := ""
	if ps.symbol != "" {
		symbolQuery = fmt.Sprintf(" symbol = '%s' and", ps.symbol)
	}

	switch month {
	case 12:
		endMonth := 01
		endYear := year + 1
		queryStatement = fmt.Sprintf(
			"SELECT %s From %s WHERE %s date >= '%s' and date < '%s' order by date ",
			pvTableFields, ps.tableName, symbolQuery, fmt.Sprintf("%4d-%02d-01", year, month), fmt.Sprintf("%4d-%02d-01", endYear, endMonth))
	default:
		endMonth := month + 1
		queryStatement = fmt.Sprintf(
			"SELECT %s From %s WHERE %s date >= '%s' and date < '%s' order by date ",
			pvTableFields, ps.tableName, symbolQuery, fmt.Sprintf("%4d-%02d-01", year, month), fmt.Sprintf("%4d-%02d-01", year, endMonth))
	}
	return ps.getRecords(queryStatement)
}

func (ps *PortfolioValueSet) getRecords(selectStatement string) error {
	rows, err := ps.pgxConn.Query(context.Background(), selectStatement)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	defer rows.Close()

	// Iterate through the result set
	for rows.Next() {
		p := PortfolioValueRecord{}
		err = rows.Scan(
			&p.Date, &p.Name, &p.Symbol, &p.Type,
			&p.Quote, &p.PriceDayChange, &p.PriceDayChangePct, &p.Shares,
			&p.CostBasis, &p.MarketValue, &p.AverageCostPerShare, &p.GainLoss12Month,
			&p.GainLoss, &p.GainLossPct)

		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		ps.Set = append(ps.Set, p)
	}

	return nil
}
