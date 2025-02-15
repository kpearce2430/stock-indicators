package app_test

import (
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/keputils/utils"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApp_GetDividends(t *testing.T) {
	key := "None"
	utils.GetEnv("POLYGON_API", key)
	if strings.Compare(key, "None") == 0 {
		t.Skip("No POLYGON_API key")
		return
	}
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{gin.Param{Key: "symbol", Value: "HD"}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	q := c.Request.URL.Query()
	c.Request.URL.RawQuery = q.Encode()
	testApp.GetDividends(c)

	responseData, err := io.ReadAll(w.Body)
	if err != nil {
		t.Log(err.Error())
		t.Fatal()
		return
	}
	if w.Code != http.StatusOK {
		t.Log(string(responseData))
		t.Errorf("got %d, want %d", w.Code, http.StatusOK)
	}
	t.Log(string(responseData))
}

func TestApp_GetAllDividends(t *testing.T) {
	//t.Skip("skipped")
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// c.Params = []gin.Param{gin.Param{Key: "symbol", Value: "HD"}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	q := c.Request.URL.Query()
	c.Request.URL.RawQuery = q.Encode()
	testApp.GetAllDividends(c)

	responseData, err := io.ReadAll(w.Body)
	if err != nil {
		t.Log(err.Error())
		t.Fatal()
		return
	}
	if w.Code != http.StatusOK {
		t.Log(string(responseData))
		t.Errorf("got %d, want %d", w.Code, http.StatusOK)
	}
	t.Log(string(responseData))
}
