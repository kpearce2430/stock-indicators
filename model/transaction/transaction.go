package transaction

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/sirupsen/logrus"
)

const (
	TransactionID               = "ID"
	TransactionDate             = "Date"
	TransactionType             = "Type"
	TransactionSecurity         = "Security"
	TransactionSecurityPayee    = "Security/Payee"
	TransactionSymbol           = "Symbol"
	TransactionAccount          = "Account"
	TransactionDescription      = "Description/Category"
	TransactionShares           = "Shares"
	TransactionInvestmentAmount = "Invest Amt"
	TransactionAmount           = "Amount"
	TransactionYear             = "Year"
	TransactionMonth            = "Month"
	TransactionTable            = "transactions"
	TransactionFields           = "id, date, type,  symbol, security, security_payee,  account, description, shares, investment_amount,amount"
)

var (
	errInvalidArguments = errors.New("invalid arguments")
	errInvalidYear      = errors.New("invalid year")
	errInvalidMonth     = errors.New("invalid month")
)

type TransactionsType string

// Transaction is an individual transaction read in from the CSV data provided.
type Transaction struct {
	Id               int              `json:"id,omitempty"`
	Date             time.Time        `json:"date,omitempty" db:"Name"`
	Type             TransactionsType `json:"type,omitempty"`
	Security         string           `json:"security,omitempty"`
	Symbol           string           `json:"symbol,omitempty"`
	SecurityPayee    string           `json:"security_payee,omitempty"`
	Description      string           `json:"description,omitempty"`
	Shares           float64          `json:"shares,omitempty"`
	InvestmentAmount float64          `json:"investment_amount,omitempty"`
	Amount           float64          `json:"amount,omitempty"`
	Account          string           `json:"account,omitempty"`
}

type TransactionSet struct {
	TransactionRows []*Transaction
	Date            time.Time
}

type TransactionLoadStatus struct {
	ID       int
	Status   bool
	Existing bool
}

func NewTransactionSet() *TransactionSet {
	return &TransactionSet{
		TransactionRows: make([]*Transaction, 0),
		Date:            time.Now(),
	}
}

func (ts *TransactionSet) Count() int {
	return len(ts.TransactionRows)
}

