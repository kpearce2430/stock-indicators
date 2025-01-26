package model

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"time"
)

var errUnexpectedNumberOfTransactions = errors.New("unexpected number of transactions found")

const TransactionFields = "id, date, type,  symbol, security, security_payee,  account, description, shares, investment_amount,amount"

// Transaction is an individual transaction read in from the CSV data provided.
type Transaction struct {
	Id               int             `json:"id,omitempty"`
	Date             time.Time       `json:"date,omitempty" db:"Name"`
	Type             TransactionType `json:"type,omitempty"`
	Security         string          `json:"security,omitempty"`
	Symbol           string          `json:"symbol,omitempty"`
	SecurityPayee    string          `json:"security_payee,omitempty"`
	Description      string          `json:"description,omitempty"`
	Shares           float64         `json:"shares,omitempty"`
	InvestmentAmount float64         `json:"investment_amount,omitempty"`
	Amount           float64         `json:"amount,omitempty"`
	Account          string          `json:"account,omitempty"`
}

type TransactionSet struct {
	TransactionRows []*Transaction
	Date            time.Time
}

// NewTransaction creates a new transaction record from headers and a CSV row.
func NewTransaction(headers []string, row []string) (*Transaction, error) {
	tr := Transaction{}

	for i, h := range headers {
		switch h {
		case "Date":
			if row[i] == "" {
				logrus.Error("Invalid Row(", len(row), ") ", row)
				return nil, fmt.Errorf("Invalid date in row %d", i)
			}
			date, err := time.Parse("1/2/2006", row[i])
			if err != nil {
				return nil, fmt.Errorf("NewEntity Date[%s]: %v", row[i], err.Error())
			}
			tr.Date = date
		case "Type":
			tr.Type = TransactionType(row[i])
		case "Security":
			tr.Security = row[i]
		case "Symbol":
			tr.Symbol = row[i]
		case "Security/Payee":
			tr.SecurityPayee = row[i]
		case "Description/Category":
			tr.Description = row[i]
		case "Shares":
			shares, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Shares: %v", err.Error())
			}
			tr.Shares = shares
		case "Invest Amt":
			iAmt, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Invest Amt: %v", err.Error())
			}
			tr.InvestmentAmount = iAmt
		case "Amount":
			amt, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Invest Amt: %v", err.Error())
			}
			tr.Amount = amt
		case "Account":
			tr.Account = row[i]
		default:
			if h != "Split" {
				fmt.Println("Skipping ", h)
			}
		}
	}

	return &tr, nil
}

func NewTransactionSet() *TransactionSet {
	return &TransactionSet{
		Date: time.Now(),
	}
}

func (ts *TransactionSet) NumberOfTransactions() int {
	return len(ts.TransactionRows)
}

