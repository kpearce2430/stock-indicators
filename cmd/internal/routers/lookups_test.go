package routers_test

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	"iex-indicators/cmd/internal/routers"
	"iex-indicators/lookups"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadLookups(t *testing.T) {

	lookupCSV := []byte("PHILIP MORRIS CO INC,PM\nCSX Corp,CSX\n3M Corp,MMM\nMicrosoft Corp,MSFT\nMMDA1,MMDA1\nPM,PM\nMondelez International Inc,MDLZ")
	t.Log(len(lookupCSV))

	router := gin.Default()
	router.POST("/lookups/:id", routers.LoadLookups)
	router.GET("/lookups/:id", routers.GetLookups)

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

}
