package app_test

/*
func TestApp_LoadHistoricalData(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(csvHistoricalDAta))
	q := c.Request.URL.Query()
	q.Add("database", databaseName)
	q.Add("symbol", "USAIX")
	q.Add("source", "random")
	c.Request.URL.RawQuery = q.Encode()
	testApp.LoadHistoricalData(c)
	if w.Code != http.StatusOK {
		t.Log("error expecting:", http.StatusOK, " got:", w.Code)
	}
	responseData, err := io.ReadAll(w.Body)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	t.Log(string(responseData))
}

*/
