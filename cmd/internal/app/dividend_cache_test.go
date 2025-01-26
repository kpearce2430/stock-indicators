package app_test

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/keputils/utils"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApp_GetDividendCache(t *testing.T) {
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
