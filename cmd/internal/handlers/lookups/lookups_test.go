package lookups_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	lookup_handlers "iex-indicators/cmd/internal/handlers/lookups"
	"iex-indicators/cmd/internal/stock_ind_router"
	couch_database "iex-indicators/couch-database"
	"iex-indicators/lookups"
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
Mondelez International Inc,MDLZ`

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

	os.Setenv("TOKEN", "Tpk_76c5b627e1d3420dbd0f2621787941ba")
	os.Setenv("IEX_URL", "sandbox.iexapis.com")
	os.Setenv("COUCHDB_URL", url)
	os.Setenv("COUCHDB_USER", "admin")
	os.Setenv("COUCHDB_PASSWORD", "password")
	os.Setenv("DATABASE_NAME", "lookups")

	//router := gin.Default()
	//// TODO: Shut down stock_ind_router...
	//
	//indicatorDatabase := couch_database.DataStore[responses.CouchIndicatorResponse]("BR")
	//if indicatorDatabase.DatabaseCreate() != true {
	//	log.Fatal("Unable to create database")
	//}
	//router.GET("/look", indicators.GetMACDRouter)
	//router.GET("/rsi", indicators.GetRsiRouter)

	fmt.Println("Starting tests")
	m.Run()
}

func TestLoadLookups(t *testing.T) {

	lookupCSV := []byte(csvLookupData)

	router := gin.Default()
	router.POST("/lookups/:id", lookup_handlers.LoadLookups)
	router.GET("/lookups/:id", lookup_handlers.GetLookups)
	router.GET("/lookups/:id/:name", lookup_handlers.GetLookupName)

	req, _ := http.NewRequest(http.MethodPost, "/lookups/1", bytes.NewBuffer(lookupCSV))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ := ioutil.ReadAll(w.Body)

	assert.Equal(t, http.StatusOK, w.Code)
	t.Log(string(responseData))

	// Now lets do get
	t.Log("Lets do get...")

	req, _ = http.NewRequest(http.MethodGet, "/lookups/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ = ioutil.ReadAll(w.Body)

	var lookupResponse lookups.LookUpSet
	err := json.Unmarshal(responseData, &lookupResponse)

	assert.Nil(t, err, "Error?")
	assert.Equal(t, http.StatusOK, w.Code)
	t.Log("Response>>", lookupResponse)
	assert.Equal(t, 7, len(lookupResponse.LookUps))

	req, _ = http.NewRequest(http.MethodGet, "/lookups/1/CSX Corp", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ = ioutil.ReadAll(w.Body)

	assert.Equal(t, http.StatusOK, w.Code)

	var status stock_ind_router.StatusObject
	err = json.Unmarshal(responseData, &status)

	assert.Nil(t, err, "Error not nil")
	assert.Equal(t, status.Symbol, "CSX")
	t.Log(status)

	req, _ = http.NewRequest(http.MethodGet, "/lookups/1/JunkieJunk", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(responseData, &status)

	assert.Nil(t, err, "Error not nil")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, status.Status, "Not Found")

	t.Log(string(responseData))

}
