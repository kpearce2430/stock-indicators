package app

import (
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// var errUnexpectedNumberOfTransactions = fmt.Errorf("unexpected number of transactions found")

func (a *App) LoadTransactionsHandler(c *gin.Context) {
	//
	if a.LookupSet == nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: "Lookup Not Loaded"})
		return
	}

	databaseName := c.DefaultQuery("database", TransactionTable)

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

	if err := model.TransactionSetLoadToDB(a.PGXConn, a.LookupSet, databaseName, rawData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: "completed"})
}
