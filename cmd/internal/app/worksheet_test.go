package app_test

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp_CreateWorksheetHandler(t *testing.T) {
	t.Skip("needs more work")
	t.Parallel()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
	testApp.CreateWorksheetHandler(c)
	if w.Code != http.StatusOK {
		t.Log("Status Not:", http.StatusOK, " Got:", w.Code)
		t.Fail()
		return
	}
}
