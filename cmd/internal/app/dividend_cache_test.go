package app_test

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp_GetDividendCache(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{gin.Param{Key: "symbol", Value: "HD"}}
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(csvPortfolioValueData))
	q := c.Request.URL.Query()
	c.Request.URL.RawQuery = q.Encode()
	testApp.GetDividendCache(c)

	responseData, err := io.ReadAll(w.Body)
	if err != nil {
		t.Log(err.Error())
		t.Fatal()
		return
	}
	assert.Equal(t, http.StatusOK, w.Code)
	t.Log(string(responseData))
}
