package app_test

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp_GetDividends(t *testing.T) {
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
