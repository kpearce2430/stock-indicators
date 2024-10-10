package symbollist_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/stock-tools/cmd/internal/handlers/symbollist"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/kpearce2430/stock-tools/postgres"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

//go:embed testdata/transactions.csv
var testTransactions []byte

//go:embed testdata/lookups.csv
var testLookups string

var lookups *model.LookUpSet
var pgxConn *pgxpool.Pool

func TestMain(m *testing.M) {
	//
	lookups = model.LoadLookupSet("1", testLookups)
	ctx := context.Background()
	postgresDBServer, _ := postgres.CreatePostgresTestServer(ctx)
	defer func() {
		if err := postgresDBServer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	pgIP, err := postgresDBServer.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	pgMappedPort, err := postgresDBServer.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatal(err)
	}

	// postgres://postgres:postgres@localhost:5432/postgres
	pgURL := fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres", pgIP, pgMappedPort.Port())
	pgxConn, err = pgxpool.New(context.Background(), pgURL)
	_ = os.Setenv("PG_DATABASE_URL", pgURL)

	ls := model.LoadLookupSet("1", testLookups)

	if err != nil {
		logrus.Fatal(err.Error())
	}
	if err := model.TransactionSetLoadToDB(pgxConn, ls, "transactions", testTransactions); err != nil {
		logrus.Fatal(err.Error())
	}
	os.Exit(m.Run())
}

func TestNewSymbolList(t *testing.T) {
	symList := symbollist.NewSymbolList(pgxConn, lookups)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", bytes.NewBuffer(testTransactions))
	symList.SymbolListGet(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be
	// responseData, _ := io.ReadAll(w.Body)
	// hist_usaix.csv.Log(string(responseData))
}

func TestSymbolList_AccountListGet(t *testing.T) {
	symList := symbollist.NewSymbolList(pgxConn, lookups)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodGet, "/", bytes.NewBuffer(testTransactions))
	symList.AccountListGet(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be
	// responseData, _ := io.ReadAll(w.Body)
	// hist_usaix.csv.Log(string(responseData))
}

func TestSymbolList_TickerInfoGet(t *testing.T) {
	symList := symbollist.NewSymbolList(pgxConn, lookups)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{
		{Key: "symbol", Value: "BMY"},
	}

	c.Request, _ = http.NewRequest(http.MethodGet, "/", bytes.NewBuffer(testTransactions))
	symList.TickerInfoGet(c)
	assert.Equal(t, 200, w.Code) // or what value you need it to be
	// responseData, _ := io.ReadAll(w.Body)
	// hist_usaix.csv.Log(string(responseData))
}