func (ts *TransactionSet) Load(rawData []byte) error {
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

func (ts *TransactionSet) LoadToDB(pgxConn *pgxpool.Pool, lookups *lookups.LookUpSet, transTable string, rawData []byte) error {
	if err := ts.Load(rawData); err != nil {
		return err
	}

	tChan := make(chan TransactionLoadStatus)
	numProcessed := 0
	newTransactions := 0
	existingTransactions := 0
	errorTransactions := 0

	logrus.Info("Number of rows :", len(ts.TransactionRows))
	for _, tr := range ts.TransactionRows {
		if tr.Type == "Payment/Deposit" {
			logrus.Debug("Skipping Payment ", tr.Id, " ", tr.Type)
			continue
		}

		if tr.Security == "" && tr.Symbol == "" {
			logrus.Debug("Skipping Blank Security and Symbol ", tr)
			continue
		}
		value, ok := lookups.GetLookUpByName(tr.Security)
		switch {
		case value == "DEAD":
			logrus.Debug("Skipping ", tr.Security, " DEAD ", tr.Id)
			continue
		case ok:
			tr.Symbol = value
		}
		if tr.Security == "" {
			logrus.Warning("Transaction [", tr, "] missing Security")
		}
		today := time.Now()
		if tr.Date.After(today) {
			logrus.Info("Skipping ", tr.Type, " - Future Transaction:", tr.Id, ":", tr.Date)
			continue
		}
		numProcessed++
		go transactionLoadToDB(tChan, pgxConn, transTable, tr)
	}

	if numProcessed == 0 {
		logrus.Warning("No transactions to process")
		return errors.New("no transactions to process")
	}

	logrus.Debug("Waiting for all transactions to be processed:", numProcessed)

	var responses []TransactionLoadStatus
	for {
		response, ok := <-tChan
		responses = append(responses, response)

		if !ok {
			logrus.Error("Something bad happened")
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
			logrus.Debug("Received all expected responses")
			break
		}
	}

	logrus.Info("In Set   : ", len(ts.TransactionRows))
	logrus.Info("Processed: ", numProcessed)
	logrus.Info("Existing : ", existingTransactions)
	logrus.Info("New      : ", newTransactions)
	if errorTransactions > 0 {
		logrus.Error("Transaction Errors:", errorTransactions)
		return fmt.Errorf("%d errors found in transaction set", errorTransactions)
	}
	return nil
}

func (ts *TransactionSet) FromDBbyId(ctx context.Context, pg *pgxpool.Pool, tableName string, id int) error {
	return ts.getTransactions(ctx, pg, fmt.Sprintf(
		"SELECT %s FROM %s WHERE id = '%d';",
		TransactionFields, tableName, id))
}

func (ts *TransactionSet) GetAll(ctx context.Context, pg *pgxpool.Pool) error {
	queryStatement := fmt.Sprintf(
		"SELECT %s From %s order by id ", TransactionFields, TransactionTable)
	return ts.getTransactions(ctx, pg, queryStatement)
}

func (ts *TransactionSet) GetTransactions(ctx context.Context, pg *pgxpool.Pool, symbol string, year, month int) error {
	now := time.Now()
	if symbol == "" && !utils.IntInRange(year, 1980, now.Year()) && !utils.IntInRange(month, 1, 12) {
		return errInvalidArguments
	}

	if year != 0 && !utils.IntInRange(year, 1980, now.Year()) {
		return errInvalidYear
	}

	if !utils.IntInRange(month, 0, 12) {
		return errInvalidMonth
	}

	needAnd := false
	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(TransactionFields)
	sb.WriteString(" FROM ")
	sb.WriteString(TransactionTable)
	sb.WriteString(" WHERE ")
	if symbol != "" {
		sb.WriteString(" symbol = '")
		sb.WriteString(symbol)
		needAnd = true
		sb.WriteString("'")
	}

	if utils.IntInRange(year, 1980, now.Year()) && utils.IntInRange(month, 1, 12) {
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

func (ts *TransactionSet) FromDBbySymbol(ctx context.Context, pg *pgxpool.Pool, tableName, symbol string) error {
	// note that type puts add types ahead of remove share.
	return ts.getTransactions(ctx, pg, fmt.Sprintf(
		"SELECT %s FROM %s WHERE symbol = '%s' ORDER BY date,id;",
		TransactionFields, tableName, symbol))
}

// GetTransactions will return the TransactionSet based on the selectStatement passed in.
func (ts *TransactionSet) getTransactions(ctx context.Context, pg *pgxpool.Pool, selectStatement string) error {
	if len(ts.TransactionRows) > 0 {
		clear(ts.TransactionRows)
	}

	logrus.Debug("Querying ", selectStatement)

	rows, err := pg.Query(ctx, selectStatement)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	defer rows.Close()

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

func (ts *TransactionSet) SymbolGetBeforeDate(ctx context.Context, pg *pgxpool.Pool, symbol string, year, month, day int) error {
	return ts.getTransactions(ctx, pg, fmt.Sprintf(
		"SELECT %s From %s WHERE symbol = '%s' and date < '%s' order by date ",
		TransactionFields, TransactionTable, symbol, fmt.Sprintf("%4d-%02d-%02d", year, month, day)))
}

func (ts *TransactionSet) ForMonth(ctx context.Context, pg *pgxpool.Pool, symbol string, year, month int) error {
	var queryStatement string
	switch month {
	case 12:
		endMonth := 01
		endYear := year + 1
		queryStatement = fmt.Sprintf(
			"SELECT %s From %s WHERE symbol = '%s' and date >= '%s' and date < '%s' order by date ",
			TransactionFields, TransactionTable, symbol, fmt.Sprintf("%4d-%02d-01", year, month), fmt.Sprintf("%4d-%02d-01", endYear, endMonth))
	default:
		endMonth := month + 1
		queryStatement = fmt.Sprintf(
			"SELECT %s From %s WHERE symbol = '%s' and date >= '%s' and date < '%s' order by date ",
			TransactionFields, TransactionTable, symbol, fmt.Sprintf("%4d-%02d-01", year, month), fmt.Sprintf("%4d-%02d-01", year, endMonth))
	}
	return ts.getTransactions(ctx, pg, queryStatement)
}

// NewTransaction creates a new transaction record from the CSV File header row and data row.
func NewTransaction(headers []string, row []string) (*Transaction, error) {
	tr := Transaction{}

	for i, h := range headers {
		switch h {
		case TransactionDate:
			if row[i] == "" {
				logrus.Error("Invalid Row(", len(row), ") ", row)
				return nil, fmt.Errorf("invalid date in row %d", i)
			}
			date, err := time.Parse("1/2/2006", row[i])
			if err != nil {
				return nil, fmt.Errorf("NewEntity Date[%s]: %v", row[i], err.Error())
			}
			tr.Date = date
		case TransactionType:
			tr.Type = TransactionsType(row[i])
		case TransactionSecurity:
			tr.Security = row[i]
		case TransactionSymbol:
			tr.Symbol = row[i]
		case TransactionSecurityPayee:
			tr.SecurityPayee = row[i]
		case TransactionDescription:
			tr.Description = row[i]
		case TransactionShares:
			shares, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Shares: %v", err.Error())
			}
			tr.Shares = shares
		case TransactionInvestmentAmount:
			iAmt, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Invest Amt: %v", err.Error())
			}
			tr.InvestmentAmount = iAmt
		case TransactionAmount:
			amt, err := utils.FloatParse(row[i])
			if err != nil {
				return nil, fmt.Errorf("NewTransactionRow Invest Amt: %v", err.Error())
			}
			tr.Amount = amt
		case TransactionAccount:
			tr.Account = row[i]
		default:
			if h != "Split" {
				fmt.Println("Skipping ", h)
			}
		}
	}
	return &tr, nil
}

// TransactionToDB inserts a Transaction object into the specified database table using a provided pgx connection and context.
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

// transactionLoadToDB
func transactionLoadToDB(tChan chan TransactionLoadStatus, pgxConn *pgxpool.Pool, transTable string, tr *Transaction) {
	ctx := context.Background()
	tSet := NewTransactionSet()
	logrus.Debug("Checking ", tr.Id)
	err := tSet.FromDBbyId(ctx, pgxConn, transTable, tr.Id)
	if err != nil {
		logrus.Error("on ", tr.Id, " : ", err.Error())
		tChan <- TransactionLoadStatus{
			ID:       tr.Id,
			Status:   false,
			Existing: false,
		}
		return
	}

	logrus.Debug("Found ", tr.Id)

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
