package app

import (
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func (a *App) LoadHistoricalData(c *gin.Context) {
	const fundHistory = "fund_history"
	quaryParams := c.Request.URL.Query()
	databaseName := quaryParams.Get("database")
	if databaseName == "" {
		databaseName = historicalDB
	}
	logrus.Info("database:", databaseName)

	symbol := quaryParams.Get("symbol")
	if symbol == "" {
		logrus.Error("missing symbol")
		c.IndentedJSON(http.StatusBadRequest, "Missing Symbol")
		return
	}
	logrus.Info("symbol:", symbol)

	source := quaryParams.Get("source")
	if source == "" {
		logrus.Error("missing source")
		c.IndentedJSON(http.StatusBadRequest, "Missing Source")
		return
	}
	logrus.Info("source:", source)

	defer func() {
		if c != nil && c.Request != nil && c.Request.Body != nil {
			if err := c.Request.Body.Close(); err != nil {
				logrus.Error(err.Error())
			}
		}
	}()

	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ds := model.NewHistoricalDataSet(a.PGXConn, fundHistory)
	if err := ds.LoadSet(string(rawData), source, symbol); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: "completed"})
}

/*
func (a *App) DeleteHistoricalData(c *gin.Context) {
	quaryParams := c.Request.URL.Query()
	databaseName := quaryParams.Get("database")
	if databaseName == "" {
		databaseName = historicalDB
	}
	logrus.Info("database:", databaseName)

	key := quaryParams.Get("key")
	if key == "" {
		logrus.Error("missing key")
		c.IndentedJSON(http.StatusBadRequest, "Missing key")
		return
	}
	logrus.Info("key:", key)
	/*
		rev, err := model.DeleteHistorical(databaseName, key)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}

	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: fmt.Sprintf("%s deleted", key)})
}
*/
