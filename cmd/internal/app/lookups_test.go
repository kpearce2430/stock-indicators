package app_test

import (
	"bytes"
	_ "embed"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

/*
func TestLoadLookups(t *testing.T) {
	//
	// Switch to test mode so you don'hist_usaix.csv get such noisy output
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{gin.Param{Key: "id", Value: "1"}}
	c.Request, _ = http.NewRequest(http.MethodPost, "/lookups/1", bytes.NewBuffer(csvLookupData))
	testApp.LoadLookups(c)
	responseData, err := io.ReadAll(w.Body)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var status model.StatusObject

	err = json.Unmarshal(responseData, &status)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	type testCase struct {
		Description string
		Params      []gin.Param
		StatusCode  int
	}

	testGetLookups := []testCase{
		{
			Description: "Get Lookups",
			Params: []gin.Param{
				{Key: "id", Value: "1"},
			},
			StatusCode: http.StatusOK,
		},
		{
			Description: "Get Non-Existent Lookups",
			Params: []gin.Param{
				{Key: "id", Value: "2"},
			},
			StatusCode: http.StatusOK,
		},
		{
			Description: "Missing URI Params",
			StatusCode:  http.StatusOK,
		},
	}

	for _, tc := range testGetLookups {
		t.Run(tc.Description, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = tc.Params
			c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			testApp.GetLookups(c)
			assert.Equal(t, tc.StatusCode, w.Code)
			if w.Code == http.StatusOK {
				responseData, err := io.ReadAll(w.Body)
				assert.Equal(t, nil, err)
				// hist_usaix.csv.Log(string(responseData))
				var status model.LookUpSet
				err = json.Unmarshal(responseData, &status)
				assert.Equal(t, nil, err)
			}
		})
	}

	testGetLookupName := []testCase{
		{
			Description: "Get All Lookups",
			Params: []gin.Param{
				{Key: "id", Value: "1"},
				{Key: "name", Value: "CSX Corp"},
			},
			StatusCode: http.StatusOK,
		},
		{
			Description: "Get All Lookups",
			Params: []gin.Param{
				{Key: "id", Value: "1"},
				{Key: "name", Value: "JunkieJunk"},
			},
			StatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range testGetLookupName {
		t.Run(tc.Description, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = tc.Params
			c.Request, err = http.NewRequest(http.MethodGet, "/", nil)
			testApp.GetLookupName(c)
			_, err := io.ReadAll(w.Body)
			assert.Equal(t, nil, err)
			assert.Equal(t, tc.StatusCode, w.Code)
		})
	}
}

*/

func TestApp_LoadLookupsToPostgres(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(csvLookupData))
	testApp.LoadLookupsToPostgres(c)
}
