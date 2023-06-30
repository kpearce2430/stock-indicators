package app_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"iex-indicators/cmd/internal/app"
	"iex-indicators/lookups"
	"iex-indicators/model"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadLookups(t *testing.T) {
	//
	// Switch to test mode so you don't get such noisy output
	gin.SetMode(gin.TestMode)

	a := app.App{
		Srv:       nil,
		LookupSet: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{gin.Param{Key: "id", Value: "1"}}
	c.Request, _ = http.NewRequest(http.MethodPost, "/lookups/1", bytes.NewBuffer(csvLookupData))
	a.LoadLookups(c)
	responseData, err := io.ReadAll(w.Body)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var status model.StatusObject

	err = json.Unmarshal(responseData, &status)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	// t.Log(string(responseData))

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
			StatusCode: http.StatusInternalServerError,
		},
		{
			Description: "Missing URI Params",
			StatusCode:  http.StatusBadRequest,
		},
	}

	for _, tc := range testGetLookups {
		func(t *testing.T, a app.App, tc testCase) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = tc.Params
			c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			a.GetLookups(c)
			assert.Equal(t, tc.StatusCode, w.Code)
			if w.Code == http.StatusOK {
				responseData, err := io.ReadAll(w.Body)
				assert.Equal(t, nil, err)
				t.Log(string(responseData))
				var status lookups.LookUpSet
				err = json.Unmarshal(responseData, &status)
				assert.Equal(t, nil, err)
			}
		}(t, a, tc)
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
		func(t *testing.T, a app.App, tc testCase) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = tc.Params
			c.Request, err = http.NewRequest(http.MethodGet, "/", nil)
			a.GetLookupName(c)
			_, err := io.ReadAll(w.Body)
			assert.Equal(t, nil, err)
			assert.Equal(t, tc.StatusCode, w.Code)
		}(t, a, tc)
	}
}
