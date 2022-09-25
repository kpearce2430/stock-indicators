package portfolio_value_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	lookup_handlers "iex-indicators/cmd/internal/handlers/lookups"
	"iex-indicators/cmd/internal/handlers/portfolio_value"
	couch_database "iex-indicators/couch-database"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var csvLookupData = `PHILIP MORRIS CO INC,PM
CSX Corp,CSX
3M Corp,MMM
Microsoft Corp,MSFT
MMDA1,MMDA1
PM,PM
JUNK INC COM,JUNK
DEAD INC COM,DEAD
Mondelez International Inc,MDLZ`

var csvPortfolioValueData string = `xInvesting - Portfolio Value - Group by Security
x
xCreated: 2021-12-11
x
xPrice and Holdings as of: 2021-12-11
x
x,Symbol,Shares,Type,Price,Price Day Change,Price Day Change (%),Cost Basis,Market Value,Average Cost Per Share,Gain/Loss 12-Month,Gain/Loss,Gain/Loss (%)
x3M Corp (MMM),MMM,"100",Stock,"177.10","0.00","0.0%","$8,039.15","$17,710.00","80.39","$308.00","$9,670.85","120.3%"
xAltria Group,MO,"1,300",Stock,"45.09","0.00","0.0%","$33,748.62","$58,617.00","25.96","$2,717.00","$24,868.38","73.7%"
xAPPLE INC COM,AAPL,"400",Stock,"179.45","0.00","0.0%","$15,606.95","$71,780.00","39.02","$22,816.00","$56,173.05","359.9%"
xJUNK INC COM,,"100",Stock,"279.45","0.00","0.0%","$15,606.95","$71,780.00","39.02","$22,816.00","$56,173.05","359.9%"
xDEAD INC COM,,"100",Stock,"279.45","0.00","0.0%","$15,606.95","$71,780.00","39.02","$22,816.00","$56,173.05","359.9%"
x
xCash,,,,,,,,"$40,405.66",,,,
xTotals,,,,,,,"$870,683.60","$1,607,023.67",,"$183,339.93","$695,934.30","79.9%"
`

// TestMain
func TestMain(m *testing.M) {

	ctx := context.Background()

	couchDBServer, _ := couch_database.CreateCouchDBServer(ctx)
	defer couchDBServer.Terminate(ctx)

	ip, err := couchDBServer.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := couchDBServer.MappedPort(ctx, "5984")
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	log.Println(url)

	os.Setenv("COUCHDB_URL", url)
	os.Setenv("COUCHDB_USER", "admin")
	os.Setenv("COUCHDB_PASSWORD", "password")
	os.Setenv("DATABASE_NAME", "pv")

	fmt.Println("Loading Lookups")

	lookupCSV := []byte(csvLookupData)

	router := gin.Default()
	router.POST("/lookups/:id", lookup_handlers.LoadLookups)

	req, _ := http.NewRequest(http.MethodPost, "/lookups/1", bytes.NewBuffer(lookupCSV))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ := ioutil.ReadAll(w.Body)

	if w.Code != 200 {
		log.Fatal("Unable to loaded lookup data", string(responseData))
	}

	fmt.Println("Starting tests")

	m.Run()
}

func TestLoadPortfolioValue(t *testing.T) {

	t.Log("Starting TestLoadPortfolioValue")
	pvData := []byte(csvPortfolioValueData)

	router := gin.Default()
	router.POST("/pv", portfolio_value.LoadPortfolioValueHandler)
	router.GET("/pv/:symbol", portfolio_value.GetPortfolioValueHandler)

	// Post the test data
	req, _ := http.NewRequest(http.MethodPost, "/pv?NOTjuldate=2022001&database=something", bytes.NewBuffer(pvData))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Get an normal value
	req, _ = http.NewRequest(http.MethodGet, "/pv/AAPL?database=something", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	responseData, err := ioutil.ReadAll(w.Body)

	assert.Nil(t, err, "Error")

	var pvResponse portfolio_value.PortfolioValueDatabaseRecord

	err = json.Unmarshal(responseData, &pvResponse)

	assert.Nil(t, err, "Error")

	t.Log(pvResponse)

	// Try JUNK where the symbol came from the lookup table.
	req, _ = http.NewRequest(http.MethodGet, "/pv/JUNK?database=something", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	responseData, err = ioutil.ReadAll(w.Body)
	assert.Nil(t, err, "Error")
	err = json.Unmarshal(responseData, &pvResponse)

	assert.Nil(t, err, "Error")

	// t.Log(pvResponse)

	// Try a Missing Record
	req, _ = http.NewRequest(http.MethodGet, "/pv/MISSING?database=something", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	responseData, err = ioutil.ReadAll(w.Body)

	// t.Log(string(responseData))

	assert.Equal(t, "\"Not Found\"", string(responseData))

}
