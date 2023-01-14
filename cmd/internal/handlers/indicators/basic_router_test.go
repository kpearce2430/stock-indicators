package indicators_test

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/keputils/couch-database"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	"iex-indicators/cmd/internal/handlers/indicators"
	"iex-indicators/cmd/internal/stock_ind_router"
	"iex-indicators/responses"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

//
//
//
var router *gin.Engine

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

	os.Setenv("TOKEN", "Tpk_76c5b627e1d3420dbd0f2621787941ba")
	os.Setenv("IEX_URL", "sandbox.iexapis.com")

	os.Setenv("PREFIX", "BR")
	os.Setenv("BR_DATABASE_NAME", "brindicator")
	os.Setenv("BR_COUCHDB_URL", url)
	os.Setenv("BR_COUCHDB_USER", "admin")
	os.Setenv("BR_COUCHDB_PASSWORD", "password")

	os.Setenv("COUCHDB_URL", url)
	os.Setenv("COUCHDB_USER", "admin")
	os.Setenv("COUCHDB_PASSWORD", "password")

	router = gin.Default()
	// TODO: Shut down stock_ind_router...

	indicatorDatabase := couch_database.DataStore[responses.CouchIndicatorResponse]("BR")
	if indicatorDatabase.DatabaseCreate() != true {
		log.Fatal("Unable to create database")
	}
	router.GET("/macd", indicators.GetMACDRouter)
	router.GET("/rsi", indicators.GetRsiRouter)

	fmt.Println("Starting tests")
	m.Run()

}

func TestGetMACDRouter(t *testing.T) {

	req, _ := http.NewRequest(http.MethodGet, "/macd?symbol=HD&indicator=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ := ioutil.ReadAll(w.Body)

	assert.NotNil(t, responseData, "Response Data was empty?")
	// assert.Equal( t, http.StatusOK,)
	// log.Println("Response>", string(responseData))

}

func commonCaller(t *testing.T, url string) []byte {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ := ioutil.ReadAll(w.Body)

	assert.NotNil(t, responseData, "Response Data was empty?")

	log.Println("responseData>>", string(responseData))

	return responseData
}

func TestGetRsiRouter(t *testing.T) {

	responseData := commonCaller(t, "/rsi?symbol=HD&indicator=false")
	assert.NotNil(t, responseData, "Response Data was empty?")

}

func TestBasicRouterUpdate(t *testing.T) {

	responseData := commonCaller(t, "/rsi?symbol=HD&indicator=false&action=update")

	assert.NotNil(t, responseData, "Response Data was empty?")

}

func checkStatus(t *testing.T, body []byte, expectedStatus string) bool {

	response := stock_ind_router.StatusObject{}

	err := json.Unmarshal(body, &response)

	if err != nil {
		t.Log("Body:", string(body))
		t.Log("Unable to unmarshall data", err.Error())
		t.Fatal("Failure to unmarshall data")
	}

	return assert.Equal(t, expectedStatus, response.Status, "Invalid response status")

}

func TestCouchDBDown(t *testing.T) {

	os.Setenv("PREFIX", "")
	os.Setenv("DATABASE_NAME", "junkjunkjunk")

	responseData := commonCaller(t, "/rsi?symbol=HD&indicatorOnly=true&action=update")

	t.Log("responseData>>", string(responseData))
	assert.NotNil(t, responseData, "Invalid Response")

	if !checkStatus(t, responseData, "CouchDB Not Up") {
		t.Fatal("Something bad happened")
	}
}

func TestNoSymbol(t *testing.T) {

	responseData := commonCaller(t, "/macd")

	assert.NotNil(t, responseData, "Invalid Response")

	checkStatus(t, responseData, "missing symbol")

}

func TestBadIndicator(t *testing.T) {

	responseData := commonCaller(t, "/rsi?symbol=HD&indicatorOnly=junk&action=update")

	assert.NotNil(t, responseData, "Response Data was empty?")

	checkStatus(t, responseData, "Invalid indicatorOnly")

}