func (tr *Transaction) String() string {
	bytes, err := json.Marshal(tr)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

func (ts *TransactionSet) String() string {
	bytes, err := json.Marshal(ts)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

func (tr *Transaction) TransactionToDB(ctx context.Context, pg *pgxpool.Pool, tableName string) error {
	insertStatement := fmt.Sprintf(
		"INSERT INTO %s( id, date, type, security, security_payee, symbol, account, description, shares, investment_amount,amount)"+
			" VALUES('%d','%s','%s','%s','%s','%s','%s','%s','%f','%f','%f');",
		tableName, tr.Id, tr.Date.Format("2006-01-02"), tr.Type, tr.Security, tr.SecurityPayee, tr.Symbol, tr.Account, tr.Description, tr.Shares, tr.InvestmentAmount, tr.Amount)
	rows, err := pg.Query(ctx, insertStatement)
	defer rows.Close()
	if err != nil {
		return err
	}
	return nil
}

func (ts *TransactionSet) LoadWithLookups(lookups *LookUpSet, rawData []byte) error {
	//
	r := csv.NewReader(strings.NewReader(string(rawData)))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1

	foundHeader := false
	var headers []string
	recordNumber := 1

	for count := 0; count < 100000; count++ {
		record, err := r.Read()

		if err == io.EOF {
			fmt.Println("found end of file:", len(ts.TransactionRows))
			return nil
		}

		if err != nil {
			fmt.Println("At ", count, " Error >", err.Error())
			return err
		}

		if !foundHeader {
			if utils.Contains(record, "Date") {
				logrus.Debug("Found Header ", record)
				foundHeader = true
				for _, r := range record[1:] {
					headers = append(headers, r)
				}
			}
			continue
		}

		// Need a better way to do this
		if len(record[1:]) != len(headers) {
			logrus.Debug("Skipping row(", record[1:], ")")
			continue
		}
		tr, err := NewTransaction(headers, record[1:])
		if err != nil {
			fmt.Println("Error:", err.Error())
			return fmt.Errorf("TR Load %s", err.Error())
		}

		value, ok := lookups.GetLookUpByName(tr.Security)
		switch {
		case value == "DEAD":
			continue
		case ok:
			tr.Symbol = value
		}

		tr.Id = recordNumber

		ts.TransactionRows = append(ts.TransactionRows, tr)
		recordNumber++
	}
	return fmt.Errorf("max records read")
}

func (ts *TransactionSet) Load(rawData []byte) error {
	//
	r := csv.NewReader(strings.NewReader(string(rawData)))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1

	foundHeader := false
	var headers []string
	recordNumber := 1

	for count := 0; count < 100000; count++ {
		record, err := r.Read()

		if err == io.EOF {
			fmt.Println("found end of file:", len(ts.TransactionRows))
			return nil
		}

		if err != nil {
			fmt.Println("At ", count, " Error >", err.Error())
			return err
		}

		if !foundHeader {
			if utils.Contains(record, "Date") {
				logrus.Debug("Found Header ", record)
				foundHeader = true
				for _, r := range record[1:] {
					headers = append(headers, r)
				}
			}
			continue
		}

		// Need a better way to do this
		if len(record[1:]) != len(headers) {
			logrus.Debug("Skipping row(", record[1:], ")")
			continue
		}
		en, err := NewTransaction(headers, record[1:])
		if err != nil {
			fmt.Println("Error:", err.Error())
			return fmt.Errorf("TR Load %s", err.Error())
		}
		en.Id = recordNumber
		recordNumber++
		ts.TransactionRows = append(ts.TransactionRows, en)
	}
	return fmt.Errorf("max records read")
}

type TransactionLoadStatus struct {
	ID       int
	Status   bool
	Existing bool
}

func transactionLoadToDB(tChan chan TransactionLoadStatus, pgxConn *pgxpool.Pool, transTable string, tr *Transaction) {

	ctx := context.Background()
	tSet := NewTransactionSet()
	err := tSet.TransactionSetFromDBbyId(ctx, pgxConn, transTable, tr.Id)
	if err != nil {
		logrus.Error("on ", tr.Id, " : ", err.Error())
		tChan <- TransactionLoadStatus{
			ID:       tr.Id,
			Status:   false,
			Existing: false,
		}
		return
	}

	switch len(tSet.TransactionRows) {
	case 0:
		logrus.Debug("No records found, adding")
		if err := tr.TransactionToDB(ctx, pgxConn, transTable); err != nil {
			logrus.Error("on ", tr.Id, " : ", err.Error())
			tChan <- TransactionLoadStatus{
				ID:       tr.Id,
				Status:   false,
				Existing: false,
			}
		}
		logrus.Debug("on ", tr.Id, " : Added")
		tChan <- TransactionLoadStatus{
			ID:       tr.Id,
			Status:   true,
			Existing: false,
		}
		return
	case 1:
		// TODO: Add More checking
		logrus.Debug("Existing Transaction:", tr.Id)
		existing := tSet.TransactionRows[0]
		if existing.Type != tr.Type {
			logrus.Errorf("> %d Transaction Mismatch %s != %s", tr.Id, existing.Type, tr.Type)
			logrus.Errorf(">> %s,%s", existing.Symbol, tr.Symbol)
			logrus.Errorf(">> %s,%s", existing.Description, tr.Description)
			tChan <- TransactionLoadStatus{
				ID:       tr.Id,
				Status:   false,
				Existing: true,
			}
			return
		}
		tChan <- TransactionLoadStatus{
			ID:       tr.Id,
			Status:   true,
			Existing: true,
		}
		return
	default:
		logrus.Error("unexpected number of transactions found:", len(tSet.TransactionRows))
		tChan <- TransactionLoadStatus{
			ID:       tr.Id,
			Status:   false,
			Existing: true,
		}
		return
	}
}

func TransactionSetLoadToDB(pgxConn *pgxpool.Pool, lookups *LookUpSet, transTable string, rawData []byte) error {
	tSet := NewTransactionSet()
	if err := tSet.Load(rawData); err != nil {
		return err
	}

	tChan := make(chan TransactionLoadStatus)
	numProcessed := 0
	newTransactions := 0
	existingTransactions := 0
	errorTransactions := 0

	logrus.Info("Number of rows :", len(tSet.TransactionRows))
	for _, tr := range tSet.TransactionRows {
		if tr.Type == "Payment/Deposit" {
			logrus.Debug("Skipping ", tr.Id, " ", tr.Type)
			continue
		}
		value, ok := lookups.GetLookUpByName(tr.Security)
		switch {
		case value == "DEAD":
			continue
		case ok:
			tr.Symbol = value
		}
		numProcessed++
		go transactionLoadToDB(tChan, pgxConn, transTable, tr)
	}

	var responses []TransactionLoadStatus
	for {
		response, ok := <-tChan
		responses = append(responses, response)

		if !ok {
			panic("Something bad happened")
		}

		switch {
		case response.Status == true && response.Existing == true:
			existingTransactions++
		case response.Status == true && response.Existing == false:
			newTransactions++
		case response.Status == false:
			errorTransactions++
		}

		if len(responses) == numProcessed {
			logrus.Info("Received all expected responses")
			break
		}
	}

	logrus.Info("In Set   : ", len(tSet.TransactionRows))
	logrus.Info("Processed: ", numProcessed)
	logrus.Info("Existing : ", existingTransactions)
	logrus.Info("New      : ", newTransactions)
	if errorTransactions > 0 {
		logrus.Error("Transaction Errors:", errorTransactions)
		return fmt.Errorf("%d errors found in transaction set", errorTransactions)
	}
	return nil
}

func (ts *TransactionSet) TransactionSetFromDBbyId(ctx context.Context, pg *pgxpool.Pool, tableName string, id int) error {
	return ts.getTransactions(ctx, pg, fmt.Sprintf(
		"SELECT %s FROM %s WHERE id = '%d';",
		TransactionFields, tableName, id))
}

func (ts *TransactionSet) TransactionSetFromDBbySymbol(ctx context.Context, pg *pgxpool.Pool, tableName, symbol string) error {
	return ts.getTransactions(ctx, pg, fmt.Sprintf(
		"SELECT %s FROM %s WHERE symbol = '%s' ORDER BY date,id;",
		TransactionFields, tableName, symbol))
}

func (ts *TransactionSet) TransactionsGetAll(ctx context.Context, pg *pgxpool.Pool) error {
	queryStatement := fmt.Sprintf(
		"SELECT %s From %s order by id ", TransactionFields, transactionTable)
	return ts.getTransactions(ctx, pg, queryStatement)
}

func (ts *TransactionSet) TransactionsAllGetBeforeDate(ctx context.Context, pg *pgxpool.Pool, year, month, day int) error {

	return ts.getTransactions(ctx, pg, fmt.Sprintf(
		"SELECT %s From %s WHERE date < '%s' order by date ",
		TransactionFields, transactionTable, fmt.Sprintf("%4d-%02d-%02d", year, month, day)))
}

func (ts *TransactionSet) TransactionsSymbolGetBeforeDate(ctx context.Context, pg *pgxpool.Pool, symbol string, year, month, day int) error {

	return ts.getTransactions(ctx, pg, fmt.Sprintf(
		"SELECT %s From %s WHERE symbol = '%s' and date < '%s' order by date ",
		TransactionFields, transactionTable, symbol, fmt.Sprintf("%4d-%02d-%02d", year, month, day)))
}

func (ts *TransactionSet) TransactionsForMonth(ctx context.Context, pg *pgxpool.Pool, symbol string, year, month int) error {
	var queryStatement string
	switch month {
	case 12:
		endMonth := 01
		endYear := year + 1
		queryStatement = fmt.Sprintf(
			"SELECT %s From %s WHERE symbol = '%s' and date >= '%s' and date < '%s' order by date ",
			TransactionFields, transactionTable, symbol, fmt.Sprintf("%4d-%02d-01", year, month), fmt.Sprintf("%4d-%02d-01", endYear, endMonth))
	default:
		endMonth := month + 1
		queryStatement = fmt.Sprintf(
			"SELECT %s From %s WHERE symbol = '%s' and date >= '%s' and date < '%s' order by date ",
			TransactionFields, transactionTable, symbol, fmt.Sprintf("%4d-%02d-01", year, month), fmt.Sprintf("%4d-%02d-01", year, endMonth))
	}
	return ts.getTransactions(ctx, pg, queryStatement)
}

func (ts *TransactionSet) GetTransactions(ctx context.Context, pg *pgxpool.Pool, symbol string, year, month int) error {
	now := time.Now()
	if symbol == "" && !intInRange(year, 1980, now.Year()) && !intInRange(month, 1, 12) {
		return errInvalidArguments
	}

	if year != 0 && !intInRange(year, 1980, now.Year()) {
		return errInvalidYear
	}

	if !intInRange(month, 0, 12) {
		return errInvalidMonth
	}

	needAnd := false
	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(TransactionFields)
	sb.WriteString(" FROM ")
	sb.WriteString(transactionTable)
	sb.WriteString(" WHERE ")
	if symbol != "" {
		sb.WriteString(" symbol = '")
		sb.WriteString(symbol)
		needAnd = true
		sb.WriteString("'")
	}

	if intInRange(year, 1980, now.Year()) && intInRange(month, 1, 12) {
		if needAnd {
			sb.WriteString(" AND ")
		}
		switch month {
		case 12:
			endMonth := 01
			endYear := year + 1
			sb.WriteString(" date >= '")
			sb.WriteString(fmt.Sprintf("%4d-%02d-01' and date < '", year, month))
			sb.WriteString(fmt.Sprintf("%4d-%02d-01'", endYear, endMonth))
			// sb.WriteString(fmt.Sprintf(" date >= '%s' and date < '%s'",))

		default:
			endMonth := month + 1
			sb.WriteString(" date >= '")
			sb.WriteString(fmt.Sprintf("%4d-%02d-01' and date < '", year, month))
			sb.WriteString(fmt.Sprintf("%4d-%02d-01'", year, endMonth))

		}
	}
	sb.WriteString(" order by date ")
	return ts.getTransactions(ctx, pg, sb.String())
}

// GetTransactions will return the TransactionSet based on the selectStatement passed in.
func (ts *TransactionSet) getTransactions(ctx context.Context, pg *pgxpool.Pool, selectStatement string) error {

	if len(ts.TransactionRows) > 0 {
		clear(ts.TransactionRows)
	}

	rows, err := pg.Query(ctx, selectStatement)
	defer rows.Close()
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	// Iterate through the result set
	for rows.Next() {
		trans := Transaction{}
		err = rows.Scan(&trans.Id, &trans.Date, &trans.Type, &trans.Symbol, &trans.Security, &trans.SecurityPayee, &trans.Account, &trans.Description, &trans.Shares, &trans.InvestmentAmount, &trans.Amount)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		ts.TransactionRows = append(ts.TransactionRows, &trans)
	}
	return nil
}
